package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimmykuang/selfskill/internal/installer"
)

func TestInstall_LocalSkillDirectory(t *testing.T) {
	srv, root := testServer(t)

	// Make a local skill source dir.
	src := filepath.Join(root, "external", "video-summary")
	if err := os.MkdirAll(src, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "SKILL.md"), []byte("---\nname: video-summary\n---\n"), 0644); err != nil {
		t.Fatal(err)
	}

	srv.Installer = installer.New(srv.SkillStore, srv.PromptStore)

	body := bytes.NewBufferString(`{"type":"skill","source":"` + src + `"}`)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/install", body)
	srv.handleInstall(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("install status %d body %s", w.Code, w.Body.String())
	}

	if !srv.SkillStore.Exists("video-summary") {
		t.Error("video-summary should be installed")
	}
}

func TestInstall_BadType(t *testing.T) {
	srv, _ := testServer(t)
	body := bytes.NewBufferString(`{"type":"banana","source":"x"}`)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/install", body)
	srv.handleInstall(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}
