package cli

import (
	"errors"
	"strings"
	"testing"
)

func TestCreateSectionGlobal(t *testing.T) {
	a := newTestApp(t)
	if out := mustRun(t, a, "create", "section", "style", "-g"); out != "section/style created in global\n" {
		t.Errorf("out = %q", out)
	}
	if out := mustRun(t, a, "get", "sections", "-g"); out != "NAME\nstyle\n" {
		t.Errorf("get out = %q", out)
	}
}

func TestCreateSectionProjectFlag(t *testing.T) {
	a := newTestApp(t)
	mustRun(t, a, "create", "project", "api")
	if out := mustRun(t, a, "create", "section", "style", "-p", "api"); out != "section/style created in project \"api\"\n" {
		t.Errorf("out = %q", out)
	}

	// -p does not auto-create.
	_, _, err := runCLI(t, a, "create", "section", "style", "-p", "none")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("err = %v, want not found", err)
	}
}

func TestCreateSectionDetectedAutoCreatesProject(t *testing.T) {
	a := newTestApp(t)
	a.cwd = gitRepo(t, "My.Repo")
	if out := mustRun(t, a, "create", "section", "style"); out != "section/style created in project \"my-repo\"\n" {
		t.Errorf("out = %q", out)
	}
	if out := mustRun(t, a, "get", "projects"); !strings.Contains(out, "my-repo") {
		t.Errorf("projects = %q", out)
	}
	if out := mustRun(t, a, "get", "sections"); out != "NAME\nstyle\n" {
		t.Errorf("sections = %q", out)
	}
}

func TestGetSectionsDetectedMissingProject(t *testing.T) {
	a := newTestApp(t)
	a.cwd = gitRepo(t, "api")
	if out := mustRun(t, a, "get", "sections"); out != "No sections found.\n" {
		t.Errorf("out = %q", out)
	}
	// get does not auto-create.
	if out := mustRun(t, a, "get", "projects"); out != "No projects found.\n" {
		t.Errorf("projects = %q", out)
	}
}

func TestSectionNotInGitRepo(t *testing.T) {
	a := newTestApp(t)
	a.cwd = t.TempDir()
	_, errOut, err := runCLI(t, a, "get", "sections")
	if !errors.Is(err, errSilent) {
		t.Fatalf("err = %v, want errSilent", err)
	}
	if !strings.Contains(errOut, "not in a git repo") || !strings.Contains(errOut, "use -p <project> or -g") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestSectionBothFlags(t *testing.T) {
	a := newTestApp(t)
	_, _, err := runCLI(t, a, "get", "sections", "-p", "api", "-g")
	if err == nil || !strings.Contains(err.Error(), "-p and -g") {
		t.Errorf("err = %v", err)
	}
}

func TestDeleteSection(t *testing.T) {
	a := newTestApp(t)
	mustRun(t, a, "create", "section", "style", "-g")
	if out := mustRun(t, a, "delete", "section", "style", "-g"); out != "section \"style\" deleted from global\n" {
		t.Errorf("out = %q", out)
	}
	_, _, err := runCLI(t, a, "delete", "section", "style", "-g")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("err = %v, want not found", err)
	}
}

func TestDeleteSectionNotEmpty(t *testing.T) {
	a := newTestApp(t)
	mustRun(t, a, "create", "project", "api")
	mustRun(t, a, "create", "section", "style", "-p", "api")
	createContext(t, a, `{"scope":"project","project":"api","section":"style","name":"go","content":"x"}`)

	_, errOut, err := runCLI(t, a, "delete", "section", "style", "-p", "api")
	if !errors.Is(err, errSilent) {
		t.Fatalf("err = %v, want errSilent", err)
	}
	if !strings.Contains(errOut, "use --force") {
		t.Errorf("stderr = %q", errOut)
	}
	mustRun(t, a, "delete", "section", "style", "-p", "api", "--force")
}
