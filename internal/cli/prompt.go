package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/kimmykuang/selfskill/internal/prompt"
	"github.com/spf13/cobra"
)

func newPromptCmd(store *prompt.Store) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prompt",
		Short: "Manage installed prompts",
	}

	cmd.AddCommand(newPromptListCmd(store))
	cmd.AddCommand(newPromptRemoveCmd(store))

	return cmd
}

func newPromptListCmd(store *prompt.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all installed prompts",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			prompts, err := store.List()
			if err != nil {
				return err
			}

			if len(prompts) == 0 {
				fmt.Fprintf(os.Stdout, "No prompts installed.\n")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "ID\tTAGS\tDESCRIPTION\n")
			for _, p := range prompts {
				tags := strings.Join(p.Tags, ", ")
				fmt.Fprintf(w, "%s\t%s\t%s\n", p.ID, tags, p.Description)
			}
			return w.Flush()
		},
	}
}

func newPromptRemoveCmd(store *prompt.Store) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <id>",
		Short: "Remove an installed prompt",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			if !store.Exists(id) {
				return fmt.Errorf("prompt %q not found", id)
			}

			if err := store.Remove(id); err != nil {
				return err
			}

			fmt.Fprintf(os.Stdout, "Removed prompt %q\n", id)
			return nil
		},
	}
}
