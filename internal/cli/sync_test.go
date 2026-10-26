package cli

import (
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rsbh/personal-context-manager/internal/client"
	"github.com/rsbh/personal-context-manager/internal/config"
	"github.com/rsbh/personal-context-manager/internal/gitsync"
	"github.com/rsbh/personal-context-manager/internal/server"
	"github.com/rsbh/personal-context-manager/internal/store"
)

// newGitTestApp returns an app with a temp home and a server with git.
// The server reads remote from the app's config, like pcm-server.
func newGitTestApp(t *testing.T) *app {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, k := range []string{"GIT_AUTHOR", "GIT_COMMITTER"} {
		t.Setenv(k+"_NAME", "pcm-test")
		t.Setenv(k+"_EMAIL", "pcm-test@example.com")
	}
	dir := config.DirIn(home)
	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.New(cfg.Path)
	if err != nil {
		t.Fatal(err)
	}
	g := gitsync.New(cfg.Path, func() (string, error) {
		c, err := config.Load(dir)
		return c.Remote, err
	})
	srv := httptest.NewServer(server.NewRouter(st, server.WithGit(g)))
	t.Cleanup(srv.Close)
	return &app{home: home, cfg: cfg, client: client.New(srv.URL)}
}

func bareRemote(t *testing.T) string {
	t.Helper()
	remote := filepath.Join(t.TempDir(), "remote.git")
	if out, err := exec.Command("git", "init", "--quiet", "--bare", "-b", "main", remote).CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	return remote
}

func TestSyncNoRemote(t *testing.T) {
	a := newGitTestApp(t)
	mustRun(t, a, "create", "section", "style", "-g")
	out := mustRun(t, a, "sync")
	want := "committed: 1 file\n" +
		"info:      remote not set, contexts are local only\n" +
		"hint:      pcm config set remote <git-url> && pcm sync\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestSyncWithRemote(t *testing.T) {
	a := newGitTestApp(t)
	mustRun(t, a, "config", "set", "remote", bareRemote(t))
	mustRun(t, a, "create", "section", "style", "-g")
	mustRun(t, a, "create", "context", "go", "-s", "style", "-g", "-m", "x")

	out := mustRun(t, a, "sync")
	if out != "committed: 2 files\npulled:    0 commits\npushed:    yes\n" {
		t.Errorf("first out = %q", out)
	}
	out = mustRun(t, a, "sync")
	if out != "committed: 0 files\npulled:    0 commits\npushed:    no\n" {
		t.Errorf("second out = %q", out)
	}
}

func TestSyncStatus(t *testing.T) {
	a := newGitTestApp(t)
	mustRun(t, a, "create", "section", "style", "-g")
	out := mustRun(t, a, "sync", "status")
	want := "branch:  main\n" +
		"remote:  none\n" +
		"changes: 1 file\n" +
		"ahead:   0\n" +
		"behind:  0\n" +
		"info:    remote not set, contexts are local only\n" +
		"hint:    pcm config set remote <git-url> && pcm sync\n" +
		"\n?? global/\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestSyncLog(t *testing.T) {
	a := newGitTestApp(t)
	if out := mustRun(t, a, "sync", "log"); out != "No commits yet.\n" {
		t.Errorf("empty out = %q", out)
	}
	mustRun(t, a, "create", "section", "style", "-g")
	mustRun(t, a, "sync")
	out := mustRun(t, a, "sync", "log", "-n", "1")
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 2 || strings.Join(strings.Fields(lines[0]), " ") != "HASH DATE MESSAGE" || !strings.HasSuffix(lines[1], "pcm sync: 1 file changed") {
		t.Errorf("out = %q", out)
	}
}

func TestSyncConflictOutput(t *testing.T) {
	a := newGitTestApp(t)
	remote := bareRemote(t)
	mustRun(t, a, "config", "set", "remote", remote)
	mustRun(t, a, "create", "section", "style", "-g")
	mustRun(t, a, "create", "context", "go", "-s", "style", "-g", "-m", "first")
	mustRun(t, a, "sync")

	// Another clone changes the same file.
	clone := filepath.Join(t.TempDir(), "clone")
	for _, args := range [][]string{{"clone", "--quiet", remote, clone}} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
	}
	writeFile(t, filepath.Join(clone, "global/style/go.md"), "remote")
	for _, args := range [][]string{{"add", "-A"}, {"commit", "--quiet", "-m", "remote"}, {"push", "--quiet", "origin", "main"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = clone
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
	}
	mustRun(t, a, "delete", "context", "go", "-s", "style", "-g")
	mustRun(t, a, "create", "context", "go", "-s", "style", "-g", "-m", "local")

	_, errOut, err := runCLI(t, a, "sync")
	if err == nil {
		t.Fatal("want error")
	}
	want := "error: sync conflict, rebase aborted\n" +
		"files: global/style/go.md\n" +
		"hint:  fix in " + a.cfg.Path + ", then pcm sync\n"
	if errOut != want {
		t.Errorf("stderr = %q, want %q", errOut, want)
	}
}
