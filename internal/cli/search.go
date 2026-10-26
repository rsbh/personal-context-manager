package cli

import (
	"errors"
	"fmt"
	"html"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rsbh/personal-context-manager/internal/client"
)

var errScopeFlags = errors.New("use only one of -p, -g and --all")

type searchFlags struct {
	scopeFlags
	all     bool
	section string
	limit   int
}

func newSearchCmd(a *app) *cobra.Command {
	var flags searchFlags
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search contexts (fuzzy, ranked; +word must match, -word must not, \"phrase\" exact)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.search(cmd, strings.Join(args, " "), flags)
		},
	}
	flags.register(cmd)
	cmd.Flags().BoolVar(&flags.all, "all", false, "search all scopes and projects")
	cmd.Flags().StringVarP(&flags.section, "section", "s", "", "section name")
	cmd.Flags().IntVar(&flags.limit, "limit", 20, "max results")
	return cmd
}

// searchParams picks what to search: -g global, -p global + project, --all
// everything. No flag: global + git repo project, or everything outside a repo.
func (a *app) searchParams(q string, flags searchFlags) (client.SearchParams, error) {
	p := client.SearchParams{Q: q, Section: flags.section, Limit: flags.limit}
	n := 0
	for _, set := range []bool{flags.global, flags.project != "", flags.all} {
		if set {
			n++
		}
	}
	if n > 1 {
		return p, errScopeFlags
	}
	if flags.all {
		return p, nil
	}
	cwd, err := a.workDir()
	if err != nil {
		return p, err
	}
	sc, err := resolveScope(flags.scopeFlags, cwd)
	switch {
	case errors.Is(err, errNotInRepo):
		return p, nil
	case err != nil:
		return p, err
	case sc.global:
		p.Scope = "global"
	default:
		p.Project, p.Global = sc.project, true
	}
	return p, nil
}

func (a *app) search(cmd *cobra.Command, q string, flags searchFlags) error {
	if strings.TrimSpace(q) == "" {
		return newUsageError(cmd, "missing query")
	}
	if flags.limit <= 0 {
		return newUsageError(cmd, "invalid limit")
	}
	p, err := a.searchParams(q, flags)
	if err != nil {
		return err
	}
	hits, err := a.client.Search(cmd.Context(), p)
	if err != nil {
		return a.handleErr(cmd, err)
	}

	out := cmd.OutOrStdout()
	if len(hits) == 0 {
		fmt.Fprintln(out, "No results found.")
		return nil
	}
	tty := isTerminal(out)
	rows := make([][]string, len(hits))
	for i, h := range hits {
		rows[i] = []string{h.Scope, h.Name, h.Section, h.Title, formatSnippet(h.Snippet, tty), fmt.Sprintf("%.2f", h.Score)}
	}
	printTable(out, []string{"SCOPE", "NAME", "SECTION", "TITLE", "SNIPPET", "SCORE"}, rows)
	return nil
}

// formatSnippet turns the API snippet (HTML escaped, matches in <mark>)
// into one line of text. Matches are bold on a terminal.
func formatSnippet(s string, tty bool) string {
	on, off := "", ""
	if tty {
		on, off = "\x1b[1m", "\x1b[0m"
	}
	// Replace marks before unescape, so an escaped "&lt;mark&gt;" in the text stays text.
	s = strings.NewReplacer("<mark>", on, "</mark>", off).Replace(s)
	return strings.Join(strings.Fields(html.UnescapeString(s)), " ")
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func newIndexCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{Use: "index", Short: "Manage the search index"}
	cmd.AddCommand(&cobra.Command{
		Use:   "rebuild",
		Short: "Rebuild the search index from md files",
		Args:  noArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			n, err := a.client.RebuildIndex(cmd.Context())
			if err != nil {
				return a.handleErr(cmd, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "index rebuilt: %s\n", plural(n, "context"))
			return nil
		},
	})
	return cmd
}
