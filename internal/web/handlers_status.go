package web

import (
	"net/http"
)

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	userLinks, projectLinks, err := s.Linker.Status()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	type linkInfoJSON struct {
		SkillName  string `json:"skillName"`
		SourcePath string `json:"sourcePath"`
		LinkPath   string `json:"linkPath"`
	}

	type statusResponse struct {
		UserLinks    []linkInfoJSON `json:"userLinks"`
		ProjectLinks []linkInfoJSON `json:"projectLinks"`
	}

	resp := statusResponse{
		UserLinks:    make([]linkInfoJSON, 0, len(userLinks)),
		ProjectLinks: make([]linkInfoJSON, 0, len(projectLinks)),
	}

	for _, l := range userLinks {
		resp.UserLinks = append(resp.UserLinks, linkInfoJSON{
			SkillName:  l.SkillName,
			SourcePath: l.SourcePath,
			LinkPath:   l.LinkPath,
		})
	}
	for _, l := range projectLinks {
		resp.ProjectLinks = append(resp.ProjectLinks, linkInfoJSON{
			SkillName:  l.SkillName,
			SourcePath: l.SourcePath,
			LinkPath:   l.LinkPath,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}
