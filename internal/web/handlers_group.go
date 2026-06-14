package web

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/kimmykuang/selfskill/internal/group"
	"github.com/kimmykuang/selfskill/internal/linker"
	"github.com/kimmykuang/selfskill/internal/service"
)

func (s *Server) handleListGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := s.GroupStore.List()
	if err != nil {
		httpFromErr(w, err)
		return
	}
	if groups == nil {
		groups = []group.Group{}
	}
	writeJSON(w, http.StatusOK, groups)
}

type createGroupBody struct {
	Name string `json:"name"`
}

func (s *Server) handleCreateGroup(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var body createGroupBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.Name == "" {
		httpError(w, http.StatusBadRequest, "name is required")
		return
	}
	if err := s.GroupStore.Create(body.Name); err != nil {
		httpError(w, http.StatusConflict, err.Error())
		return
	}
	g, err := s.GroupStore.Get(body.Name)
	if err != nil {
		httpFromErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, g)
}

func (s *Server) handleGetGroup(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	g, err := s.GroupStore.Get(name)
	if err != nil {
		httpError(w, http.StatusNotFound, "group not found")
		return
	}
	writeJSON(w, http.StatusOK, g)
}

type updateGroupBody struct {
	Description string   `json:"description"`
	Skills      []string `json:"skills"`
	Plugins     []string `json:"plugins"`
	Prompts     []string `json:"prompts"`
}

func (s *Server) handleUpdateGroup(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var body updateGroupBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	g, err := s.GroupStore.Get(name)
	if err != nil {
		httpError(w, http.StatusNotFound, "group not found")
		return
	}
	g.Description = body.Description
	g.Skills = body.Skills
	g.Plugins = body.Plugins
	g.Prompts = body.Prompts
	if err := s.GroupStore.SaveReplace(g); err != nil {
		httpFromErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, g)
}

func (s *Server) handleDeleteGroup(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.GroupStore.Delete(name); err != nil {
		httpError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleLoadGroup(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = string(linker.ScopeUser)
	}
	report, err := s.Svc.LoadGroup(name, linker.Scope(scope))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			httpError(w, http.StatusNotFound, err.Error())
			return
		}
		httpFromErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleUnloadGroup(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = string(linker.ScopeUser)
	}
	report, err := s.Svc.UnloadGroup(name, linker.Scope(scope))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			httpError(w, http.StatusNotFound, err.Error())
			return
		}
		httpFromErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}
