package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rsbh/personal-context-manager/internal/skill"
)

func newSkillCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Manage the Claude Code skill",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "install",
			Short: "Install the skill into ~/.claude/skills/pcm",
			Args:  noArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				home, err := a.homeDir()
				if err != nil {
					return err
				}
				file, err := skill.Install(home)
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "skill/pcm installed in %s\n", file)
				return nil
			},
		},
		&cobra.Command{
			Use:   "uninstall",
			Short: "Remove the skill from ~/.claude/skills/pcm",
			Args:  noArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				home, err := a.homeDir()
				if err != nil {
					return err
				}
				removed, err := skill.Uninstall(home)
				if err != nil {
					return err
				}
				if !removed {
					fmt.Fprintln(cmd.OutOrStdout(), "skill/pcm not installed")
					return nil
				}
				fmt.Fprintf(cmd.OutOrStdout(), "skill/pcm uninstalled from %s\n", skill.Dir(home))
				return nil
			},
		},
		&cobra.Command{
			Use:   "show",
			Short: "Print the skill",
			Args:  noArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				_, err := cmd.OutOrStdout().Write(skill.Content())
				return err
			},
		},
	)
	return cmd
}
