package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/rsbh/personal-context-manager/internal/client"
)

// k8s style verbs: pcm <verb> <resource> [name] [flags].
// Each resource adds its subcommands to these verbs.

func newGetCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{Use: "get", Short: "Show resources"}
	cmd.AddCommand(newGetProjectCmd(a))
	return cmd
}

func newCreateCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{Use: "create", Short: "Create a resource"}
	cmd.AddCommand(newCreateProjectCmd(a))
	return cmd
}

func newDeleteCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{Use: "delete", Short: "Delete a resource"}
	cmd.AddCommand(newDeleteProjectCmd(a))
	return cmd
}

// handleErr prints the not-reachable lines for a server that is down.
// Other errors return as is, and Execute prints them.
func (a *app) handleErr(cmd *cobra.Command, err error) error {
	if errors.Is(err, client.ErrNotReachable) {
		fields, _ := healthFields(cmd.Context(), a.client)
		printFields(cmd.OutOrStdout(), fields)
		return errSilent
	}
	return err
}

// printTable prints a header and rows, aligned like kubectl get.
func printTable(w io.Writer, header []string, rows [][]string) {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, strings.Join(header, "\t"))
	for _, r := range rows {
		fmt.Fprintln(tw, strings.Join(r, "\t"))
	}
	tw.Flush()
}
