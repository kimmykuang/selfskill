package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kimmykuang/selfskill/internal/plugin"
	"github.com/spf13/cobra"
)

func newPluginCmd(pluginStore *plugin.Store, registry *plugin.Registry, loader *plugin.Loader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage Claude Code plugins",
	}

	cmd.AddCommand(newPluginAddMarketplaceCmd(registry))
	cmd.AddCommand(newPluginListAvailableCmd(registry))
	cmd.AddCommand(newPluginInstallCmd(pluginStore, registry))
	cmd.AddCommand(newPluginLoadCmd(loader))
	cmd.AddCommand(newPluginUnloadCmd(loader))
	cmd.AddCommand(newPluginListCmd(pluginStore))
	cmd.AddCommand(newPluginSkillsCmd(pluginStore))
	cmd.AddCommand(newPluginUpdateCmd(pluginStore))
	cmd.AddCommand(newPluginRemoveCmd(pluginStore, loader))

	return cmd
}

func newPluginAddMarketplaceCmd(registry *plugin.Registry) *cobra.Command {
	return &cobra.Command{
		Use:   "add-marketplace <git-url>",
		Short: "Add a marketplace by cloning its git repo",
		Long: `Clone a marketplace repo to ~/ss/marketplaces/<name>/.
After adding, use 'ss plugin list-available' to see available plugins.

Examples:
  ss plugin add-marketplace https://github.com/anthropics/claude-plugins-official.git
  ss plugin add-marketplace https://github.com/anthropics/ai-minions-skills.git`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := registry.AddMarketplace(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Added marketplace %q\n", name)
			fmt.Fprintf(os.Stdout, "Run 'ss plugin list-available %s' to see available plugins.\n", name)
			return nil
		},
	}
}

func newPluginListAvailableCmd(registry *plugin.Registry) *cobra.Command {
	return &cobra.Command{
		Use:   "list-available [marketplace]",
		Short: "List plugins available for install from marketplaces",
		Long: `List plugins available in added marketplaces.
If a marketplace name is given, only show plugins from that marketplace.
Otherwise show plugins from all marketplaces.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var marketplaces []string
			if len(args) == 1 {
				marketplaces = []string{args[0]}
			} else {
				var err error
				marketplaces, err = registry.ListMarketplaces()
				if err != nil {
					return err
				}
			}

			if len(marketplaces) == 0 {
				fmt.Fprintf(os.Stdout, "No marketplaces added. Use 'ss plugin add-marketplace <git-url>' first.\n")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "PLUGIN\tMARKETPLACE\tDESCRIPTION\n")
			for _, mp := range marketplaces {
				mf, err := registry.LoadMarketplace(mp)
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: cannot read marketplace %q: %v\n", mp, err)
					continue
				}
				for _, p := range mf.Plugins {
					desc := p.Description
					if len(desc) > 60 {
						desc = desc[:57] + "..."
					}
					fmt.Fprintf(w, "%s\t%s\t%s\n", p.Name, mp, desc)
				}
			}
			return w.Flush()
		},
	}
}

func newPluginInstallCmd(store *plugin.Store, registry *plugin.Registry) *cobra.Command {
	return &cobra.Command{
		Use:   "install <plugin>@<marketplace>",
		Short: "Download a plugin to ~/ss/plugins/ (does not load)",
		Long: `Download a plugin from a marketplace into ss management.
The plugin is cloned from its git source but not loaded into CC.
Use 'ss plugin load' to activate it.

Examples:
  ss plugin install superpowers@claude-plugins-official
  ss plugin install bilibili-backend-platform@ai-minions-skills`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := args[0]
			pluginName, marketplace := parsePluginArg(input)
			if marketplace == "" {
				return fmt.Errorf("marketplace required: use format <plugin>@<marketplace>")
			}

			// Look up plugin in marketplace
			mp, err := registry.FindPlugin(pluginName, marketplace)
			if err != nil {
				return err
			}

			if registry.IsLocalPlugin(mp) {
				// Local plugin — copy from marketplace repo
				srcPath, err := registry.LocalPluginPath(mp, marketplace)
				if err != nil {
					return err
				}
				version := "latest"
				p, err := store.InstallFromLocal(srcPath, marketplace, pluginName, version)
				if err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Installed %s@%s (local) -> %s\n", pluginName, marketplace, p.InstallPath)
			} else {
				// Remote plugin — git clone
				gitURL, err := registry.ResolveGitURL(mp, marketplace)
				if err != nil {
					return err
				}
				// Use "latest" as version for now; could detect tags
				version := "latest"
				p, err := store.InstallFromGit(gitURL, marketplace, pluginName, version)
				if err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Installed %s@%s -> %s\n", pluginName, marketplace, p.InstallPath)
			}

			return nil
		},
	}
}

func newPluginLoadCmd(loader *plugin.Loader) *cobra.Command {
	return &cobra.Command{
		Use:   "load <plugin>",
		Short: "Load a plugin into Claude Code (create symlink + register)",
		Long: `Make a plugin visible to CC by creating a symlink in the CC cache
directory and adding an entry to installed_plugins.json.
The plugin must be installed first via 'ss plugin install'.

Examples:
  ss plugin load superpowers
  ss plugin load bilibili-backend-platform@ai-minions-skills`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loader.Load(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Loaded plugin %q into CC\n", args[0])
			return nil
		},
	}
}

func newPluginUnloadCmd(loader *plugin.Loader) *cobra.Command {
	return &cobra.Command{
		Use:   "unload <plugin>",
		Short: "Unload a plugin from Claude Code (remove symlink + deregister)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := loader.Unload(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Unloaded plugin %q from CC\n", args[0])
			return nil
		},
	}
}

func newPluginListCmd(store *plugin.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all plugins managed by ss",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			plugins, err := store.List()
			if err != nil {
				return err
			}

			if len(plugins) == 0 {
				fmt.Fprintf(os.Stdout, "No plugins installed. Use 'ss plugin install' to add one.\n")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "PLUGIN\tMARKETPLACE\tVERSION\tSTATE\n")
			for _, p := range plugins {
				state := "installed"
				if store.IsLoaded(p.FullName()) {
					state = "loaded"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Name, p.Marketplace, p.Version, state)
			}
			return w.Flush()
		},
	}
}

func newPluginSkillsCmd(store *plugin.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "skills <plugin>",
		Short: "List skills contained in a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skills, err := store.ListSkills(args[0])
			if err != nil {
				return err
			}

			if len(skills) == 0 {
				fmt.Fprintf(os.Stdout, "No skills found in plugin %q\n", args[0])
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "SKILL\n")
			for _, s := range skills {
				fmt.Fprintf(w, "%s\n", s)
			}
			return w.Flush()
		},
	}
}

func newPluginUpdateCmd(store *plugin.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "update <plugin>",
		Short: "Update a plugin to the latest version (git pull)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := store.Update(args[0]); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Updated plugin %q\n", args[0])
			return nil
		},
	}
}

func newPluginRemoveCmd(store *plugin.Store, _ *plugin.Loader) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <plugin>",
		Short: "Remove a plugin from ss (must be unloaded first)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Check if loaded
			if store.IsLoaded(name) {
				return fmt.Errorf("plugin %q is currently loaded; run 'ss plugin unload %s' first", name, name)
			}

			if err := store.Remove(name); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Removed plugin %q\n", name)
			return nil
		},
	}
}

func parsePluginArg(input string) (name, marketplace string) {
	for i := len(input) - 1; i >= 0; i-- {
		if input[i] == '@' {
			return input[:i], input[i+1:]
		}
	}
	return input, ""
}
