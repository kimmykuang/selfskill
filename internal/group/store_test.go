package group

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore_CreateAndGet(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	if err := store.Create("testgroup"); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	g, err := store.Get("testgroup")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if g.Name != "testgroup" {
		t.Errorf("expected name=testgroup, got %s", g.Name)
	}
	if len(g.Skills) != 0 {
		t.Errorf("expected empty skills, got %v", g.Skills)
	}
}

func TestStore_CreateDuplicate(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	store.Create("dup")
	if err := store.Create("dup"); err == nil {
		t.Fatal("expected error on duplicate create")
	}
}

func TestStore_AddAndRemoveSkill(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	store.Create("mygroup")

	if err := store.AddSkill("mygroup", "skill-a"); err != nil {
		t.Fatalf("AddSkill failed: %v", err)
	}
	if err := store.AddSkill("mygroup", "skill-b"); err != nil {
		t.Fatalf("AddSkill failed: %v", err)
	}
	// Deduplicate
	if err := store.AddSkill("mygroup", "skill-a"); err != nil {
		t.Fatalf("AddSkill dedup failed: %v", err)
	}

	g, _ := store.Get("mygroup")
	if len(g.Skills) != 2 {
		t.Fatalf("expected 2 skills, got %d: %v", len(g.Skills), g.Skills)
	}

	if err := store.RemoveSkill("mygroup", "skill-a"); err != nil {
		t.Fatalf("RemoveSkill failed: %v", err)
	}
	g, _ = store.Get("mygroup")
	if len(g.Skills) != 1 || g.Skills[0] != "skill-b" {
		t.Errorf("expected [skill-b], got %v", g.Skills)
	}
}

func TestStore_List(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	store.Create("group-a")
	store.Create("group-b")

	groups, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
}

func TestStore_Delete(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	store.Create("to-delete")
	if err := store.Delete("to-delete"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := store.Get("to-delete"); err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestStore_OldYAMLWithoutPromptsField(t *testing.T) {
	dir := t.TempDir()
	// Write a group YAML in the pre-Phase-3 shape — no `prompts:` field at all.
	path := filepath.Join(dir, "legacy.yaml")
	yamlData := []byte("name: legacy\ndescription: pre-phase3\nskills: [a, b]\nplugins: [x]\n")
	if err := os.WriteFile(path, yamlData, 0644); err != nil {
		t.Fatal(err)
	}

	s := NewStore(dir)
	g, err := s.Get("legacy")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if g.Name != "legacy" || len(g.Skills) != 2 || len(g.Plugins) != 1 {
		t.Errorf("legacy fields lost: %+v", g)
	}
	if len(g.Prompts) != 0 {
		t.Errorf("Prompts should be empty, got %v", g.Prompts)
	}
}
