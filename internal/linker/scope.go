package linker

// Scope represents where skills are activated.
type Scope string

const (
	ScopeUser    Scope = "user"
	ScopeProject Scope = "project"
)

// LinkResult represents the outcome of a link/unlink operation for one skill.
type LinkResult struct {
	SkillName string
	Action    string // "linked", "skipped", "removed", "error"
	Detail    string
}

// LinkInfo represents an active symlink in a scope directory.
type LinkInfo struct {
	SkillName  string
	SourcePath string // ~/ss/skills/<name>/
	LinkPath   string // target symlink path
}

// SkillEntry represents any skill visible to CC, regardless of how it was installed.
type SkillEntry struct {
	Name   string
	Path   string
	Source string // "ss", "manual", "plugin"
}

// FullStatus contains all skills visible to CC across all sources.
type FullStatus struct {
	UserScope    []SkillEntry
	ProjectScope []SkillEntry
	Plugins      []SkillEntry
}

