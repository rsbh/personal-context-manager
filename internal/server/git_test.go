package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/rsbh/personal-context-manager/internal/gitsync"
	"github.com/rsbh/personal-context-manager/internal/store"
)

func gitEnv(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, k := range []string{"GIT_AUTHOR", "GIT_COMMITTER"} {
		t.Setenv(k+"_NAME", "pcm-test")
		t.Setenv(k+"_EMAIL", "pcm-test@example.com")
	}
}

// newGitServer returns a router with git and a bare remote.
func newGitServer(t *testing.T) (http.Handler, string, string) {
	t.Helper()
	gitEnv(t)
	root := filepath.Join(t.TempDir(), "repo")
	st, err := store.New(root)
	if err != nil {
		t.Fatal(err)
	}
	remote := filepath.Join(t.TempDir(), "remote.git")
	if out, err := exec.Command("git", "init", "--quiet", "--bare", "-b", "main", remote).CombinedOutput(); err != nil {
		t.Fatalf("git init bare: %v: %s", err, out)
	}
	url := ""
	g := gitsync.New(root, func() (string, error) { return url, nil })
	h := NewRouter(st, WithGit(g))
	t.Cleanup(func() { url = "" })
	return h, root, remote
}

func TestStoreInitMain(t *testing.T) {
	gitEnv(t)
	root := t.TempDir()
	if _, err := store.New(root); err != nil {
		t.Fatal(err)
	}
	head, err := os.ReadFile(filepath.Join(root, ".git", "HEAD"))
	if err != nil {
		t.Fatal(err)
	}
	if string(head) != "ref: refs/heads/main\n" {
		t.Errorf("HEAD = %q", head)
	}
}

func TestGitEndpoints(t *testing.T) {
	h, _, _ := newGitServer(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/sections", `{"scope":"global","name":"style"}`), http.StatusCreated)

	rec := do(t, h, http.MethodGet, "/api/v1/git/status", "")
	expectStatus(t, rec, http.StatusOK)
	st := decodeBody[gitsync.Status](t, rec)
	if st.Branch != "main" || len(st.Changes) == 0 {
		t.Errorf("status = %+v", st)
	}

	rec = do(t, h, http.MethodGet, "/api/v1/git/log", "")
	expectStatus(t, rec, http.StatusOK)
	if got := decodeBody[[]gitsync.Commit](t, rec); len(got) != 0 {
		t.Errorf("empty log = %+v", got)
	}

	rec = do(t, h, http.MethodPost, "/api/v1/git/sync", "")
	expectStatus(t, rec, http.StatusOK)
	if res := decodeBody[gitsync.SyncResult](t, rec); res.Committed != 1 || res.Pushed {
		t.Errorf("sync = %+v", res)
	}

	rec = do(t, h, http.MethodGet, "/api/v1/git/log?n=5", "")
	expectStatus(t, rec, http.StatusOK)
	if got := decodeBody[[]gitsync.Commit](t, rec); len(got) != 1 {
		t.Errorf("log = %+v", got)
	}
	expectStatus(t, do(t, h, http.MethodGet, "/api/v1/git/log?n=abc", ""), http.StatusBadRequest)
}

func TestGitSyncConflictResponse(t *testing.T) {
	rec := httpRecorder()
	writeGitError(rec, &gitsync.ConflictError{Files: []string{"global/style/go.md"}})
	expectStatus(t, rec, http.StatusConflict)
	got := decodeBody[map[string]any](t, rec)
	if got["error"] != "sync conflict, rebase aborted" {
		t.Errorf("body = %v", got)
	}
	if files, _ := got["files"].([]any); len(files) != 1 || files[0] != "global/style/go.md" {
		t.Errorf("files = %v", got["files"])
	}
}

func TestGitNotConfigured(t *testing.T) {
	h, _ := newTestServer(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/git/sync", ""), http.StatusNotImplemented)
}

func TestGitSyncSurvivesClientCancel(t *testing.T) {
	h, _, _ := newGitServer(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/sections", `{"scope":"global","name":"style"}`), http.StatusCreated)

	// The client is gone before git runs. Sync must still finish.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/git/sync", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	expectStatus(t, rec, http.StatusOK)
	if res := decodeBody[gitsync.SyncResult](t, rec); res.Committed != 1 {
		t.Errorf("sync = %+v", res)
	}
}
