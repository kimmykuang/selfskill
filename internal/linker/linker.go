package linker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kimmykuang/selfskill/internal/config"
)

// Linker manages symlinks between ~/ss/skills/ and CC skills directories.
type Linker struct {
	skillsSourceDir string // ~/ss/skills/
}

func New(skillsSourceDir string) *Linker {
	return &Linker{skillsSourceDir: skillsSourceDir}
}

// Load creates symlinks for the given skills in the target scope directory.
func (l *Linker) Load(skills []string, scope Scope) ([]LinkResult, error) {
	targetDir, err := l.TargetDir(scope)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("creating target dir: %w", err)
	}

	var results []LinkResult
	for _, name := range skills {
		r := l.linkOne(name, targetDir)
		results = append(results, r)
	}
	return results, nil
}

// Unload removes symlinks for the given skills from the target scope directory.
func (l *Linker) Unload(skills []string, scope Scope) ([]LinkResult, error) {
	targetDir, err := l.TargetDir(scope)
	if err != nil {
		return nil, err
	}

	var results []LinkResult
	for _, name := range skills {
		r := l.unlinkOne(name, targetDir)
		results = append(results, r)
	}
	return results, nil
}

// UnloadAll removes all ss-managed symlinks from the target scope directory.
func (l *Linker) UnloadAll(scope Scope) ([]LinkResult, error) {
	targetDir, err := l.TargetDir(scope)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var results []LinkResult
	for _, e := range entries {
		linkPath := filepath.Join(targetDir, e.Name())
		if !isSymlink(linkPath) {
			continue
		}
		target, err := os.Readlink(linkPath)
		if err != nil {
			continue
		}
		if l.isManagedTarget(target) {
			if err := os.Remove(linkPath); err != nil {
				results = append(results, LinkResult{SkillName: e.Name(), Action: "error", Detail: err.Error()})
			} else {
				results = append(results, LinkResult{SkillName: e.Name(), Action: "removed"})
			}
		}
	}
	return results, nil
}

// Status returns all active ss-managed symlinks in both scopes.
func (l *Linker) Status() (userLinks, projectLinks []LinkInfo, err error) {
	userLinks, _ = l.scanScope(ScopeUser)
	projectLinks, _ = l.scanScope(ScopeProject)
	return userLinks, projectLinks, nil
}

// FullStatus returns all skills visible to CC, including non-ss-managed ones and plugins.
func (l *Linker) FullStatus() (*FullStatus, error) {
	status := &FullStatus{}
	status.UserScope = l.scanAllEntries(ScopeUser)
	status.ProjectScope = l.scanAllEntries(ScopeProject)
	status.Plugins = l.scanPlugins()
	return status, nil
}

// TargetDir resolves the symlink target directory for a given scope.
func (l *Linker) TargetDir(scope Scope) (string, error) {
	switch scope {
	case ScopeUser:
		return config.UserSkillsDir(), nil
	case ScopeProject:
		return config.ProjectSkillsDir(), nil
	default:
		return "", fmt.Errorf("unknown scope: %s", scope)
	}
}

func (l *Linker) linkOne(name, targetDir string) LinkResult {
	sourceDir := filepath.Join(l.skillsSourceDir, name)
	linkPath := filepath.Join(targetDir, name)

	// Check source exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return LinkResult{SkillName: name, Action: "error", Detail: "source not found"}
	}

	// Check if link already exists
	info, err := os.Lstat(linkPath)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			target, _ := os.Readlink(linkPath)
			absTarget, _ := filepath.Abs(target)
			absSource, _ := filepath.Abs(sourceDir)
			if absTarget == absSource {
				return LinkResult{SkillName: name, Action: "skipped", Detail: "already linked"}
			}
			return LinkResult{SkillName: name, Action: "error", Detail: fmt.Sprintf("link exists pointing to %s", target)}
		}
		return LinkResult{SkillName: name, Action: "error", Detail: "path exists and is not a symlink"}
	}

	if err := os.Symlink(sourceDir, linkPath); err != nil {
		return LinkResult{SkillName: name, Action: "error", Detail: err.Error()}
	}
	return LinkResult{SkillName: name, Action: "linked"}
}

func (l *Linker) unlinkOne(name, targetDir string) LinkResult {
	linkPath := filepath.Join(targetDir, name)

	if !isSymlink(linkPath) {
		return LinkResult{SkillName: name, Action: "skipped", Detail: "not a symlink"}
	}

	target, err := os.Readlink(linkPath)
	if err != nil {
		return LinkResult{SkillName: name, Action: "error", Detail: err.Error()}
	}

	if !l.isManagedTarget(target) {
		return LinkResult{SkillName: name, Action: "skipped", Detail: "not managed by ss"}
	}

	if err := os.Remove(linkPath); err != nil {
		return LinkResult{SkillName: name, Action: "error", Detail: err.Error()}
	}
	return LinkResult{SkillName: name, Action: "removed"}
}

func (l *Linker) scanScope(scope Scope) ([]LinkInfo, error) {
	targetDir, err := l.TargetDir(scope)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return nil, nil
	}

	var links []LinkInfo
	for _, e := range entries {
		linkPath := filepath.Join(targetDir, e.Name())
		if !isSymlink(linkPath) {
			continue
		}
		target, err := os.Readlink(linkPath)
		if err != nil {
			continue
		}
		if l.isManagedTarget(target) {
			links = append(links, LinkInfo{
				SkillName:  e.Name(),
				SourcePath: target,
				LinkPath:   linkPath,
			})
		}
	}
	return links, nil
}

func (l *Linker) isManagedTarget(target string) bool {
	absTarget, _ := filepath.Abs(target)
	absSource, _ := filepath.Abs(l.skillsSourceDir)
	return strings.HasPrefix(absTarget, absSource)
}

func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// scanAllEntries returns all skill entries in a scope directory, regardless of type.
func (l *Linker) scanAllEntries(scope Scope) []SkillEntry {
	targetDir, err := l.TargetDir(scope)
	if err != nil {
		return nil
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return nil
	}

	var result []SkillEntry
	for _, e := range entries {
		entryPath := filepath.Join(targetDir, e.Name())
		entry := SkillEntry{Name: e.Name()}

		if isSymlink(entryPath) {
			target, err := os.Readlink(entryPath)
			if err != nil {
				continue
			}
			entry.Path = target
			if l.isManagedTarget(target) {
				entry.Source = "ss"
			} else {
				entry.Source = "manual"
			}
		} else {
			entry.Path = entryPath
			entry.Source = "manual"
		}
		result = append(result, entry)
	}
	return result
}

// scanPlugins scans ~/.claude/plugins/cache/ for installed plugin skills.
func (l *Linker) scanPlugins() []SkillEntry {
	home, _ := os.UserHomeDir()
	pluginsDir := filepath.Join(home, ".claude", "plugins", "cache")

	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil
	}

	var result []SkillEntry
	for _, provider := range entries {
		if !provider.IsDir() {
			continue
		}
		providerPath := filepath.Join(pluginsDir, provider.Name())
		// Walk provider -> plugin -> version -> skills/
		l.walkPluginProvider(providerPath, &result)
	}
	return result
}

func (l *Linker) walkPluginProvider(providerPath string, result *[]SkillEntry) {
	plugins, err := os.ReadDir(providerPath)
	if err != nil {
		return
	}
	for _, plugin := range plugins {
		if !plugin.IsDir() {
			continue
		}
		pluginPath := filepath.Join(providerPath, plugin.Name())
		versions, err := os.ReadDir(pluginPath)
		if err != nil {
			continue
		}
		for _, ver := range versions {
			if !ver.IsDir() {
				continue
			}
			skillsPath := filepath.Join(pluginPath, ver.Name(), "skills")
			skills, err := os.ReadDir(skillsPath)
			if err != nil {
				continue
			}
			for _, sk := range skills {
				if !sk.IsDir() {
					continue
				}
				skillDir := filepath.Join(skillsPath, sk.Name())
				// Verify SKILL.md exists
				if _, err := os.Stat(filepath.Join(skillDir, "SKILL.md")); err != nil {
					continue
				}
				*result = append(*result, SkillEntry{
					Name:   fmt.Sprintf("%s/%s/%s", filepath.Base(providerPath), plugin.Name(), sk.Name()),
					Path:   skillDir,
					Source: "plugin",
				})
			}
		}
	}
}

