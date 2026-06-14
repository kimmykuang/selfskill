package cc_compat

import (
	"os"
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

func TestJSONRegistry_AddListRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "installed_plugins.json")
	r := NewJSONRegistry(path)

	entry := Entry{
		Scope:        "user",
		InstallPath:  "/tmp/pkg/superpowers/5.1.0",
		Version:      "5.1.0",
		InstalledAt:  "2026-01-01T00:00:00Z",
		LastUpdated:  "2026-01-01T00:00:00Z",
		GitCommitSha: "abc123",
	}

	if err := r.Add("superpowers@mp", entry); err != nil {
		t.Fatalf("Add: %v", err)
	}

	all, err := r.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	got, ok := all["superpowers@mp"]
	if !ok || len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0].Version != "5.1.0" {
		t.Errorf("version = %q, want 5.1.0", got[0].Version)
	}

	if err := r.Remove("superpowers@mp"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	all, _ = r.List()
	if _, ok := all["superpowers@mp"]; ok {
		t.Fatal("entry not removed")
	}
}

func TestJSONRegistry_AddPreservesInstalledAt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "installed_plugins.json")
	r := NewJSONRegistry(path)

	first := Entry{Scope: "user", InstallPath: "/p", Version: "1", InstalledAt: "2026-01-01T00:00:00Z", LastUpdated: "2026-01-01T00:00:00Z"}
	if err := r.Add("x@mp", first); err != nil {
		t.Fatalf("first Add: %v", err)
	}

	// Re-add without InstalledAt; previous one should be preserved.
	second := Entry{Scope: "user", InstallPath: "/p", Version: "2", LastUpdated: "2026-02-01T00:00:00Z"}
	if err := r.Add("x@mp", second); err != nil {
		t.Fatalf("second Add: %v", err)
	}

	all, _ := r.List()
	got := all["x@mp"][0]
	if got.InstalledAt != "2026-01-01T00:00:00Z" {
		t.Errorf("InstalledAt = %q, want preserved 2026-01-01...", got.InstalledAt)
	}
	if got.Version != "2" {
		t.Errorf("Version = %q, want 2", got.Version)
	}
}

func TestJSONRegistry_RemoveMissingIsNoOp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "installed_plugins.json")
	r := NewJSONRegistry(path)

	if err := r.Remove("nope@mp"); err != nil {
		t.Errorf("Remove on missing should be no-op, got %v", err)
	}
}

func TestFakeRegistry(t *testing.T) {
	f := NewFakeRegistry()

	if err := f.Add("a@mp", Entry{Scope: "user", InstallPath: "/x", Version: "1"}); err != nil {
		t.Fatal(err)
	}
	all, _ := f.List()
	if _, ok := all["a@mp"]; !ok {
		t.Fatal("Add did not record entry")
	}
	if err := f.Remove("a@mp"); err != nil {
		t.Fatal(err)
	}
	all, _ = f.List()
	if _, ok := all["a@mp"]; ok {
		t.Fatal("Remove did not delete entry")
	}
}

func TestJSONRegistry_AddDefaultsInstalledAtFromLastUpdated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "installed_plugins.json")
	r := NewJSONRegistry(path)

	// First install: no prior entry, no InstalledAt — should default to LastUpdated.
	if err := r.Add("first@mp", Entry{Scope: "user", InstallPath: "/p", Version: "1", LastUpdated: "2026-06-01T00:00:00Z"}); err != nil {
		t.Fatal(err)
	}
	all, _ := r.List()
	got := all["first@mp"][0]
	if got.InstalledAt != "2026-06-01T00:00:00Z" {
		t.Errorf("InstalledAt = %q, want default to LastUpdated 2026-06-01...", got.InstalledAt)
	}
}

func TestFakeRegistry_AddDefaultsInstalledAtFromLastUpdated(t *testing.T) {
	f := NewFakeRegistry()

	// First install: no prior entry, no InstalledAt — should default to LastUpdated.
	if err := f.Add("first@mp", Entry{Scope: "user", InstallPath: "/p", Version: "1", LastUpdated: "2026-06-01T00:00:00Z"}); err != nil {
		t.Fatal(err)
	}
	all, _ := f.List()
	got := all["first@mp"][0]
	if got.InstalledAt != "2026-06-01T00:00:00Z" {
		t.Errorf("InstalledAt = %q, want default to LastUpdated 2026-06-01...", got.InstalledAt)
	}
}

func TestJSONRegistry_PreservesUnrelatedPlugins(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "installed_plugins.json")

	// Pre-seed with a CC-written entry.
	seed := []byte(`{"version":2,"plugins":{"cc-native@mp":[{"scope":"user","installPath":"/cc/x","version":"1.0.0","installedAt":"2026-01-01T00:00:00Z","lastUpdated":"2026-01-01T00:00:00Z"}]}}`)
	if err := os.WriteFile(path, seed, 0644); err != nil {
		t.Fatal(err)
	}

	r := NewJSONRegistry(path)
	if err := r.Add("ours@mp", Entry{Scope: "user", InstallPath: "/ours", Version: "0.1", InstalledAt: "2026-06-01T00:00:00Z", LastUpdated: "2026-06-01T00:00:00Z"}); err != nil {
		t.Fatal(err)
	}

	all, _ := r.List()
	if _, ok := all["cc-native@mp"]; !ok {
		t.Error("CC-native entry was clobbered")
	}
	if _, ok := all["ours@mp"]; !ok {
		t.Error("our entry missing")
	}
}
