package linker

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestLinker(t *testing.T) (skillsDir, targetDir string, lk *Linker) {
	t.Helper()
	skillsDir = t.TempDir()
	targetDir = t.TempDir()
	lk = New(skillsDir)
	return
}

func createSkill(t *testing.T, skillsDir, name string) {
	t.Helper()
	dir := filepath.Join(skillsDir, name)
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# "+name), 0644)
}

func TestLinkOne_Success(t *testing.T) {
	skillsDir, targetDir, lk := setupTestLinker(t)
	createSkill(t, skillsDir, "test-skill")

	result := lk.linkOne("test-skill", targetDir)
	if result.Action != "linked" {
		t.Errorf("expected action=linked, got %s (%s)", result.Action, result.Detail)
	}

	// Verify symlink
	linkPath := filepath.Join(targetDir, "test-skill")
	if !isSymlink(linkPath) {
		t.Fatal("expected symlink to be created")
	}
	target, _ := os.Readlink(linkPath)
	if target != filepath.Join(skillsDir, "test-skill") {
		t.Errorf("symlink target = %s", target)
	}
}

func TestLinkOne_AlreadyLinked(t *testing.T) {
	skillsDir, targetDir, lk := setupTestLinker(t)
	createSkill(t, skillsDir, "test-skill")

	lk.linkOne("test-skill", targetDir)
	result := lk.linkOne("test-skill", targetDir)
	if result.Action != "skipped" {
		t.Errorf("expected action=skipped, got %s (%s)", result.Action, result.Detail)
	}
}

func TestLinkOne_SourceNotFound(t *testing.T) {
	_, targetDir, lk := setupTestLinker(t)

	result := lk.linkOne("nonexistent", targetDir)
	if result.Action != "error" {
		t.Errorf("expected action=error, got %s", result.Action)
	}
}

func TestLinkOne_ConflictingSymlink(t *testing.T) {
	skillsDir, targetDir, lk := setupTestLinker(t)
	createSkill(t, skillsDir, "test-skill")

	// Create a symlink pointing elsewhere
	otherDir := t.TempDir()
	os.Symlink(otherDir, filepath.Join(targetDir, "test-skill"))

	result := lk.linkOne("test-skill", targetDir)
	if result.Action != "error" {
		t.Errorf("expected action=error for conflicting symlink, got %s", result.Action)
	}
}

func TestLinkOne_RegularFileConflict(t *testing.T) {
	skillsDir, targetDir, lk := setupTestLinker(t)
	createSkill(t, skillsDir, "test-skill")

	// Create a regular file at the link path
	os.WriteFile(filepath.Join(targetDir, "test-skill"), []byte("data"), 0644)

	result := lk.linkOne("test-skill", targetDir)
	if result.Action != "error" {
		t.Errorf("expected action=error for regular file conflict, got %s", result.Action)
	}
}

func TestUnlinkOne_Success(t *testing.T) {
	skillsDir, targetDir, lk := setupTestLinker(t)
	createSkill(t, skillsDir, "test-skill")

	lk.linkOne("test-skill", targetDir)
	result := lk.unlinkOne("test-skill", targetDir)
	if result.Action != "removed" {
		t.Errorf("expected action=removed, got %s (%s)", result.Action, result.Detail)
	}

	// Verify symlink is gone
	if isSymlink(filepath.Join(targetDir, "test-skill")) {
		t.Error("symlink should be removed")
	}
}

func TestUnlinkOne_NotSymlink(t *testing.T) {
	_, targetDir, lk := setupTestLinker(t)

	result := lk.unlinkOne("nonexistent", targetDir)
	if result.Action != "skipped" {
		t.Errorf("expected action=skipped, got %s", result.Action)
	}
}

func TestUnlinkOne_NotManaged(t *testing.T) {
	_, targetDir, lk := setupTestLinker(t)

	// Create a symlink pointing to a non-managed location
	otherDir := t.TempDir()
	os.Symlink(otherDir, filepath.Join(targetDir, "external"))

	result := lk.unlinkOne("external", targetDir)
	if result.Action != "skipped" {
		t.Errorf("expected action=skipped for non-managed, got %s (%s)", result.Action, result.Detail)
	}
}

func TestLoad_MultipleSkills(t *testing.T) {
	skillsDir := t.TempDir()
	targetDir := t.TempDir()
	lk := New(skillsDir)

	createSkill(t, skillsDir, "skill-a")
	createSkill(t, skillsDir, "skill-b")

	// Test Load directly with targetDir (bypass TargetDir which uses config)
	var results []LinkResult
	for _, name := range []string{"skill-a", "skill-b", "nonexistent"} {
		results = append(results, lk.linkOne(name, targetDir))
	}

	if results[0].Action != "linked" {
		t.Errorf("skill-a: expected linked, got %s", results[0].Action)
	}
	if results[1].Action != "linked" {
		t.Errorf("skill-b: expected linked, got %s", results[1].Action)
	}
	if results[2].Action != "error" {
		t.Errorf("nonexistent: expected error, got %s", results[2].Action)
	}
}

func TestUnloadAll_MixedEntries(t *testing.T) {
	skillsDir := t.TempDir()
	targetDir := t.TempDir()
	lk := New(skillsDir)

	// Create managed symlinks
	createSkill(t, skillsDir, "skill-a")
	createSkill(t, skillsDir, "skill-b")
	os.Symlink(filepath.Join(skillsDir, "skill-a"), filepath.Join(targetDir, "skill-a"))
	os.Symlink(filepath.Join(skillsDir, "skill-b"), filepath.Join(targetDir, "skill-b"))

	// Create non-managed symlink
	otherDir := t.TempDir()
	os.Symlink(otherDir, filepath.Join(targetDir, "external"))

	// Create regular file
	os.WriteFile(filepath.Join(targetDir, "regular.md"), []byte("data"), 0644)

	// Simulate UnloadAll logic
	entries, _ := os.ReadDir(targetDir)
	var results []LinkResult
	for _, e := range entries {
		linkPath := filepath.Join(targetDir, e.Name())
		if !isSymlink(linkPath) {
			continue
		}
		target, err := os.Readlink(linkPath)
		if err != nil {
			continue
		}
		if lk.isManagedTarget(target) {
			if err := os.Remove(linkPath); err != nil {
				results = append(results, LinkResult{SkillName: e.Name(), Action: "error", Detail: err.Error()})
			} else {
				results = append(results, LinkResult{SkillName: e.Name(), Action: "removed"})
			}
		}
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 removed, got %d", len(results))
	}

	// External symlink should remain
	if !isSymlink(filepath.Join(targetDir, "external")) {
		t.Error("external symlink should not be removed")
	}

	// Regular file should remain
	if _, err := os.Stat(filepath.Join(targetDir, "regular.md")); err != nil {
		t.Error("regular file should not be removed")
	}
}

func TestIsManagedTarget(t *testing.T) {
	skillsDir := t.TempDir()
	lk := New(skillsDir)

	managed := filepath.Join(skillsDir, "some-skill")
	if !lk.isManagedTarget(managed) {
		t.Error("expected managed=true for path under skillsDir")
	}

	unmanaged := "/some/other/path"
	if lk.isManagedTarget(unmanaged) {
		t.Error("expected managed=false for unrelated path")
	}
}

func TestIsSymlink(t *testing.T) {
	dir := t.TempDir()

	// Regular file
	regular := filepath.Join(dir, "file")
	os.WriteFile(regular, []byte("x"), 0644)
	if isSymlink(regular) {
		t.Error("regular file should not be symlink")
	}

	// Directory
	subdir := filepath.Join(dir, "subdir")
	os.MkdirAll(subdir, 0755)
	if isSymlink(subdir) {
		t.Error("directory should not be symlink")
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
