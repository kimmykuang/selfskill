package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/kimmykuang/selfskill/internal/config"
)

// InstalledPluginEntry matches the CC installed_plugins.json entry format.
type InstalledPluginEntry struct {
	Scope        string `json:"scope"`
	ProjectPath  string `json:"projectPath,omitempty"`
	InstallPath  string `json:"installPath"`
	Version      string `json:"version"`
	InstalledAt  string `json:"installedAt"`
	LastUpdated  string `json:"lastUpdated"`
	GitCommitSha string `json:"gitCommitSha,omitempty"`
}

// InstalledPluginsFile represents the CC installed_plugins.json structure.
type InstalledPluginsFile struct {
	Version int                             `json:"version"`
	Plugins map[string][]InstalledPluginEntry `json:"plugins"`
}

// Loader manages symlinks and installed_plugins.json for loading plugins into CC.
type Loader struct {
	store *Store
}

func NewLoader(store *Store) *Loader {
	return &Loader{store: store}
}

// Load makes a plugin visible to CC by creating a symlink and writing JSON.
func (l *Loader) Load(name string) error {
	p, err := l.store.Get(name)
	if err != nil {
		return err
	}

	// Create symlink: ~/.claude/plugins/cache/<marketplace>/<plugin>/<version> -> ~/ss/plugins/...
	ccCachePath := filepath.Join(config.CCPluginsCacheDir(), p.Marketplace, p.Name, p.Version)
	if err := l.createSymlink(p.InstallPath, ccCachePath); err != nil {
		return err
	}

	// Write entry to installed_plugins.json
	if err := l.addJSONEntry(p); err != nil {
		// Rollback symlink on JSON failure
		os.Remove(ccCachePath)
		return err
	}

	return nil
}

// Unload removes a plugin from CC by removing the symlink and JSON entry.
func (l *Loader) Unload(name string) error {
	p, err := l.store.Get(name)
	if err != nil {
		return err
	}

	// Remove symlink
	ccCachePath := filepath.Join(config.CCPluginsCacheDir(), p.Marketplace, p.Name, p.Version)
	if isSymlink(ccCachePath) {
		if err := os.Remove(ccCachePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing symlink: %w", err)
		}
		// Clean up empty parent dirs in cache
		pluginDir := filepath.Dir(ccCachePath)
		removeIfEmpty(pluginDir)
		mpDir := filepath.Dir(pluginDir)
		removeIfEmpty(mpDir)
	}

	// Remove JSON entry
	return l.removeJSONEntry(p)
}

// createSymlink creates the symlink with parent directories.
func (l *Loader) createSymlink(source, link string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(link), 0755); err != nil {
		return fmt.Errorf("creating cache dir: %w", err)
	}

	// Check if link already exists
	info, err := os.Lstat(link)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			target, _ := os.Readlink(link)
			absTarget, _ := filepath.Abs(target)
			absSource, _ := filepath.Abs(source)
			if absTarget == absSource {
				return nil // already linked correctly
			}
			// Remove existing symlink pointing elsewhere
			os.Remove(link)
		} else {
			return fmt.Errorf("path %s exists and is not a symlink", link)
		}
	}

	return os.Symlink(source, link)
}

// addJSONEntry adds or updates the plugin entry in installed_plugins.json.
func (l *Loader) addJSONEntry(p *Plugin) error {
	return l.withJSONLock(func(f *InstalledPluginsFile) error {
		key := p.FullName()
		now := time.Now().UTC().Format(time.RFC3339Nano)

		// Preserve InstalledAt if entry already exists
		installedAt := now
		if existing, ok := f.Plugins[key]; ok && len(existing) > 0 && existing[0].InstalledAt != "" {
			installedAt = existing[0].InstalledAt
		}

		entry := InstalledPluginEntry{
			Scope:        "user",
			InstallPath:  filepath.Join(config.CCPluginsCacheDir(), p.Marketplace, p.Name, p.Version),
			Version:      p.Version,
			InstalledAt:  installedAt,
			LastUpdated:  now,
			GitCommitSha: p.GitCommitSha,
		}

		f.Plugins[key] = []InstalledPluginEntry{entry}
		return nil
	})
}

// removeJSONEntry removes the plugin entry from installed_plugins.json.
func (l *Loader) removeJSONEntry(p *Plugin) error {
	return l.withJSONLock(func(f *InstalledPluginsFile) error {
		delete(f.Plugins, p.FullName())
		return nil
	})
}

// withJSONLock performs a locked read-modify-write on installed_plugins.json.
func (l *Loader) withJSONLock(fn func(*InstalledPluginsFile) error) error {
	path := config.CCInstalledPluginsFile()

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	lockPath := path + ".lock"
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("creating lock file: %w", err)
	}
	defer lockFile.Close()

	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("acquiring lock: %w", err)
	}
	defer syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)

	f, err := l.readJSON()
	if err != nil {
		return err
	}

	if err := fn(f); err != nil {
		return err
	}

	return l.writeJSON(f)
}

// readJSON reads and parses installed_plugins.json.
func (l *Loader) readJSON() (*InstalledPluginsFile, error) {
	path := config.CCInstalledPluginsFile()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &InstalledPluginsFile{
				Version: 2,
				Plugins: make(map[string][]InstalledPluginEntry),
			}, nil
		}
		return nil, fmt.Errorf("reading installed plugins: %w", err)
	}

	var f InstalledPluginsFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parsing installed plugins: %w", err)
	}
	if f.Plugins == nil {
		f.Plugins = make(map[string][]InstalledPluginEntry)
	}
	return &f, nil
}

// writeJSON atomically writes installed_plugins.json.
func (l *Loader) writeJSON(f *InstalledPluginsFile) error {
	path := config.CCInstalledPluginsFile()
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	// Write to temp file then rename for atomicity
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

func removeIfEmpty(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	if len(entries) == 0 {
		os.Remove(dir)
	}
}
