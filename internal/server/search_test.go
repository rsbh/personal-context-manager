package server

import (
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/rsbh/personal-context-manager/internal/search"
	"github.com/rsbh/personal-context-manager/internal/store"
)

// newSearchServer returns a router with a search index and a global
// section "style" with the context "go".
func newSearchServer(t *testing.T) (http.Handler, string) {
	t.Helper()
	root := t.TempDir()
	st, err := store.New(root)
	if err != nil {
		t.Fatal(err)
	}
	idx, _, err := search.Open(filepath.Join(t.TempDir(), "index"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { idx.Close() })
	h := NewRouter(st, WithSearch(idx))
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/sections", `{"scope":"global","name":"style"}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/contexts",
		`{"scope":"global","section":"style","name":"go","content":"Use gofmt."}`), http.StatusCreated)
	return h, root
}

// hitIDs runs a search and returns the hit ids.
func hitIDs(t *testing.T, h http.Handler, query string) []string {
	t.Helper()
	rec := do(t, h, http.MethodGet, "/api/v1/search?"+query, "")
	expectStatus(t, rec, http.StatusOK)
	hits := decodeBody[[]search.Hit](t, rec)
	ids := make([]string, len(hits))
	for i, hit := range hits {
		ids[i] = hit.ID
	}
	return ids
}

func TestSearchNotConfigured(t *testing.T) {
	h, _ := newTestServer(t)
	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/api/v1/search?q=go"},
		{http.MethodPost, "/api/v1/index/rebuild"},
	} {
		rec := do(t, h, tc.method, tc.path, "")
		expectStatus(t, rec, http.StatusNotImplemented)
		if e := decodeBody[errorResponse](t, rec); e.Error != "search not configured" {
			t.Errorf("%s %s: error = %q", tc.method, tc.path, e.Error)
		}
	}
}

func TestSearchHit(t *testing.T) {
	h, _ := newSearchServer(t)
	rec := do(t, h, http.MethodGet, "/api/v1/search?q=gofmt", "")
	expectStatus(t, rec, http.StatusOK)
	hits := decodeBody[[]search.Hit](t, rec)
	if len(hits) != 1 || hits[0].ID != "global/style/go" || hits[0].Title != "go" {
		t.Errorf("hits = %+v", hits)
	}
}

func TestSearchNoHitIsEmptyArray(t *testing.T) {
	h, _ := newSearchServer(t)
	rec := do(t, h, http.MethodGet, "/api/v1/search?q=zzzqqq", "")
	expectStatus(t, rec, http.StatusOK)
	if got := rec.Body.String(); got != "[]\n" {
		t.Errorf("body = %q, want []", got)
	}
}

func TestSearchBadRequest(t *testing.T) {
	h, _ := newSearchServer(t)
	for path, want := range map[string]string{
		"/api/v1/search":                "missing q",
		"/api/v1/search?q=%20":          "missing q",
		"/api/v1/search?q=go&limit=abc": "invalid limit",
		"/api/v1/search?q=go&limit=0":   "invalid limit",
		"/api/v1/search?q=go&limit=-1":  "invalid limit",
	} {
		rec := do(t, h, http.MethodGet, path, "")
		expectStatus(t, rec, http.StatusBadRequest)
		if e := decodeBody[errorResponse](t, rec); e.Error != want {
			t.Errorf("%s: error = %q, want %q", path, e.Error, want)
		}
	}
}

func TestSearchFilters(t *testing.T) {
	h, _ := newSearchServer(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/projects", `{"name":"api"}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/sections", `{"scope":"project","project":"api","name":"style"}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/contexts",
		`{"scope":"project","project":"api","section":"style","name":"go","content":"Run gofmt."}`), http.StatusCreated)

	if got := hitIDs(t, h, "q=gofmt&project=api"); !slices.Equal(got, []string{"projects/api/style/go"}) {
		t.Errorf("project=api: %v", got)
	}
	if got := hitIDs(t, h, "q=gofmt&scope=global&section=style"); !slices.Equal(got, []string{"global/style/go"}) {
		t.Errorf("scope=global: %v", got)
	}
	if got := hitIDs(t, h, "q=gofmt&limit=1"); len(got) != 1 {
		t.Errorf("limit=1: %v", got)
	}
	got := hitIDs(t, h, "q=gofmt&project=api&global=true")
	slices.Sort(got)
	if want := []string{"global/style/go", "projects/api/style/go"}; !slices.Equal(got, want) {
		t.Errorf("project=api&global=true: %v, want %v", got, want)
	}
}

func TestSearchFollowsWrites(t *testing.T) {
	h, _ := newSearchServer(t)

	expectStatus(t, do(t, h, http.MethodPut, "/api/v1/contexts/global/style/go", `{"content":"Use goimports."}`), http.StatusOK)
	if got := hitIDs(t, h, "q=gofmt"); len(got) != 0 {
		t.Errorf("after update, old word: %v", got)
	}
	if got := hitIDs(t, h, "q=goimports"); len(got) != 1 {
		t.Errorf("after update, new word: %v", got)
	}

	expectStatus(t, do(t, h, http.MethodDelete, "/api/v1/contexts/global/style/go", ""), http.StatusNoContent)
	if got := hitIDs(t, h, "q=goimports"); len(got) != 0 {
		t.Errorf("after delete: %v", got)
	}
}

func TestSearchAfterForceDelete(t *testing.T) {
	h, _ := newSearchServer(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/projects", `{"name":"api"}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/sections", `{"scope":"project","project":"api","name":"style"}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/contexts",
		`{"scope":"project","project":"api","section":"style","name":"go","content":"Run gofmt."}`), http.StatusCreated)

	expectStatus(t, do(t, h, http.MethodDelete, "/api/v1/projects/api?force=true", ""), http.StatusNoContent)
	if got := hitIDs(t, h, "q=gofmt"); !slices.Equal(got, []string{"global/style/go"}) {
		t.Errorf("after project delete: %v", got)
	}

	expectStatus(t, do(t, h, http.MethodDelete, "/api/v1/sections/style?scope=global&force=true", ""), http.StatusNoContent)
	if got := hitIDs(t, h, "q=gofmt"); len(got) != 0 {
		t.Errorf("after section delete: %v", got)
	}
}

func TestIndexRebuild(t *testing.T) {
	h, root := newSearchServer(t)
	if err := os.WriteFile(filepath.Join(root, "global", "style", "rs.md"), []byte("Use rustfmt."), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := hitIDs(t, h, "q=rustfmt"); len(got) != 0 {
		t.Errorf("before rebuild: %v", got)
	}

	rec := do(t, h, http.MethodPost, "/api/v1/index/rebuild", "")
	expectStatus(t, rec, http.StatusOK)
	if got := decodeBody[rebuildResponse](t, rec); got.Indexed != 2 {
		t.Errorf("indexed = %d, want 2", got.Indexed)
	}
	if got := hitIDs(t, h, "q=rustfmt"); !slices.Equal(got, []string{"global/style/rs"}) {
		t.Errorf("after rebuild: %v", got)
	}
}
