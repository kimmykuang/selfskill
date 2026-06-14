// Package cc_compat is the only place in ss that touches CC's private
// installed_plugins.json file. Centralizing the file format here means a
// CC schema change is a one-file fix, not a fleet refactor.
package cc_compat

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// Entry mirrors a single object in installed_plugins.json under
// the "plugins"[name] array. Fields match CC's current (version: 2) format.
type Entry struct {
	Scope        string `json:"scope"`
	ProjectPath  string `json:"projectPath,omitempty"`
	InstallPath  string `json:"installPath"`
	Version      string `json:"version"`
	InstalledAt  string `json:"installedAt"`
	LastUpdated  string `json:"lastUpdated"`
	GitCommitSha string `json:"gitCommitSha,omitempty"`
}

// file mirrors the on-disk structure.
type file struct {
	Version int                `json:"version"`
	Plugins map[string][]Entry `json:"plugins"`
}

// Registry is the stable interface ss uses internally; behind it lives
// whatever CC currently expects on disk.
type Registry interface {
	List() (map[string][]Entry, error)
	Add(name string, entry Entry) error
	Remove(name string) error
}

// JSONRegistry implements Registry against the current (version: 2) JSON file.
type JSONRegistry struct {
	path string
}

func NewJSONRegistry(path string) *JSONRegistry {
	return &JSONRegistry{path: path}
}

// List returns a snapshot of every entry. Missing file returns an empty map.
func (r *JSONRegistry) List() (map[string][]Entry, error) {
	var out map[string][]Entry
	if err := r.withReadLock(func(f *file) error {
		out = f.Plugins
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

// Add inserts or replaces the entry list for a single plugin name.
// If entry.InstalledAt is empty, the existing InstalledAt is preserved;
// on first install with no prior entry, InstalledAt defaults to LastUpdated.
func (r *JSONRegistry) Add(name string, entry Entry) error {
	return r.withLock(func(f *file) error {
		if entry.InstalledAt == "" {
			if existing, ok := f.Plugins[name]; ok && len(existing) > 0 && existing[0].InstalledAt != "" {
				entry.InstalledAt = existing[0].InstalledAt
			} else {
				entry.InstalledAt = entry.LastUpdated
			}
		}
		f.Plugins[name] = []Entry{entry}
		return nil
	})
}

// Remove deletes the entry list for a single plugin name.
// Removing a non-existent name is a no-op (matches map delete).
func (r *JSONRegistry) Remove(name string) error {
	return r.withLock(func(f *file) error {
		delete(f.Plugins, name)
		return nil
	})
}

// withReadLock acquires a shared lock for read-only access. Use withLock for mutations.
func (r *JSONRegistry) withReadLock(fn func(*file) error) error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0755); err != nil {
		return err
	}

	lockPath := r.path + ".lock"
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("creating lock file: %w", err)
	}
	defer lockFile.Close()

	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_SH); err != nil {
		return fmt.Errorf("acquiring read lock: %w", err)
	}
	defer func() { _ = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN) }()

	f, err := r.read()
	if err != nil {
		return err
	}
	return fn(f)
}

func (r *JSONRegistry) withLock(fn func(*file) error) error {
	if err := os.MkdirAll(filepath.Dir(r.path), 0755); err != nil {
		return err
	}

	lockPath := r.path + ".lock"
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("creating lock file: %w", err)
	}
	defer lockFile.Close()

	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("acquiring lock: %w", err)
	}
	defer func() { _ = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN) }()

	f, err := r.read()
	if err != nil {
		return err
	}
	if err := fn(f); err != nil {
		return err
	}
	return r.write(f)
}

func (r *JSONRegistry) read() (*file, error) {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &file{Version: 2, Plugins: map[string][]Entry{}}, nil
		}
		return nil, fmt.Errorf("reading installed plugins: %w", err)
	}

	var f file
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parsing installed plugins: %w", err)
	}
	if f.Plugins == nil {
		f.Plugins = map[string][]Entry{}
	}
	return &f, nil
}

func (r *JSONRegistry) write(f *file) error {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp := r.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}

	renamed := false
	defer func() {
		if !renamed {
			_ = os.Remove(tmp)
		}
	}()

	if err := os.Rename(tmp, r.path); err != nil {
		return err
	}
	renamed = true
	return nil
}
