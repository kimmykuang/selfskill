package web

import (
	"net/http"

	"github.com/kimmykuang/selfskill/internal/service"
)

func (s *Server) handleListSkills(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := service.SkillFilter{
		Query: q.Get("q"),
	}
	switch q.Get("origin") {
	case "local":
		filter.Origin = service.OriginLocal
	case "github":
		filter.Origin = service.OriginGitHub
	}
	skills, err := s.Svc.ListSkills(filter)
	if err != nil {
		httpFromErr(w, err)
		return
	}
	if skills == nil {
		skills = []service.SkillView{}
	}
	writeJSON(w, http.StatusOK, skills)
}

func (s *Server) handleGetSkill(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	sk, err := s.SkillStore.Get(name)
	if err != nil {
		httpError(w, http.StatusNotFound, "skill not found")
		return
	}
	writeJSON(w, http.StatusOK, sk)
}
