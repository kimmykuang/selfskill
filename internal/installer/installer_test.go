package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kimmykuang/selfskill/internal/prompt"
	"github.com/kimmykuang/selfskill/internal/skill"
)

func TestInstallSkill_LocalDir(t *testing.T) {
	skillsDir := t.TempDir()
	promptsDir := t.TempDir()
	ss := skill.NewStore(skillsDir)
	ps := prompt.NewStore(promptsDir)
	inst := New(ss, ps)

	// Create a source skill directory
	srcDir := t.TempDir()
	os.MkdirAll(filepath.Join(srcDir, "my-skill", "scripts"), 0755)
	os.WriteFile(filepath.Join(srcDir, "my-skill", "SKILL.md"), []byte(`---
name: my-skill
version: 1.0.0
---
# My Skill
`), 0644)
	os.WriteFile(filepath.Join(srcDir, "my-skill", "scripts", "run.sh"), []byte("#!/bin/bash"), 0755)

	results, err := inst.InstallSkill(filepath.Join(srcDir, "my-skill"), false)
	if err != nil {
		t.Fatalf("InstallSkill failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Action != "installed" {
		t.Errorf("expected installed, got %s (%s)", results[0].Action, results[0].Detail)
	}

	// Verify installed
	if !ss.Exists("my-skill") {
		t.Error("skill should exist after install")
	}
}

func TestInstallSkill_ParentDir(t *testing.T) {
	skillsDir := t.TempDir()
	promptsDir := t.TempDir()
	ss := skill.NewStore(skillsDir)
	ps := prompt.NewStore(promptsDir)
	inst := New(ss, ps)

	// Create parent with two skill subdirs
	srcDir := t.TempDir()
	os.MkdirAll(filepath.Join(srcDir, "skill-a"), 0755)
	os.WriteFile(filepath.Join(srcDir, "skill-a", "SKILL.md"), []byte("# A\n"), 0644)
	os.MkdirAll(filepath.Join(srcDir, "skill-b"), 0755)
	os.WriteFile(filepath.Join(srcDir, "skill-b", "SKILL.md"), []byte("# B\n"), 0644)
	// Non-skill dir should be ignored
	os.MkdirAll(filepath.Join(srcDir, "not-a-skill"), 0755)
	os.WriteFile(filepath.Join(srcDir, "not-a-skill", "README.md"), []byte("# Not a skill"), 0644)

	results, err := inst.InstallSkill(srcDir, false)
	if err != nil {
		t.Fatalf("InstallSkill failed: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestInstallSkill_AlreadyExists(t *testing.T) {
	skillsDir := t.TempDir()
	promptsDir := t.TempDir()
	ss := skill.NewStore(skillsDir)
	ps := prompt.NewStore(promptsDir)
	inst := New(ss, ps)

	// Pre-install
	ss.SaveFromContent("existing", []byte("# Existing\n"))

	srcDir := t.TempDir()
	os.MkdirAll(filepath.Join(srcDir, "existing"), 0755)
	os.WriteFile(filepath.Join(srcDir, "existing", "SKILL.md"), []byte("# New\n"), 0644)

	// Without force
	results, _ := inst.InstallSkill(filepath.Join(srcDir, "existing"), false)
	if results[0].Action != "error" {
		t.Errorf("expected error without force, got %s", results[0].Action)
	}

	// With force
	results, _ = inst.InstallSkill(filepath.Join(srcDir, "existing"), true)
	if results[0].Action != "installed" {
		t.Errorf("expected installed with force, got %s (%s)", results[0].Action, results[0].Detail)
	}
}

func TestInstallPrompt_LocalFile(t *testing.T) {
	skillsDir := t.TempDir()
	promptsDir := t.TempDir()
	ss := skill.NewStore(skillsDir)
	ps := prompt.NewStore(promptsDir)
	inst := New(ss, ps)

	// Create a prompt file
	srcFile := filepath.Join(t.TempDir(), "wuxia.md")
	os.WriteFile(srcFile, []byte(`---
id: wuxia-video
tags: [video, wuxia]
description: 武侠视频总结
---
请总结这个武侠视频的内容。
`), 0644)

	results, err := inst.InstallPrompt(srcFile, false)
	if err != nil {
		t.Fatalf("InstallPrompt failed: %v", err)
	}
	if len(results) != 1 || results[0].Action != "installed" {
		t.Errorf("unexpected result: %v", results)
	}
	if !ps.Exists("wuxia-video") {
		t.Error("prompt should exist after install")
	}
}

func TestResolveGitHubURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"https://github.com/user/repo/blob/main/skills/test.md",
			"https://raw.githubusercontent.com/user/repo/main/skills/test.md",
		},
		{
			"https://example.com/file.md",
			"https://example.com/file.md",
		},
	}

	for _, tt := range tests {
		got := resolveGitHubURL(tt.input)
		if got != tt.expected {
			t.Errorf("resolveGitHubURL(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
