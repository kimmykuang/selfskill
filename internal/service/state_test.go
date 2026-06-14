package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kimmykuang/selfskill/internal/linker"
	"github.com/kimmykuang/selfskill/internal/plugin/cc_compat"
)

func TestIsSkillLoaded(t *testing.T) {
	svc, root := testService(t)
	writeSkill(t, root, "linked", "name: linked\n")
	writeSkill(t, root, "unlinked", "name: unlinked\n")

	// Load "linked" via the real linker into the user-scope tempdir.
	results, err := svc.deps.Linker.Load([]string{"linked"}, linker.ScopeUser)
	if err != nil {
		t.Fatalf("linker.Load: %v", err)
	}
	if len(results) == 0 || results[0].Action != "linked" {
		t.Fatalf("linker did not link: %+v", results)
	}

	if !svc.IsSkillLoaded("linked", linker.ScopeUser) {
		t.Errorf("linked should be reported loaded")
	}
	if svc.IsSkillLoaded("unlinked", linker.ScopeUser) {
		t.Errorf("unlinked should be reported not loaded")
	}
}

func TestIsPluginLoaded(t *testing.T) {
	svc, _ := testService(t)
	// Drop an entry directly into the fake registry.
	reg := svc.deps.Registry.(*cc_compat.FakeRegistry)
	if err := reg.Add("alpha@mp", cc_compat.Entry{Scope: "user", InstallPath: "/x", Version: "1"}); err != nil {
		t.Fatal(err)
	}

	if !svc.IsPluginLoaded("alpha@mp") {
		t.Error("alpha@mp should be loaded")
	}
	if svc.IsPluginLoaded("beta@mp") {
		t.Error("beta@mp should not be loaded")
	}
}

// Sanity check: IsSkillLoaded scoped lookup uses linker, not file existence.
func TestIsSkillLoaded_DoesNotPanicWithoutClaudeDir(t *testing.T) {
	root := t.TempDir()
	// Intentionally do NOT create the claude/skills dir.
	deps := Deps{
		Linker:   linker.New(filepath.Join(root, "no-such-dir")),
		Registry: cc_compat.NewFakeRegistry(),
	}
	svc := New(deps)

	// Should return false, not panic.
	if svc.IsSkillLoaded("anything", linker.ScopeUser) {
		t.Error("missing scope dir should report not loaded")
	}

	// And the file at the link path obviously is not there:
	if _, err := os.Stat(filepath.Join(root, "no-such-dir", "anything")); err == nil {
		t.Error("scope dir should not exist")
	}
}
