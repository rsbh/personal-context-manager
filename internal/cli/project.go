package cli

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/rsbh/personal-context-manager/internal/client"
)

func newGetProjectCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:     "project [name]",
		Aliases: []string{"projects"},
		Short:   "List projects, or show one project",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var names []string
			if len(args) == 1 {
				name, err := a.client.GetProject(cmd.Context(), args[0])
				if err != nil {
					return a.handleErr(cmd, err)
				}
				names = []string{name}
			} else {
				var err error
				if names, err = a.client.ListProjects(cmd.Context()); err != nil {
					return a.handleErr(cmd, err)
				}
			}

			out := cmd.OutOrStdout()
			if len(names) == 0 {
				fmt.Fprintln(out, "No projects found.")
				return nil
			}
			rows := make([][]string, len(names))
			for i, n := range names {
				rows[i] = []string{n}
			}
			printTable(out, []string{"NAME"}, rows)
			return nil
		},
	}
}

func newCreateProjectCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "project <name>",
		Short: "Create a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := a.client.CreateProject(cmd.Context(), args[0]); err != nil {
				return a.handleErr(cmd, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "project/%s created\n", args[0])
			return nil
		},
	}
}

func newDeleteProjectCmd(a *app) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "project <name>",
		Short: "Delete a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := a.client.DeleteProject(cmd.Context(), args[0], force)
			var apiErr *client.APIError
			if errors.As(err, &apiErr) && apiErr.Status == http.StatusConflict {
				printFields(cmd.ErrOrStderr(), []field{
					{"error", apiErr.Message},
					{"hint", "use --force to delete the project and its contents"},
				})
				return errSilent
			}
			if err != nil {
				return a.handleErr(cmd, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "project %q deleted\n", args[0])
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "delete the project and its contents")
	return cmd
}
