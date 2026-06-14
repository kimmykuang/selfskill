package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kimmykuang/selfskill/internal/config"
	"github.com/kimmykuang/selfskill/internal/plugin/cc_compat"
)

// Loader manages symlinks into ~/.claude/plugins/cache/ and delegates
// installed_plugins.json maintenance to a cc_compat.Registry.
type Loader struct {
	store    *Store
	registry cc_compat.Registry
}

// NewLoader uses the default JSON registry at config.CCInstalledPluginsFile().
func NewLoader(store *Store) *Loader {
	return &Loader{
		store:    store,
		registry: cc_compat.NewJSONRegistry(config.CCInstalledPluginsFile()),
	}
}

// NewLoaderWithRegistry is a constructor for tests that want to inject a
// fake registry; production code should use NewLoader.
func NewLoaderWithRegistry(store *Store, registry cc_compat.Registry) *Loader {
	return &Loader{store: store, registry: registry}
}

// Load makes a plugin visible to CC by creating a symlink and writing JSON.
func (l *Loader) Load(name string) error {
	p, err := l.store.Get(name)
	if err != nil {
		return err
	}

	ccCachePath := filepath.Join(config.CCPluginsCacheDir(), p.Marketplace, p.Name, p.Version)
	if err := l.createSymlink(p.InstallPath, ccCachePath); err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	entry := cc_compat.Entry{
		Scope:        "user",
		InstallPath:  ccCachePath,
		Version:      p.Version,
		InstalledAt:  now,
		LastUpdated:  now,
		GitCommitSha: p.GitCommitSha,
	}
	if err := l.registry.Add(p.FullName(), entry); err != nil {
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

	ccCachePath := filepath.Join(config.CCPluginsCacheDir(), p.Marketplace, p.Name, p.Version)
	if isSymlink(ccCachePath) {
		if err := os.Remove(ccCachePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing symlink: %w", err)
		}
		pluginDir := filepath.Dir(ccCachePath)
		removeIfEmpty(pluginDir)
		mpDir := filepath.Dir(pluginDir)
		removeIfEmpty(mpDir)
	}

	return l.registry.Remove(p.FullName())
}

// createSymlink creates the symlink with parent directories.
func (l *Loader) createSymlink(source, link string) error {
	if err := os.MkdirAll(filepath.Dir(link), 0755); err != nil {
		return fmt.Errorf("creating cache dir: %w", err)
	}

	info, err := os.Lstat(link)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			target, _ := os.Readlink(link)
			absTarget, _ := filepath.Abs(target)
			absSource, _ := filepath.Abs(source)
			if absTarget == absSource {
				return nil
			}
			os.Remove(link)
		} else {
			return fmt.Errorf("path %s exists and is not a symlink", link)
		}
	}

	return os.Symlink(source, link)
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
