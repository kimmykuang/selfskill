package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/kimmykuang/selfskill/internal/group"
	"github.com/spf13/cobra"
)

func newGroupCmd(store *group.Store) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Manage skill groups",
	}

	cmd.AddCommand(newGroupCreateCmd(store))
	cmd.AddCommand(newGroupAddCmd(store))
	cmd.AddCommand(newGroupRemoveCmd(store))
	cmd.AddCommand(newGroupListCmd(store))
	cmd.AddCommand(newGroupDeleteCmd(store))

	return cmd
}

func newGroupCreateCmd(store *group.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new skill group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := store.Create(name); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Created group %q\n", name)
			return nil
		},
	}
}

func newGroupAddCmd(store *group.Store) *cobra.Command {
	var isPlugin bool
	var isPrompt bool
	cmd := &cobra.Command{
		Use:   "add <group> <skill-or-plugin-or-prompt>",
		Short: "Add a skill, plugin, or prompt to a group",
		Long: `Add a skill, plugin, or prompt to a group.

Examples:
  ss group add mygroup video-summary          # add a skill
  ss group add mygroup --plugin superpowers   # add a plugin
  ss group add mygroup --prompt wuxia         # add a prompt`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName := args[0]
			itemName := args[1]

			if isPlugin && isPrompt {
				return fmt.Errorf("--plugin and --prompt are mutually exclusive")
			}

			switch {
			case isPlugin:
				if err := store.AddPlugin(groupName, itemName); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Added plugin %q to group %q\n", itemName, groupName)
			case isPrompt:
				if err := store.AddPrompt(groupName, itemName); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Added prompt %q to group %q\n", itemName, groupName)
			default:
				if err := store.AddSkill(groupName, itemName); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Added skill %q to group %q\n", itemName, groupName)
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&isPlugin, "plugin", "p", false, "Add as a plugin instead of a skill")
	cmd.Flags().BoolVar(&isPrompt, "prompt", false, "Add as a prompt instead of a skill")
	return cmd
}

func newGroupRemoveCmd(store *group.Store) *cobra.Command {
	var isPlugin bool
	var isPrompt bool
	cmd := &cobra.Command{
		Use:   "remove <group> <skill-or-plugin-or-prompt>",
		Short: "Remove a skill, plugin, or prompt from a group",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName := args[0]
			itemName := args[1]

			if isPlugin && isPrompt {
				return fmt.Errorf("--plugin and --prompt are mutually exclusive")
			}

			switch {
			case isPlugin:
				if err := store.RemovePlugin(groupName, itemName); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Removed plugin %q from group %q\n", itemName, groupName)
			case isPrompt:
				if err := store.RemovePrompt(groupName, itemName); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Removed prompt %q from group %q\n", itemName, groupName)
			default:
				if err := store.RemoveSkill(groupName, itemName); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Removed skill %q from group %q\n", itemName, groupName)
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&isPlugin, "plugin", "p", false, "Remove a plugin instead of a skill")
	cmd.Flags().BoolVar(&isPrompt, "prompt", false, "Remove a prompt instead of a skill")
	return cmd
}

func newGroupListCmd(store *group.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all skill groups",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			groups, err := store.List()
			if err != nil {
				return err
			}

			if len(groups) == 0 {
				fmt.Fprintf(os.Stdout, "No groups defined.\n")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "NAME\tSKILLS\tPLUGINS\tPROMPTS\n")
			for _, g := range groups {
				skills := strings.Join(g.Skills, ", ")
				if skills == "" {
					skills = "(none)"
				}
				plugins := strings.Join(g.Plugins, ", ")
				if plugins == "" {
					plugins = "(none)"
				}
				prompts := strings.Join(g.Prompts, ", ")
				if prompts == "" {
					prompts = "(none)"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", g.Name, skills, plugins, prompts)
			}
			return w.Flush()
		},
	}
}

func newGroupDeleteCmd(store *group.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a skill group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := store.Delete(name); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Deleted group %q\n", name)
			return nil
		},
	}
}
