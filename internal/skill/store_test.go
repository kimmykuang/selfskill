package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore_SaveAndGet(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	// Create a source skill directory
	srcDir := t.TempDir()
	os.MkdirAll(filepath.Join(srcDir, "scripts"), 0755)
	os.MkdirAll(filepath.Join(srcDir, "assets"), 0755)
	os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte(`---
name: test-skill
version: 1.0.0
description: A test skill
---
# Test Skill
`), 0644)
	os.WriteFile(filepath.Join(srcDir, "scripts", "run.sh"), []byte("#!/bin/bash\necho hello"), 0755)

	// Save
	if err := store.Save("test-skill", srcDir); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Exists
	if !store.Exists("test-skill") {
		t.Fatal("expected skill to exist")
	}

	// Get
	sk, err := store.Get("test-skill")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if sk.Name != "test-skill" {
		t.Errorf("expected name=test-skill, got %s", sk.Name)
	}
	if sk.Version != "1.0.0" {
		t.Errorf("expected version=1.0.0, got %s", sk.Version)
	}
	if !sk.HasScripts {
		t.Error("expected HasScripts=true")
	}
	if !sk.HasAssets {
		t.Error("expected HasAssets=true")
	}

	// Verify script is executable
	info, _ := os.Stat(filepath.Join(dir, "test-skill", "scripts", "run.sh"))
	if info.Mode()&0111 == 0 {
		t.Error("expected script to be executable")
	}
}

func TestStore_List(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	store.SaveFromContent("skill-a", []byte("---\nversion: 1.0.0\n---\n# A\n"))
	store.SaveFromContent("skill-b", []byte("---\nversion: 2.0.0\n---\n# B\n"))

	skills, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}
}

func TestStore_Remove(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	store.SaveFromContent("to-remove", []byte("# Remove me\n"))
	if !store.Exists("to-remove") {
		t.Fatal("expected skill to exist before removal")
	}

	if err := store.Remove("to-remove"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	if store.Exists("to-remove") {
		t.Fatal("expected skill to not exist after removal")
	}
}
