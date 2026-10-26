package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAll(t *testing.T) {
	root := t.TempDir()
	st, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.CreateSection(ScopeGlobal, "", "style"); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Create(Context{Scope: ScopeGlobal, Section: "style", Name: "go",
		Frontmatter: Frontmatter{Title: "Go"}, Content: "Use gofmt.\n"}); err != nil {
		t.Fatal(err)
	}
	// A file written outside pcm, without frontmatter.
	if err := os.WriteFile(filepath.Join(root, "global", "style", "rs.md"), []byte("Use rustfmt.\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	all, err := st.All()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("All = %d contexts, want 2", len(all))
	}
	if all[0].ID != "global/style/go" || all[0].Frontmatter.Title != "Go" || all[0].Content != "Use gofmt.\n" {
		t.Errorf("all[0] = %+v", all[0])
	}
	if all[1].ID != "global/style/rs" || all[1].Content != "Use rustfmt.\n" {
		t.Errorf("all[1] = %+v", all[1])
	}
}
