package compare

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kimmykuang/selfskill/internal/plugin"
	"github.com/kimmykuang/selfskill/internal/skill"
)

func TestCompareSkill_NoSource(t *testing.T) {
	skillsDir := t.TempDir()
	skillName := "no-source"
	skDir := filepath.Join(skillsDir, skillName)
	if err := os.MkdirAll(skDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := []byte("---\nname: no-source\nversion: 1.0.0\ndescription: a skill\n---\n# body\n")
	if err := os.WriteFile(filepath.Join(skDir, "SKILL.md"), content, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	pluginStore := plugin.NewStore(t.TempDir())
	skillStore := skill.NewStore(skillsDir)
	c := New(pluginStore, skillStore)

	var buf bytes.Buffer
	err := c.CompareSkill(skillName, &buf)
	if err == nil {
		t.Fatal("expected error for skill without source, got nil")
	}
	if !strings.Contains(err.Error(), "no source recorded") {
		t.Fatalf("expected error to contain 'no source recorded', got %v", err)
	}
}

func TestCompare_NotFound(t *testing.T) {
	pluginStore := plugin.NewStore(t.TempDir())
	skillStore := skill.NewStore(t.TempDir())
	c := New(pluginStore, skillStore)

	err := c.Compare("ghost", io.Discard)
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
	if !strings.Contains(err.Error(), "no plugin or skill") {
		t.Fatalf("expected error to contain 'no plugin or skill', got %v", err)
	}
}
