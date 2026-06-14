package web

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/kimmykuang/selfskill/internal/compare"
	"github.com/kimmykuang/selfskill/internal/group"
	"github.com/kimmykuang/selfskill/internal/installer"
	"github.com/kimmykuang/selfskill/internal/linker"
	"github.com/kimmykuang/selfskill/internal/plugin"
	"github.com/kimmykuang/selfskill/internal/prompt"
	"github.com/kimmykuang/selfskill/internal/service"
	"github.com/kimmykuang/selfskill/internal/skill"
)

// Server is the HTTP server for the web UI.
type Server struct {
	SkillStore   *skill.Store
	PromptStore  *prompt.Store
	PluginStore  *plugin.Store
	GroupStore   *group.Store
	PluginLoader *plugin.Loader
	Linker       *linker.Linker
	Installer    *installer.Installer
	Registry     *plugin.Registry
	Comparer     *compare.Comparer
	Svc          *service.Service
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

	// Phase 4 — group endpoints.
	mux.HandleFunc("GET /api/groups", s.handleListGroups)
	mux.HandleFunc("POST /api/groups", s.handleCreateGroup)
	mux.HandleFunc("GET /api/groups/{name}", s.handleGetGroup)
	mux.HandleFunc("PUT /api/groups/{name}", s.handleUpdateGroup)
	mux.HandleFunc("DELETE /api/groups/{name}", s.handleDeleteGroup)
	mux.HandleFunc("POST /api/groups/{name}/load", s.handleLoadGroup)
	mux.HandleFunc("POST /api/groups/{name}/unload", s.handleUnloadGroup)

	// Phase 4 — install endpoint.
	mux.HandleFunc("POST /api/install", s.handleInstall)

	// Phase 4 — marketplace endpoints.
	mux.HandleFunc("GET /api/marketplaces", s.handleListMarketplaces)
	mux.HandleFunc("POST /api/marketplaces", s.handleAddMarketplace)
	mux.HandleFunc("GET /api/marketplaces/{name}/plugins", s.handleListMarketplacePlugins)
	mux.HandleFunc("POST /api/plugins/install", s.handleInstallPlugin)

	// Phase 4 — search.
	mux.HandleFunc("GET /api/search", s.handleSearch)

	// Phase 4 — skill detail.
	mux.HandleFunc("PUT /api/skills/{name}", s.handleUpdateSkill)
	mux.HandleFunc("GET /api/skills/{name}/diff", s.handleSkillDiff)
	mux.HandleFunc("GET /api/skills/{name}/files", s.handleSkillFiles)
	mux.HandleFunc("GET /api/skills/{name}/files/{path...}", s.handleSkillFile)

	// Static files
	mux.Handle("/", http.FileServer(http.FS(s.StaticFS)))

	addr := fmt.Sprintf("127.0.0.1:%d", s.Port)
	return http.ListenAndServe(addr, mux)
}
