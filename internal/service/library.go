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
