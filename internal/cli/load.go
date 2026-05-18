package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kimmykuang/selfskill/internal/group"
	"github.com/kimmykuang/selfskill/internal/linker"
	"github.com/kimmykuang/selfskill/internal/plugin"
	"github.com/spf13/cobra"
)

func newLoadCmd(groupStore *group.Store, lnk *linker.Linker, pluginLoader *plugin.Loader) *cobra.Command {
	var scope string

	cmd := &cobra.Command{
		Use:   "load <group>",
		Short: "Load a skill group into the active scope",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName := args[0]

			g, err := groupStore.Get(groupName)
			if err != nil {
				return err
			}

			if len(g.Skills) == 0 && len(g.Plugins) == 0 {
				fmt.Fprintf(os.Stdout, "Group %q is empty.\n", groupName)
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "ITEM\tTYPE\tACTION\tDETAIL\n")

			// Load skills via linker
			if len(g.Skills) > 0 {
				s := linker.Scope(scope)
				results, err := lnk.Load(g.Skills, s)
				if err != nil {
					return err
				}
				for _, r := range results {
					fmt.Fprintf(w, "%s\tskill\t%s\t%s\n", r.SkillName, r.Action, r.Detail)
				}
			}

			// Load plugins via plugin loader
			for _, pName := range g.Plugins {
				if err := pluginLoader.Load(pName); err != nil {
					fmt.Fprintf(w, "%s\tplugin\terror\t%s\n", pName, err.Error())
				} else {
					fmt.Fprintf(w, "%s\tplugin\tloaded\t\n", pName)
				}
			}

			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "user", "Target scope for skills: user or project")
	return cmd
}
