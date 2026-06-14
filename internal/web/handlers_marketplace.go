package web

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleListMarketplaces(w http.ResponseWriter, r *http.Request) {
	if s.Registry == nil {
		writeJSON(w, http.StatusOK, []string{})
		return
	}
	mps, err := s.Registry.ListMarketplaces()
	if err != nil {
		httpFromErr(w, err)
		return
	}
	if mps == nil {
		mps = []string{}
	}
	writeJSON(w, http.StatusOK, mps)
}

type addMarketplaceBody struct {
	URL string `json:"url"`
}

func (s *Server) handleAddMarketplace(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var body addMarketplaceBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.URL == "" {
		httpError(w, http.StatusBadRequest, "url is required")
		return
	}
	name, err := s.Registry.AddMarketplace(body.URL)
	if err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"name": name})
}

func (s *Server) handleListMarketplacePlugins(w http.ResponseWriter, r *http.Request) {
	mp := r.PathValue("name")
	mf, err := s.Registry.LoadMarketplace(mp)
	if err != nil {
		httpError(w, http.StatusNotFound, err.Error())
		return
	}
	type pluginListing struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	out := make([]pluginListing, 0, len(mf.Plugins))
	for _, p := range mf.Plugins {
		out = append(out, pluginListing{Name: p.Name, Description: p.Description})
	}
	writeJSON(w, http.StatusOK, out)
}

type installPluginBody struct {
	Plugin      string `json:"plugin"`
	Marketplace string `json:"marketplace"`
}

func (s *Server) handleInstallPlugin(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var body installPluginBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.Plugin == "" || body.Marketplace == "" {
		httpError(w, http.StatusBadRequest, "plugin and marketplace are required")
		return
	}
	mp, err := s.Registry.FindPlugin(body.Plugin, body.Marketplace)
	if err != nil {
		httpError(w, http.StatusNotFound, err.Error())
		return
	}
	if s.Registry.IsLocalPlugin(mp) {
		src, err := s.Registry.LocalPluginPath(mp, body.Marketplace)
		if err != nil {
			httpError(w, http.StatusBadRequest, err.Error())
			return
		}
		p, err := s.PluginStore.InstallFromLocal(src, body.Marketplace, body.Plugin, "latest")
		if err != nil {
			httpError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, p)
		return
	}
	gitURL, err := s.Registry.ResolveGitURL(mp, body.Marketplace)
	if err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	p, err := s.PluginStore.InstallFromGit(gitURL, body.Marketplace, body.Plugin, "latest")
	if err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, p)
}
