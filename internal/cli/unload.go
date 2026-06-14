package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kimmykuang/selfskill/internal/linker"
	"github.com/kimmykuang/selfskill/internal/service"
	"github.com/spf13/cobra"
)

func newUnloadCmd(svc *service.Service, lnk *linker.Linker) *cobra.Command {
	var scope string
	var all bool

	cmd := &cobra.Command{
		Use:   "unload [group]",
		Short: "Unload a skill group or all skills from the active scope",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s := linker.Scope(scope)

			if all {
				results, err := lnk.UnloadAll(s)
				if err != nil {
					return err
				}
				return printSkillResults(os.Stdout, results)
			}

			if len(args) != 1 {
				return fmt.Errorf("unload requires a group name (or --all)")
			}
			report, err := svc.UnloadGroup(args[0], s)
			if err != nil {
				return err
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
	cmd.Flags().BoolVar(&all, "all", false, "Unload all ss-managed skills in scope")
	return cmd
}

func printSkillResults(w *os.File, results []linker.LinkResult) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "SKILL\tACTION\tDETAIL\n")
	for _, r := range results {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", r.SkillName, r.Action, r.Detail)
	}
	return tw.Flush()
}
