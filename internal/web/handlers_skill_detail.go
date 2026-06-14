package web

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimmykuang/selfskill/internal/frontmatter"
)

type updateSkillBody struct {
	Description string                 `json:"description,omitempty"`
	Body        string                 `json:"body,omitempty"`
	Frontmatter map[string]interface{} `json:"frontmatter,omitempty"`
}

func (s *Server) handleUpdateSkill(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !s.SkillStore.Exists(name) {
		httpError(w, http.StatusNotFound, "skill not found")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var body updateSkillBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	mdPath := filepath.Join(s.SkillStore.SkillDir(name), "SKILL.md")
	raw, err := os.ReadFile(mdPath)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	meta, currentBody, err := frontmatter.Parse(raw)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if meta == nil {
		meta = map[string]interface{}{}
	}
	if body.Description != "" {
		meta["description"] = body.Description
	}
	if body.Frontmatter != nil {
		for k, v := range body.Frontmatter {
			meta[k] = v
		}
	}
	newBody := currentBody
	if body.Body != "" {
		newBody = body.Body
	}
	out, err := frontmatter.Marshal(meta, newBody)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := os.WriteFile(mdPath, out, 0644); err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) handleSkillDiff(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if s.Comparer == nil {
		httpError(w, http.StatusInternalServerError, "comparer not configured")
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if err := s.Comparer.CompareSkill(name, w); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s *Server) handleSkillFiles(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !s.SkillStore.Exists(name) {
		httpError(w, http.StatusNotFound, "skill not found")
		return
	}
	root := s.SkillStore.SkillDir(name)
	var files []string
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		files = append(files, rel)
		return nil
	})
	if walkErr != nil {
		httpError(w, http.StatusInternalServerError, walkErr.Error())
		return
	}
	if files == nil {
		files = []string{}
	}
	writeJSON(w, http.StatusOK, files)
}

func (s *Server) handleSkillFile(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	rel := r.PathValue("path")
	if !s.SkillStore.Exists(name) {
		httpError(w, http.StatusNotFound, "skill not found")
		return
	}
	cleaned := filepath.Clean(rel)
	if strings.HasPrefix(cleaned, "..") || filepath.IsAbs(cleaned) || strings.Contains(cleaned, "..") {
		httpError(w, http.StatusBadRequest, "path traversal disallowed")
		return
	}
	path := filepath.Join(s.SkillStore.SkillDir(name), cleaned)
	data, err := os.ReadFile(path)
	if err != nil {
		httpError(w, http.StatusNotFound, "file not found")
		return
	}
	w.Header().Set("Content-Type", contentTypeFor(cleaned))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func contentTypeFor(path string) string {
	switch ext := strings.ToLower(filepath.Ext(path)); ext {
	case ".md":
		return "text/markdown; charset=utf-8"
	case ".json":
		return "application/json"
	case ".yaml", ".yml":
		return "application/yaml"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	default:
		return "text/plain; charset=utf-8"
	}
}
