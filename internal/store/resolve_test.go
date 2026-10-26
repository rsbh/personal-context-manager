package store

import "testing"

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func mustCreate(t *testing.T, s *Store, c Context) {
	t.Helper()
	if _, err := s.Create(c); err != nil {
		t.Fatalf("create %+v: %v", c, err)
	}
}

func ids(cs []Context) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.ID
	}
	return out
}

func TestResolve(t *testing.T) {
	s := newTestStore(t)
	for _, sec := range []string{"review", "style"} {
		if err := s.CreateSection(ScopeGlobal, "", sec); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.CreateProject("api"); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateSection(ScopeProject, "api", "style"); err != nil {
		t.Fatal(err)
	}
	mustCreate(t, s, Context{Scope: ScopeGlobal, Section: "style", Name: "go", Content: "global go"})
	mustCreate(t, s, Context{Scope: ScopeGlobal, Section: "style", Name: "py", Content: "global py"})
	mustCreate(t, s, Context{Scope: ScopeGlobal, Section: "review", Name: "pr", Content: "global pr"})
	mustCreate(t, s, Context{Scope: ScopeProject, Project: "api", Section: "style", Name: "go", Content: "api go"})

	tests := []struct {
		name             string
		project, section string
		want             []string
	}{
		{"global only", "", "", []string{"global/review/pr", "global/style/go", "global/style/py"}},
		{"project overrides global", "api", "", []string{"global/review/pr", "global/style/py", "projects/api/style/go"}},
		{"section filter", "api", "style", []string{"global/style/py", "projects/api/style/go"}},
		{"missing project gives global", "none", "", []string{"global/review/pr", "global/style/go", "global/style/py"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.Resolve(tt.project, tt.section)
			if err != nil {
				t.Fatal(err)
			}
			if g := ids(got); len(g) != len(tt.want) || !equal(g, tt.want) {
				t.Errorf("ids = %v, want %v", g, tt.want)
			}
			for _, c := range got {
				if c.Content == "" {
					t.Errorf("%s: content is empty", c.ID)
				}
			}
		})
	}
}

func TestResolveInvalidName(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.Resolve("Bad/Name", ""); err == nil {
		t.Error("want error for invalid project")
	}
	if _, err := s.Resolve("", "Bad"); err == nil {
		t.Error("want error for invalid section")
	}
}

func equal(a, b []string) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
