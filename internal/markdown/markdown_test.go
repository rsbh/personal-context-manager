package markdown

import (
	"bytes"
	"testing"

	"github.com/rsbh/personal-context-manager/internal/store"
)

func TestWrite(t *testing.T) {
	var buf bytes.Buffer
	Write(&buf, []store.Context{
		{Scope: "global", Section: "style", Name: "go", Frontmatter: store.Frontmatter{Description: "Go style"}, Content: "Use gofmt.\n"},
		{Scope: "project", Project: "api", Section: "review", Name: "pr", Content: "Check errors."},
	})
	want := "# Context\n\n" +
		"## style/go (global)\n> Go style\n\nUse gofmt.\n\n" +
		"## review/pr (project \"api\")\n\nCheck errors.\n"
	if buf.String() != want {
		t.Errorf("out =\n%s\nwant =\n%s", buf.String(), want)
	}
}

func TestWriteEmpty(t *testing.T) {
	var buf bytes.Buffer
	Write(&buf, nil)
	if buf.String() != "No contexts found.\n" {
		t.Errorf("out = %q", buf.String())
	}
}
