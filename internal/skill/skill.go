package skill

// Skill represents a skill directory with metadata from SKILL.md frontmatter.
type Skill struct {
	Name        string // directory name
	Version     string
	Description string
	Source      string
	DirPath     string // absolute path to ~/ss/skills/<name>/
	HasAssets   bool
	HasRef      bool
	HasTests    bool
	HasScripts  bool
}
