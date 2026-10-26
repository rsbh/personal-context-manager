package cli

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rsbh/personal-context-manager/internal/client"
	"github.com/rsbh/personal-context-manager/internal/server"
	"github.com/rsbh/personal-context-manager/internal/store"
)

// newTestApp returns an app that talks to a real router with a temp store.
func newTestApp(t *testing.T) *app {
	t.Helper()
	st, err := store.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(server.NewRouter(st))
	t.Cleanup(srv.Close)
	return &app{client: client.New(srv.URL)}
}

// runCLI runs pcm with args and returns stdout, stderr and the error.
func runCLI(t *testing.T, a *app, args ...string) (string, string, error) {
	t.Helper()
	root := newRootCmd(a)
	var out, errOut bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), errOut.String(), err
}

func mustRun(t *testing.T, a *app, args ...string) string {
	t.Helper()
	out, errOut, err := runCLI(t, a, args...)
	if err != nil {
		t.Fatalf("pcm %s: %v, stderr %q", strings.Join(args, " "), err, errOut)
	}
	return out
}

func TestGetProjectsEmpty(t *testing.T) {
	a := newTestApp(t)
	if out := mustRun(t, a, "get", "projects"); out != "No projects found.\n" {
		t.Errorf("out = %q", out)
	}
}

func TestCreateAndGetProjects(t *testing.T) {
	a := newTestApp(t)
	if out := mustRun(t, a, "create", "project", "web"); out != "project/web created\n" {
		t.Errorf("create out = %q", out)
	}
	mustRun(t, a, "create", "project", "api")

	if out := mustRun(t, a, "get", "projects"); out != "NAME\napi\nweb\n" {
		t.Errorf("list out = %q", out)
	}
	if out := mustRun(t, a, "get", "project", "api"); out != "NAME\napi\n" {
		t.Errorf("get out = %q", out)
	}
}

func TestCreateProjectExists(t *testing.T) {
	a := newTestApp(t)
	mustRun(t, a, "create", "project", "api")
	_, _, err := runCLI(t, a, "create", "project", "api")
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("err = %v, want already exists", err)
	}
}

func TestGetProjectNotFound(t *testing.T) {
	a := newTestApp(t)
	_, _, err := runCLI(t, a, "get", "project", "none")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("err = %v, want not found", err)
	}
}

func TestDeleteProject(t *testing.T) {
	a := newTestApp(t)
	mustRun(t, a, "create", "project", "api")
	if out := mustRun(t, a, "delete", "project", "api"); out != "project \"api\" deleted\n" {
		t.Errorf("out = %q", out)
	}
}

func TestDeleteProjectNotEmpty(t *testing.T) {
	a := newTestApp(t)
	mustRun(t, a, "create", "project", "api")
	mustRun(t, a, "create", "section", "style", "-p", "api")

	_, errOut, err := runCLI(t, a, "delete", "project", "api")
	if !errors.Is(err, errSilent) {
		t.Fatalf("err = %v, want errSilent", err)
	}
	if !strings.Contains(errOut, "not empty") || !strings.Contains(errOut, "use --force") {
		t.Errorf("stderr = %q", errOut)
	}
	mustRun(t, a, "delete", "project", "api", "--force")
}

func TestProjectNotReachable(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()

	out, _, err := runCLI(t, &app{client: client.New(url)}, "get", "projects")
	if !errors.Is(err, errSilent) {
		t.Fatalf("err = %v, want errSilent", err)
	}
	if !strings.Contains(out, "hint:   run `pcm server start`") {
		t.Errorf("out = %q", out)
	}
}

// createContext uses the API directly. The CLI has no context commands yet.
func createContext(t *testing.T, a *app, body string) {
	t.Helper()
	resp, err := http.Post(a.client.BaseURL()+"/api/v1/contexts", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create context: status %d", resp.StatusCode)
	}
}
