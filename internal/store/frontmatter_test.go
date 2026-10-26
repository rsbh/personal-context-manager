package store

import "testing"

func TestRender(t *testing.T) {
	got, err := Render(Frontmatter{Title: "Go", Tags: []string{"go", "fmt"}}, "Use gofmt.\n")
	if err != nil {
		t.Fatal(err)
	}
	want := "---\ntitle: Go\ntags:\n  - go\n  - fmt\n---\nUse gofmt.\n"
	if string(got) != want {
		t.Errorf("Render = %q, want %q", got, want)
	}

	got, err = Render(Frontmatter{}, "plain")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "plain" {
		t.Errorf("Render without frontmatter = %q", got)
	}
}
