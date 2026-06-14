package cc_compat

import (
	"path/filepath"
	"testing"
)

func TestJSONRegistry_ListMissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "installed_plugins.json")

	r := NewJSONRegistry(path)
	entries, err := r.List()
	if err != nil {
		t.Fatalf("List on missing file: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty list, got %d", len(entries))
	}
}
