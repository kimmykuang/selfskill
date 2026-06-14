package group

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Store manages group YAML files under a base directory.
type Store struct {
	dir string // ~/ss/groups/
}

func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

// Create creates a new empty group.
func (s *Store) Create(name string) error {
	path := s.groupPath(name)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("group %q already exists", name)
	}

	g := &Group{
		Name:   name,
		Skills: []string{},
	}
	return s.save(g)
}

// Get reads a group by name.
func (s *Store) Get(name string) (*Group, error) {
	path := s.groupPath(name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("group %q not found", name)
		}
		return nil, err
	}

	var g Group
	if err := yaml.Unmarshal(data, &g); err != nil {
		return nil, fmt.Errorf("parsing group %q: %w", name, err)
	}
	return &g, nil
}

// List returns all groups.
func (s *Store) List() ([]Group, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var groups []Group
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".yaml")
		g, err := s.Get(name)
		if err != nil {
			continue
		}
		groups = append(groups, *g)
	}
	return groups, nil
}

// Delete removes a group config file.
func (s *Store) Delete(name string) error {
	path := s.groupPath(name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("group %q not found", name)
	}
	return os.Remove(path)
}

// AddSkill adds a skill to a group (deduplicates).
func (s *Store) AddSkill(groupName, skillName string) error {
	g, err := s.Get(groupName)
	if err != nil {
		return err
	}

	for _, sk := range g.Skills {
		if sk == skillName {
			return nil // already in group
		}
	}

	g.Skills = append(g.Skills, skillName)
	return s.save(g)
}

// RemoveSkill removes a skill from a group.
func (s *Store) RemoveSkill(groupName, skillName string) error {
	g, err := s.Get(groupName)
	if err != nil {
		return err
	}

	filtered := g.Skills[:0]
	for _, sk := range g.Skills {
		if sk != skillName {
			filtered = append(filtered, sk)
		}
	}
	g.Skills = filtered
	return s.save(g)
}

// AddPlugin adds a plugin to a group (deduplicates).
func (s *Store) AddPlugin(groupName, pluginName string) error {
	g, err := s.Get(groupName)
	if err != nil {
		return err
	}

	for _, p := range g.Plugins {
		if p == pluginName {
			return nil // already in group
		}
	}

	g.Plugins = append(g.Plugins, pluginName)
	return s.save(g)
}

// RemovePlugin removes a plugin from a group.
func (s *Store) RemovePlugin(groupName, pluginName string) error {
	g, err := s.Get(groupName)
	if err != nil {
		return err
	}

	filtered := g.Plugins[:0]
	for _, p := range g.Plugins {
		if p != pluginName {
			filtered = append(filtered, p)
		}
	}
	g.Plugins = filtered
	return s.save(g)
}

// AddPrompt adds a prompt id to a group (deduplicates).
func (s *Store) AddPrompt(groupName, promptID string) error {
	g, err := s.Get(groupName)
	if err != nil {
		return err
	}

	for _, id := range g.Prompts {
		if id == promptID {
			return nil // already in group
		}
	}

	g.Prompts = append(g.Prompts, promptID)
	return s.save(g)
}

// RemovePrompt removes a prompt id from a group.
func (s *Store) RemovePrompt(groupName, promptID string) error {
	g, err := s.Get(groupName)
	if err != nil {
		return err
	}

	filtered := g.Prompts[:0]
	for _, id := range g.Prompts {
		if id != promptID {
			filtered = append(filtered, id)
		}
	}
	g.Prompts = filtered
	return s.save(g)
}

func (s *Store) save(g *Group) error {
	data, err := yaml.Marshal(g)
	if err != nil {
		return err
	}
	return os.WriteFile(s.groupPath(g.Name), data, 0644)
}

func (s *Store) groupPath(name string) string {
	return filepath.Join(s.dir, name+".yaml")
}
