package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCreateAndGetContext(t *testing.T) {
	a := newTestApp(t)
	mustRun(t, a, "create", "section", "style", "-g")

	out := mustRun(t, a, "create", "context", "go", "-s", "style", "-g", "-m", "Use gofmt.",
		"--title", "Go", "--description", "Go style", "--tag", "go", "--tag", "fmt")
	if out != "context/go created in global (section \"style\")\n" {
		t.Errorf("create out = %q", out)
	}

	out = mustRun(t, a, "get", "context", "go", "-s", "style", "-g")
	want := "---\ntitle: Go\ndescription: Go style\ntags:\n  - go\n  - fmt\n---\nUse gofmt.\n"
	if out != want {
		t.Errorf("get out = %q, want %q", out, want)
	}
}

func TestCreateContextDetectedProject(t *testing.T) {
	a := newTestApp(t)
	a.cwd = gitRepo(t, "api")
	mustRun(t, a, "create", "section", "style")
	out := mustRun(t, a, "create", "context", "go", "-s", "style", "-m", "x")
	if out != "context/go created in project \"api\" (section \"style\")\n" {
		t.Errorf("out = %q", out)
	}
}

func TestCreateContextMissingFlags(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"create", "context", "go", "-g", "-m", "x"}, "missing section: use -s <section>"},
		{[]string{"create", "context", "go", "-s", "style", "-g"}, "missing content: use -m <text> or -f <file>"},
		{[]string{"create", "context", "go", "-s", "style", "-g", "-m", "x", "-f", "a.md"}, "use -m or -f, not both"},
		{[]string{"get", "context", "go", "-g"}, "missing section: use -s <section>"},
		{[]string{"delete", "context", "go", "-g"}, "missing section: use -s <section>"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			_, _, err := runCLI(t, &app{cwd: t.TempDir()}, tt.args...)
			var ue *usageError
			if !errors.As(err, &ue) || ue.msg != tt.want {
				t.Errorf("err = %v, want usageError %q", err, tt.want)
			}
		})
	}
}

func TestCreateContextFromFile(t *testing.T) {
	a := newTestApp(t)
	mustRun(t, a, "create", "section", "style", "-g")
	file := filepath.Join(t.TempDir(), "rules.md")
	if err := os.WriteFile(file, []byte("Use gofmt.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	mustRun(t, a, "create", "context", "go", "-s", "style", "-g", "-f", file)
	if out := mustRun(t, a, "get", "context", "go", "-s", "style", "-g"); out != "Use gofmt.\n" {
		t.Errorf("get out = %q", out)
	}
}

func TestCreateContextFromStdin(t *testing.T) {
	a := newTestApp(t)
	mustRun(t, a, "create", "section", "style", "-g")

	root := newRootCmd(a)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader("# Go\n- Use gofmt.\n"))
	root.SetArgs([]string{"create", "context", "go", "-s", "style", "-g", "-f", "-"})
	if err := root.Execute(); err != nil {
		t.Fatalf("create: %v, out %q", err, out.String())
	}

	if got := mustRun(t, a, "get", "context", "go", "-s", "style", "-g"); got != "# Go\n- Use gofmt.\n" {
		t.Errorf("get out = %q", got)
	}
}

func TestCreateContextMissingFile(t *testing.T) {
	a := newTestApp(t)
	mustRun(t, a, "create", "section", "style", "-g")
	file := filepath.Join(t.TempDir(), "none.md")

	_, _, err := runCLI(t, a, "create", "context", "go", "-s", "style", "-g", "-f", file)
	if err == nil || !strings.Contains(err.Error(), "none.md") {
		t.Errorf("err = %v, want error with none.md", err)
	}
}

func TestGetContexts(t *testing.T) {
	a := newTestApp(t)
	if out := mustRun(t, a, "get", "contexts", "-g"); out != "No contexts found.\n" {
		t.Errorf("empty out = %q", out)
	}

	mustRun(t, a, "create", "section", "style", "-g")
	mustRun(t, a, "create", "section", "review", "-g")
	mustRun(t, a, "create", "context", "go", "-s", "style", "-g", "-m", "x", "--title", "Go", "--description", "Go style")
	mustRun(t, a, "create", "context", "pr", "-s", "review", "-g", "-m", "x")

	out := mustRun(t, a, "get", "contexts", "-g")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Fatalf("lines = %q", lines)
	}
	if f := strings.Fields(lines[0]); strings.Join(f, " ") != "NAME SECTION TITLE DESCRIPTION UPDATED" {
		t.Errorf("header = %q", lines[0])
	}
	// Sorted by id: review/pr before style/go.
	if !strings.HasPrefix(lines[1], "pr ") || !strings.HasPrefix(lines[2], "go ") || !strings.Contains(lines[2], "Go style") {
		t.Errorf("rows = %q", lines[1:])
	}

	out = mustRun(t, a, "get", "contexts", "-g", "-s", "review")
	if strings.Contains(out, "go ") || !strings.Contains(out, "pr ") {
		t.Errorf("filtered out = %q", out)
	}
}

func TestGetContextsDetectedMissingProject(t *testing.T) {
	a := newTestApp(t)
	a.cwd = gitRepo(t, "api")
	if out := mustRun(t, a, "get", "contexts"); out != "project: api (from git repo)\n\nNo contexts found.\n" {
		t.Errorf("out = %q", out)
	}
}

func TestGetContextNotFound(t *testing.T) {
	a := newTestApp(t)
	mustRun(t, a, "create", "section", "style", "-g")
	_, _, err := runCLI(t, a, "get", "context", "none", "-s", "style", "-g")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("err = %v, want not found", err)
	}
}

func TestDeleteContext(t *testing.T) {
	a := newTestApp(t)
	mustRun(t, a, "create", "section", "style", "-g")
	mustRun(t, a, "create", "context", "go", "-s", "style", "-g", "-m", "x")
	out := mustRun(t, a, "delete", "context", "go", "-s", "style", "-g")
	if out != "context \"go\" deleted from global (section \"style\")\n" {
		t.Errorf("out = %q", out)
	}
}

func TestAge(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	tests := map[time.Duration]string{
		10 * time.Second: "10s",
		5 * time.Minute:  "5m",
		3 * time.Hour:    "3h",
		50 * time.Hour:   "2d",
		-time.Minute:     "0s",
	}
	for d, want := range tests {
		if got := age(now.Add(-d), now); got != want {
			t.Errorf("age(%s) = %q, want %q", d, got, want)
		}
	}
}
