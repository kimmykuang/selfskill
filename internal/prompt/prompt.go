package prompt

// Prompt represents a prompt file with metadata from frontmatter.
type Prompt struct {
	ID          string
	Tags        []string
	Description string
	Body        string
	FilePath    string // absolute path on disk
}
