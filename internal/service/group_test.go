package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kimmykuang/selfskill/internal/linker"
)

func TestLoadGroup_LinksSkills(t *testing.T) {
	svc, root := testService(t)
	writeSkill(t, root, "alpha", "name: alpha\n")
	writeSkill(t, root, "beta", "name: beta\n")

	if err := svc.deps.Groups.Create("g1"); err != nil {
		t.Fatal(err)
	}
	if err := svc.deps.Groups.AddSkill("g1", "alpha"); err != nil {
		t.Fatal(err)
	}
	if err := svc.deps.Groups.AddSkill("g1", "beta"); err != nil {
		t.Fatal(err)
	}

	res, err := svc.LoadGroup("g1", linker.ScopeUser)
	if err != nil {
		t.Fatalf("LoadGroup: %v", err)
	}
	if len(res.Skills) != 2 {
		t.Fatalf("expected 2 skill results, got %d", len(res.Skills))
	}
	for _, r := range res.Skills {
		if r.Action != "linked" && r.Action != "skipped" {
			t.Errorf("unexpected action %q for %s", r.Action, r.SkillName)
		}
	}

	if !svc.IsSkillLoaded("alpha", linker.ScopeUser) {
		t.Error("alpha not loaded after LoadGroup")
	}
}

func TestLoadGroup_NotFound(t *testing.T) {
	svc, _ := testService(t)
	_, err := svc.LoadGroup("missing", linker.ScopeUser)
	if err == nil {
		t.Fatal("expected error on missing group")
	}
}

func TestUnloadGroup_RemovesSkillLinks(t *testing.T) {
	svc, root := testService(t)
	writeSkill(t, root, "alpha", "name: alpha\n")
	if err := svc.deps.Groups.Create("g1"); err != nil {
		t.Fatal(err)
	}
	if err := svc.deps.Groups.AddSkill("g1", "alpha"); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.LoadGroup("g1", linker.ScopeUser); err != nil {
		t.Fatal(err)
	}
	if !svc.IsSkillLoaded("alpha", linker.ScopeUser) {
		t.Fatal("setup: alpha should be loaded")
	}

	if _, err := svc.UnloadGroup("g1", linker.ScopeUser); err != nil {
		t.Fatalf("UnloadGroup: %v", err)
	}
	if svc.IsSkillLoaded("alpha", linker.ScopeUser) {
		t.Error("alpha still loaded after UnloadGroup")
	}

	linkPath := filepath.Join(root, "claude/skills", "alpha")
	if _, err := os.Lstat(linkPath); err == nil {
		t.Errorf("link path %s should be gone", linkPath)
	}
}
