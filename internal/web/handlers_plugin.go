package web

import (
	"net/http"
)

type pluginJSON struct {
	Name        string `json:"name"`
	Marketplace string `json:"marketplace"`
	Version     string `json:"version"`
	InstallPath string `json:"installPath"`
	State       string `json:"state"` // "installed" or "loaded"
}

func (s *Server) handleListPlugins(w http.ResponseWriter, r *http.Request) {
	plugins, err := s.PluginStore.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	result := make([]pluginJSON, 0, len(plugins))
	for _, p := range plugins {
		state := "installed"
		if s.PluginStore.IsLoaded(p.FullName()) {
			state = "loaded"
		}
		result = append(result, pluginJSON{
			Name:        p.Name,
			Marketplace: p.Marketplace,
			Version:     p.Version,
			InstallPath: p.InstallPath,
			State:       state,
		})
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetPlugin(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	p, err := s.PluginStore.Get(name)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "plugin not found"})
		return
	}

	state := "installed"
	if s.PluginStore.IsLoaded(p.FullName()) {
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
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "loaded"})
}

func (s *Server) handlePluginUnload(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.PluginLoader.Unload(name); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "unloaded"})
}
