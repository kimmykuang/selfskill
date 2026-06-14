package service

import (
	"fmt"

	"github.com/kimmykuang/selfskill/internal/linker"
)

// SkillLoadResult mirrors linker.LinkResult but is decoupled so callers
// don't need to import linker. (linker.LinkResult re-exposed.)
type SkillLoadResult = linker.LinkResult

// PluginLoadResult is what LoadGroup reports per plugin attempt.
type PluginLoadResult struct {
	Name   string
	Action string // "loaded" | "error"
	Detail string
}

// GroupLoadReport is the structured outcome of LoadGroup / UnloadGroup.
type GroupLoadReport struct {
	GroupName string
	Skills    []SkillLoadResult
	Plugins   []PluginLoadResult
}

// LoadGroup loads all skills (via linker) and plugins (via plugin loader)
// listed in the group. Per-item failures do not abort: each is reported.
// Group not found returns ErrNotFound.
func (s *Service) LoadGroup(name string, scope linker.Scope) (*GroupLoadReport, error) {
	g, err := s.deps.Groups.Get(name)
	if err != nil {
		return nil, fmt.Errorf("%w: group %q: %v", ErrNotFound, name, err)
	}

	report := &GroupLoadReport{GroupName: name}

	if len(g.Skills) > 0 {
		results, err := s.deps.Linker.Load(g.Skills, scope)
		if err != nil {
			return nil, fmt.Errorf("loading skills: %w", err)
		}
		report.Skills = results
	}

	for _, p := range g.Plugins {
		if err := s.deps.PluginLoader.Load(p); err != nil {
			report.Plugins = append(report.Plugins, PluginLoadResult{Name: p, Action: "error", Detail: err.Error()})
			continue
		}
		report.Plugins = append(report.Plugins, PluginLoadResult{Name: p, Action: "loaded"})
	}

	return report, nil
}

// UnloadGroup is the inverse of LoadGroup.
func (s *Service) UnloadGroup(name string, scope linker.Scope) (*GroupLoadReport, error) {
	g, err := s.deps.Groups.Get(name)
	if err != nil {
		return nil, fmt.Errorf("%w: group %q: %v", ErrNotFound, name, err)
	}

	report := &GroupLoadReport{GroupName: name}

	if len(g.Skills) > 0 {
		results, err := s.deps.Linker.Unload(g.Skills, scope)
		if err != nil {
			return nil, fmt.Errorf("unloading skills: %w", err)
		}
		report.Skills = results
	}

	for _, p := range g.Plugins {
		if err := s.deps.PluginLoader.Unload(p); err != nil {
			report.Plugins = append(report.Plugins, PluginLoadResult{Name: p, Action: "error", Detail: err.Error()})
			continue
		}
		report.Plugins = append(report.Plugins, PluginLoadResult{Name: p, Action: "unloaded"})
	}

	return report, nil
}
