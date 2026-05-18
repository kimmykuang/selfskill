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
	cmd := &cobra.Command{
		Use:   "add <group> <skill-or-plugin>",
		Short: "Add a skill or plugin to a group",
		Long: `Add a skill or plugin to a group.

Examples:
  ss group add mygroup video-summary          # add a skill
  ss group add mygroup --plugin superpowers   # add a plugin`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName := args[0]
			itemName := args[1]

			if isPlugin {
				if err := store.AddPlugin(groupName, itemName); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Added plugin %q to group %q\n", itemName, groupName)
			} else {
				if err := store.AddSkill(groupName, itemName); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Added skill %q to group %q\n", itemName, groupName)
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&isPlugin, "plugin", "p", false, "Add as a plugin instead of a skill")
	return cmd
}

func newGroupRemoveCmd(store *group.Store) *cobra.Command {
	var isPlugin bool
	cmd := &cobra.Command{
		Use:   "remove <group> <skill-or-plugin>",
		Short: "Remove a skill or plugin from a group",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName := args[0]
			itemName := args[1]

			if isPlugin {
				if err := store.RemovePlugin(groupName, itemName); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Removed plugin %q from group %q\n", itemName, groupName)
			} else {
				if err := store.RemoveSkill(groupName, itemName); err != nil {
					return err
				}
				fmt.Fprintf(os.Stdout, "Removed skill %q from group %q\n", itemName, groupName)
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&isPlugin, "plugin", "p", false, "Remove a plugin instead of a skill")
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
			fmt.Fprintf(w, "NAME\tSKILLS\tPLUGINS\n")
			for _, g := range groups {
				skills := strings.Join(g.Skills, ", ")
				if skills == "" {
					skills = "(none)"
				}
				plugins := strings.Join(g.Plugins, ", ")
				if plugins == "" {
					plugins = "(none)"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", g.Name, skills, plugins)
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
