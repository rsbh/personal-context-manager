package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/rsbh/personal-context-manager/internal/store"
)

func TestPrintMarkdown(t *testing.T) {
	var buf bytes.Buffer
	printMarkdown(&buf, []store.Context{
		{Scope: "global", Section: "style", Name: "go", Frontmatter: store.Frontmatter{Description: "Go style"}, Content: "Use gofmt.\n"},
		{Scope: "project", Project: "api", Section: "review", Name: "pr", Content: "Check errors."},
	})
	want := "# Context\n\n" +
		"## style/go (global)\n> Go style\n\nUse gofmt.\n\n" +
		"## review/pr (project \"api\")\n\nCheck errors.\n"
	if buf.String() != want {
		t.Errorf("out =\n%s\nwant =\n%s", buf.String(), want)
	}
}

func TestGetContextsMarkdown(t *testing.T) {
	a := newTestApp(t)
	mustRun(t, a, "create", "section", "style", "-g")
	mustRun(t, a, "create", "context", "go", "-s", "style", "-g", "-m", "global go")
	mustRun(t, a, "create", "context", "py", "-s", "style", "-g", "-m", "global py")
	a.cwd = gitRepo(t, "api")
	mustRun(t, a, "create", "section", "style")
	mustRun(t, a, "create", "context", "go", "-s", "style", "-m", "api go")

	out := mustRun(t, a, "get", "contexts", "-o", "md")
	if !strings.Contains(out, "## style/py (global)") || !strings.Contains(out, "## style/go (project \"api\")") {
		t.Errorf("out = %q", out)
	}
	if strings.Contains(out, "global go") {
		t.Errorf("global go not replaced: %q", out)
	}

	out = mustRun(t, a, "get", "contexts", "-o", "md", "-g")
	if !strings.Contains(out, "global go") || strings.Contains(out, "api go") {
		t.Errorf("-g out = %q", out)
	}
}

func TestGetContextsMarkdownNotInRepo(t *testing.T) {
	a := newTestApp(t)
	mustRun(t, a, "create", "section", "style", "-g")
	mustRun(t, a, "create", "context", "go", "-s", "style", "-g", "-m", "x")
	a.cwd = t.TempDir()
	if out := mustRun(t, a, "get", "contexts", "-o", "md"); !strings.Contains(out, "## style/go (global)") {
		t.Errorf("out = %q", out)
	}
}

func TestGetContextsMarkdownEmpty(t *testing.T) {
	a := newTestApp(t)
	if out := mustRun(t, a, "get", "contexts", "-o", "md", "-g"); out != "No contexts found.\n" {
		t.Errorf("out = %q", out)
	}
}

func TestGetContextsInvalidOutput(t *testing.T) {
	_, _, err := runCLI(t, &app{cwd: t.TempDir()}, "get", "contexts", "-o", "json")
	var ue *usageError
	if !errors.As(err, &ue) || ue.msg != `invalid output format "json": use table or md` {
		t.Errorf("err = %v", err)
	}
}

func TestGetContextMarkdownWithName(t *testing.T) {
	_, _, err := runCLI(t, &app{cwd: t.TempDir()}, "get", "context", "go", "-s", "style", "-o", "md")
	var ue *usageError
	if !errors.As(err, &ue) || ue.msg != "-o md works only for a list, without a name" {
		t.Errorf("err = %v", err)
	}
}
