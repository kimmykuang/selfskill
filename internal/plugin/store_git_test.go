package plugin

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// setupBareRepo creates a bare git repo backed by a working tree containing
// the supplied files. It returns a file:// URL pointing at the bare repo, the
// HEAD commit sha, and the path to the underlying work tree (so callers can
// add further commits and push them back to the bare).
//
// The whole layout lives under t.TempDir() so cleanup is automatic.
func setupBareRepo(t *testing.T, files map[string]string) (bareURL, sha, workDir string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	base := t.TempDir()
	work := filepath.Join(base, "work")
	bare := filepath.Join(base, "bare")

	run := func(name string, args ...string) {
		cmd := exec.Command(name, args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s %v: %v\n%s", name, args, err, out)
		}
	}

	if err := os.MkdirAll(work, 0755); err != nil {
		t.Fatalf("mkdir work: %v", err)
	}

	run("git", "-C", work, "init", "-q", "-b", "main")
	run("git", "-C", work, "config", "user.email", "ci@example.com")
	run("git", "-C", work, "config", "user.name", "ci")
	run("git", "-C", work, "config", "commit.gpgsign", "false")

	for path, content := range files {
		full := filepath.Join(work, path)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", full, err)
		}
	}

	run("git", "-C", work, "add", ".")
	run("git", "-C", work, "commit", "-q", "-m", "init")
	run("git", "clone", "--bare", "-q", work, bare)

	out, err := exec.Command("git", "-C", work, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatalf("rev-parse HEAD: %v", err)
	}
	sha = strings.TrimSpace(string(out))
	bareURL = "file://" + bare
	workDir = work
	return
}

// pushNewCommit adds a new file in workDir, commits it, and pushes the new
// HEAD to the bare repo. Returns the new HEAD sha.
func pushNewCommit(t *testing.T, workDir, bareDir, relPath, content string) string {
	t.Helper()
	run := func(name string, args ...string) {
		cmd := exec.Command(name, args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s %v: %v\n%s", name, args, err, out)
		}
	}

	full := filepath.Join(workDir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", full, err)
	}
	run("git", "-C", workDir, "add", ".")
	run("git", "-C", workDir, "commit", "-q", "-m", "update")
	run("git", "-C", workDir, "push", "-q", bareDir, "main:main")

	out, err := exec.Command("git", "-C", workDir, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatalf("rev-parse HEAD: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func bareDirFromURL(url string) string {
	return strings.TrimPrefix(url, "file://")
}

func TestStore_InstallFromGit_WritesMetaAndDropsGit(t *testing.T) {
	files := map[string]string{
		"README.md":             "# myplugin\n",
		"plugin.json":           `{"name":"myplugin"}`,
		"skills/skill1/SKILL.md": "# skill1\n",
	}
	bareURL, headSha, _ := setupBareRepo(t, files)

	store := NewStore(t.TempDir())

	before := time.Now().Add(-time.Second).UTC()
	p, err := store.InstallFromGit(bareURL, "mp", "myplugin", "latest")
	if err != nil {
		t.Fatalf("InstallFromGit: %v", err)
	}

	if p.Name != "myplugin" || p.Marketplace != "mp" || p.Version != "latest" {
		t.Errorf("plugin fields mismatch: %+v", p)
	}
	if p.GitCommitSha != headSha {
		t.Errorf("GitCommitSha=%q, want %q", p.GitCommitSha, headSha)
	}
	if p.InstallPath == "" {
		t.Fatalf("empty InstallPath")
	}

	// .ss-meta.json exists and matches.
	metaPath := filepath.Join(p.InstallPath, metaFileName)
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("read meta: %v", err)
	}
	var meta PluginMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("unmarshal meta: %v", err)
	}
	if meta.URL != bareURL {
		t.Errorf("meta.URL=%q, want %q", meta.URL, bareURL)
	}
	if meta.Marketplace != "mp" {
		t.Errorf("meta.Marketplace=%q, want mp", meta.Marketplace)
	}
	if meta.CommitSha != headSha {
		t.Errorf("meta.CommitSha=%q, want %q", meta.CommitSha, headSha)
	}
	if meta.Version != "latest" {
		t.Errorf("meta.Version=%q, want latest", meta.Version)
	}
	installedAt, err := time.Parse(time.RFC3339, meta.InstalledAt)
	if err != nil {
		t.Errorf("meta.InstalledAt %q is not RFC3339: %v", meta.InstalledAt, err)
	} else if installedAt.Before(before) {
		t.Errorf("meta.InstalledAt %v is before %v", installedAt, before)
	}

	// .git was dropped.
	if _, err := os.Stat(filepath.Join(p.InstallPath, ".git")); !os.IsNotExist(err) {
		t.Errorf(".git still present at install path: err=%v", err)
	}

	// Original files were materialised.
	for _, rel := range []string{"README.md", "plugin.json", "skills/skill1/SKILL.md"} {
		if _, err := os.Stat(filepath.Join(p.InstallPath, rel)); err != nil {
			t.Errorf("expected %s in install path: %v", rel, err)
		}
	}
}

func TestStore_GetAndList_BackfillRemoteURL(t *testing.T) {
	bareURL, headSha, _ := setupBareRepo(t, map[string]string{
		"README.md": "# r\n",
	})

	store := NewStore(t.TempDir())
	if _, err := store.InstallFromGit(bareURL, "mp", "myplugin", "latest"); err != nil {
		t.Fatalf("InstallFromGit: %v", err)
	}

	got, err := store.Get("myplugin@mp")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.RemoteURL != bareURL {
		t.Errorf("Get RemoteURL=%q, want %q", got.RemoteURL, bareURL)
	}
	if got.GitCommitSha != headSha {
		t.Errorf("Get GitCommitSha=%q, want %q", got.GitCommitSha, headSha)
	}

	plugins, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("List len=%d, want 1", len(plugins))
	}
	if plugins[0].RemoteURL != bareURL {
		t.Errorf("List[0].RemoteURL=%q, want %q", plugins[0].RemoteURL, bareURL)
	}
	if plugins[0].GitCommitSha != headSha {
		t.Errorf("List[0].GitCommitSha=%q, want %q", plugins[0].GitCommitSha, headSha)
	}
}

func TestStore_GetRemoteURL_FromMockInstall(t *testing.T) {
	bareURL, _, _ := setupBareRepo(t, map[string]string{
		"README.md": "# r\n",
	})

	store := NewStore(t.TempDir())
	if _, err := store.InstallFromGit(bareURL, "mp", "myplugin", "latest"); err != nil {
		t.Fatalf("InstallFromGit: %v", err)
	}

	got, err := store.GetRemoteURL("myplugin@mp")
	if err != nil {
		t.Fatalf("GetRemoteURL: %v", err)
	}
	if got != bareURL {
		t.Errorf("GetRemoteURL=%q, want %q", got, bareURL)
	}
}

func TestStore_Update_FromMockRepo(t *testing.T) {
	bareURL, headSha, workDir := setupBareRepo(t, map[string]string{
		"README.md": "# r\n",
	})

	store := NewStore(t.TempDir())
	p, err := store.InstallFromGit(bareURL, "mp", "myplugin", "latest")
	if err != nil {
		t.Fatalf("InstallFromGit: %v", err)
	}
	originalInstalledAt, err := time.Parse(time.RFC3339, mustReadMeta(t, p.InstallPath).InstalledAt)
	if err != nil {
		t.Fatalf("parse installedAt: %v", err)
	}

	// Sleep so the post-update RFC3339 timestamp differs at second resolution.
	time.Sleep(1100 * time.Millisecond)

	newSha := pushNewCommit(t, workDir, bareDirFromURL(bareURL), "NEWFILE.md", "new\n")
	if newSha == headSha {
		t.Fatalf("expected new sha to differ from initial %q", headSha)
	}

	if err := store.Update("myplugin@mp"); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if _, err := os.Stat(filepath.Join(p.InstallPath, "NEWFILE.md")); err != nil {
		t.Errorf("expected NEWFILE.md after update: %v", err)
	}
	if _, err := os.Stat(filepath.Join(p.InstallPath, ".git")); !os.IsNotExist(err) {
		t.Errorf(".git still present after update: err=%v", err)
	}

	meta := mustReadMeta(t, p.InstallPath)
	if meta.CommitSha != newSha {
		t.Errorf("meta.CommitSha=%q, want %q", meta.CommitSha, newSha)
	}
	if meta.URL != bareURL {
		t.Errorf("meta.URL=%q, want %q (should be preserved)", meta.URL, bareURL)
	}
	updatedAt, err := time.Parse(time.RFC3339, meta.InstalledAt)
	if err != nil {
		t.Fatalf("parse updated installedAt: %v", err)
	}
	if updatedAt.Before(originalInstalledAt) {
		t.Errorf("updated installedAt %v is before original %v", updatedAt, originalInstalledAt)
	}
}

func TestStore_Update_NoRemoteURL(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	pluginDir := filepath.Join(dir, "mp", "noremote", "v1")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	meta := &PluginMeta{
		URL:         "",
		Marketplace: "mp",
		Version:     "v1",
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, metaFileName), data, 0644); err != nil {
		t.Fatalf("write meta: %v", err)
	}

	err = store.Update("noremote@mp")
	if err == nil {
		t.Fatal("expected error when meta has empty url")
	}
	if !strings.Contains(err.Error(), "no remote url") {
		t.Errorf("expected error to mention 'no remote url', got: %v", err)
	}
}

func mustReadMeta(t *testing.T, installPath string) *PluginMeta {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(installPath, metaFileName))
	if err != nil {
		t.Fatalf("read meta: %v", err)
	}
	var m PluginMeta
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal meta: %v", err)
	}
	return &m
}
