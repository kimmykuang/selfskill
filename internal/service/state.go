package service

import "github.com/kimmykuang/selfskill/internal/linker"

// IsSkillLoaded reports whether `name` is currently linked into `scope`'s
// skills directory by ss. Skills linked manually by the user are NOT
// reported as loaded — the linker only counts symlinks pointing at our
// own skills source dir.
func (s *Service) IsSkillLoaded(name string, scope linker.Scope) bool {
	if s.deps.Linker == nil {
		return false
	}
	userLinks, projectLinks, err := s.deps.Linker.Status()
	if err != nil {
		return false
	}
	target := userLinks
	if scope == linker.ScopeProject {
		target = projectLinks
	}
	for _, l := range target {
		if l.SkillName == name {
			return true
		}
	}
	return false
}

// IsPluginLoaded reports whether `name` (the FullName form, e.g.
// "superpowers@claude-plugins-official") has an entry in cc_compat.Registry.
func (s *Service) IsPluginLoaded(name string) bool {
	if s.deps.Registry == nil {
		return false
	}
	all, err := s.deps.Registry.List()
	if err != nil {
		return false
	}
	_, ok := all[name]
	return ok
}
