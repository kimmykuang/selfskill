package group

// Group represents a named collection of skills, plugins, and prompts.
type Group struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Skills      []string `yaml:"skills"`
	Plugins     []string `yaml:"plugins,omitempty"`
	Prompts     []string `yaml:"prompts,omitempty"`
}
