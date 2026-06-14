package web

import (
	"encoding/json"
	"net/http"
)

type installBody struct {
	Type   string `json:"type"`             // "skill" | "prompt"
	Source string `json:"source"`           // local path or URL
	Force  bool   `json:"force,omitempty"`
}

func (s *Server) handleInstall(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var body installBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.Source == "" {
		httpError(w, http.StatusBadRequest, "source is required")
		return
	}
	if body.Type != "skill" && body.Type != "prompt" {
		httpError(w, http.StatusBadRequest, `type must be "skill" or "prompt"`)
		return
	}
	if s.Installer == nil {
		httpError(w, http.StatusInternalServerError, "installer not configured")
		return
	}

	switch body.Type {
	case "skill":
		results, err := s.Installer.InstallSkill(body.Source, body.Force)
		if err != nil {
			httpError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, results)
	case "prompt":
		results, err := s.Installer.InstallPrompt(body.Source, body.Force)
		if err != nil {
			httpError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, results)
	}
}
