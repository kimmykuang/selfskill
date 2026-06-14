package cli

import (
	"github.com/kimmykuang/selfskill/internal/compare"
	"github.com/spf13/cobra"
)

func newCompareCmd(c *compare.Comparer) *cobra.Command {
	var asPlugin, asSkill bool
	cmd := &cobra.Command{
		Use:   "compare <name>",
		Short: "Compare a locally installed plugin/skill against its remote source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch {
			case asPlugin:
				return c.ComparePlugin(args[0], cmd.OutOrStdout())
			case asSkill:
				return c.CompareSkill(args[0], cmd.OutOrStdout())
			default:
				return c.Compare(args[0], cmd.OutOrStdout())
			}
		},
	}
	cmd.Flags().BoolVar(&asPlugin, "plugin", false, "Compare as plugin")
	cmd.Flags().BoolVar(&asSkill, "skill", false, "Compare as skill")
	return cmd
}
