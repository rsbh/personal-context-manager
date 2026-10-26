package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestNormalizeName(t *testing.T) {
	tests := map[string]string{
		"api":           "api",
		"My.Repo":       "my-repo",
		"personal_ctx":  "personal_ctx",
		"a  b..c":       "a-b-c",
		"-.Leading.-":   "leading",
		"_under":        "under",
		"ünicode-repo":  "nicode-repo",
		"...":           "",
		"Project-2026!": "project-2026",
	}
	for in, want := range tests {
		if got := normalizeName(in); got != want {
			t.Errorf("normalizeName(%q) = %q, want %q", in, got, want)
		}
	}
}

// gitRepo creates a git repo folder named name and returns its path.
func gitRepo(t *testing.T, name string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "init", "--quiet")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	return dir
}

func TestResolveScope(t *testing.T) {
	repo := gitRepo(t, "My.Repo")
	plain := t.TempDir()

	tests := []struct {
		name    string
		flags   scopeFlags
		cwd     string
		want    scope
		wantErr error
	}{
		{"global", scopeFlags{global: true}, plain, scope{global: true}, nil},
		{"project flag", scopeFlags{project: "api"}, plain, scope{project: "api"}, nil},
		{"detect from root", scopeFlags{}, repo, scope{project: "my-repo", detected: true}, nil},
		{"detect from sub dir", scopeFlags{}, filepath.Join(repo, "sub"), scope{project: "my-repo", detected: true}, nil},
		{"not in git repo", scopeFlags{}, plain, scope{}, errNotInRepo},
		{"both flags", scopeFlags{project: "api", global: true}, plain, scope{}, errBothScopes},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveScope(tt.flags, tt.cwd)
			if err != tt.wantErr {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("scope = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestScopeString(t *testing.T) {
	if got := (scope{global: true}).String(); got != "global" {
		t.Errorf("global = %q", got)
	}
	if got := (scope{project: "api"}).String(); got != `project "api"` {
		t.Errorf("project = %q", got)
	}
}
