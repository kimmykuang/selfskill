package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kimmykuang/selfskill/internal/group"
	"github.com/kimmykuang/selfskill/internal/linker"
	"github.com/kimmykuang/selfskill/internal/plugin"
	"github.com/kimmykuang/selfskill/internal/plugin/cc_compat"
	"github.com/kimmykuang/selfskill/internal/prompt"
	"github.com/kimmykuang/selfskill/internal/skill"
)

// helper for tests
func testService(t *testing.T) (*Service, string) {
	t.Helper()
	root := t.TempDir()
	for _, sub := range []string{"skills", "prompts", "groups", "plugins", "claude/skills"} {
		if err := os.MkdirAll(filepath.Join(root, sub), 0755); err != nil {
			t.Fatal(err)
		}
	}
	pluginStore := plugin.NewStore(filepath.Join(root, "plugins"))
	reg := cc_compat.NewFakeRegistry()
	deps := Deps{
		Skills:       skill.NewStore(filepath.Join(root, "skills")),
		Prompts:      prompt.NewStore(filepath.Join(root, "prompts")),
		Groups:       group.NewStore(filepath.Join(root, "groups")),
		Plugins:      pluginStore,
		Linker:       linker.New(filepath.Join(root, "claude/skills")),
		PluginLoader: plugin.NewLoaderWithRegistry(pluginStore, reg),
		Registry:     reg,
	}
	return New(deps), root
}

func writeSkill(t *testing.T, root, name, frontmatter string) {
	t.Helper()
	dir := filepath.Join(root, "skills", name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	body := "---\n" + frontmatter + "---\n\n# " + name + "\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestListSkills_NoFilter(t *testing.T) {
	svc, root := testService(t)
	writeSkill(t, root, "alpha", "name: alpha\ndescription: first\n")
	writeSkill(t, root, "beta", "name: beta\ndescription: second\nsource: https://example.com/x.md\n")

	skills, err := svc.ListSkills(SkillFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 2 {
		t.Fatalf("got %d skills, want 2", len(skills))
	}
}

func TestListSkills_FilterByOrigin(t *testing.T) {
	svc, root := testService(t)
	writeSkill(t, root, "local-one", "name: local-one\n")
	writeSkill(t, root, "remote-one", "name: remote-one\nsource: https://example.com/r.md\n")

	got, err := svc.ListSkills(SkillFilter{Origin: OriginLocal})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "local-one" {
		t.Fatalf("local filter returned %+v", got)
	}

	got, err = svc.ListSkills(SkillFilter{Origin: OriginGitHub})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "remote-one" {
		t.Fatalf("github filter returned %+v", got)
	}
}

func TestListSkills_FilterByKeyword(t *testing.T) {
	svc, root := testService(t)
	writeSkill(t, root, "video-summary", "name: video-summary\ndescription: summarize videos\n")
	writeSkill(t, root, "code-review", "name: code-review\ndescription: review code\n")

	got, err := svc.ListSkills(SkillFilter{Query: "video"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "video-summary" {
		t.Fatalf("keyword filter returned %+v", got)
	}
}

func writePrompt(t *testing.T, root, id, frontmatter string) {
	t.Helper()
	body := "---\n" + frontmatter + "---\n\nbody for " + id + "\n"
	path := filepath.Join(root, "prompts", id+".md")
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestListPrompts_FilterByTag(t *testing.T) {
	svc, root := testService(t)
	writePrompt(t, root, "wuxia", "id: wuxia\ndescription: martial arts\ntags: [视频, 武侠]\n")
	writePrompt(t, root, "tech", "id: tech\ndescription: technology\ntags: [视频, 技术]\n")
	writePrompt(t, root, "other", "id: other\ndescription: misc\n")

	got, err := svc.ListPrompts(PromptFilter{Tags: []string{"武侠"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "wuxia" {
		t.Fatalf("tag filter returned %+v", got)
	}
}

func TestListPrompts_FilterByQuery(t *testing.T) {
	svc, root := testService(t)
	writePrompt(t, root, "alpha", "id: alpha\ndescription: about cats\n")
	writePrompt(t, root, "beta", "id: beta\ndescription: about dogs\n")

	got, err := svc.ListPrompts(PromptFilter{Query: "cats"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "alpha" {
		t.Fatalf("query filter returned %+v", got)
	}
}

func TestListPlugins_StateAware(t *testing.T) {
	svc, root := testService(t)
	// Manufacture two plugin install dirs.
	for _, name := range []string{"alpha", "beta"} {
		dir := filepath.Join(root, "plugins", "mp", name, "1.0.0")
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "plugin.json"), []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	all, err := svc.ListPlugins(PluginFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("got %d plugins, want 2", len(all))
	}
	// Both should be "installed" (not loaded). State is set by the service.
	for _, p := range all {
		if p.State != PluginStateInstalled {
			t.Errorf("plugin %s state = %q, want installed", p.Name, p.State)
		}
	}
}
