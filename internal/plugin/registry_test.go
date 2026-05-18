package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRegistry_ListMarketplacesEmpty(t *testing.T) {
	dir := t.TempDir()
	r := &Registry{dir: dir}

	names, err := r.ListMarketplaces()
	if err != nil {
		t.Fatalf("ListMarketplaces failed: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("expected 0 marketplaces, got %d", len(names))
	}
}

func TestRegistry_ListMarketplaces(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "mp1"), 0755)
	os.MkdirAll(filepath.Join(dir, "mp2"), 0755)
	os.WriteFile(filepath.Join(dir, "not-a-dir"), []byte("x"), 0644)

	r := &Registry{dir: dir}
	names, err := r.ListMarketplaces()
	if err != nil {
		t.Fatalf("ListMarketplaces failed: %v", err)
	}
	if len(names) != 2 {
		t.Errorf("expected 2 marketplaces, got %d", len(names))
	}
}

func TestRegistry_LoadMarketplace(t *testing.T) {
	dir := t.TempDir()
	mpDir := filepath.Join(dir, "test-mp", ".claude-plugin")
	os.MkdirAll(mpDir, 0755)

	mf := MarketplaceFile{
		Name:        "test-mp",
		Description: "Test marketplace",
		Plugins: []MarketplacePlugin{
			{Name: "plugin-a", Description: "Plugin A", Source: "./plugins/a"},
			{Name: "plugin-b", Description: "Plugin B", Source: map[string]interface{}{"source": "url", "url": "https://github.com/user/repo.git"}},
		},
	}
	data, _ := json.Marshal(mf)
	os.WriteFile(filepath.Join(mpDir, "marketplace.json"), data, 0644)

	r := &Registry{dir: dir}
	loaded, err := r.LoadMarketplace("test-mp")
	if err != nil {
		t.Fatalf("LoadMarketplace failed: %v", err)
	}
	if loaded.Name != "test-mp" {
		t.Errorf("name = %s, want test-mp", loaded.Name)
	}
	if len(loaded.Plugins) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(loaded.Plugins))
	}
}

func TestRegistry_LoadMarketplaceRootJSON(t *testing.T) {
	dir := t.TempDir()
	mpDir := filepath.Join(dir, "root-mp")
	os.MkdirAll(mpDir, 0755)

	mf := MarketplaceFile{Name: "root-mp", Plugins: []MarketplacePlugin{{Name: "p1"}}}
	data, _ := json.Marshal(mf)
	os.WriteFile(filepath.Join(mpDir, "marketplace.json"), data, 0644)

	r := &Registry{dir: dir}
	loaded, err := r.LoadMarketplace("root-mp")
	if err != nil {
		t.Fatalf("LoadMarketplace (root) failed: %v", err)
	}
	if loaded.Name != "root-mp" {
		t.Errorf("name = %s, want root-mp", loaded.Name)
	}
}

func TestRegistry_LoadMarketplaceNotFound(t *testing.T) {
	dir := t.TempDir()
	r := &Registry{dir: dir}

	_, err := r.LoadMarketplace("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent marketplace")
	}
}

func TestRegistry_FindPlugin(t *testing.T) {
	dir := t.TempDir()
	mpDir := filepath.Join(dir, "mp", ".claude-plugin")
	os.MkdirAll(mpDir, 0755)

	mf := MarketplaceFile{
		Plugins: []MarketplacePlugin{
			{Name: "alpha", Description: "Alpha plugin"},
			{Name: "beta", Description: "Beta plugin"},
		},
	}
	data, _ := json.Marshal(mf)
	os.WriteFile(filepath.Join(mpDir, "marketplace.json"), data, 0644)

	r := &Registry{dir: dir}

	p, err := r.FindPlugin("beta", "mp")
	if err != nil {
		t.Fatalf("FindPlugin failed: %v", err)
	}
	if p.Name != "beta" {
		t.Errorf("name = %s, want beta", p.Name)
	}

	_, err = r.FindPlugin("nonexistent", "mp")
	if err == nil {
		t.Fatal("expected error for nonexistent plugin")
	}
}

func TestRegistry_ResolveGitURL(t *testing.T) {
	r := &Registry{}

	// Remote plugin with URL
	mp := &MarketplacePlugin{
		Name:   "remote",
		Source: map[string]interface{}{"url": "https://github.com/user/repo.git"},
	}
	url, err := r.ResolveGitURL(mp, "mp")
	if err != nil {
		t.Fatalf("ResolveGitURL failed: %v", err)
	}
	if url != "https://github.com/user/repo.git" {
		t.Errorf("url = %s", url)
	}

	// Local plugin (string source)
	local := &MarketplacePlugin{Name: "local", Source: "./plugins/local"}
	_, err = r.ResolveGitURL(local, "mp")
	if err == nil {
		t.Fatal("expected error for local plugin")
	}

	// Source object without url field
	noURL := &MarketplacePlugin{Name: "bad", Source: map[string]interface{}{"source": "url"}}
	_, err = r.ResolveGitURL(noURL, "mp")
	if err == nil {
		t.Fatal("expected error for source without url")
	}
}

func TestRegistry_IsLocalPlugin(t *testing.T) {
	r := &Registry{}

	local := &MarketplacePlugin{Source: "./plugins/x"}
	if !r.IsLocalPlugin(local) {
		t.Error("expected IsLocalPlugin=true for string source")
	}

	remote := &MarketplacePlugin{Source: map[string]interface{}{"url": "https://x.git"}}
	if r.IsLocalPlugin(remote) {
		t.Error("expected IsLocalPlugin=false for object source")
	}
}

func TestRegistry_LocalPluginPath(t *testing.T) {
	r := &Registry{dir: "/base/marketplaces"}

	mp := &MarketplacePlugin{Name: "test", Source: "./plugins/test"}
	path, err := r.LocalPluginPath(mp, "my-mp")
	if err != nil {
		t.Fatalf("LocalPluginPath failed: %v", err)
	}
	want := "/base/marketplaces/my-mp/plugins/test"
	if path != want {
		t.Errorf("path = %s, want %s", path, want)
	}

	// Non-local plugin should error
	remote := &MarketplacePlugin{Name: "r", Source: map[string]interface{}{"url": "x"}}
	_, err = r.LocalPluginPath(remote, "mp")
	if err == nil {
		t.Fatal("expected error for non-local plugin")
	}
}

func TestRepoNameFromURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://github.com/anthropics/claude-plugins-official.git", "claude-plugins-official"},
		{"https://github.com/user/repo", "repo"},
		{"git@github.com:user/my-repo.git", "my-repo"},
		{"git@github.com:org/repo", "repo"},
		{"simple-name", "simple-name"},
	}
	for _, tt := range tests {
		got := repoNameFromURL(tt.input)
		if got != tt.want {
			t.Errorf("repoNameFromURL(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRegistry_FindMarketplaceJSON(t *testing.T) {
	dir := t.TempDir()
	r := &Registry{dir: dir}

	// Test .claude-plugin/marketplace.json (priority)
	mpDir := filepath.Join(dir, "mp1")
	os.MkdirAll(filepath.Join(mpDir, ".claude-plugin"), 0755)
	os.WriteFile(filepath.Join(mpDir, ".claude-plugin", "marketplace.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(mpDir, "marketplace.json"), []byte("{}"), 0644)

	path, err := r.findMarketplaceJSON("mp1")
	if err != nil {
		t.Fatalf("findMarketplaceJSON failed: %v", err)
	}
	if filepath.Base(filepath.Dir(path)) != ".claude-plugin" {
		t.Errorf("expected .claude-plugin path, got %s", path)
	}

	// Test root marketplace.json fallback
	mp2Dir := filepath.Join(dir, "mp2")
	os.MkdirAll(mp2Dir, 0755)
	os.WriteFile(filepath.Join(mp2Dir, "marketplace.json"), []byte("{}"), 0644)

	path, err = r.findMarketplaceJSON("mp2")
	if err != nil {
		t.Fatalf("findMarketplaceJSON (root) failed: %v", err)
	}
	if filepath.Base(path) != "marketplace.json" {
		t.Errorf("expected marketplace.json, got %s", path)
	}

	// Test not found
	os.MkdirAll(filepath.Join(dir, "mp3"), 0755)
	_, err = r.findMarketplaceJSON("mp3")
	if err == nil {
		t.Fatal("expected error when no marketplace.json")
	}
}
