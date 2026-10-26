package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

const (
	outputTable = "table"
	outputMD    = "md"
)

func checkOutput(cmd *cobra.Command, output string, args []string) error {
	switch output {
	case outputTable:
		return nil
	case outputMD:
		if len(args) > 0 {
			return newUsageError(cmd, "-o md works only for a list, without a name")
		}
		return nil
	default:
		return newUsageError(cmd, "invalid output format %q: use table or md", output)
	}
}

// resolveContexts prints merged global + project contexts as markdown.
// Not in a git repo and no flag: global only.
func (a *app) resolveContexts(cmd *cobra.Command, flags contextFlags) error {
	cwd, err := a.workDir()
	if err != nil {
		return err
	}
	sc, err := resolveScope(flags.scopeFlags, cwd)
	switch {
	case errors.Is(err, errNotInRepo):
		sc = scope{global: true}
	case err != nil:
		return err
	}

	md, err := a.client.ResolveMarkdown(cmd.Context(), sc.project, flags.section)
	if err != nil {
		return a.handleErr(cmd, err)
	}
	fmt.Fprint(cmd.OutOrStdout(), md)
	return nil
}
