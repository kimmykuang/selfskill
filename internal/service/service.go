// Package service is the shared use-case layer between CLI and Web.
// It composes the file-tree stores, the symlink linker, and the plugin
// loader into named operations like LoadGroup, ListSkills, Search.
package service

import (
	"github.com/kimmykuang/selfskill/internal/group"
	"github.com/kimmykuang/selfskill/internal/linker"
	"github.com/kimmykuang/selfskill/internal/plugin"
	"github.com/kimmykuang/selfskill/internal/plugin/cc_compat"
	"github.com/kimmykuang/selfskill/internal/prompt"
	"github.com/kimmykuang/selfskill/internal/skill"
)

// Deps bundles the dependencies a Service needs. Constructed once in
// cli/root.go for production and per-test in service tests.
type Deps struct {
	Skills       *skill.Store
	Prompts      *prompt.Store
	Groups       *group.Store
	Plugins      *plugin.Store
	Linker       *linker.Linker
	PluginLoader *plugin.Loader
	Registry     cc_compat.Registry
}

// Service is the shared use-case API. Its methods take and return plain
// Go types — no HTTP shapes, no CLI shapes.
type Service struct {
	deps Deps
}

func New(deps Deps) *Service {
	return &Service{deps: deps}
}
