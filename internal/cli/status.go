package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kimmykuang/selfskill/internal/linker"
	"github.com/spf13/cobra"
)

func newStatusCmd(lnk *linker.Linker) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show currently loaded skills in all scopes",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			status, err := lnk.FullStatus()
			if err != nil {
				return err
			}

			if len(status.UserScope) == 0 && len(status.ProjectScope) == 0 && len(status.Plugins) == 0 {
				fmt.Fprintf(os.Stdout, "No skills found.\n")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

			if len(status.UserScope) > 0 {
				fmt.Fprintf(w, "USER SCOPE (~/.claude/skills/):\n")
				fmt.Fprintf(w, "  NAME\tSOURCE\tPATH\n")
				for _, e := range status.UserScope {
					fmt.Fprintf(w, "  %s\t[%s]\t%s\n", e.Name, e.Source, e.Path)
				}
			}

			if len(status.ProjectScope) > 0 {
				if len(status.UserScope) > 0 {
					fmt.Fprintf(w, "\n")
				}
				fmt.Fprintf(w, "PROJECT SCOPE (./.claude/skills/):\n")
				fmt.Fprintf(w, "  NAME\tSOURCE\tPATH\n")
				for _, e := range status.ProjectScope {
					fmt.Fprintf(w, "  %s\t[%s]\t%s\n", e.Name, e.Source, e.Path)
				}
			}

			if len(status.Plugins) > 0 {
				if len(status.UserScope) > 0 || len(status.ProjectScope) > 0 {
					fmt.Fprintf(w, "\n")
				}
				fmt.Fprintf(w, "PLUGINS (~/.claude/plugins/):\n")
				fmt.Fprintf(w, "  NAME\tPATH\n")
				for _, e := range status.Plugins {
					fmt.Fprintf(w, "  %s\t%s\n", e.Name, e.Path)
				}
			}

			return w.Flush()
		},
	}
}
