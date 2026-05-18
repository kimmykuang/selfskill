package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	homeDir     string
	homeDirOnce sync.Once
)

func userHome() string {
	homeDirOnce.Do(func() {
		var err error
		homeDir, err = os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "fatal: cannot determine home directory: %v\n", err)
			os.Exit(1)
		}
	})
	return homeDir
}

// SSHome returns the root directory for ss data: ~/ss/
func SSHome() string {
	return filepath.Join(userHome(), "ss")
}

// SkillsDir returns the skills storage directory: ~/ss/skills/
func SkillsDir() string {
	return filepath.Join(SSHome(), "skills")
}

// PromptsDir returns the prompts storage directory: ~/ss/prompts/
func PromptsDir() string {
	return filepath.Join(SSHome(), "prompts")
}

// GroupsDir returns the groups config directory: ~/ss/groups/
func GroupsDir() string {
	return filepath.Join(SSHome(), "groups")
}

// PluginsDir returns the plugins storage directory: ~/ss/plugins/
func PluginsDir() string {
	return filepath.Join(SSHome(), "plugins")
}

// MarketplacesDir returns the ss marketplaces directory: ~/ss/marketplaces/
func MarketplacesDir() string {
	return filepath.Join(SSHome(), "marketplaces")
}

// CCPluginsCacheDir returns the CC plugins cache directory: ~/.claude/plugins/cache/
func CCPluginsCacheDir() string {
	return filepath.Join(userHome(), ".claude", "plugins", "cache")
}

// CCInstalledPluginsFile returns the CC installed plugins JSON path.
func CCInstalledPluginsFile() string {
	return filepath.Join(userHome(), ".claude", "plugins", "installed_plugins.json")
}

// ConfigFile returns the global config file path: ~/ss/ss.yaml
func ConfigFile() string {
	return filepath.Join(SSHome(), "ss.yaml")
}

// UserSkillsDir returns the CC user-scope skills directory: ~/.claude/skills/
func UserSkillsDir() string {
	return filepath.Join(userHome(), ".claude", "skills")
}

// ProjectSkillsDir returns the CC project-scope skills directory: <cwd>/.claude/skills/
func ProjectSkillsDir() string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, ".claude", "skills")
}

// EnsureDirs creates all required directories if they don't exist.
func EnsureDirs() error {
	dirs := []string{
		SSHome(),
		SkillsDir(),
		PromptsDir(),
		GroupsDir(),
		PluginsDir(),
		MarketplacesDir(),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	return nil
}
