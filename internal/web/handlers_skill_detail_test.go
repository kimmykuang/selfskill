package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateSkill_RewritesFrontmatterAndBody(t *testing.T) {
	srv, root := testServer(t)
	dir := filepath.Join(root, "skills", "alpha")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: alpha\ndescription: old\n---\n\nold body\n"), 0644); err != nil {
		t.Fatal(err)
	}

	body := bytes.NewBufferString(`{"description":"new","body":"new body"}`)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api/skills/alpha", body)
	r.SetPathValue("name", "alpha")
	srv.handleUpdateSkill(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}

	got, _ := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if !strings.Contains(string(got), "description: new") {
		t.Errorf("frontmatter not updated:\n%s", got)
	}
	if !strings.Contains(string(got), "new body") {
		t.Errorf("body not updated:\n%s", got)
	}
}

func TestSkillFiles_ListsSubtree(t *testing.T) {
	srv, root := testServer(t)
	dir := filepath.Join(root, "skills", "alpha", "assets")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "skills", "alpha", "SKILL.md"), []byte("---\nname: alpha\n---\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "logo.png"), []byte("\x89PNG"), 0644); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/skills/alpha/files", nil)
	r.SetPathValue("name", "alpha")
	srv.handleSkillFiles(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "SKILL.md") || !strings.Contains(body, "assets/logo.png") {
		t.Errorf("file listing missing entries:\n%s", body)
	}
}

func TestSkillFile_RejectsTraversal(t *testing.T) {
	srv, root := testServer(t)
	dir := filepath.Join(root, "skills", "alpha")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: alpha\n---\n"), 0644); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/skills/alpha/files/..%2Fsecret", nil)
	r.SetPathValue("name", "alpha")
	r.SetPathValue("path", "../secret")
	srv.handleSkillFile(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("traversal should 400, got %d", w.Code)
	}
}
