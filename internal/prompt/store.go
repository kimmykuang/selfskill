package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimmykuang/selfskill/internal/frontmatter"
)

// Store manages prompt files under a base directory.
type Store struct {
	dir string // ~/ss/prompts/
}

func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

// List returns all installed prompts.
func (s *Store) List() ([]Prompt, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var prompts []Prompt
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		p, err := s.loadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			continue
		}
		prompts = append(prompts, *p)
	}
	return prompts, nil
}

// Get reads a prompt by ID.
func (s *Store) Get(id string) (*Prompt, error) {
	path := filepath.Join(s.dir, id+".md")
	return s.loadFile(path)
}

// Remove deletes a prompt file.
func (s *Store) Remove(id string) error {
	path := filepath.Join(s.dir, id+".md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("prompt %q not found", id)
	}
	return os.Remove(path)
}

// Save writes a prompt to disk.
func (s *Store) Save(p *Prompt) error {
	meta := map[string]interface{}{
		"id":          p.ID,
		"tags":        p.Tags,
		"description": p.Description,
	}
	data, err := frontmatter.Marshal(meta, p.Body)
	if err != nil {
		return err
	}
	path := filepath.Join(s.dir, p.ID+".md")
	return os.WriteFile(path, data, 0644)
}

// Search filters prompts by query string and/or tags.
func (s *Store) Search(query string, tags []string) ([]Prompt, error) {
	all, err := s.List()
	if err != nil {
		return nil, err
	}

	var results []Prompt
	for _, p := range all {
		if query != "" && !containsIgnoreCase(p.ID, query) && !containsIgnoreCase(p.Description, query) && !containsIgnoreCase(p.Body, query) {
			continue
		}
		if len(tags) > 0 && !hasAnyTag(p.Tags, tags) {
			continue
		}
		results = append(results, p)
	}
	return results, nil
}

// Exists checks if a prompt is installed.
func (s *Store) Exists(id string) bool {
	path := filepath.Join(s.dir, id+".md")
	_, err := os.Stat(path)
	return err == nil
}

func (s *Store) loadFile(path string) (*Prompt, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	meta, body, err := frontmatter.Parse(data)
	if err != nil {
		return nil, err
	}

	p := &Prompt{
		Body:     body,
		FilePath: path,
	}

	// Derive ID from filename if not in frontmatter
	base := filepath.Base(path)
	p.ID = strings.TrimSuffix(base, ".md")

	if meta != nil {
		if v, ok := meta["id"].(string); ok && v != "" {
			p.ID = v
		}
		if v, ok := meta["description"].(string); ok {
			p.Description = v
		}
		if v, ok := meta["tags"].([]interface{}); ok {
			for _, t := range v {
				if ts, ok := t.(string); ok {
					p.Tags = append(p.Tags, ts)
				}
			}
		}
	}

	return p, nil
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func hasAnyTag(have, want []string) bool {
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
