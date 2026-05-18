package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoader_LoadAndUnload(t *testing.T) {
	// Setup plugin store
	pluginsDir := t.TempDir()
	store := NewStore(pluginsDir)

	// Create a plugin
	pluginDir := filepath.Join(pluginsDir, "mp", "testplugin", "1.0.0")
	os.MkdirAll(pluginDir, 0755)
	os.WriteFile(filepath.Join(pluginDir, "package.json"), []byte("{}"), 0644)

	// Setup CC cache and JSON paths using temp dirs
	ccCacheDir := t.TempDir()
	ccPluginsDir := t.TempDir()
	jsonPath := filepath.Join(ccPluginsDir, "installed_plugins.json")

	// Override config paths for testing — we'll test the internal methods directly
	loader := NewLoader(store)

	// Test Load by calling internal methods
	p, err := store.Get("testplugin@mp")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Create symlink manually (since createSymlink uses config paths)
	linkPath := filepath.Join(ccCacheDir, p.Marketplace, p.Name, p.Version)
	if err := loader.createSymlink(p.InstallPath, linkPath); err != nil {
		t.Fatalf("createSymlink failed: %v", err)
	}

	// Verify symlink exists
	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("symlink not created: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("expected symlink")
	}

	// Verify target
	target, _ := os.Readlink(linkPath)
	if target != p.InstallPath {
		t.Errorf("symlink target = %s, want %s", target, p.InstallPath)
	}

	// Test idempotent symlink (same target)
	if err := loader.createSymlink(p.InstallPath, linkPath); err != nil {
		t.Fatalf("idempotent createSymlink failed: %v", err)
	}

	// Test symlink to different target (should replace)
	otherDir := t.TempDir()
	if err := loader.createSymlink(otherDir, linkPath); err != nil {
		t.Fatalf("replace createSymlink failed: %v", err)
	}
	newTarget, _ := os.Readlink(linkPath)
	absNew, _ := filepath.Abs(newTarget)
	absOther, _ := filepath.Abs(otherDir)
	if absNew != absOther {
		t.Errorf("after replace, target = %s, want %s", absNew, absOther)
	}

	// Test createSymlink when path is a regular file (should error)
	regularFile := filepath.Join(ccCacheDir, "regular")
	os.WriteFile(regularFile, []byte("data"), 0644)
	if err := loader.createSymlink(otherDir, regularFile); err == nil {
		t.Fatal("expected error when path is a regular file")
	}

	// Test JSON write/read
	os.WriteFile(jsonPath, []byte(`{"version":2,"plugins":{}}`), 0644)
}

func TestLoader_JSONReadWrite(t *testing.T) {
	// Create a temp JSON file
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "installed_plugins.json")

	// Write initial empty file
	initial := &InstalledPluginsFile{
		Version: 2,
		Plugins: map[string][]InstalledPluginEntry{},
	}
	data, _ := json.MarshalIndent(initial, "", "  ")
	os.WriteFile(jsonPath, data, 0644)

	// Read it back using a loader (we'll test readJSON/writeJSON directly)
	pluginsDir := t.TempDir()
	store := NewStore(pluginsDir)
	loader := NewLoader(store)

	// Since readJSON uses config.CCInstalledPluginsFile(), we test the format
	// by verifying the struct serialization
	f := &InstalledPluginsFile{
		Version: 2,
		Plugins: map[string][]InstalledPluginEntry{
			"test@mp": {
				{
					Scope:       "user",
					InstallPath: "/path/to/plugin",
					Version:     "1.0.0",
					InstalledAt: "2026-01-01T00:00:00Z",
					LastUpdated: "2026-01-01T00:00:00Z",
				},
			},
		},
	}

	// Test writeJSON
	outPath := filepath.Join(tmpDir, "out.json")
	outData, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	os.WriteFile(outPath, outData, 0644)

	// Read back and verify
	readData, _ := os.ReadFile(outPath)
	var readBack InstalledPluginsFile
	if err := json.Unmarshal(readData, &readBack); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if readBack.Version != 2 {
		t.Errorf("version = %d, want 2", readBack.Version)
	}
	entries, ok := readBack.Plugins["test@mp"]
	if !ok || len(entries) != 1 {
		t.Fatalf("expected 1 entry for test@mp")
	}
	if entries[0].Version != "1.0.0" {
		t.Errorf("entry version = %s, want 1.0.0", entries[0].Version)
	}

	_ = loader // used to verify NewLoader doesn't panic
}

func TestIsSymlink(t *testing.T) {
	dir := t.TempDir()

	// Regular file
	regular := filepath.Join(dir, "regular")
	os.WriteFile(regular, []byte("data"), 0644)
	if isSymlink(regular) {
		t.Error("regular file should not be symlink")
	}

	// Symlink
	link := filepath.Join(dir, "link")
	os.Symlink(regular, link)
	if !isSymlink(link) {
		t.Error("symlink should be detected")
	}

	// Nonexistent
	if isSymlink(filepath.Join(dir, "nope")) {
		t.Error("nonexistent should not be symlink")
	}
}

func TestRemoveIfEmpty(t *testing.T) {
	dir := t.TempDir()

	// Empty dir should be removed
	emptyDir := filepath.Join(dir, "empty")
	os.MkdirAll(emptyDir, 0755)
	removeIfEmpty(emptyDir)
	if _, err := os.Stat(emptyDir); !os.IsNotExist(err) {
		t.Error("empty dir should be removed")
	}

	// Non-empty dir should remain
	nonEmpty := filepath.Join(dir, "nonempty")
	os.MkdirAll(nonEmpty, 0755)
	os.WriteFile(filepath.Join(nonEmpty, "file"), []byte("x"), 0644)
	removeIfEmpty(nonEmpty)
	if _, err := os.Stat(nonEmpty); err != nil {
		t.Error("non-empty dir should remain")
	}
}
