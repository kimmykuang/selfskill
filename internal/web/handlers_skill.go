package web

import (
	"net/http"

	"github.com/kimmykuang/selfskill/internal/skill"
)

func (s *Server) handleListSkills(w http.ResponseWriter, r *http.Request) {
	skills, err := s.SkillStore.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if skills == nil {
		skills = []skill.Skill{}
	}
	writeJSON(w, http.StatusOK, skills)
}

func (s *Server) handleGetSkill(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	sk, err := s.SkillStore.Get(name)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "skill not found"})
		return
	}
	writeJSON(w, http.StatusOK, sk)
}
