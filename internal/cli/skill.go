package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kimmykuang/selfskill/internal/skill"
	"github.com/spf13/cobra"
)

func newSkillCmd(store *skill.Store) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Manage installed skills",
	}

	cmd.AddCommand(newSkillListCmd(store))
	cmd.AddCommand(newSkillRemoveCmd(store))

	return cmd
}

func newSkillListCmd(store *skill.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all installed skills",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			skills, err := store.List()
			if err != nil {
				return err
			}

			if len(skills) == 0 {
				fmt.Fprintf(os.Stdout, "No skills installed.\n")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "NAME\tVERSION\tDESCRIPTION\n")
			for _, s := range skills {
				fmt.Fprintf(w, "%s\t%s\t%s\n", s.Name, s.Version, s.Description)
			}
			return w.Flush()
		},
	}
}

func newSkillRemoveCmd(store *skill.Store) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove an installed skill",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if !store.Exists(name) {
				return fmt.Errorf("skill %q not found", name)
			}

			if !force {
				fmt.Fprintf(os.Stdout, "Removing skill %q...\n", name)
			}

			if err := store.Remove(name); err != nil {
				return err
			}

			fmt.Fprintf(os.Stdout, "Removed skill %q\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")
	return cmd
}
