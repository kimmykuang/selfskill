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

func newUnloadCmd(groupStore *group.Store, lnk *linker.Linker, pluginLoader *plugin.Loader) *cobra.Command {
	var (
		scope string
		all   bool
	)

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

				if len(results) == 0 {
					fmt.Fprintf(os.Stdout, "No skills loaded in %s scope.\n", scope)
					return nil
				}

				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintf(w, "SKILL\tACTION\tDETAIL\n")
				for _, r := range results {
					fmt.Fprintf(w, "%s\t%s\t%s\n", r.SkillName, r.Action, r.Detail)
				}
				return w.Flush()
			}

			if len(args) == 0 {
				return fmt.Errorf("provide a group name or use --all")
			}

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

			// Unload skills
			if len(g.Skills) > 0 {
				results, err := lnk.Unload(g.Skills, s)
				if err != nil {
					return err
				}
				for _, r := range results {
					fmt.Fprintf(w, "%s\tskill\t%s\t%s\n", r.SkillName, r.Action, r.Detail)
				}
			}

			// Unload plugins
			for _, pName := range g.Plugins {
				if err := pluginLoader.Unload(pName); err != nil {
					fmt.Fprintf(w, "%s\tplugin\terror\t%s\n", pName, err.Error())
				} else {
					fmt.Fprintf(w, "%s\tplugin\tunloaded\t\n", pName)
				}
			}

			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "user", "Target scope for skills: user or project")
	cmd.Flags().BoolVar(&all, "all", false, "Unload all skills from the scope")
	return cmd
}
