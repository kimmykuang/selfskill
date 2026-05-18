package prompt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore_SaveAndGet(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	p := &Prompt{
		ID:          "test-prompt",
		Tags:        []string{"video", "summary"},
		Description: "A test prompt",
		Body:        "# Test\n\nDo something.\n",
	}

	if err := store.Save(p); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, err := store.Get("test-prompt")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.ID != "test-prompt" {
		t.Errorf("expected id=test-prompt, got %s", got.ID)
	}
	if got.Description != "A test prompt" {
		t.Errorf("expected description mismatch, got %s", got.Description)
	}
	if len(got.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(got.Tags))
	}
	if got.Body != "# Test\n\nDo something.\n" {
		t.Errorf("unexpected body: %q", got.Body)
	}
}

func TestStore_List(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	store.Save(&Prompt{ID: "a", Tags: []string{"t1"}, Body: "body a"})
	store.Save(&Prompt{ID: "b", Tags: []string{"t2"}, Body: "body b"})

	prompts, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(prompts) != 2 {
		t.Fatalf("expected 2 prompts, got %d", len(prompts))
	}
}

func TestStore_Search(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	store.Save(&Prompt{ID: "wuxia-video", Tags: []string{"video", "wuxia"}, Body: "武侠视频"})
	store.Save(&Prompt{ID: "tech-video", Tags: []string{"video", "tech"}, Body: "技术视频"})
	store.Save(&Prompt{ID: "cooking", Tags: []string{"food"}, Body: "做饭"})

	// Search by tag
	results, _ := store.Search("", []string{"wuxia"})
	if len(results) != 1 || results[0].ID != "wuxia-video" {
		t.Errorf("tag search failed: %v", results)
	}

	// Search by query
	results, _ = store.Search("技术", nil)
	if len(results) != 1 || results[0].ID != "tech-video" {
		t.Errorf("query search failed: %v", results)
	}
}

func TestStore_Remove(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	store.Save(&Prompt{ID: "to-remove", Body: "remove me"})
	if !store.Exists("to-remove") {
		t.Fatal("expected prompt to exist")
	}

	// Verify file exists
	if _, err := os.Stat(filepath.Join(dir, "to-remove.md")); err != nil {
		t.Fatalf("file should exist: %v", err)
	}

	if err := store.Remove("to-remove"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	if store.Exists("to-remove") {
		t.Fatal("expected prompt to not exist after removal")
	}
}
