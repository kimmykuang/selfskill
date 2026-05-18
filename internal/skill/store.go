package skill

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/kimmykuang/selfskill/internal/frontmatter"
)

const skillFile = "SKILL.md"

// Store manages skill directories under a base directory.
type Store struct {
	dir string // ~/ss/skills/
}

func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

// List returns all installed skills.
func (s *Store) List() ([]Skill, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var skills []Skill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		sk, err := s.Get(e.Name())
		if err != nil {
			continue // skip invalid entries
		}
		skills = append(skills, *sk)
	}
	return skills, nil
}

// Get reads a skill by name.
func (s *Store) Get(name string) (*Skill, error) {
	dirPath := filepath.Join(s.dir, name)
	mdPath := filepath.Join(dirPath, skillFile)

	data, err := os.ReadFile(mdPath)
	if err != nil {
		return nil, fmt.Errorf("reading skill %q: %w", name, err)
	}

	meta, _, err := frontmatter.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parsing skill %q frontmatter: %w", name, err)
	}

	sk := &Skill{
		Name:    name,
		DirPath: dirPath,
	}
	if meta != nil {
		if v, ok := meta["version"].(string); ok {
			sk.Version = v
		}
		if v, ok := meta["description"].(string); ok {
			sk.Description = v
		}
		if v, ok := meta["source"].(string); ok {
			sk.Source = v
		}
	}

	// Check subdirectories
	sk.HasAssets = dirExists(filepath.Join(dirPath, "assets"))
	sk.HasRef = dirExists(filepath.Join(dirPath, "reference"))
	sk.HasTests = dirExists(filepath.Join(dirPath, "tests"))
	sk.HasScripts = dirExists(filepath.Join(dirPath, "scripts"))

	return sk, nil
}

// Remove deletes a skill directory.
func (s *Store) Remove(name string) error {
	dirPath := filepath.Join(s.dir, name)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return fmt.Errorf("skill %q not found", name)
	}
	return os.RemoveAll(dirPath)
}

// Save copies an entire source directory as a skill.
func (s *Store) Save(name string, srcDir string) error {
	destDir := filepath.Join(s.dir, name)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	if err := copyDir(srcDir, destDir); err != nil {
		os.RemoveAll(destDir) // clean up partial copy
		return err
	}
	return nil
}

// SaveFromContent creates a skill directory with just a SKILL.md file.
func (s *Store) SaveFromContent(name string, content []byte) error {
	destDir := filepath.Join(s.dir, name)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(destDir, skillFile), content, 0644)
}

// Exists checks if a skill is installed.
func (s *Store) Exists(name string) bool {
	mdPath := filepath.Join(s.dir, name, skillFile)
	_, err := os.Stat(mdPath)
	return err == nil
}

// SkillDir returns the absolute path for a skill.
func (s *Store) SkillDir(name string) string {
	return filepath.Join(s.dir, name)
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		return copyFile(path, destPath)
	})
}

func copyFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
