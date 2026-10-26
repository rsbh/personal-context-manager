package cli

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	errNotInRepo  = errors.New("not in a git repo")
	errBothScopes = errors.New("use only one of -p and -g")
)

// scopeFlags are -p and -g, like kubectl -n.
type scopeFlags struct {
	project string
	global  bool
}

func (f *scopeFlags) register(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&f.project, "project", "p", "", "project name (default: current git repo)")
	cmd.Flags().BoolVarP(&f.global, "global", "g", false, "use global scope")
}

// scope is global or one project. detected is true when the project comes from the git repo.
type scope struct {
	global   bool
	project  string
	detected bool
}

func (s scope) String() string {
	if s.global {
		return "global"
	}
	return fmt.Sprintf("project %q", s.project)
}

// api returns the scope and project values for the API.
func (s scope) api() (string, string) {
	if s.global {
		return "global", ""
	}
	return "project", s.project
}

// resolveScope picks the scope from flags, else from the git repo root of cwd.
func resolveScope(f scopeFlags, cwd string) (scope, error) {
	switch {
	case f.global && f.project != "":
		return scope{}, errBothScopes
	case f.global:
		return scope{global: true}, nil
	case f.project != "":
		return scope{project: f.project}, nil
	}
	root, err := gitRoot(cwd)
	if err != nil {
		return scope{}, errNotInRepo
	}
	name := normalizeName(filepath.Base(root))
	if name == "" {
		return scope{}, fmt.Errorf("repo name %q is not a valid project name, use -p <project>", filepath.Base(root))
	}
	return scope{project: name, detected: true}, nil
}

func gitRoot(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

var (
	invalidChars = regexp.MustCompile(`[^a-z0-9_-]+`)
	dashes       = regexp.MustCompile(`-+`)
)

// normalizeName makes a valid pcm name: My.Repo -> my-repo.
func normalizeName(s string) string {
	s = invalidChars.ReplaceAllString(strings.ToLower(s), "-")
	s = dashes.ReplaceAllString(s, "-")
	return strings.Trim(s, "-_")
}

// scope resolves the scope for cmd. It prints a hint when not in a git repo.
func (a *app) scope(cmd *cobra.Command, f scopeFlags) (scope, error) {
	cwd, err := a.workDir()
	if err != nil {
		return scope{}, err
	}
	s, err := resolveScope(f, cwd)
	if errors.Is(err, errNotInRepo) {
		printFields(cmd.ErrOrStderr(), []field{
			{"error", err.Error()},
			{"hint", "use -p <project> or -g"},
		})
		return scope{}, errSilent
	}
	return s, err
}
