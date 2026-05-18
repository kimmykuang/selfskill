package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kimmykuang/selfskill/internal/prompt"
)

func (s *Server) handleListPrompts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	tagsParam := r.URL.Query().Get("tags")

	var tags []string
	if tagsParam != "" {
		tags = strings.Split(tagsParam, ",")
	}

	var prompts []prompt.Prompt
	var err error

	if query != "" || len(tags) > 0 {
		prompts, err = s.PromptStore.Search(query, tags)
	} else {
		prompts, err = s.PromptStore.List()
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if prompts == nil {
		prompts = []prompt.Prompt{}
	}
	writeJSON(w, http.StatusOK, prompts)
}

func (s *Server) handleGetPrompt(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := s.PromptStore.Get(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "prompt not found"})
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) handleCreatePrompt(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB limit
	var p prompt.Prompt
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if p.ID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
		return
	}
	if s.PromptStore.Exists(p.ID) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "prompt already exists"})
		return
	}
	if err := s.PromptStore.Save(&p); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (s *Server) handleUpdatePrompt(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB limit
	var p prompt.Prompt
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	p.ID = id

	if !s.PromptStore.Exists(id) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "prompt not found"})
		return
	}
	if err := s.PromptStore.Save(&p); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) handleDeletePrompt(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := s.PromptStore.Remove(id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "prompt not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
