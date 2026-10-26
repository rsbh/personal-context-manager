package cli

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/rsbh/personal-context-manager/internal/client"
)

func newGetSectionCmd(a *app) *cobra.Command {
	var flags scopeFlags
	cmd := &cobra.Command{
		Use:     "section",
		Aliases: []string{"sections"},
		Short:   "List sections in a scope",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			sc, err := a.scope(cmd, flags)
			if err != nil {
				return err
			}
			scopeName, project := sc.api()
			names, err := a.client.ListSections(cmd.Context(), scopeName, project)
			var apiErr *client.APIError
			// A detected project that does not exist yet has no sections.
			if sc.detected && errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound {
				names, err = nil, nil
			}
			if err != nil {
				return a.handleErr(cmd, err)
			}

			out := cmd.OutOrStdout()
			if len(names) == 0 {
				fmt.Fprintln(out, "No sections found.")
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
	flags.register(cmd)
	return cmd
}

func newCreateSectionCmd(a *app) *cobra.Command {
	var flags scopeFlags
	cmd := &cobra.Command{
		Use:   "section <name>",
		Short: "Create a section in a scope",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sc, err := a.scope(cmd, flags)
			if err != nil {
				return err
			}
			if sc.detected {
				if err := a.ensureProject(cmd, sc.project); err != nil {
					return err
				}
			}
			scopeName, project := sc.api()
			if err := a.client.CreateSection(cmd.Context(), scopeName, project, args[0]); err != nil {
				return a.handleErr(cmd, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "section/%s created in %s\n", args[0], sc)
			return nil
		},
	}
	flags.register(cmd)
	return cmd
}

func newDeleteSectionCmd(a *app) *cobra.Command {
	var (
		flags scopeFlags
		force bool
	)
	cmd := &cobra.Command{
		Use:   "section <name>",
		Short: "Delete a section from a scope",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sc, err := a.scope(cmd, flags)
			if err != nil {
				return err
			}
			scopeName, project := sc.api()
			err = a.client.DeleteSection(cmd.Context(), scopeName, project, args[0], force)
			if err != nil {
				return a.deleteErr(cmd, err, "section")
			}
			fmt.Fprintf(cmd.OutOrStdout(), "section %q deleted from %s\n", args[0], sc)
			return nil
		},
	}
	flags.register(cmd)
	cmd.Flags().BoolVar(&force, "force", false, "delete the section and its contexts")
	return cmd
}

// ensureProject creates the project if it does not exist.
func (a *app) ensureProject(cmd *cobra.Command, name string) error {
	err := a.client.CreateProject(cmd.Context(), name)
	var apiErr *client.APIError
	if errors.As(err, &apiErr) && apiErr.Status == http.StatusConflict {
		return nil
	}
	if err != nil {
		return a.handleErr(cmd, err)
	}
	return nil
}
