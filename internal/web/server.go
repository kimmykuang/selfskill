package web

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/kimmykuang/selfskill/internal/linker"
	"github.com/kimmykuang/selfskill/internal/plugin"
	"github.com/kimmykuang/selfskill/internal/prompt"
	"github.com/kimmykuang/selfskill/internal/skill"
)

// Server is the HTTP server for the web UI.
type Server struct {
	SkillStore   *skill.Store
	PromptStore  *prompt.Store
	PluginStore  *plugin.Store
	PluginLoader *plugin.Loader
	Linker       *linker.Linker
	StaticFS     fs.FS
	Port         int
}

// Start starts the HTTP server on the configured port.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/prompts", s.handleListPrompts)
	mux.HandleFunc("GET /api/prompts/{id}", s.handleGetPrompt)
	mux.HandleFunc("POST /api/prompts", s.handleCreatePrompt)
	mux.HandleFunc("PUT /api/prompts/{id}", s.handleUpdatePrompt)
	mux.HandleFunc("DELETE /api/prompts/{id}", s.handleDeletePrompt)
	mux.HandleFunc("GET /api/skills", s.handleListSkills)
	mux.HandleFunc("GET /api/skills/{name}", s.handleGetSkill)
	mux.HandleFunc("GET /api/plugins", s.handleListPlugins)
	mux.HandleFunc("GET /api/plugins/{name}", s.handleGetPlugin)
	mux.HandleFunc("GET /api/plugins/{name}/skills", s.handlePluginSkills)
	mux.HandleFunc("POST /api/plugins/{name}/load", s.handlePluginLoad)
	mux.HandleFunc("POST /api/plugins/{name}/unload", s.handlePluginUnload)
	mux.HandleFunc("GET /api/status", s.handleStatus)

	// Static files
	mux.Handle("/", http.FileServer(http.FS(s.StaticFS)))

	addr := fmt.Sprintf("127.0.0.1:%d", s.Port)
	return http.ListenAndServe(addr, mux)
}
