package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kimmykuang/selfskill/internal/config"
)

// MarketplacePlugin represents a plugin entry in marketplace.json.
type MarketplacePlugin struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Source      interface{} `json:"source"` // string or object
	Category    string      `json:"category"`
}

// MarketplaceFile represents the marketplace.json structure.
type MarketplaceFile struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Plugins     []MarketplacePlugin `json:"plugins"`
}

// Registry manages marketplace repos under ~/ss/marketplaces/ and reads their metadata.
type Registry struct {
	dir string // ~/ss/marketplaces/
}

func NewRegistry() *Registry {
	return &Registry{dir: config.MarketplacesDir()}
}

// AddMarketplace clones a marketplace git repo into ~/ss/marketplaces/<name>/.
// The name is derived from the repo (last path segment without .git).
func (r *Registry) AddMarketplace(gitURL string) (string, error) {
	name := repoNameFromURL(gitURL)
	destDir := filepath.Join(r.dir, name)

	if _, err := os.Stat(destDir); err == nil {
		return name, fmt.Errorf("marketplace %q already exists at %s", name, destDir)
	}

	cmd := exec.Command("git", "clone", "--depth", "1", gitURL, destDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.RemoveAll(destDir)
		return "", fmt.Errorf("git clone marketplace: %w", err)
	}

	// Verify marketplace.json exists
	if _, err := r.findMarketplaceJSON(name); err != nil {
		os.RemoveAll(destDir)
		return "", fmt.Errorf("cloned repo has no marketplace.json: %w", err)
	}

	return name, nil
}

// RemoveMarketplace removes a marketplace repo.
func (r *Registry) RemoveMarketplace(name string) error {
	destDir := filepath.Join(r.dir, name)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		return fmt.Errorf("marketplace %q not found", name)
	}
	return os.RemoveAll(destDir)
}

// UpdateMarketplace pulls the latest changes for a marketplace repo.
func (r *Registry) UpdateMarketplace(name string) error {
	destDir := filepath.Join(r.dir, name)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		return fmt.Errorf("marketplace %q not found", name)
	}
	cmd := exec.Command("git", "-C", destDir, "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ListMarketplaces returns available marketplace names from ~/ss/marketplaces/.
func (r *Registry) ListMarketplaces() ([]string, error) {
	entries, err := os.ReadDir(r.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// LoadMarketplace reads and parses a marketplace.json file.
func (r *Registry) LoadMarketplace(name string) (*MarketplaceFile, error) {
	path, err := r.findMarketplaceJSON(name)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading marketplace %q: %w", name, err)
	}

	var mf MarketplaceFile
	if err := json.Unmarshal(data, &mf); err != nil {
		return nil, fmt.Errorf("parsing marketplace %q: %w", name, err)
	}
	return &mf, nil
}

// FindPlugin looks up a plugin in a specific marketplace.
func (r *Registry) FindPlugin(pluginName, marketplace string) (*MarketplacePlugin, error) {
	mf, err := r.LoadMarketplace(marketplace)
	if err != nil {
		return nil, err
	}

	for _, p := range mf.Plugins {
		if p.Name == pluginName {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("plugin %q not found in marketplace %q", pluginName, marketplace)
}

// ResolveGitURL extracts the git clone URL from a plugin's source field.
func (r *Registry) ResolveGitURL(mp *MarketplacePlugin, marketplace string) (string, error) {
	switch src := mp.Source.(type) {
	case string:
		// Relative path — local plugin in the marketplace repo
		return "", fmt.Errorf("plugin %q uses local source %q (part of marketplace repo, not independently cloneable)", mp.Name, src)
	case map[string]interface{}:
		if url, ok := src["url"].(string); ok {
			return url, nil
		}
		return "", fmt.Errorf("plugin %q source object has no url field", mp.Name)
	default:
		return "", fmt.Errorf("plugin %q has unknown source format", mp.Name)
	}
}

// IsLocalPlugin returns true if the plugin source is a relative path (part of marketplace repo).
func (r *Registry) IsLocalPlugin(mp *MarketplacePlugin) bool {
	_, ok := mp.Source.(string)
	return ok
}

// LocalPluginPath returns the absolute path to a local plugin within the marketplace repo.
func (r *Registry) LocalPluginPath(mp *MarketplacePlugin, marketplace string) (string, error) {
	src, ok := mp.Source.(string)
	if !ok {
		return "", fmt.Errorf("plugin %q is not a local plugin", mp.Name)
	}
	mpRepoDir := filepath.Join(r.dir, marketplace)
	return filepath.Join(mpRepoDir, src), nil
}

// findMarketplaceJSON locates the marketplace.json in a marketplace repo.
// Checks both .claude-plugin/marketplace.json and marketplace.json at root.
func (r *Registry) findMarketplaceJSON(name string) (string, error) {
	repoDir := filepath.Join(r.dir, name)

	// Try .claude-plugin/marketplace.json first (CC convention)
	path := filepath.Join(repoDir, ".claude-plugin", "marketplace.json")
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	// Try root marketplace.json
	path = filepath.Join(repoDir, "marketplace.json")
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("no marketplace.json found in %s", repoDir)
}

// repoNameFromURL extracts the repo name from a git URL.
// Handles both HTTPS and SSH-style URLs:
//   "https://github.com/anthropics/claude-plugins-official.git" -> "claude-plugins-official"
//   "git@github.com:anthropics/claude-plugins-official.git" -> "claude-plugins-official"
func repoNameFromURL(rawURL string) string {
	// Find the last path segment
	var base string
	if idx := strings.LastIndex(rawURL, "/"); idx >= 0 {
		base = rawURL[idx+1:]
	} else if idx := strings.LastIndex(rawURL, ":"); idx >= 0 {
		// SSH-style: git@github.com:user/repo.git
		remainder := rawURL[idx+1:]
		if slashIdx := strings.LastIndex(remainder, "/"); slashIdx >= 0 {
			base = remainder[slashIdx+1:]
		} else {
			base = remainder
		}
	} else {
		base = rawURL
	}
	return strings.TrimSuffix(base, ".git")
}
