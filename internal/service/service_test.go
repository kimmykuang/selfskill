package service

import (
	"testing"

	"github.com/kimmykuang/selfskill/internal/group"
	"github.com/kimmykuang/selfskill/internal/linker"
	"github.com/kimmykuang/selfskill/internal/plugin"
	"github.com/kimmykuang/selfskill/internal/plugin/cc_compat"
	"github.com/kimmykuang/selfskill/internal/prompt"
	"github.com/kimmykuang/selfskill/internal/skill"
)

func TestNew_ReturnsNonNil(t *testing.T) {
	dir := t.TempDir()
	svc := New(Deps{
		Skills:       skill.NewStore(dir + "/skills"),
		Prompts:      prompt.NewStore(dir + "/prompts"),
		Groups:       group.NewStore(dir + "/groups"),
		Plugins:      plugin.NewStore(dir + "/plugins"),
		Linker:       linker.New(dir + "/skills"),
		PluginLoader: plugin.NewLoaderWithRegistry(plugin.NewStore(dir+"/plugins"), cc_compat.NewFakeRegistry()),
		Registry:     cc_compat.NewFakeRegistry(),
	})
	if svc == nil {
		t.Fatal("New returned nil")
	}
}
