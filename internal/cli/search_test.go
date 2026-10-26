package cli

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/rsbh/personal-context-manager/internal/client"
	"github.com/rsbh/personal-context-manager/internal/search"
	"github.com/rsbh/personal-context-manager/internal/server"
	"github.com/rsbh/personal-context-manager/internal/store"
)

// newSearchApp returns an app with a search index and contexts:
// global/style/go, projects/api/style/apigo, projects/web/style/webgo.
// It runs outside a git repo. It also returns the repo root.
func newSearchApp(t *testing.T) (*app, string) {
	t.Helper()
	root := t.TempDir()
	st, err := store.New(root)
	if err != nil {
		t.Fatal(err)
	}
	idx, _, err := search.Open(filepath.Join(t.TempDir(), "index"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { idx.Close() })
	srv := httptest.NewServer(server.NewRouter(st, server.WithSearch(idx)))
	t.Cleanup(srv.Close)
	a := &app{client: client.New(srv.URL), cwd: t.TempDir()}

	mustRun(t, a, "create", "section", "style", "-g")
	mustRun(t, a, "create", "context", "go", "-s", "style", "-g", "-m", "Use gofmt <always>.\nSecond line.", "--title", "Go")
	for _, p := range []string{"api", "web"} {
		mustRun(t, a, "create", "project", p)
		mustRun(t, a, "create", "section", "style", "-p", p)
		mustRun(t, a, "create", "context", p+"go", "-s", "style", "-p", p, "-m", "Run gofmt in "+p+".")
	}
	return a, root
}

// searchRows returns the table rows, without the header, as "scope/name".
func searchRows(t *testing.T, out string) []string {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if got := strings.Join(strings.Fields(lines[0]), " "); got != "SCOPE NAME SECTION TITLE SNIPPET SCORE" {
		t.Fatalf("header = %q", lines[0])
	}
	var rows []string
	for _, l := range lines[1:] {
		f := strings.Fields(l)
		rows = append(rows, f[0]+"/"+f[1])
	}
	slices.Sort(rows)
	return rows
}

func TestSearchScopes(t *testing.T) {
	a, _ := newSearchApp(t)
	repo := gitRepo(t, "api")
	for _, tc := range []struct {
		name string
		cwd  string
		args []string
		want []string
	}{
		{"not in repo: all", "", nil, []string{"global/go", "project/apigo", "project/webgo"}},
		{"in repo: global + project", repo, nil, []string{"global/go", "project/apigo"}},
		{"-p: global + project", "", []string{"-p", "web"}, []string{"global/go", "project/webgo"}},
		{"-g: global only", repo, []string{"-g"}, []string{"global/go"}},
		{"--all in repo", repo, []string{"--all"}, []string{"global/go", "project/apigo", "project/webgo"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cwd != "" {
				a.cwd = tc.cwd
			} else {
				a.cwd = t.TempDir()
			}
			out := mustRun(t, a, append([]string{"search", "gofmt"}, tc.args...)...)
			if got := searchRows(t, out); !slices.Equal(got, tc.want) {
				t.Errorf("rows = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSearchScopeFlagsConflict(t *testing.T) {
	a, _ := newSearchApp(t)
	for _, args := range [][]string{{"-p", "api", "--all"}, {"-g", "--all"}, {"-p", "api", "-g"}} {
		_, _, err := runCLI(t, a, append([]string{"search", "gofmt"}, args...)...)
		if err == nil || !strings.Contains(err.Error(), "use only one of -p, -g and --all") {
			t.Errorf("%v: err = %v", args, err)
		}
	}
}

func TestSearchRow(t *testing.T) {
	a, _ := newSearchApp(t)
	out := mustRun(t, a, "search", "gofmt", "-g")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 {
		t.Fatalf("out = %q", out)
	}
	row := lines[1]
	// Not a TTY: no <mark>, no ANSI, HTML decoded, one line.
	if !strings.Contains(row, "Use gofmt <always>. Second line.") {
		t.Errorf("snippet in row = %q", row)
	}
	if !strings.Contains(row, "Go") || !strings.Contains(row, "style") {
		t.Errorf("row = %q", row)
	}
	f := strings.Fields(row)
	if score := f[len(f)-1]; len(score) < 4 || score[len(score)-3] != '.' {
		t.Errorf("score = %q, want 2 decimals", score)
	}
}

func TestSearchJoinsArgs(t *testing.T) {
	a, _ := newSearchApp(t)
	out := mustRun(t, a, "search", "+gofmt", "+web", "--all")
	if got := searchRows(t, out); !slices.Equal(got, []string{"project/webgo"}) {
		t.Errorf("rows = %v", got)
	}
}

func TestSearchSection(t *testing.T) {
	a, _ := newSearchApp(t)
	mustRun(t, a, "create", "section", "review", "-g")
	mustRun(t, a, "create", "context", "rv", "-s", "review", "-g", "-m", "Check gofmt.")
	out := mustRun(t, a, "search", "gofmt", "-s", "review")
	if got := searchRows(t, out); !slices.Equal(got, []string{"global/rv"}) {
		t.Errorf("rows = %v", got)
	}
}

func TestSearchLimit(t *testing.T) {
	a, _ := newSearchApp(t)
	out := mustRun(t, a, "search", "gofmt", "--limit", "1")
	if got := searchRows(t, out); len(got) != 1 {
		t.Errorf("rows = %v", got)
	}
	for _, n := range []string{"0", "-1"} {
		if _, _, err := runCLI(t, a, "search", "gofmt", "--limit", n); err == nil || !strings.Contains(err.Error(), "invalid limit") {
			t.Errorf("limit %s: err = %v", n, err)
		}
	}
}

func TestSearchNoResults(t *testing.T) {
	a, _ := newSearchApp(t)
	if out := mustRun(t, a, "search", "zzzqqq"); out != "No results found.\n" {
		t.Errorf("out = %q", out)
	}
}

func TestSearchMissingQuery(t *testing.T) {
	a, _ := newSearchApp(t)
	for _, args := range [][]string{{"search"}, {"search", " "}} {
		if _, _, err := runCLI(t, a, args...); err == nil || !strings.Contains(err.Error(), "missing query") {
			t.Errorf("%q: err = %v", args, err)
		}
	}
}

func TestIndexRebuild(t *testing.T) {
	a, root := newSearchApp(t)
	if err := os.WriteFile(filepath.Join(root, "global", "style", "rs.md"), []byte("Use rustfmt."), 0o600); err != nil {
		t.Fatal(err)
	}
	if out := mustRun(t, a, "index", "rebuild"); out != "index rebuilt: 4 contexts\n" {
		t.Errorf("out = %q", out)
	}
	out := mustRun(t, a, "search", "rustfmt")
	if got := searchRows(t, out); !slices.Equal(got, []string{"global/rs"}) {
		t.Errorf("rows = %v", got)
	}
}

func TestFormatSnippet(t *testing.T) {
	in := "Use <mark>gofmt</mark> &lt;now&gt;.\n\tNext  line."
	if got, want := formatSnippet(in, false), "Use gofmt <now>. Next line."; got != want {
		t.Errorf("no tty = %q, want %q", got, want)
	}
	if got, want := formatSnippet(in, true), "Use \x1b[1mgofmt\x1b[0m <now>. Next line."; got != want {
		t.Errorf("tty = %q, want %q", got, want)
	}
}
