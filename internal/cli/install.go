package cli

import (
	"fmt"
	"os"

	"github.com/kimmykuang/selfskill/internal/installer"
	"github.com/spf13/cobra"
)

func newInstallCmd(inst *installer.Installer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install skills or prompts from a source",
	}

	cmd.AddCommand(newInstallSkillCmd(inst))
	cmd.AddCommand(newInstallPromptCmd(inst))

	return cmd
}

func newInstallSkillCmd(inst *installer.Installer) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "skill <source>",
		Short: "Install a skill from a local path or URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			source := args[0]
			results, err := inst.InstallSkill(source, force)
			if err != nil {
				return err
			}

			for _, r := range results {
				switch r.Action {
				case "installed":
					fmt.Fprintf(os.Stdout, "Installed skill %q -> %s\n", r.Name, r.Path)
				case "error":
					fmt.Fprintf(os.Stdout, "Error installing %q: %s\n", r.Name, r.Detail)
				default:
					fmt.Fprintf(os.Stdout, "%s: %s %s\n", r.Name, r.Action, r.Detail)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing skill")
	return cmd
}

func newInstallPromptCmd(inst *installer.Installer) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "prompt <source>",
		Short: "Install a prompt from a local path or URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			source := args[0]
			results, err := inst.InstallPrompt(source, force)
			if err != nil {
				return err
			}

			for _, r := range results {
				switch r.Action {
				case "installed":
					fmt.Fprintf(os.Stdout, "Installed prompt %q -> %s\n", r.Name, r.Path)
				case "error":
					fmt.Fprintf(os.Stdout, "Error installing %q: %s\n", r.Name, r.Detail)
				default:
					fmt.Fprintf(os.Stdout, "%s: %s %s\n", r.Name, r.Action, r.Detail)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing prompt")
	return cmd
}
