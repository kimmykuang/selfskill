package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoader_CreateSymlinkLifecycle(t *testing.T) {
	pluginsDir := t.TempDir()
	store := NewStore(pluginsDir)

	pluginDir := filepath.Join(pluginsDir, "mp", "testplugin", "1.0.0")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "package.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	loader := NewLoader(store)
	p, err := store.Get("testplugin@mp")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	ccCacheDir := t.TempDir()
	linkPath := filepath.Join(ccCacheDir, p.Marketplace, p.Name, p.Version)

	if err := loader.createSymlink(p.InstallPath, linkPath); err != nil {
		t.Fatalf("createSymlink: %v", err)
	}
	info, err := os.Lstat(linkPath)
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("symlink not created")
	}

	// Idempotent
	if err := loader.createSymlink(p.InstallPath, linkPath); err != nil {
		t.Fatalf("idempotent createSymlink: %v", err)
	}

	// Replace
	other := t.TempDir()
	if err := loader.createSymlink(other, linkPath); err != nil {
		t.Fatalf("replace createSymlink: %v", err)
	}
	got, _ := os.Readlink(linkPath)
	gotAbs, _ := filepath.Abs(got)
	wantAbs, _ := filepath.Abs(other)
	if gotAbs != wantAbs {
		t.Errorf("after replace, target = %s, want %s", gotAbs, wantAbs)
	}

	// Regular file should error
	regular := filepath.Join(ccCacheDir, "regular")
	if err := os.WriteFile(regular, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := loader.createSymlink(other, regular); err == nil {
		t.Fatal("expected error on non-symlink path")
	}
}

func TestIsSymlink(t *testing.T) {
	dir := t.TempDir()

	regular := filepath.Join(dir, "regular")
	os.WriteFile(regular, []byte("data"), 0644)
	if isSymlink(regular) {
		t.Error("regular file should not be symlink")
	}

	link := filepath.Join(dir, "link")
	os.Symlink(regular, link)
	if !isSymlink(link) {
		t.Error("symlink should be detected")
	}

	if isSymlink(filepath.Join(dir, "nope")) {
		t.Error("nonexistent should not be symlink")
	}
}

func TestRemoveIfEmpty(t *testing.T) {
	dir := t.TempDir()

	emptyDir := filepath.Join(dir, "empty")
	os.MkdirAll(emptyDir, 0755)
	removeIfEmpty(emptyDir)
	if _, err := os.Stat(emptyDir); !os.IsNotExist(err) {
		t.Error("empty dir should be removed")
	}

	nonEmpty := filepath.Join(dir, "nonempty")
	os.MkdirAll(nonEmpty, 0755)
	os.WriteFile(filepath.Join(nonEmpty, "file"), []byte("x"), 0644)
	removeIfEmpty(nonEmpty)
	if _, err := os.Stat(nonEmpty); err != nil {
		t.Error("non-empty dir should remain")
	}
}
