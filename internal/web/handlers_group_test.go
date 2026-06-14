package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/kimmykuang/selfskill/internal/group"
	"github.com/kimmykuang/selfskill/internal/linker"
	"github.com/kimmykuang/selfskill/internal/plugin"
	"github.com/kimmykuang/selfskill/internal/plugin/cc_compat"
	"github.com/kimmykuang/selfskill/internal/prompt"
	"github.com/kimmykuang/selfskill/internal/service"
	"github.com/kimmykuang/selfskill/internal/skill"
)

// testServer constructs a Server backed by tempdir stores. Reused by other
// handler tests in this package.
func testServer(t *testing.T) (*Server, string) {
	t.Helper()
	root := t.TempDir()
	for _, sub := range []string{"skills", "prompts", "groups", "plugins", "claude/skills"} {
		if err := os.MkdirAll(filepath.Join(root, sub), 0755); err != nil {
			t.Fatal(err)
		}
	}
	skills := skill.NewStore(filepath.Join(root, "skills"))
	prompts := prompt.NewStore(filepath.Join(root, "prompts"))
	groups := group.NewStore(filepath.Join(root, "groups"))
	plugins := plugin.NewStore(filepath.Join(root, "plugins"))
	reg := cc_compat.NewFakeRegistry()
	pluginLoader := plugin.NewLoaderWithRegistry(plugins, reg)
	lnk := linker.NewWithDirs(filepath.Join(root, "skills"), filepath.Join(root, "claude/skills"), filepath.Join(root, "claude/skills"))

	svc := service.New(service.Deps{
		Skills: skills, Prompts: prompts, Groups: groups, Plugins: plugins,
		Linker: lnk, PluginLoader: pluginLoader, Registry: reg,
	})

	return &Server{
		SkillStore: skills, PromptStore: prompts, PluginStore: plugins,
		GroupStore: groups, PluginLoader: pluginLoader, Linker: lnk,
		Svc: svc,
	}, root
}

func TestGroupAPI_CreateListGet(t *testing.T) {
	srv, _ := testServer(t)

	// Create group via POST.
	body := bytes.NewBufferString(`{"name":"work"}`)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/groups", body)
	srv.handleCreateGroup(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("create status %d body %s", w.Code, w.Body.String())
	}

	// List groups.
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/api/groups", nil)
	srv.handleListGroups(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("list status %d body %s", w.Code, w.Body.String())
	}
	var listed []group.Group
	if err := json.Unmarshal(w.Body.Bytes(), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].Name != "work" {
		t.Fatalf("listed = %+v", listed)
	}

	// Get specific group.
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/api/groups/work", nil)
	r.SetPathValue("name", "work")
	srv.handleGetGroup(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("get status %d body %s", w.Code, w.Body.String())
	}

	// Get missing group → 404.
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/api/groups/missing", nil)
	r.SetPathValue("name", "missing")
	srv.handleGetGroup(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("missing status = %d, want 404", w.Code)
	}
}

func TestGroupAPI_UpdateAndDelete(t *testing.T) {
	srv, _ := testServer(t)

	// Pre-create.
	if err := srv.GroupStore.Create("work"); err != nil {
		t.Fatal(err)
	}

	// PUT replaces skills/plugins/prompts.
	body := bytes.NewBufferString(`{"description":"day job","skills":["a","b"],"plugins":["sp"],"prompts":["p1"]}`)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api/groups/work", body)
	r.SetPathValue("name", "work")
	srv.handleUpdateGroup(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("update status %d body %s", w.Code, w.Body.String())
	}

	g, err := srv.GroupStore.Get("work")
	if err != nil {
		t.Fatal(err)
	}
	if g.Description != "day job" || len(g.Skills) != 2 || len(g.Plugins) != 1 || len(g.Prompts) != 1 {
		t.Fatalf("after PUT, group = %+v", g)
	}

	// DELETE removes it.
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodDelete, "/api/groups/work", nil)
	r.SetPathValue("name", "work")
	srv.handleDeleteGroup(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("delete status %d", w.Code)
	}

	if _, err := srv.GroupStore.Get("work"); err == nil {
		t.Fatal("group should be gone after delete")
	}
}

func TestGroupAPI_LoadAndUnload(t *testing.T) {
	srv, root := testServer(t)

	// Set up a skill and a group containing it.
	skillDir := filepath.Join(root, "skills", "alpha")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: alpha\n---\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := srv.GroupStore.Create("g1"); err != nil {
		t.Fatal(err)
	}
	if err := srv.GroupStore.AddSkill("g1", "alpha"); err != nil {
		t.Fatal(err)
	}

	// POST load.
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/groups/g1/load", nil)
	r.SetPathValue("name", "g1")
	srv.handleLoadGroup(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("load status %d body %s", w.Code, w.Body.String())
	}

	if !srv.Svc.IsSkillLoaded("alpha", linker.ScopeUser) {
		t.Error("alpha should be loaded after POST load")
	}

	// POST unload.
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/api/groups/g1/unload", nil)
	r.SetPathValue("name", "g1")
	srv.handleUnloadGroup(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("unload status %d body %s", w.Code, w.Body.String())
	}
	if srv.Svc.IsSkillLoaded("alpha", linker.ScopeUser) {
		t.Error("alpha should not be loaded after POST unload")
	}
}
