package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/rsbh/personal-context-manager/internal/store"
)

func TestContextID(t *testing.T) {
	if got := ContextID("global", "", "style", "go"); got != "global/style/go" {
		t.Errorf("global id = %q", got)
	}
	if got := ContextID("project", "api", "style", "go"); got != "projects/api/style/go" {
		t.Errorf("project id = %q", got)
	}
}

func TestListContexts(t *testing.T) {
	f := &fakeAPI{status: http.StatusOK, resp: `[{"id":"global/style/go","scope":"global","section":"style","name":"go","frontmatter":{"title":"Go"}}]`}
	got, err := f.start(t).ListContexts(context.Background(), "global", "", "style")
	if err != nil {
		t.Fatal(err)
	}
	f.expect(t, http.MethodGet, "/api/v1/contexts", "scope=global&section=style", "")
	if len(got) != 1 || got[0].Name != "go" || got[0].Frontmatter.Title != "Go" {
		t.Errorf("contexts = %+v", got)
	}
}

func TestListContextsNoSection(t *testing.T) {
	f := &fakeAPI{status: http.StatusOK, resp: `[]`}
	if _, err := f.start(t).ListContexts(context.Background(), "project", "api", ""); err != nil {
		t.Fatal(err)
	}
	f.expect(t, http.MethodGet, "/api/v1/contexts", "project=api&scope=project", "")
}

func TestGetContext(t *testing.T) {
	f := &fakeAPI{status: http.StatusOK, resp: `{"id":"global/style/go","name":"go","content":"Use gofmt."}`}
	got, err := f.start(t).GetContext(context.Background(), "global/style/go")
	if err != nil {
		t.Fatal(err)
	}
	f.expect(t, http.MethodGet, "/api/v1/contexts/global/style/go", "", "")
	if got.Content != "Use gofmt." {
		t.Errorf("context = %+v", got)
	}
}

func TestCreateContext(t *testing.T) {
	f := &fakeAPI{status: http.StatusCreated, resp: `{"id":"global/style/go"}`}
	_, err := f.start(t).CreateContext(context.Background(), store.Context{
		Scope:       "global",
		Section:     "style",
		Name:        "go",
		Frontmatter: store.Frontmatter{Title: "Go", Tags: []string{"go"}},
		Content:     "Use gofmt.",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := `{"scope":"global","section":"style","name":"go","frontmatter":{"title":"Go","tags":["go"]},"content":"Use gofmt."}` + "\n"
	f.expect(t, http.MethodPost, "/api/v1/contexts", "", want)
}

func TestDeleteContext(t *testing.T) {
	f := &fakeAPI{status: http.StatusNoContent}
	if err := f.start(t).DeleteContext(context.Background(), "projects/api/style/go"); err != nil {
		t.Fatal(err)
	}
	f.expect(t, http.MethodDelete, "/api/v1/contexts/projects/api/style/go", "", "")
}

func TestResolve(t *testing.T) {
	f := &fakeAPI{status: http.StatusOK, resp: `[{"id":"global/style/go","content":"x"}]`}
	got, err := f.start(t).Resolve(context.Background(), "api", "style")
	if err != nil {
		t.Fatal(err)
	}
	f.expect(t, http.MethodGet, "/api/v1/resolve", "project=api&section=style", "")
	if len(got) != 1 || got[0].Content != "x" {
		t.Errorf("resolve = %+v", got)
	}

	f = &fakeAPI{status: http.StatusOK, resp: `[]`}
	if _, err := f.start(t).Resolve(context.Background(), "", ""); err != nil {
		t.Fatal(err)
	}
	f.expect(t, http.MethodGet, "/api/v1/resolve", "", "")
}
