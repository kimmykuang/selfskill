package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kimmykuang/selfskill/internal/config"
)

const metaFileName = ".ss-meta.json"

// Store manages plugin source files under ~/ss/plugins/.
type Store struct {
	dir string // ~/ss/plugins/
}

func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

// List returns all installed plugins across all marketplaces.
func (s *Store) List() ([]Plugin, error) {
	marketplaces, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var plugins []Plugin
	for _, mp := range marketplaces {
		if !mp.IsDir() {
			continue
		}
		mpDir := filepath.Join(s.dir, mp.Name())
		pluginDirs, err := os.ReadDir(mpDir)
		if err != nil {
			continue
		}
		for _, pd := range pluginDirs {
			if !pd.IsDir() {
				continue
			}
			// Find version subdirectory
			versions, err := os.ReadDir(filepath.Join(mpDir, pd.Name()))
			if err != nil {
				continue
			}
			for _, v := range versions {
				if !v.IsDir() {
					continue
				}
				p := Plugin{
					Name:        pd.Name(),
					Marketplace: mp.Name(),
					Version:     v.Name(),
					InstallPath: filepath.Join(mpDir, pd.Name(), v.Name()),
				}
				applyMeta(&p)
				plugins = append(plugins, p)
			}
		}
	}
	return plugins, nil
}

// Get returns a plugin by name. Accepts "superpowers" or "superpowers@claude-plugins-official".
func (s *Store) Get(name string) (*Plugin, error) {
	pluginName, marketplace := parseName(name)

	// Direct lookup if marketplace is specified
	if marketplace != "" {
		mpDir := filepath.Join(s.dir, marketplace, pluginName)
		versions, err := os.ReadDir(mpDir)
		if err != nil {
			return nil, fmt.Errorf("plugin %q not found in %s", name, s.dir)
		}
		for _, v := range versions {
			if v.IsDir() {
				p := &Plugin{
					Name:        pluginName,
					Marketplace: marketplace,
					Version:     v.Name(),
					InstallPath: filepath.Join(mpDir, v.Name()),
				}
				applyMeta(p)
				return p, nil
			}
		}
		return nil, fmt.Errorf("plugin %q not found in %s", name, s.dir)
	}

	// Scan all marketplaces when marketplace is unspecified
	marketplaces, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("plugin %q not found in %s", name, s.dir)
	}
	for _, mp := range marketplaces {
		if !mp.IsDir() {
			continue
		}
		mpDir := filepath.Join(s.dir, mp.Name(), pluginName)
		versions, err := os.ReadDir(mpDir)
		if err != nil {
			continue
		}
		for _, v := range versions {
			if v.IsDir() {
				p := &Plugin{
					Name:        pluginName,
					Marketplace: mp.Name(),
					Version:     v.Name(),
					InstallPath: filepath.Join(mpDir, v.Name()),
				}
				applyMeta(p)
				return p, nil
			}
		}
	}
	return nil, fmt.Errorf("plugin %q not found in %s", name, s.dir)
}

// Exists checks if a plugin is installed.
func (s *Store) Exists(name string) bool {
	_, err := s.Get(name)
	return err == nil
}

// PluginDir returns the expected install path for a plugin.
func (s *Store) PluginDir(marketplace, name, version string) string {
	return filepath.Join(s.dir, marketplace, name, version)
}

// Remove deletes a plugin's source files.
func (s *Store) Remove(name string) error {
	p, err := s.Get(name)
	if err != nil {
		return err
	}

	// Remove the version directory
	if err := os.RemoveAll(p.InstallPath); err != nil {
		return err
	}

	// Clean up empty parent directories
	pluginDir := filepath.Dir(p.InstallPath)
	removeIfEmpty(pluginDir)
	mpDir := filepath.Dir(pluginDir)
	removeIfEmpty(mpDir)
	return nil
}

// IsLoaded checks if a plugin is currently loaded (has symlink in CC cache).
func (s *Store) IsLoaded(name string) bool {
	p, err := s.Get(name)
	if err != nil {
		return false
	}

	// Check if symlink exists in CC cache pointing to our install path
	ccCachePath := filepath.Join(config.CCPluginsCacheDir(), p.Marketplace, p.Name, p.Version)
	info, err := os.Lstat(ccCachePath)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// parseName splits "superpowers@claude-plugins-official" into name and marketplace.
func parseName(input string) (name, marketplace string) {
	parts := strings.SplitN(input, "@", 2)
	name = parts[0]
	if len(parts) == 2 {
		marketplace = parts[1]
	}
	return
}

// InstallFromGit clones a plugin from a git URL into ~/ss/plugins/<marketplace>/<plugin>/<version>/.
func (s *Store) InstallFromGit(gitURL, marketplace, pluginName, version string) (*Plugin, error) {
	if err := validateGitURL(gitURL); err != nil {
		return nil, err
	}

	destDir := s.PluginDir(marketplace, pluginName, version)

	// Check if already installed
	if _, err := os.Stat(destDir); err == nil {
		return nil, fmt.Errorf("plugin %s@%s version %s already installed at %s", pluginName, marketplace, version, destDir)
	}

	if err := os.MkdirAll(filepath.Dir(destDir), 0755); err != nil {
		return nil, fmt.Errorf("creating plugin dir: %w", err)
	}

	// Git clone
	var cmd *exec.Cmd
	if version == "" || version == "latest" {
		// No specific version — shallow clone default branch
		cmd = exec.Command("git", "clone", "--depth", "1", gitURL, destDir)
	} else if isTag(version) {
		cmd = exec.Command("git", "clone", "--depth", "1", "--branch", version, gitURL, destDir)
	} else {
		cmd = exec.Command("git", "clone", gitURL, destDir)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Clean up on failure
		os.RemoveAll(destDir)
		return nil, fmt.Errorf("git clone failed: %w", err)
	}

	// If version is a commit sha, checkout it
	if version != "" && version != "latest" && !isTag(version) && len(version) >= 7 {
		checkout := exec.Command("git", "-C", destDir, "checkout", version)
		checkout.Stdout = os.Stdout
		checkout.Stderr = os.Stderr
		if err := checkout.Run(); err != nil {
			os.RemoveAll(destDir)
			return nil, fmt.Errorf("git checkout %s failed: %w", version, err)
		}
	}

	// Get the actual commit sha before .git is removed.
	sha := getGitSha(destDir)

	// Persist install metadata.
	meta := &PluginMeta{
		URL:         gitURL,
		Marketplace: marketplace,
		CommitSha:   sha,
		Version:     version,
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := writeMeta(destDir, meta); err != nil {
		os.RemoveAll(destDir)
		return nil, fmt.Errorf("writing plugin meta: %w", err)
	}

	// Drop the .git directory — we no longer rely on it.
	if err := os.RemoveAll(filepath.Join(destDir, ".git")); err != nil {
		os.RemoveAll(destDir)
		return nil, fmt.Errorf("removing .git: %w", err)
	}

	return &Plugin{
		Name:         pluginName,
		Marketplace:  marketplace,
		Version:      version,
		InstallPath:  destDir,
		GitCommitSha: sha,
		RemoteURL:    gitURL,
	}, nil
}

// InstallFromLocal copies a local plugin directory into ~/ss/plugins/<marketplace>/<plugin>/<version>/.
func (s *Store) InstallFromLocal(srcDir, marketplace, pluginName, version string) (*Plugin, error) {
	destDir := s.PluginDir(marketplace, pluginName, version)

	if _, err := os.Stat(destDir); err == nil {
		return nil, fmt.Errorf("plugin %s@%s version %s already installed", pluginName, marketplace, version)
	}

	if err := os.MkdirAll(filepath.Dir(destDir), 0755); err != nil {
		return nil, fmt.Errorf("creating plugin dir: %w", err)
	}

	// Copy directory using cp -a
	cmd := exec.Command("cp", "-a", srcDir, destDir)
	if err := cmd.Run(); err != nil {
		os.RemoveAll(destDir)
		return nil, fmt.Errorf("copying plugin: %w", err)
	}

	// Read sha from a possibly-copied .git before deleting it.
	sha := getGitSha(destDir)

	meta := &PluginMeta{
		URL:         "",
		Marketplace: marketplace,
		CommitSha:   sha,
		Version:     version,
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := writeMeta(destDir, meta); err != nil {
		os.RemoveAll(destDir)
		return nil, fmt.Errorf("writing plugin meta: %w", err)
	}

	// Drop .git if it was copied along.
	if err := os.RemoveAll(filepath.Join(destDir, ".git")); err != nil {
		os.RemoveAll(destDir)
		return nil, fmt.Errorf("removing .git: %w", err)
	}

	return &Plugin{
		Name:         pluginName,
		Marketplace:  marketplace,
		Version:      version,
		InstallPath:  destDir,
		GitCommitSha: sha,
	}, nil
}

// GetMeta loads .ss-meta.json for a plugin. Returns (nil, nil) if the file does not exist.
func (s *Store) GetMeta(name string) (*PluginMeta, error) {
	p, err := s.Get(name)
	if err != nil {
		return nil, err
	}
	return readMeta(p.InstallPath)
}

// GetRemoteURL returns the remote git URL recorded in .ss-meta.json.
func (s *Store) GetRemoteURL(name string) (string, error) {
	meta, err := s.GetMeta(name)
	if err != nil {
		return "", err
	}
	if meta == nil || meta.URL == "" {
		return "", fmt.Errorf("no remote url for plugin %q", name)
	}
	return meta.URL, nil
}

// Update refreshes a plugin by re-cloning its remote into a temp dir and
// swapping the contents of the install path.
func (s *Store) Update(name string) error {
	p, err := s.Get(name)
	if err != nil {
		return err
	}

	meta, err := readMeta(p.InstallPath)
	if err != nil {
		return fmt.Errorf("reading plugin meta: %w", err)
	}
	if meta == nil || meta.URL == "" {
		return fmt.Errorf("plugin %q has no remote url, cannot update", name)
	}

	if err := validateGitURL(meta.URL); err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "ss-plugin-update-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cloneDir := filepath.Join(tmpDir, "src")
	cmd := exec.Command("git", "clone", "--depth", "1", meta.URL, cloneDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	// Capture sha before stripping .git.
	sha := getGitSha(cloneDir)

	// Drop .git from the freshly cloned tree before copying.
	if err := os.RemoveAll(filepath.Join(cloneDir, ".git")); err != nil {
		return fmt.Errorf("removing .git from clone: %w", err)
	}

	// Wipe the install path, preserving .ss-meta.json so we can rewrite it.
	entries, err := os.ReadDir(p.InstallPath)
	if err != nil {
		return fmt.Errorf("reading install path: %w", err)
	}
	for _, e := range entries {
		if e.Name() == metaFileName {
			continue
		}
		if err := os.RemoveAll(filepath.Join(p.InstallPath, e.Name())); err != nil {
			return fmt.Errorf("clearing install path: %w", err)
		}
	}

	// Copy the cloned contents into the install path.
	srcEntries, err := os.ReadDir(cloneDir)
	if err != nil {
		return fmt.Errorf("reading clone dir: %w", err)
	}
	for _, e := range srcEntries {
		src := filepath.Join(cloneDir, e.Name())
		dst := filepath.Join(p.InstallPath, e.Name())
		cp := exec.Command("cp", "-a", src, dst)
		if err := cp.Run(); err != nil {
			return fmt.Errorf("copying %s: %w", e.Name(), err)
		}
	}

	// Refresh meta with the new sha and timestamp; preserve url/marketplace/version.
	meta.CommitSha = sha
	meta.InstalledAt = time.Now().UTC().Format(time.RFC3339)
	if err := writeMeta(p.InstallPath, meta); err != nil {
		return fmt.Errorf("writing plugin meta: %w", err)
	}

	return nil
}

// ListSkills returns skill directories within a plugin.
func (s *Store) ListSkills(name string) ([]string, error) {
	p, err := s.Get(name)
	if err != nil {
		return nil, err
	}

	skillsDir := filepath.Join(p.InstallPath, "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("no skills directory in plugin %q", name)
	}

	var skills []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		mdPath := filepath.Join(skillsDir, e.Name(), "SKILL.md")
		if _, err := os.Stat(mdPath); err == nil {
			skills = append(skills, e.Name())
		}
	}
	return skills, nil
}

func isTag(version string) bool {
	// Simple heuristic: if it looks like a semver or has dots, treat as tag
	return strings.Contains(version, ".")
}

func getGitSha(dir string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func validateGitURL(url string) error {
	if strings.HasPrefix(url, "-") {
		return fmt.Errorf("invalid git URL: %q", url)
	}
	if !strings.HasPrefix(url, "https://") &&
		!strings.HasPrefix(url, "http://") &&
		!strings.HasPrefix(url, "git@") &&
		!strings.HasPrefix(url, "file://") {
		return fmt.Errorf("only https://, http://, git@ and file:// URLs are supported, got: %q", url)
	}
	return nil
}

// writeMeta writes .ss-meta.json into installDir.
func writeMeta(installDir string, meta *PluginMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(installDir, metaFileName), data, 0644)
}

// readMeta loads .ss-meta.json from installDir. Returns (nil, nil) when the file is absent.
func readMeta(installDir string) (*PluginMeta, error) {
	path := filepath.Join(installDir, metaFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var m PluginMeta
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", metaFileName, err)
	}
	return &m, nil
}

// applyMeta backfills RemoteURL and GitCommitSha from .ss-meta.json when present.
func applyMeta(p *Plugin) {
	meta, err := readMeta(p.InstallPath)
	if err != nil || meta == nil {
		return
	}
	p.RemoteURL = meta.URL
	if p.GitCommitSha == "" {
		p.GitCommitSha = meta.CommitSha
	}
}
