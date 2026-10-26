package client

import (
	"context"
	"net/http"
	"testing"
)

func TestSearch(t *testing.T) {
	f := &fakeAPI{status: http.StatusOK, resp: `[{"id":"global/style/go","name":"go","snippet":"Use <mark>gofmt</mark>.","score":1.5}]`}
	got, err := f.start(t).Search(context.Background(), SearchParams{Q: "gofmt fmt", Project: "api", Global: true, Section: "style", Limit: 5})
	if err != nil {
		t.Fatal(err)
	}
	f.expect(t, http.MethodGet, "/api/v1/search", "global=true&limit=5&project=api&q=gofmt+fmt&section=style", "")
	if len(got) != 1 || got[0].ID != "global/style/go" || got[0].Score != 1.5 {
		t.Errorf("hits = %+v", got)
	}
}

func TestSearchMinimal(t *testing.T) {
	for _, tc := range []struct {
		p     SearchParams
		query string
	}{
		{SearchParams{Q: "go"}, "q=go"},
		{SearchParams{Q: "go", Scope: "global"}, "q=go&scope=global"},
	} {
		f := &fakeAPI{status: http.StatusOK, resp: `[]`}
		if _, err := f.start(t).Search(context.Background(), tc.p); err != nil {
			t.Fatal(err)
		}
		f.expect(t, http.MethodGet, "/api/v1/search", tc.query, "")
	}
}

func TestRebuildIndex(t *testing.T) {
	f := &fakeAPI{status: http.StatusOK, resp: `{"indexed":3}`}
	n, err := f.start(t).RebuildIndex(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	f.expect(t, http.MethodPost, "/api/v1/index/rebuild", "", "")
	if n != 3 {
		t.Errorf("indexed = %d, want 3", n)
	}
}
