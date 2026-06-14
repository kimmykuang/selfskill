package compare

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kimmykuang/selfskill/internal/frontmatter"
	"github.com/kimmykuang/selfskill/internal/plugin"
	"github.com/kimmykuang/selfskill/internal/skill"
)

// Comparer compares locally installed plugins/skills against their remote source.
type Comparer struct {
	pluginStore *plugin.Store
	skillStore  *skill.Store
}

// New constructs a Comparer.
func New(ps *plugin.Store, ss *skill.Store) *Comparer {
	return &Comparer{pluginStore: ps, skillStore: ss}
}

// ComparePlugin clones the plugin's recorded remote into a temp dir and diffs
// it against the local install path.
func (c *Comparer) ComparePlugin(name string, w io.Writer) error {
	p, err := c.pluginStore.Get(name)
	if err != nil {
		return err
	}
	url, err := c.pluginStore.GetRemoteURL(name)
	if err != nil {
		return err
	}

	tmp, err := os.MkdirTemp("", "ss-compare-plugin-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	cloneDir := filepath.Join(tmp, "remote")
	cmd := exec.Command("git", "clone", "--depth", "1", url, cloneDir)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	// Drop the remote .git so it does not pollute the diff.
	if err := os.RemoveAll(filepath.Join(cloneDir, ".git")); err != nil {
		return fmt.Errorf("removing remote .git: %w", err)
	}

	return runDiff(p.InstallPath, cloneDir, w)
}

// CompareSkill downloads the skill's recorded source and diffs it against the
// local SKILL.md. Currently supports single-file http(s) .md sources.
func (c *Comparer) CompareSkill(name string, w io.Writer) error {
	sk, err := c.skillStore.Get(name)
	if err != nil {
		return err
	}
	if sk.Source == "" {
		return fmt.Errorf("skill %q has no source recorded; cannot compare", name)
	}

	if isGitURL(sk.Source) {
		return fmt.Errorf("skill compare from git URL not yet supported")
	}

	body, err := httpGet(sk.Source)
	if err != nil {
		return err
	}

	tmp, err := os.MkdirTemp("", "ss-compare-skill-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	remotePath := filepath.Join(tmp, "SKILL.md")
	if err := os.WriteFile(remotePath, body, 0644); err != nil {
		return fmt.Errorf("writing remote skill: %w", err)
	}

	// Strip the local source field before diffing so it does not pollute output.
	localPath := filepath.Join(sk.DirPath, "SKILL.md")
	localData, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("reading local skill: %w", err)
	}
	meta, bodyText, err := frontmatter.Parse(localData)
	if err != nil {
		return fmt.Errorf("parsing local skill frontmatter: %w", err)
	}
	if meta != nil {
		delete(meta, "source")
	}
	stripped, err := frontmatter.Marshal(meta, bodyText)
	if err != nil {
		return fmt.Errorf("marshaling local skill: %w", err)
	}
	localStrippedPath := filepath.Join(tmp, "local-SKILL.md")
	if err := os.WriteFile(localStrippedPath, stripped, 0644); err != nil {
		return fmt.Errorf("writing local skill copy: %w", err)
	}

	return runDiff(localStrippedPath, remotePath, w)
}

// Compare resolves name as a plugin first, then as a skill.
func (c *Comparer) Compare(name string, w io.Writer) error {
	if c.pluginStore != nil && c.pluginStore.Exists(name) {
		return c.ComparePlugin(name, w)
	}
	if c.skillStore != nil && c.skillStore.Exists(name) {
		return c.CompareSkill(name, w)
	}
	return fmt.Errorf("no plugin or skill named %q found", name)
}

// runDiff invokes `diff -ru` between two paths, writing output to w.
// diff exit code 0 = identical, 1 = differences (treated as success), >1 = error.
func runDiff(local, remote string, w io.Writer) error {
	cmd := exec.Command("diff", "-ru", "--exclude=.git", "--exclude=.ss-meta.json", local, remote)
	cmd.Stdout = w
	cmd.Stderr = w
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			if ee.ExitCode() == 1 {
				return nil
			}
		}
		return fmt.Errorf("diff failed: %w", err)
	}
	return nil
}

func isGitURL(s string) bool {
	return strings.HasSuffix(s, ".git") || strings.HasPrefix(s, "git@")
}

func httpGet(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http GET %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http GET %s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
