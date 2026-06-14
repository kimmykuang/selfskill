package web

import (
	"net/http"

	"github.com/kimmykuang/selfskill/internal/service"
)

type pluginJSON struct {
	Name        string `json:"name"`
	Marketplace string `json:"marketplace"`
	Version     string `json:"version"`
	InstallPath string `json:"installPath"`
	State       string `json:"state"` // "installed" or "loaded"
}

func (s *Server) handleListPlugins(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := service.PluginFilter{
		Query: q.Get("q"),
	}
	switch q.Get("state") {
	case "loaded":
		filter.State = service.PluginStateLoaded
	case "installed":
		filter.State = service.PluginStateInstalled
	}
	plugins, err := s.Svc.ListPlugins(filter)
	if err != nil {
		httpFromErr(w, err)
		return
	}
	out := make([]pluginJSON, 0, len(plugins))
	for _, p := range plugins {
		out = append(out, pluginJSON{
			Name:        p.Name,
			Marketplace: p.Marketplace,
			Version:     p.Version,
			InstallPath: p.InstallPath,
			State:       string(p.State),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleGetPlugin(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	p, err := s.PluginStore.Get(name)
	if err != nil {
		httpError(w, http.StatusNotFound, "plugin not found")
		return
	}
	state := "installed"
	if s.Svc != nil && s.Svc.IsPluginLoaded(p.FullName()) {
		state = "loaded"
	}
	writeJSON(w, http.StatusOK, pluginJSON{
		Name:        p.Name,
		Marketplace: p.Marketplace,
		Version:     p.Version,
		InstallPath: p.InstallPath,
		State:       state,
	})
}

func (s *Server) handlePluginSkills(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	skills, err := s.PluginStore.ListSkills(name)
	if err != nil {
		writeJSON(w, http.StatusOK, []string{})
		return
	}
	if skills == nil {
		skills = []string{}
	}
	writeJSON(w, http.StatusOK, skills)
}

func (s *Server) handlePluginLoad(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.PluginLoader.Load(name); err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "loaded"})
}

func (s *Server) handlePluginUnload(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.PluginLoader.Unload(name); err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "unloaded"})
}
