package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimmykuang/selfskill/internal/service"
)

func TestSearch_AcrossSkillsAndPrompts(t *testing.T) {
	srv, root := testServer(t)

	// Add a skill matching "video".
	dir := filepath.Join(root, "skills", "video-summary")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: video-summary\ndescription: video stuff\n---\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Add a prompt matching "video".
	if err := os.WriteFile(filepath.Join(root, "prompts", "video-prompt.md"), []byte("---\nid: video-prompt\ndescription: about videos\n---\n\nbody\n"), 0644); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/search?q=video", nil)
	srv.handleSearch(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var hits []service.Hit
	if err := json.Unmarshal(w.Body.Bytes(), &hits); err != nil {
		t.Fatal(err)
	}
	if len(hits) != 2 {
		t.Fatalf("got %d hits, want 2", len(hits))
	}
}

func TestSearch_EmptyQueryReturnsEmpty(t *testing.T) {
	srv, _ := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/search?q=", nil)
	srv.handleSearch(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	var hits []service.Hit
	if err := json.Unmarshal(w.Body.Bytes(), &hits); err != nil {
		t.Fatal(err)
	}
	if len(hits) != 0 {
		t.Errorf("expected empty hits, got %d", len(hits))
	}
}

func TestListSkills_FilterByOrigin(t *testing.T) {
	srv, root := testServer(t)
	mk := func(name, fm string) {
		dir := filepath.Join(root, "skills", name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\n"+fm+"---\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	mk("local-x", "name: local-x\n")
	mk("remote-x", "name: remote-x\nsource: https://example.com/x.md\n")

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/skills?origin=local", nil)
	srv.handleListSkills(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "local-x") || strings.Contains(body, "remote-x") {
		t.Errorf("origin=local filter wrong; body=%s", body)
	}
}
