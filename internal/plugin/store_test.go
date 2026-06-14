package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestStore_ListEmpty(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	plugins, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(plugins))
	}
}

func TestStore_ListWithPlugins(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	// Create plugin directory structure
	pluginDir := filepath.Join(dir, "test-marketplace", "test-plugin", "1.0.0")
	os.MkdirAll(pluginDir, 0755)
	os.WriteFile(filepath.Join(pluginDir, "package.json"), []byte("{}"), 0644)

	plugins, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0].Name != "test-plugin" {
		t.Errorf("expected name=test-plugin, got %s", plugins[0].Name)
	}
	if plugins[0].Marketplace != "test-marketplace" {
		t.Errorf("expected marketplace=test-marketplace, got %s", plugins[0].Marketplace)
	}
	if plugins[0].Version != "1.0.0" {
		t.Errorf("expected version=1.0.0, got %s", plugins[0].Version)
	}
}

func TestStore_GetWithMarketplace(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	pluginDir := filepath.Join(dir, "mp1", "myplugin", "2.0.0")
	os.MkdirAll(pluginDir, 0755)

	p, err := store.Get("myplugin@mp1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if p.Name != "myplugin" || p.Marketplace != "mp1" || p.Version != "2.0.0" {
		t.Errorf("unexpected plugin: %+v", p)
	}
}

func TestStore_GetWithoutMarketplace(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	pluginDir := filepath.Join(dir, "mp1", "myplugin", "1.0.0")
	os.MkdirAll(pluginDir, 0755)

	p, err := store.Get("myplugin")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if p.Name != "myplugin" || p.Marketplace != "mp1" {
		t.Errorf("unexpected plugin: %+v", p)
	}
}

func TestStore_GetNotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, err := store.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent plugin")
	}
}

func TestStore_Exists(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	os.MkdirAll(filepath.Join(dir, "mp", "exists", "v1"), 0755)

	if !store.Exists("exists@mp") {
		t.Error("expected Exists=true")
	}
	if store.Exists("nope") {
		t.Error("expected Exists=false for nonexistent")
	}
}

func TestStore_Remove(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	pluginDir := filepath.Join(dir, "mp", "removeme", "v1")
	os.MkdirAll(pluginDir, 0755)
	os.WriteFile(filepath.Join(pluginDir, "file.txt"), []byte("data"), 0644)

	if err := store.Remove("removeme@mp"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Version dir should be gone
	if _, err := os.Stat(pluginDir); !os.IsNotExist(err) {
		t.Error("expected plugin dir to be removed")
	}
	// Empty parent dirs should be cleaned up
	if _, err := os.Stat(filepath.Join(dir, "mp", "removeme")); !os.IsNotExist(err) {
		t.Error("expected empty parent dir to be removed")
	}
}

func TestStore_RemoveNotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	err := store.Remove("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent plugin")
	}
}

func TestStore_ListSkills(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	pluginDir := filepath.Join(dir, "mp", "myplugin", "v1")
	skillsDir := filepath.Join(pluginDir, "skills")
	os.MkdirAll(filepath.Join(skillsDir, "skill-a"), 0755)
	os.MkdirAll(filepath.Join(skillsDir, "skill-b"), 0755)
	os.MkdirAll(filepath.Join(skillsDir, "not-a-skill"), 0755)
	os.WriteFile(filepath.Join(skillsDir, "skill-a", "SKILL.md"), []byte("# A"), 0644)
	os.WriteFile(filepath.Join(skillsDir, "skill-b", "SKILL.md"), []byte("# B"), 0644)
	// not-a-skill has no SKILL.md

	skills, err := store.ListSkills("myplugin@mp")
	if err != nil {
		t.Fatalf("ListSkills failed: %v", err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}
}

func TestStore_ListSkillsNoDir(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	pluginDir := filepath.Join(dir, "mp", "noskills", "v1")
	os.MkdirAll(pluginDir, 0755)

	_, err := store.ListSkills("noskills@mp")
	if err == nil {
		t.Fatal("expected error when no skills directory")
	}
}

func TestStore_PluginDir(t *testing.T) {
	store := NewStore("/base")
	got := store.PluginDir("mp", "plug", "v1")
	want := "/base/mp/plug/v1"
	if got != want {
		t.Errorf("PluginDir = %s, want %s", got, want)
	}
}

func TestParseName(t *testing.T) {
	tests := []struct {
		input      string
		wantName   string
		wantMarket string
	}{
		{"superpowers@claude-plugins-official", "superpowers", "claude-plugins-official"},
		{"superpowers", "superpowers", ""},
		{"a@b@c", "a", "b@c"},
	}
	for _, tt := range tests {
		name, market := parseName(tt.input)
		if name != tt.wantName || market != tt.wantMarket {
			t.Errorf("parseName(%q) = (%q, %q), want (%q, %q)", tt.input, name, market, tt.wantName, tt.wantMarket)
		}
	}
}

func TestIsTag(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"5.1.0", true},
		{"v1.0", true},
		{"latest", false},
		{"abc123f", false},
	}
	for _, tt := range tests {
		got := isTag(tt.input)
		if got != tt.want {
			t.Errorf("isTag(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestValidateGitURL(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"https://github.com/user/repo.git", false},
		{"http://example.com/repo.git", false},
		{"git@github.com:user/repo.git", false},
		{"file:///tmp/repo.git", false},
		{"--upload-pack=evil", true},
		{"/local/path", true},
	}
	for _, tt := range tests {
		err := validateGitURL(tt.url)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateGitURL(%q) err=%v, wantErr=%v", tt.url, err, tt.wantErr)
		}
	}
}

func TestStore_GetMeta_NotPresent(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	pluginDir := filepath.Join(dir, "mp", "no-meta", "v1")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	meta, err := store.GetMeta("no-meta@mp")
	if err != nil {
		t.Fatalf("GetMeta failed: %v", err)
	}
	if meta != nil {
		t.Errorf("expected nil meta, got %+v", meta)
	}
}

func TestStore_GetMeta_Parses(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	pluginDir := filepath.Join(dir, "mp", "with-meta", "v1")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	want := &PluginMeta{
		URL:         "https://example.com/repo.git",
		Marketplace: "mp",
		CommitSha:   "deadbeef",
		Version:     "v1",
		InstalledAt: "2026-05-27T00:00:00Z",
	}
	data, err := json.MarshalIndent(want, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, metaFileName), data, 0644); err != nil {
		t.Fatalf("write meta: %v", err)
	}

	got, err := store.GetMeta("with-meta@mp")
	if err != nil {
		t.Fatalf("GetMeta failed: %v", err)
	}
	if got == nil {
		t.Fatal("expected meta, got nil")
	}
	if got.URL != want.URL || got.Marketplace != want.Marketplace ||
		got.CommitSha != want.CommitSha || got.Version != want.Version ||
		got.InstalledAt != want.InstalledAt {
		t.Errorf("meta mismatch: got %+v want %+v", got, want)
	}
}

func TestStore_GetRemoteURL_NoMeta(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	pluginDir := filepath.Join(dir, "mp", "bare", "v1")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, err := store.GetRemoteURL("bare@mp")
	if err == nil {
		t.Fatal("expected error when meta missing")
	}
}

func TestStore_GetRemoteURL_EmptyURL(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	pluginDir := filepath.Join(dir, "mp", "local", "v1")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	meta := &PluginMeta{Marketplace: "mp", Version: "v1"}
	data, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(filepath.Join(pluginDir, metaFileName), data, 0644); err != nil {
		t.Fatalf("write meta: %v", err)
	}

	_, err := store.GetRemoteURL("local@mp")
	if err == nil {
		t.Fatal("expected error when meta has empty url")
	}
}

func TestStore_ListBackfillsMeta(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	pluginDir := filepath.Join(dir, "mp", "ploog", "v1")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	meta := &PluginMeta{
		URL:         "https://example.com/ploog.git",
		Marketplace: "mp",
		CommitSha:   "abc1234",
		Version:     "v1",
		InstalledAt: "2026-05-27T00:00:00Z",
	}
	data, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(filepath.Join(pluginDir, metaFileName), data, 0644); err != nil {
		t.Fatalf("write meta: %v", err)
	}

	plugins, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0].RemoteURL != meta.URL {
		t.Errorf("expected RemoteURL=%q, got %q", meta.URL, plugins[0].RemoteURL)
	}
	if plugins[0].GitCommitSha != meta.CommitSha {
		t.Errorf("expected GitCommitSha=%q, got %q", meta.CommitSha, plugins[0].GitCommitSha)
	}
}

func TestStore_GetBackfillsMeta(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	pluginDir := filepath.Join(dir, "mp", "gp", "v1")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	meta := &PluginMeta{
		URL:         "https://example.com/gp.git",
		Marketplace: "mp",
		CommitSha:   "feedface",
		Version:     "v1",
		InstalledAt: "2026-05-27T00:00:00Z",
	}
	data, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(filepath.Join(pluginDir, metaFileName), data, 0644); err != nil {
		t.Fatalf("write meta: %v", err)
	}

	p, err := store.Get("gp@mp")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if p.RemoteURL != meta.URL {
		t.Errorf("expected RemoteURL=%q, got %q", meta.URL, p.RemoteURL)
	}
	if p.GitCommitSha != meta.CommitSha {
		t.Errorf("expected GitCommitSha=%q, got %q", meta.CommitSha, p.GitCommitSha)
	}
}

func TestStore_Update_NoMeta(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	pluginDir := filepath.Join(dir, "mp", "stale", "v1")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	err := store.Update("stale@mp")
	if err == nil {
		t.Fatal("expected error when no meta present")
	}
}