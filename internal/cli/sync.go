package cli

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rsbh/personal-context-manager/internal/client"
)

var remoteNotSet = []field{
	{"info", "remote not set, contexts are local only"},
	{"hint", "pcm config set remote <git-url> && pcm sync"},
}

func newSyncCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Commit, pull and push the context repo",
		Args:  noArgs,
		RunE:  a.sync,
	}
	var n int
	logCmd := &cobra.Command{
		Use:   "log",
		Short: "Show recent sync commits",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return a.syncLog(cmd, n)
		},
	}
	logCmd.Flags().IntVarP(&n, "number", "n", 10, "number of commits")
	cmd.AddCommand(
		&cobra.Command{
			Use:   "status",
			Short: "Show context repo changes and remote",
			Args:  noArgs,
			RunE:  a.syncStatus,
		},
		logCmd,
	)
	return cmd
}

func (a *app) sync(cmd *cobra.Command, _ []string) error {
	res, err := a.client.Sync(cmd.Context())
	var apiErr *client.APIError
	if errors.As(err, &apiErr) && apiErr.Status == http.StatusConflict {
		printFields(cmd.ErrOrStderr(), []field{
			{"error", apiErr.Message},
			{"files", strings.Join(apiErr.Files, ", ")},
			{"hint", "fix in " + a.cfg.Path + ", then pcm sync"},
		})
		return errSilent
	}
	if err != nil {
		return a.handleErr(cmd, err)
	}

	fields := []field{{"committed", plural(res.Committed, "file")}}
	if res.Remote == "" {
		fields = append(fields, remoteNotSet...)
	} else {
		pushed := "no"
		if res.Pushed {
			pushed = "yes"
		}
		fields = append(fields, field{"pulled", plural(res.Pulled, "commit")}, field{"pushed", pushed})
	}
	printFields(cmd.OutOrStdout(), fields)
	return nil
}

func (a *app) syncStatus(cmd *cobra.Command, _ []string) error {
	st, err := a.client.GitStatus(cmd.Context())
	if err != nil {
		return a.handleErr(cmd, err)
	}
	remote := st.Remote
	if remote == "" {
		remote = "none"
	}
	fields := []field{
		{"branch", st.Branch},
		{"remote", remote},
		{"changes", plural(len(st.Changes), "file")},
		{"ahead", strconv.Itoa(st.Ahead)},
		{"behind", strconv.Itoa(st.Behind)},
	}
	switch {
	case st.Remote == "":
		fields = append(fields, remoteNotSet...)
	case st.Ahead > 0:
		fields = append(fields, field{"hint", "run pcm sync to push"})
	}
	out := cmd.OutOrStdout()
	printFields(out, fields)
	if len(st.Changes) > 0 {
		fmt.Fprintln(out)
		for _, c := range st.Changes {
			fmt.Fprintln(out, c)
		}
	}
	return nil
}

func (a *app) syncLog(cmd *cobra.Command, n int) error {
	if n < 1 {
		return newUsageError(cmd, "invalid -n %d: use a number >= 1", n)
	}
	commits, err := a.client.GitLog(cmd.Context(), n)
	if err != nil {
		return a.handleErr(cmd, err)
	}
	out := cmd.OutOrStdout()
	if len(commits) == 0 {
		fmt.Fprintln(out, "No commits yet.")
		return nil
	}
	rows := make([][]string, len(commits))
	for i, c := range commits {
		rows[i] = []string{c.Hash, c.Date, c.Message}
	}
	printTable(out, []string{"HASH", "DATE", "MESSAGE"}, rows)
	return nil
}

// plural returns "1 file" or "2 files".
func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return fmt.Sprintf("%d %ss", n, word)
}
