package installer

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimmykuang/selfskill/internal/frontmatter"
	"github.com/kimmykuang/selfskill/internal/prompt"
	"github.com/kimmykuang/selfskill/internal/skill"
)

func newInstallerWithTempDirs(t *testing.T) (*Installer, *skill.Store, *prompt.Store) {
	t.Helper()
	ss := skill.NewStore(t.TempDir())
	ps := prompt.NewStore(t.TempDir())
	return New(ss, ps), ss, ps
}

func readSkillMD(t *testing.T, ss *skill.Store, name string) (map[string]any, string) {
	t.Helper()
	path := filepath.Join(ss.SkillDir(name), "SKILL.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read SKILL.md: %v", err)
	}
	meta, body, err := frontmatter.Parse(data)
	if err != nil {
		t.Fatalf("parse frontmatter: %v", err)
	}
	return meta, body
}

func TestInstallSkill_FromMockURL_WritesSourceFrontmatter(t *testing.T) {
	mockBody := "---\nname: foo\nversion: 1.0.0\ndescription: a foo skill\n---\n# Foo\n\nThis is the foo skill.\n"
	mux := http.NewServeMux()
	mux.HandleFunc("/skills/foo.md", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(mockBody))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	inst, ss, _ := newInstallerWithTempDirs(t)
	url := server.URL + "/skills/foo.md"

	results, err := inst.InstallSkill(url, false)
	if err != nil {
		t.Fatalf("InstallSkill: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Action != "installed" {
		t.Fatalf("expected installed, got %s (%s)", results[0].Action, results[0].Detail)
	}
	if results[0].Name != "foo" {
		t.Errorf("expected name=foo, got %q", results[0].Name)
	}

	if !ss.Exists("foo") {
		t.Fatalf("skill foo not found on disk")
	}

	meta, body, err := frontmatter.Parse(mustReadFile(t, filepath.Join(ss.SkillDir("foo"), "SKILL.md")))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got, _ := meta["name"].(string); got != "foo" {
		t.Errorf("name=%q, want foo", got)
	}
	if got, _ := meta["version"].(string); got != "1.0.0" {
		t.Errorf("version=%q, want 1.0.0", got)
	}
	if got, _ := meta["source"].(string); got != url {
		t.Errorf("source=%q, want %q", got, url)
	}
	if !strings.Contains(body, "This is the foo skill.") {
		t.Errorf("body missing original content; got: %q", body)
	}
}

func TestInstallSkill_FromMockURL_GitHubBlobRewrite(t *testing.T) {
	// We can't fake the github.com hostname with httptest, so this case
	// verifies the simpler invariant: the URL the user supplied (whatever
	// form) is what gets persisted as `source`, even when origURL == url.
	mockBody := "---\nname: bar\nversion: 0.1.0\n---\n# Bar\n"
	mux := http.NewServeMux()
	mux.HandleFunc("/skills/bar.md", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockBody))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	inst, ss, _ := newInstallerWithTempDirs(t)
	userURL := server.URL + "/skills/bar.md"

	if _, err := inst.InstallSkill(userURL, false); err != nil {
		t.Fatalf("InstallSkill: %v", err)
	}
	meta, _ := readSkillMD(t, ss, "bar")
	if got, _ := meta["source"].(string); got != userURL {
		t.Errorf("source=%q, want %q (original URL)", got, userURL)
	}
}

func TestInstallSkill_FromMockURL_AlreadyExists(t *testing.T) {
	v1 := "---\nname: dup\nversion: \"1\"\n---\n# v1\n"
	v2 := "---\nname: dup\nversion: \"2\"\n---\n# v2 updated\n"
	current := v1
	mux := http.NewServeMux()
	mux.HandleFunc("/skills/dup.md", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(current))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	inst, ss, _ := newInstallerWithTempDirs(t)
	url := server.URL + "/skills/dup.md"

	if r, err := inst.InstallSkill(url, false); err != nil || r[0].Action != "installed" {
		t.Fatalf("first install: %v %+v", err, r)
	}

	current = v2
	r, err := inst.InstallSkill(url, false)
	if err != nil {
		t.Fatalf("second install (no force): %v", err)
	}
	if r[0].Action != "error" || !strings.Contains(r[0].Detail, "already exists") {
		t.Errorf("expected already exists error, got %+v", r[0])
	}

	r, err = inst.InstallSkill(url, true)
	if err != nil {
		t.Fatalf("third install (force): %v", err)
	}
	if r[0].Action != "installed" {
		t.Fatalf("force install action=%s detail=%s", r[0].Action, r[0].Detail)
	}
	meta, body := readSkillMD(t, ss, "dup")
	if got, _ := meta["version"].(string); got != "2" {
		t.Errorf("version=%q, want 2 after force overwrite", got)
	}
	if !strings.Contains(body, "v2 updated") {
		t.Errorf("body=%q, want v2 content", body)
	}
}

func TestInstallSkill_FromMockURL_HTTPError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/skills/broken.md", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	inst, _, _ := newInstallerWithTempDirs(t)
	_, err := inst.InstallSkill(server.URL+"/skills/broken.md", false)
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should mention status 500, got %v", err)
	}
}

func TestInstallSkill_FromMockURL_FrontmatterWithoutName(t *testing.T) {
	// No `name` in frontmatter — installer should fall back to URL filename.
	mockBody := "---\nversion: 0.1.0\ndescription: nameless\n---\nbody\n"
	mux := http.NewServeMux()
	mux.HandleFunc("/skills/noname.md", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockBody))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	inst, ss, _ := newInstallerWithTempDirs(t)
	url := server.URL + "/skills/noname.md"

	r, err := inst.InstallSkill(url, false)
	if err != nil {
		t.Fatalf("InstallSkill: %v", err)
	}
	if r[0].Name != "noname" {
		t.Errorf("name=%q, want noname (from URL)", r[0].Name)
	}
	if !ss.Exists("noname") {
		t.Fatalf("skill noname not on disk")
	}
	meta, _ := readSkillMD(t, ss, "noname")
	if got, _ := meta["source"].(string); got != url {
		t.Errorf("source=%q, want %q", got, url)
	}
}

func TestInstallPrompt_FromMockURL(t *testing.T) {
	mockBody := "---\nid: p1\ntags: [t1, t2]\ndescription: prompt one\n---\nplease do the thing.\n"
	mux := http.NewServeMux()
	mux.HandleFunc("/prompts/p1.md", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockBody))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	inst, _, ps := newInstallerWithTempDirs(t)
	url := server.URL + "/prompts/p1.md"

	r, err := inst.InstallPrompt(url, false)
	if err != nil {
		t.Fatalf("InstallPrompt: %v", err)
	}
	if len(r) != 1 || r[0].Action != "installed" {
		t.Fatalf("unexpected result: %+v", r)
	}
	if !ps.Exists("p1") {
		t.Fatalf("prompt p1 not on disk")
	}
	p, err := ps.Get("p1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if p.Description != "prompt one" {
		t.Errorf("description=%q", p.Description)
	}
	if len(p.Tags) != 2 || p.Tags[0] != "t1" || p.Tags[1] != "t2" {
		t.Errorf("tags=%v, want [t1 t2]", p.Tags)
	}
	if !strings.Contains(p.Body, "please do the thing.") {
		t.Errorf("body=%q", p.Body)
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}
