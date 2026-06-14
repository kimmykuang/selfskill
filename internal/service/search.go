package service

// Hit is a unified search result across skill / prompt / plugin.
type Hit struct {
	Kind        string // "skill" | "prompt" | "plugin"
	Name        string // name or id
	Description string
}

// Search returns hits across all asset types whose name or description
// matches the query (case-insensitive substring). Empty query returns no hits.
func (s *Service) Search(query string) ([]Hit, error) {
	if query == "" {
		return nil, nil
	}
	var hits []Hit

	skills, err := s.ListSkills(SkillFilter{Query: query})
	if err != nil {
		return nil, err
	}
	for _, sk := range skills {
		hits = append(hits, Hit{Kind: "skill", Name: sk.Name, Description: sk.Description})
	}

	prompts, err := s.ListPrompts(PromptFilter{Query: query})
	if err != nil {
		return nil, err
	}
	for _, p := range prompts {
		hits = append(hits, Hit{Kind: "prompt", Name: p.ID, Description: p.Description})
	}

	plugins, err := s.ListPlugins(PluginFilter{Query: query})
	if err != nil {
		return nil, err
	}
	for _, p := range plugins {
		hits = append(hits, Hit{Kind: "plugin", Name: p.Name + "@" + p.Marketplace})
	}

	return hits, nil
}
