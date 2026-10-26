package cli

import (
	"strings"
	"testing"
)

// rowsOf returns the table rows as fields, without the header.
func rowsOf(t *testing.T, out string) [][]string {
	t.Helper()
	var rows [][]string
	header := false
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		f := strings.Fields(line)
		if len(f) > 0 && f[0] == "SCOPE" {
			header = true
			if strings.Join(f, " ") != "SCOPE NAME SECTION TITLE DESCRIPTION UPDATED" {
				t.Errorf("header = %q", line)
			}
			continue
		}
		if header {
			rows = append(rows, f)
		}
	}
	if !header {
		t.Fatalf("no table header in %q", out)
	}
	return rows
}

func setupGlobalAndProject(t *testing.T) *app {
	t.Helper()
	a := newTestApp(t)
	mustRun(t, a, "create", "section", "style", "-g")
	mustRun(t, a, "create", "context", "go", "-s", "style", "-g", "-m", "global go")
	mustRun(t, a, "create", "context", "py", "-s", "style", "-g", "-m", "global py")
	mustRun(t, a, "create", "project", "web")
	mustRun(t, a, "create", "section", "style", "-p", "web")
	mustRun(t, a, "create", "context", "go", "-s", "style", "-p", "web", "-m", "web go")
	mustRun(t, a, "create", "context", "ts", "-s", "style", "-p", "web", "-m", "web ts")
	return a
}

func TestGetContextsCurrentDetected(t *testing.T) {
	a := setupGlobalAndProject(t)
	a.cwd = gitRepo(t, "web")

	out := mustRun(t, a, "get", "contexts")
	if !strings.HasPrefix(out, "project: web (from git repo)\n\n") {
		t.Errorf("out = %q", out)
	}
	// Global first, then project. Project go replaces global go.
	got := []string{}
	for _, r := range rowsOf(t, out) {
		got = append(got, r[0]+"/"+r[1])
	}
	if want := "global/py project/go project/ts"; strings.Join(got, " ") != want {
		t.Errorf("rows = %v, want %s", got, want)
	}
}

func TestGetContextsCurrentProjectFlag(t *testing.T) {
	a := setupGlobalAndProject(t)
	a.cwd = t.TempDir()
	out := mustRun(t, a, "get", "contexts", "-p", "web", "-s", "style")
	if !strings.HasPrefix(out, "project: web\n\n") {
		t.Errorf("out = %q", out)
	}
	if n := len(rowsOf(t, out)); n != 3 {
		t.Errorf("rows = %d, want 3", n)
	}
}

func TestGetContextsCurrentNotInRepo(t *testing.T) {
	a := setupGlobalAndProject(t)
	a.cwd = t.TempDir()
	out := mustRun(t, a, "get", "contexts")
	if !strings.HasPrefix(out, "project: none (not in a git repo)\n\n") {
		t.Errorf("out = %q", out)
	}
	for _, r := range rowsOf(t, out) {
		if r[0] != "global" {
			t.Errorf("row %v is not global", r)
		}
	}
}

func TestGetContextsGlobalUnchanged(t *testing.T) {
	a := setupGlobalAndProject(t)
	out := mustRun(t, a, "get", "contexts", "-g")
	if strings.Contains(out, "project:") || strings.Contains(out, "SCOPE") {
		t.Errorf("-g out = %q", out)
	}
}
