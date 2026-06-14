package compare

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimmykuang/selfskill/internal/frontmatter"
	"github.com/kimmykuang/selfskill/internal/plugin"
	"github.com/kimmykuang/selfskill/internal/skill"
)

// setupBareRepo creates a bare git repo backed by a working tree containing
// the supplied files. It returns a file:// URL pointing at the bare repo and
// the HEAD commit sha. The whole layout lives under t.TempDir() so cleanup is
// automatic.
//
// Adapted from internal/plugin/store_git_test.go to keep this package's tests
// self-contained without exporting test helpers across packages.
func setupBareRepo(t *testing.T, files map[string]string) (bareURL, sha string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	base := t.TempDir()
	work := filepath.Join(base, "work")
	bare := filepath.Join(base, "bare")

	if err := os.MkdirAll(work, 0755); err != nil {
		t.Fatalf("mkdir work: %v", err)
	}

	gitCfg := []string{
		"-c", "user.email=ci@example.com",
		"-c", "user.name=ci",
		"-c", "commit.gpgsign=false",
	}
	run := func(args ...string) {
		full := append(append([]string{}, gitCfg...), args...)
		cmd := exec.Command("git", full...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	run("-C", work, "init", "-q", "-b", "main")

	for path, content := range files {
		full := filepath.Join(work, path)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", full, err)
		}
	}

	run("-C", work, "add", ".")
	run("-C", work, "commit", "-q", "-m", "init")
	run("clone", "--bare", "-q", work, bare)

	out, err := exec.Command("git", "-C", work, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatalf("rev-parse HEAD: %v", err)
	}
	sha = strings.TrimSpace(string(out))
	bareURL = "file://" + bare
	return
}

func newComparerWithStores(t *testing.T) (*Comparer, *plugin.Store, *skill.Store) {
	t.Helper()
	ps := plugin.NewStore(t.TempDir())
	ss := skill.NewStore(t.TempDir())
	return New(ps, ss), ps, ss
}

// TestCompareSkill_FromMockURL_RawFrontmatterOrder reproduces the regression
// where a freshly-installed remote skill produced a spurious diff because the
// installer rewrote the local frontmatter through yaml.Marshal (alphabetical
// keys) while the remote stayed in its original "name first, description
// second" shape. The fix normalizes both sides through frontmatter.Parse +
// frontmatter.Marshal before diffing, so the diff is empty regardless of
// the remote file's original key order.
func TestCompareSkill_FromMockURL_RawFrontmatterOrder(t *testing.T) {
	// Remote is a hand-written file with frontmatter in "name, description"
	// order — exactly what a human would write.
	remoteRaw := []byte("---\nname: ordered\ndescription: order test\n---\n\nbody\n")

	mux := http.NewServeMux()
	mux.HandleFunc("/skills/ordered.md", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(remoteRaw)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	sourceURL := server.URL + "/skills/ordered.md"

	// Local copy: same content, but with `source` injected and re-marshaled
	// through yaml.Marshal — keys end up alphabetical (description, name, source).
	localMeta := map[string]any{
		"name":        "ordered",
		"description": "order test",
		"source":      sourceURL,
	}
	localBytes, err := frontmatter.Marshal(localMeta, "\nbody\n")
	if err != nil {
		t.Fatalf("marshal local: %v", err)
	}

	c, _, ss := newComparerWithStores(t)
	if err := ss.SaveFromContent("ordered", localBytes); err != nil {
		t.Fatalf("SaveFromContent: %v", err)
	}

	var buf bytes.Buffer
	if err := c.CompareSkill("ordered", &buf); err != nil {
		t.Fatalf("CompareSkill: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty diff after normalization, got:\n%s", buf.String())
	}
}

func TestCompareSkill_FromMockURL_Identical(t *testing.T) {
	base := map[string]any{
		"description": "a skill",
		"name":        "x",
		"version":     "1.0.0",
	}
	body := "# body\n"
	mockBytes, err := frontmatter.Marshal(base, body)
	if err != nil {
		t.Fatalf("marshal mock body: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/skills/x.md", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write(mockBytes)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	sourceURL := server.URL + "/skills/x.md"
	localMeta := map[string]any{
		"description": "a skill",
		"name":        "x",
		"source":      sourceURL,
		"version":     "1.0.0",
	}
	localBytes, err := frontmatter.Marshal(localMeta, body)
	if err != nil {
		t.Fatalf("marshal local body: %v", err)
	}

	c, _, ss := newComparerWithStores(t)
	if err := ss.SaveFromContent("x", localBytes); err != nil {
		t.Fatalf("SaveFromContent: %v", err)
	}

	var buf bytes.Buffer
	if err := c.CompareSkill("x", &buf); err != nil {
		t.Fatalf("CompareSkill: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty diff, got:\n%s", buf.String())
	}
}

func TestCompareSkill_FromMockURL_HasDiff(t *testing.T) {
	mockBytes := []byte("---\nname: x\nversion: 1.0.0\n---\n# remote body\n")

	mux := http.NewServeMux()
	mux.HandleFunc("/skills/x.md", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(mockBytes)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	sourceURL := server.URL + "/skills/x.md"
	localBytes := []byte("---\nname: x\nsource: " + sourceURL + "\nversion: 1.0.0\n---\n# local body different\n")

	c, _, ss := newComparerWithStores(t)
	if err := ss.SaveFromContent("x", localBytes); err != nil {
		t.Fatalf("SaveFromContent: %v", err)
	}

	var buf bytes.Buffer
	if err := c.CompareSkill("x", &buf); err != nil {
		t.Fatalf("CompareSkill returned err: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "---") {
		t.Errorf("expected unified diff '---' marker, got:\n%s", out)
	}
	if !strings.Contains(out, "+++") {
		t.Errorf("expected unified diff '+++' marker, got:\n%s", out)
	}
	if !strings.Contains(out, "local body different") && !strings.Contains(out, "remote body") {
		t.Errorf("expected diff body content in output, got:\n%s", out)
	}
}

func TestCompareSkill_HTTPError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/skills/broken.md", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	sourceURL := server.URL + "/skills/broken.md"
	localBytes := []byte("---\nname: broken\nsource: " + sourceURL + "\n---\n# body\n")

	c, _, ss := newComparerWithStores(t)
	if err := ss.SaveFromContent("broken", localBytes); err != nil {
		t.Fatalf("SaveFromContent: %v", err)
	}

	var buf bytes.Buffer
	err := c.CompareSkill("broken", &buf)
	if err == nil {
		t.Fatal("expected error from HTTP 500, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected error to mention 500, got: %v", err)
	}
}

func TestCompareSkill_GitURL_Unsupported(t *testing.T) {
	cases := []struct {
		name   string
		source string
	}{
		{"https-git", "https://github.com/x/y.git"},
		{"ssh-git", "git@github.com:x/y.git"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			localBytes := []byte("---\nname: gitskill\nsource: " + tc.source + "\n---\n# body\n")

			c, _, ss := newComparerWithStores(t)
			if err := ss.SaveFromContent("gitskill", localBytes); err != nil {
				t.Fatalf("SaveFromContent: %v", err)
			}

			var buf bytes.Buffer
			err := c.CompareSkill("gitskill", &buf)
			if err == nil {
				t.Fatal("expected error for git URL source, got nil")
			}
			if !strings.Contains(err.Error(), "not yet supported") {
				t.Errorf("expected 'not yet supported' in error, got: %v", err)
			}
		})
	}
}

func TestComparePlugin_FromMockRepo_Identical(t *testing.T) {
	files := map[string]string{
		"README.md":              "# myplugin\n",
		"plugin.json":            `{"name":"myplugin"}`,
		"skills/skill1/SKILL.md": "# skill1\n",
	}
	bareURL, _ := setupBareRepo(t, files)

	c, ps, _ := newComparerWithStores(t)
	if _, err := ps.InstallFromGit(bareURL, "mp", "myplugin", "latest"); err != nil {
		t.Fatalf("InstallFromGit: %v", err)
	}

	var buf bytes.Buffer
	if err := c.ComparePlugin("myplugin@mp", &buf); err != nil {
		t.Fatalf("ComparePlugin: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no diff, got:\n%s", buf.String())
	}
}

func TestComparePlugin_FromMockRepo_HasDiff(t *testing.T) {
	files := map[string]string{
		"README.md":   "# myplugin\n",
		"plugin.json": `{"name":"myplugin"}`,
	}
	bareURL, _ := setupBareRepo(t, files)

	c, ps, _ := newComparerWithStores(t)
	p, err := ps.InstallFromGit(bareURL, "mp", "myplugin", "latest")
	if err != nil {
		t.Fatalf("InstallFromGit: %v", err)
	}

	readmePath := filepath.Join(p.InstallPath, "README.md")
	f, err := os.OpenFile(readmePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("open README: %v", err)
	}
	if _, err := f.WriteString("// local edit\n"); err != nil {
		f.Close()
		t.Fatalf("append README: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close README: %v", err)
	}

	var buf bytes.Buffer
	if err := c.ComparePlugin("myplugin@mp", &buf); err != nil {
		t.Fatalf("ComparePlugin: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "local edit") {
		t.Errorf("expected diff to mention 'local edit', got:\n%s", out)
	}
	if !strings.Contains(out, "+++") || !strings.Contains(out, "---") {
		t.Errorf("expected unified diff markers, got:\n%s", out)
	}
}

func TestComparePlugin_NoMeta(t *testing.T) {
	pluginsDir := t.TempDir()
	installPath := filepath.Join(pluginsDir, "mp", "noremote", "v1")
	if err := os.MkdirAll(installPath, 0755); err != nil {
		t.Fatalf("mkdir installPath: %v", err)
	}
	if err := os.WriteFile(filepath.Join(installPath, "README.md"), []byte("# r\n"), 0644); err != nil {
		t.Fatalf("write README: %v", err)
	}

	ps := plugin.NewStore(pluginsDir)
	ss := skill.NewStore(t.TempDir())
	c := New(ps, ss)

	var buf bytes.Buffer
	err := c.ComparePlugin("noremote@mp", &buf)
	if err == nil {
		t.Fatal("expected error when plugin has no recorded remote url")
	}
	if !strings.Contains(err.Error(), "no remote url") {
		t.Errorf("expected 'no remote url' in error, got: %v", err)
	}
}

func TestCompare_AutoDetect_Plugin(t *testing.T) {
	files := map[string]string{
		"README.md": "# myplugin\n",
	}
	bareURL, _ := setupBareRepo(t, files)

	c, ps, _ := newComparerWithStores(t)
	if _, err := ps.InstallFromGit(bareURL, "mp", "myplugin", "latest"); err != nil {
		t.Fatalf("InstallFromGit: %v", err)
	}

	var buf bytes.Buffer
	if err := c.Compare("myplugin@mp", &buf); err != nil {
		t.Fatalf("Compare auto-detect plugin: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no diff via auto-detect plugin, got:\n%s", buf.String())
	}
}

func TestCompare_AutoDetect_Skill(t *testing.T) {
	base := map[string]any{
		"description": "a skill",
		"name":        "auto",
		"version":     "1.0.0",
	}
	body := "# body\n"
	mockBytes, err := frontmatter.Marshal(base, body)
	if err != nil {
		t.Fatalf("marshal mock: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/skills/auto.md", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(mockBytes)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	sourceURL := server.URL + "/skills/auto.md"
	localMeta := map[string]any{
		"description": "a skill",
		"name":        "auto",
		"source":      sourceURL,
		"version":     "1.0.0",
	}
	localBytes, err := frontmatter.Marshal(localMeta, body)
	if err != nil {
		t.Fatalf("marshal local: %v", err)
	}

	c, _, ss := newComparerWithStores(t)
	if err := ss.SaveFromContent("auto", localBytes); err != nil {
		t.Fatalf("SaveFromContent: %v", err)
	}

	var buf bytes.Buffer
	if err := c.Compare("auto", &buf); err != nil {
		t.Fatalf("Compare auto-detect skill: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no diff via auto-detect skill, got:\n%s", buf.String())
	}
}
