package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rsbh/personal-context-manager/internal/store"
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

	contexts, err := a.client.Resolve(cmd.Context(), sc.project, flags.section)
	if err != nil {
		return a.handleErr(cmd, err)
	}
	if len(contexts) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No contexts found.")
		return nil
	}
	printMarkdown(cmd.OutOrStdout(), contexts)
	return nil
}

// printMarkdown prints one heading per context, the description as a quote, then the content.
func printMarkdown(w io.Writer, contexts []store.Context) {
	fmt.Fprintln(w, "# Context")
	for _, c := range contexts {
		sc := scope{global: c.Scope == store.ScopeGlobal, project: c.Project}
		fmt.Fprintf(w, "\n## %s/%s (%s)\n", c.Section, c.Name, sc)
		if d := c.Frontmatter.Description; d != "" {
			fmt.Fprintf(w, "> %s\n", d)
		}
		fmt.Fprintf(w, "\n%s\n", strings.TrimRight(c.Content, "\n"))
	}
}
