package service

import (
	"strings"

	"github.com/kimmykuang/selfskill/internal/skill"
)

// Origin is a skill's source classification, derived (not persisted) from
// the `source` frontmatter field. Empty source -> OriginLocal; else OriginGitHub.
type Origin string

const (
	OriginAny    Origin = ""
	OriginLocal  Origin = "local"
	OriginGitHub Origin = "github"
)

// SkillFilter narrows ListSkills results.
type SkillFilter struct {
	Origin Origin // empty => no origin filter
	Query  string // case-insensitive substring match against Name and Description
	// Tags/State filters land in later phases when skill tags exist on disk.
}

// SkillView is what ListSkills returns. It augments skill.Skill with the
// derived Origin so callers can render or filter on it.
type SkillView struct {
	skill.Skill
	Origin Origin
}

// ListSkills returns all skills matching f. Order matches store order.
func (s *Service) ListSkills(f SkillFilter) ([]SkillView, error) {
	all, err := s.deps.Skills.List()
	if err != nil {
		return nil, err
	}
	out := make([]SkillView, 0, len(all))
	for _, sk := range all {
		v := SkillView{Skill: sk, Origin: deriveOrigin(sk.Source)}
		if !matchesSkillFilter(v, f) {
			continue
		}
		out = append(out, v)
	}
	return out, nil
}

func deriveOrigin(source string) Origin {
	if source == "" {
		return OriginLocal
	}
	return OriginGitHub
}

func matchesSkillFilter(v SkillView, f SkillFilter) bool {
	if f.Origin != OriginAny && v.Origin != f.Origin {
		return false
	}
	if f.Query != "" {
		q := strings.ToLower(f.Query)
		if !strings.Contains(strings.ToLower(v.Name), q) &&
			!strings.Contains(strings.ToLower(v.Description), q) {
			return false
		}
	}
	return true
}

// PromptFilter narrows ListPrompts.
type PromptFilter struct {
	Tags  []string // OR semantics: prompt matches if it has ANY listed tag
	Query string
}

// PluginState reflects whether a plugin is currently visible to CC.
type PluginState string

const (
	PluginStateInstalled PluginState = "installed"
	PluginStateLoaded    PluginState = "loaded"
)

// PluginFilter narrows ListPlugins.
type PluginFilter struct {
	State PluginState // empty => any
	Query string
}

// PluginView augments plugin.Plugin with derived state.
type PluginView struct {
	Name         string
	Marketplace  string
	Version      string
	InstallPath  string
	GitCommitSha string
	RemoteURL    string
	State        PluginState
}

// ListPrompts returns prompts matching f.
func (s *Service) ListPrompts(f PromptFilter) ([]promptView, error) {
	prompts, err := s.deps.Prompts.List()
	if err != nil {
		return nil, err
	}
	out := make([]promptView, 0, len(prompts))
	for _, p := range prompts {
		v := promptView{ID: p.ID, Description: p.Description, Tags: p.Tags, Body: p.Body, FilePath: p.FilePath}
		if !matchesPromptFilter(v, f) {
			continue
		}
		out = append(out, v)
	}
	return out, nil
}

// promptView is exported as PromptView via type alias for naming clarity.
type promptView struct {
	ID          string
	Description string
	Tags        []string
	Body        string
	FilePath    string
}

// PromptView is the public type returned by ListPrompts.
type PromptView = promptView

func matchesPromptFilter(v promptView, f PromptFilter) bool {
	if len(f.Tags) > 0 && !hasAnyTag(v.Tags, f.Tags) {
		return false
	}
	if f.Query != "" {
		q := strings.ToLower(f.Query)
		if !strings.Contains(strings.ToLower(v.ID), q) &&
			!strings.Contains(strings.ToLower(v.Description), q) &&
			!strings.Contains(strings.ToLower(v.Body), q) {
			return false
		}
	}
	return true
}

func hasAnyTag(have, want []string) bool {
	if len(have) == 0 {
		return false
	}
	set := make(map[string]bool, len(have))
	for _, t := range have {
		set[strings.ToLower(t)] = true
	}
	for _, t := range want {
		if set[strings.ToLower(t)] {
			return true
		}
	}
	return false
}

// ListPlugins returns plugins matching f, with derived state from the registry.
func (s *Service) ListPlugins(f PluginFilter) ([]PluginView, error) {
	plugins, err := s.deps.Plugins.List()
	if err != nil {
		return nil, err
	}
	loadedSet, err := s.loadedPluginSet()
	if err != nil {
		return nil, err
	}
	out := make([]PluginView, 0, len(plugins))
	for _, p := range plugins {
		state := PluginStateInstalled
		if loadedSet[p.FullName()] {
			state = PluginStateLoaded
		}
		v := PluginView{
			Name:         p.Name,
			Marketplace:  p.Marketplace,
			Version:      p.Version,
			InstallPath:  p.InstallPath,
			GitCommitSha: p.GitCommitSha,
			RemoteURL:    p.RemoteURL,
			State:        state,
		}
		if !matchesPluginFilter(v, f) {
			continue
		}
		out = append(out, v)
	}
	return out, nil
}

func (s *Service) loadedPluginSet() (map[string]bool, error) {
	if s.deps.Registry == nil {
		return map[string]bool{}, nil
	}
	all, err := s.deps.Registry.List()
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(all))
	for k := range all {
		out[k] = true
	}
	return out, nil
}

func matchesPluginFilter(v PluginView, f PluginFilter) bool {
	if f.State != "" && v.State != f.State {
		return false
	}
	if f.Query != "" {
		q := strings.ToLower(f.Query)
		if !strings.Contains(strings.ToLower(v.Name), q) &&
			!strings.Contains(strings.ToLower(v.Marketplace), q) {
			return false
		}
	}
	return true
}
