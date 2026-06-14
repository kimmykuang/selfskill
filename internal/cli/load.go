package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kimmykuang/selfskill/internal/linker"
	"github.com/kimmykuang/selfskill/internal/service"
	"github.com/spf13/cobra"
)

func newLoadCmd(svc *service.Service) *cobra.Command {
	var scope string

	cmd := &cobra.Command{
		Use:   "load <group>",
		Short: "Load a skill group into the active scope",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			report, err := svc.LoadGroup(args[0], linker.Scope(scope))
			if err != nil {
				return err
			}
			if len(report.Skills) == 0 && len(report.Plugins) == 0 {
				fmt.Fprintf(os.Stdout, "Group %q is empty.\n", args[0])
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "ITEM\tTYPE\tACTION\tDETAIL\n")
			for _, r := range report.Skills {
				fmt.Fprintf(w, "%s\tskill\t%s\t%s\n", r.SkillName, r.Action, r.Detail)
			}
			for _, r := range report.Plugins {
				fmt.Fprintf(w, "%s\tplugin\t%s\t%s\n", r.Name, r.Action, r.Detail)
			}
			return w.Flush()
		},
	}
	cmd.Flags().StringVar(&scope, "scope", "user", "Target scope for skills: user or project")
	return cmd
}
