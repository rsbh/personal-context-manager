// Package gitsync syncs the context repo with a git remote. It runs the git CLI,
// so it uses the user's git config, SSH agent and credential helpers.
package gitsync

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

const branchMain = "main"

// ConflictError means pull found a conflict. The rebase is aborted.
type ConflictError struct {
	Files []string
}

func (e *ConflictError) Error() string { return "sync conflict, rebase aborted" }

// Status is the git state of the repo.
type Status struct {
	Branch  string   `json:"branch"`
	Remote  string   `json:"remote"`
	Changes []string `json:"changes"` // git status --porcelain lines
	Ahead   int      `json:"ahead"`
	Behind  int      `json:"behind"`
}

// SyncResult is the result of a Sync.
type SyncResult struct {
	Committed int    `json:"committed"` // files
	Pulled    int    `json:"pulled"`    // commits
	Pushed    bool   `json:"pushed"`
	Remote    string `json:"remote"`
}

// Commit is one entry of the git log.
type Commit struct {
	Hash    string `json:"hash"`
	Date    string `json:"date"`
	Message string `json:"message"`
}

// Repo runs git operations on the context repo. Calls are serialized.
type Repo struct {
	root   string
	remote func() (string, error) // read on each call, so a config change needs no restart
	mu     sync.Mutex
}

// New returns a Repo for the git repo at root. remote returns the configured remote URL.
func New(root string, remote func() (string, error)) *Repo {
	return &Repo{root: root, remote: remote}
}

// Sync commits all changes, pulls with rebase, then pushes.
// Without a remote it only commits.
func (r *Repo) Sync(ctx context.Context) (SyncResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	remote, err := r.remote()
	if err != nil {
		return SyncResult{}, err
	}
	res := SyncResult{Remote: remote}
	branch, err := r.ensureMain(ctx)
	if err != nil {
		return res, err
	}

	if _, err := r.git(ctx, "add", "-A"); err != nil {
		return res, err
	}
	staged, err := r.git(ctx, "diff", "--cached", "--name-only")
	if err != nil {
		return res, err
	}
	if res.Committed = len(lines(staged)); res.Committed > 0 {
		if _, err := r.git(ctx, "commit", "--quiet", "-m", commitMessage(res.Committed)); err != nil {
			return res, err
		}
	}
	if remote == "" {
		return res, nil
	}

	if err := r.setOrigin(ctx, remote); err != nil {
		return res, err
	}
	if _, err := r.git(ctx, "fetch", "--quiet", "origin"); err != nil {
		return res, err
	}
	upstream := "origin/" + branch
	remoteHasBranch := r.ok(ctx, "rev-parse", "--verify", "--quiet", "refs/remotes/"+upstream)

	if remoteHasBranch {
		if !r.hasCommits(ctx) {
			// Nothing local yet: take the remote branch.
			if _, err := r.git(ctx, "checkout", "--quiet", "-B", branch, upstream); err != nil {
				return res, err
			}
			res.Pulled, err = r.count(ctx, "HEAD")
			return res, err
		}
		if res.Pulled, err = r.count(ctx, "HEAD.."+upstream); err != nil {
			return res, err
		}
		if res.Pulled > 0 {
			if _, err := r.git(ctx, "rebase", "--quiet", upstream); err != nil {
				return res, r.abortRebase(ctx)
			}
		}
	}

	if !r.hasCommits(ctx) {
		return res, nil
	}
	ahead := "HEAD"
	if remoteHasBranch {
		ahead = upstream + "..HEAD"
	}
	n, err := r.count(ctx, ahead)
	if err != nil {
		return res, err
	}
	if n > 0 {
		if _, err := r.git(ctx, "push", "--quiet", "-u", "origin", branch); err != nil {
			return res, err
		}
		res.Pushed = true
	}
	return res, nil
}

// Status is local only: ahead and behind are from the last fetch.
func (r *Repo) Status(ctx context.Context) (Status, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	remote, err := r.remote()
	if err != nil {
		return Status{}, err
	}
	branch, err := r.ensureMain(ctx)
	if err != nil {
		return Status{}, err
	}
	st := Status{Branch: branch, Remote: remote, Changes: []string{}}
	porcelain, err := r.git(ctx, "status", "--porcelain")
	if err != nil {
		return st, err
	}
	st.Changes = append(st.Changes, lines(porcelain)...)

	if remote == "" || !r.hasCommits(ctx) {
		return st, nil
	}
	upstream := "origin/" + branch
	if !r.ok(ctx, "rev-parse", "--verify", "--quiet", "refs/remotes/"+upstream) {
		// Never pushed: all local commits are ahead.
		st.Ahead, err = r.count(ctx, "HEAD")
		return st, err
	}
	out, err := r.git(ctx, "rev-list", "--left-right", "--count", "HEAD..."+upstream)
	if err != nil {
		return st, err
	}
	if f := strings.Fields(out); len(f) == 2 {
		st.Ahead, _ = strconv.Atoi(f[0])
		st.Behind, _ = strconv.Atoi(f[1])
	}
	return st, nil
}

// Log returns the last n commits, newest first. No commits gives an empty list.
func (r *Repo) Log(ctx context.Context, n int) ([]Commit, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := []Commit{}
	if !r.hasCommits(ctx) {
		return out, nil
	}
	log, err := r.git(ctx, "log", "-n", strconv.Itoa(n), "--date=format:%Y-%m-%d %H:%M", "--format=%h%x09%ad%x09%s")
	if err != nil {
		return nil, err
	}
	for _, line := range lines(log) {
		if f := strings.SplitN(line, "\t", 3); len(f) == 3 {
			out = append(out, Commit{Hash: f[0], Date: f[1], Message: f[2]})
		}
	}
	return out, nil
}

// ensureMain names an unborn branch main. It returns the current branch.
func (r *Repo) ensureMain(ctx context.Context) (string, error) {
	out, err := r.git(ctx, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	branch := strings.TrimSpace(out)
	if branch != branchMain && !r.hasCommits(ctx) {
		if _, err := r.git(ctx, "symbolic-ref", "HEAD", "refs/heads/"+branchMain); err != nil {
			return "", err
		}
		branch = branchMain
	}
	return branch, nil
}

func (r *Repo) setOrigin(ctx context.Context, url string) error {
	current, err := r.git(ctx, "remote", "get-url", "origin")
	switch {
	case err != nil:
		_, err = r.git(ctx, "remote", "add", "origin", url)
	case strings.TrimSpace(current) != url:
		_, err = r.git(ctx, "remote", "set-url", "origin", url)
	}
	return err
}

func (r *Repo) abortRebase(ctx context.Context) error {
	out, _ := r.git(ctx, "diff", "--name-only", "--diff-filter=U")
	files := lines(out)
	if _, err := r.git(ctx, "rebase", "--abort"); err != nil {
		return err
	}
	return &ConflictError{Files: files}
}

func (r *Repo) hasCommits(ctx context.Context) bool {
	return r.ok(ctx, "rev-parse", "--verify", "--quiet", "HEAD")
}

func (r *Repo) count(ctx context.Context, rng string) (int, error) {
	out, err := r.git(ctx, "rev-list", "--count", rng)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(out))
}

func (r *Repo) ok(ctx context.Context, args ...string) bool {
	_, err := r.git(ctx, args...)
	return err == nil
}

// git runs git in the repo and returns stdout. The error has stderr.
func (r *Repo) git(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.root
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func lines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimRight(l, "\r"); strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}

func commitMessage(files int) string {
	if files == 1 {
		return "pcm sync: 1 file changed"
	}
	return fmt.Sprintf("pcm sync: %d files changed", files)
}
