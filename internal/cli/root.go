package cli

import (
	"io/fs"

	"github.com/kimmykuang/selfskill/internal/compare"
	"github.com/kimmykuang/selfskill/internal/config"
	"github.com/kimmykuang/selfskill/internal/group"
	"github.com/kimmykuang/selfskill/internal/installer"
	"github.com/kimmykuang/selfskill/internal/linker"
	"github.com/kimmykuang/selfskill/internal/plugin"
	"github.com/kimmykuang/selfskill/internal/prompt"
	"github.com/kimmykuang/selfskill/internal/skill"
	"github.com/spf13/cobra"
)

func NewRootCmd(version string, staticFS fs.FS) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ss",
		Short:   "Skills & Prompt management tool for Claude Code",
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return config.EnsureDirs()
		},
	}

	// Initialize stores
	skillStore := skill.NewStore(config.SkillsDir())
	promptStore := prompt.NewStore(config.PromptsDir())
	groupStore := group.NewStore(config.GroupsDir())
	pluginStore := plugin.NewStore(config.PluginsDir())

	// Initialize services
	inst := installer.New(skillStore, promptStore)
	lnk := linker.New(config.SkillsDir())
	registry := plugin.NewRegistry()
	pluginLoader := plugin.NewLoader(pluginStore)
	cmpr := compare.New(pluginStore, skillStore)

	// Register commands
	cmd.AddCommand(newInstallCmd(inst))
	cmd.AddCommand(newSkillCmd(skillStore))
	cmd.AddCommand(newPromptCmd(promptStore))
	cmd.AddCommand(newGroupCmd(groupStore))
	cmd.AddCommand(newLoadCmd(groupStore, lnk, pluginLoader))
	cmd.AddCommand(newUnloadCmd(groupStore, lnk, pluginLoader))
	cmd.AddCommand(newStatusCmd(lnk))
	cmd.AddCommand(newPluginCmd(pluginStore, registry, pluginLoader))
	cmd.AddCommand(newCompareCmd(cmpr))
	cmd.AddCommand(newWebCmd(skillStore, promptStore, pluginStore, pluginLoader, lnk, staticFS))

	return cmd
}
