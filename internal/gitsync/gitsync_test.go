package gitsync

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
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

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return string(out)
}

func write(t *testing.T, root, path, content string) {
	t.Helper()
	file := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// setup returns a local repo (branch from init) and a bare remote path.
func setup(t *testing.T, initBranch string) (string, string) {
	t.Helper()
	gitEnv(t)
	base := t.TempDir()
	local := filepath.Join(base, "local")
	remote := filepath.Join(base, "remote.git")
	run(t, base, "init", "--quiet", "-b", initBranch, local)
	run(t, base, "init", "--quiet", "--bare", "-b", "main", remote)
	return local, remote
}

func fixedRemote(url string) func() (string, error) {
	return func() (string, error) { return url, nil }
}

// otherClonePush clones remote, writes a file and pushes.
func otherClonePush(t *testing.T, remote, path, content string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "clone")
	run(t, filepath.Dir(dir), "clone", "--quiet", remote, dir)
	write(t, dir, path, content)
	run(t, dir, "add", "-A")
	run(t, dir, "commit", "--quiet", "-m", "other")
	run(t, dir, "push", "--quiet", "origin", "main")
}

func TestSyncWithoutRemote(t *testing.T) {
	local, _ := setup(t, "main")
	write(t, local, "global/style/go.md", "Use gofmt.")
	write(t, local, "global/style/.gitkeep", "")
	r := New(local, fixedRemote(""))

	res, err := r.Sync(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Committed != 2 || res.Pushed || res.Pulled != 0 || res.Remote != "" {
		t.Errorf("result = %+v", res)
	}
	if msg := run(t, local, "log", "-1", "--format=%s"); strings.TrimSpace(msg) != "pcm sync: 2 files changed" {
		t.Errorf("commit message = %q", msg)
	}
}

func TestSyncPushPullAndNothing(t *testing.T) {
	local, remote := setup(t, "main")
	r := New(local, fixedRemote(remote))
	ctx := context.Background()

	write(t, local, "global/style/go.md", "Use gofmt.")
	res, err := r.Sync(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if res.Committed != 1 || !res.Pushed {
		t.Errorf("first sync = %+v", res)
	}
	if files := run(t, local, "--git-dir", remote, "ls-tree", "-r", "--name-only", "main"); !strings.Contains(files, "global/style/go.md") {
		t.Errorf("remote files = %q", files)
	}

	res, err = r.Sync(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if res.Committed != 0 || res.Pushed || res.Pulled != 0 {
		t.Errorf("nothing sync = %+v", res)
	}

	otherClonePush(t, remote, "global/style/py.md", "Use black.")
	res, err = r.Sync(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if res.Pulled != 1 {
		t.Errorf("pull sync = %+v", res)
	}
	if _, err := os.Stat(filepath.Join(local, "global/style/py.md")); err != nil {
		t.Errorf("pulled file missing: %v", err)
	}
}

func TestSyncConflict(t *testing.T) {
	local, remote := setup(t, "main")
	r := New(local, fixedRemote(remote))
	ctx := context.Background()

	write(t, local, "global/style/go.md", "Use gofmt.")
	if _, err := r.Sync(ctx); err != nil {
		t.Fatal(err)
	}
	otherClonePush(t, remote, "global/style/go.md", "remote rule")
	write(t, local, "global/style/go.md", "local rule")

	_, err := r.Sync(ctx)
	var ce *ConflictError
	if !errors.As(err, &ce) {
		t.Fatalf("err = %v, want *ConflictError", err)
	}
	if !slices.Equal(ce.Files, []string{"global/style/go.md"}) {
		t.Errorf("files = %v", ce.Files)
	}
	// Rebase aborted: local commit kept, no rebase in progress.
	if _, err := os.Stat(filepath.Join(local, ".git", "rebase-merge")); !os.IsNotExist(err) {
		t.Errorf("rebase still in progress: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(local, "global/style/go.md"))
	if string(data) != "local rule" {
		t.Errorf("local file = %q", data)
	}
}

func TestSyncRenamesUnbornBranchToMain(t *testing.T) {
	local, remote := setup(t, "master")
	write(t, local, "a.md", "x")
	if _, err := New(local, fixedRemote(remote)).Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if b := strings.TrimSpace(run(t, local, "symbolic-ref", "--short", "HEAD")); b != "main" {
		t.Errorf("branch = %q", b)
	}
}

func TestSyncUpdatesRemoteURL(t *testing.T) {
	local, remote := setup(t, "main")
	run(t, local, "remote", "add", "origin", "/old/url.git")
	write(t, local, "a.md", "x")
	if _, err := New(local, fixedRemote(remote)).Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if url := strings.TrimSpace(run(t, local, "remote", "get-url", "origin")); url != remote {
		t.Errorf("origin = %q", url)
	}
}

func TestStatus(t *testing.T) {
	local, _ := setup(t, "main")
	write(t, local, "global/style/go.md", "x")
	st, err := New(local, fixedRemote("")).Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if st.Branch != "main" || st.Remote != "" || !slices.Equal(st.Changes, []string{"?? global/"}) {
		t.Errorf("status = %+v", st)
	}
}

func TestStatusAheadBehind(t *testing.T) {
	local, remote := setup(t, "main")
	r := New(local, fixedRemote(remote))
	write(t, local, "a.md", "x")
	if _, err := r.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	write(t, local, "b.md", "y")
	run(t, local, "add", "-A")
	run(t, local, "commit", "--quiet", "-m", "local")

	st, err := r.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if st.Remote != remote || st.Ahead != 1 || st.Behind != 0 {
		t.Errorf("status = %+v", st)
	}
}

func TestLog(t *testing.T) {
	local, _ := setup(t, "main")
	r := New(local, fixedRemote(""))
	commits, err := r.Log(context.Background(), 10)
	if err != nil || len(commits) != 0 {
		t.Fatalf("empty log = %v, %v", commits, err)
	}

	write(t, local, "a.md", "x")
	if _, err := r.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	commits, err = r.Log(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 1 || commits[0].Message != "pcm sync: 1 file changed" || len(commits[0].Hash) < 7 || commits[0].Date == "" {
		t.Errorf("log = %+v", commits)
	}
}

func TestStatusAheadWithoutUpstream(t *testing.T) {
	local, remote := setup(t, "main")
	write(t, local, "a.md", "x")
	if _, err := New(local, fixedRemote("")).Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	// Remote set after the first local commit: origin/main does not exist yet.
	st, err := New(local, fixedRemote(remote)).Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if st.Ahead != 1 || st.Behind != 0 {
		t.Errorf("status = %+v, want ahead 1", st)
	}
}
