package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rsbh/personal-context-manager/internal/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.New(t.TempDir())
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	return st
}

func newTestServer(t *testing.T) (http.Handler, string) {
	t.Helper()
	root := t.TempDir()
	st, err := store.New(root)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	return NewRouter(st), root
}

func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, r))
	return rec
}

func expectStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, want, rec.Body.String())
	}
}

func decodeBody[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode body %q: %v", rec.Body.String(), err)
	}
	return v
}

func TestHealth(t *testing.T) {
	h, _ := newTestServer(t)
	rec := do(t, h, http.MethodGet, "/health", "")
	expectStatus(t, rec, http.StatusOK)
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"status":"ok"}` {
		t.Errorf("body = %q, want %q", got, `{"status":"ok"}`)
	}
}

func TestNewRunsGitInit(t *testing.T) {
	_, root := newTestServer(t)
	if _, err := os.Stat(filepath.Join(root, ".git")); err != nil {
		t.Fatalf(".git missing: %v", err)
	}
}

func TestProjects(t *testing.T) {
	h, root := newTestServer(t)

	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/projects", `{"name":"pcm"}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/projects", `{"name":"pcm"}`), http.StatusConflict)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/projects", `{"name":"../x"}`), http.StatusBadRequest)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/projects", `{bad`), http.StatusBadRequest)

	if _, err := os.Stat(filepath.Join(root, "projects", "pcm", ".gitkeep")); err != nil {
		t.Errorf(".gitkeep missing: %v", err)
	}

	rec := do(t, h, http.MethodGet, "/api/v1/projects", "")
	expectStatus(t, rec, http.StatusOK)
	if got := decodeBody[[]string](t, rec); len(got) != 1 || got[0] != "pcm" {
		t.Errorf("projects = %v, want [pcm]", got)
	}

	rec = do(t, h, http.MethodGet, "/api/v1/projects/pcm", "")
	expectStatus(t, rec, http.StatusOK)
	if got := decodeBody[map[string]string](t, rec); got["name"] != "pcm" {
		t.Errorf("project = %v, want name pcm", got)
	}
	expectStatus(t, do(t, h, http.MethodGet, "/api/v1/projects/none", ""), http.StatusNotFound)
	expectStatus(t, do(t, h, http.MethodGet, "/api/v1/projects/Bad", ""), http.StatusBadRequest)

	// Not empty: needs force.
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/sections", `{"scope":"project","project":"pcm","name":"style"}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodDelete, "/api/v1/projects/pcm", ""), http.StatusConflict)
	expectStatus(t, do(t, h, http.MethodDelete, "/api/v1/projects/pcm?force=true", ""), http.StatusNoContent)
	expectStatus(t, do(t, h, http.MethodDelete, "/api/v1/projects/pcm", ""), http.StatusNotFound)
}

func TestSections(t *testing.T) {
	h, _ := newTestServer(t)

	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/sections", `{"scope":"global","name":"style"}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/sections", `{"scope":"global","name":"style"}`), http.StatusConflict)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/sections", `{"scope":"bad","name":"style"}`), http.StatusBadRequest)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/sections", `{"scope":"project","project":"none","name":"style"}`), http.StatusNotFound)

	// Per scope: project section list is separate from global.
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/projects", `{"name":"pcm"}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/sections", `{"scope":"project","project":"pcm","name":"review"}`), http.StatusCreated)

	rec := do(t, h, http.MethodGet, "/api/v1/sections?scope=global", "")
	expectStatus(t, rec, http.StatusOK)
	if got := decodeBody[[]string](t, rec); len(got) != 1 || got[0] != "style" {
		t.Errorf("global sections = %v, want [style]", got)
	}
	rec = do(t, h, http.MethodGet, "/api/v1/sections?scope=project&project=pcm", "")
	expectStatus(t, rec, http.StatusOK)
	if got := decodeBody[[]string](t, rec); len(got) != 1 || got[0] != "review" {
		t.Errorf("project sections = %v, want [review]", got)
	}

	expectStatus(t, do(t, h, http.MethodDelete, "/api/v1/sections/style?scope=global", ""), http.StatusNoContent)
	expectStatus(t, do(t, h, http.MethodDelete, "/api/v1/sections/style?scope=global", ""), http.StatusNotFound)
}

func TestContexts(t *testing.T) {
	h, root := newTestServer(t)

	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/sections", `{"scope":"global","name":"style"}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/projects", `{"name":"pcm"}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/sections", `{"scope":"project","project":"pcm","name":"style"}`), http.StatusCreated)

	// Create.
	body := `{"scope":"global","section":"style","name":"backend",
		"frontmatter":{"title":"Backend","description":"Go style","tags":["go"]},
		"content":"# Go\nUse gofmt.\n"}`
	rec := do(t, h, http.MethodPost, "/api/v1/contexts", body)
	expectStatus(t, rec, http.StatusCreated)
	c := decodeBody[store.Context](t, rec)
	if c.ID != "global/style/backend" || c.Frontmatter.Title != "Backend" || c.Content != "# Go\nUse gofmt.\n" {
		t.Errorf("created = %+v", c)
	}
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/contexts", body), http.StatusConflict)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/contexts",
		`{"scope":"global","section":"none","name":"x","content":""}`), http.StatusNotFound)

	// File on disk has frontmatter.
	data, err := os.ReadFile(filepath.Join(root, "global", "style", "backend.md"))
	if err != nil {
		t.Fatal(err)
	}
	want := "---\ntitle: Backend\ndescription: Go style\ntags:\n  - go\n---\n# Go\nUse gofmt.\n"
	if string(data) != want {
		t.Errorf("file = %q", data)
	}

	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/contexts",
		`{"scope":"project","project":"pcm","section":"style","name":"ui","content":"React"}`), http.StatusCreated)

	// Get with raw slashes and with %2F.
	for _, path := range []string{"/api/v1/contexts/global/style/backend", "/api/v1/contexts/global%2Fstyle%2Fbackend"} {
		rec = do(t, h, http.MethodGet, path, "")
		expectStatus(t, rec, http.StatusOK)
		if got := decodeBody[store.Context](t, rec); got.Content != "# Go\nUse gofmt.\n" || len(got.Frontmatter.Tags) != 1 {
			t.Errorf("get %s = %+v", path, got)
		}
	}
	expectStatus(t, do(t, h, http.MethodGet, "/api/v1/contexts/global/style/none", ""), http.StatusNotFound)
	expectStatus(t, do(t, h, http.MethodGet, "/api/v1/contexts/global/../../etc", ""), http.StatusBadRequest)

	// List with filters. No content in list.
	rec = do(t, h, http.MethodGet, "/api/v1/contexts", "")
	expectStatus(t, rec, http.StatusOK)
	list := decodeBody[[]store.Context](t, rec)
	if len(list) != 2 || list[0].Content != "" {
		t.Errorf("list = %+v", list)
	}
	rec = do(t, h, http.MethodGet, "/api/v1/contexts?scope=project&project=pcm", "")
	expectStatus(t, rec, http.StatusOK)
	if list := decodeBody[[]store.Context](t, rec); len(list) != 1 || list[0].ID != "projects/pcm/style/ui" {
		t.Errorf("filtered list = %+v", list)
	}

	// Update. Empty frontmatter removes the header.
	rec = do(t, h, http.MethodPut, "/api/v1/contexts/global/style/backend", `{"content":"new"}`)
	expectStatus(t, rec, http.StatusOK)
	if got := decodeBody[store.Context](t, rec); got.Content != "new" || got.Frontmatter.Title != "" {
		t.Errorf("updated = %+v", got)
	}
	expectStatus(t, do(t, h, http.MethodPut, "/api/v1/contexts/global/style/none", `{"content":"x"}`), http.StatusNotFound)

	// Delete.
	expectStatus(t, do(t, h, http.MethodDelete, "/api/v1/contexts/global/style/backend", ""), http.StatusNoContent)
	expectStatus(t, do(t, h, http.MethodDelete, "/api/v1/contexts/global/style/backend", ""), http.StatusNotFound)
}

func TestResolve(t *testing.T) {
	h, _ := newTestServer(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/sections", `{"scope":"global","name":"style"}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/projects", `{"name":"api"}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/sections", `{"scope":"project","project":"api","name":"style"}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/contexts", `{"scope":"global","section":"style","name":"go","content":"global"}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/contexts", `{"scope":"global","section":"style","name":"py","content":"py"}`), http.StatusCreated)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/contexts", `{"scope":"project","project":"api","section":"style","name":"go","content":"api"}`), http.StatusCreated)

	rec := do(t, h, http.MethodGet, "/api/v1/resolve?project=api&section=style", "")
	expectStatus(t, rec, http.StatusOK)
	got := decodeBody[[]store.Context](t, rec)
	if len(got) != 2 || got[0].ID != "global/style/py" || got[1].ID != "projects/api/style/go" || got[1].Content != "api" {
		t.Errorf("resolve = %+v", got)
	}

	rec = do(t, h, http.MethodGet, "/api/v1/resolve", "")
	expectStatus(t, rec, http.StatusOK)
	if got := decodeBody[[]store.Context](t, rec); len(got) != 2 {
		t.Errorf("global resolve = %+v", got)
	}

	expectStatus(t, do(t, h, http.MethodGet, "/api/v1/resolve?project=Bad", ""), http.StatusBadRequest)
}

func TestResolveFormats(t *testing.T) {
	h, _ := newTestServer(t)
	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/sections", `{"scope":"global","name":"style"}`), http.StatusCreated)

	get := func(path, accept string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if accept != "" {
			req.Header.Set("Accept", accept)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}
	const md = "text/markdown; charset=utf-8"

	// Empty markdown.
	rec := get("/api/v1/resolve?format=md", "")
	expectStatus(t, rec, http.StatusOK)
	if rec.Body.String() != "No contexts found.\n" {
		t.Errorf("empty body = %q", rec.Body.String())
	}

	expectStatus(t, do(t, h, http.MethodPost, "/api/v1/contexts", `{"scope":"global","section":"style","name":"go","content":"Use gofmt."}`), http.StatusCreated)

	tests := []struct {
		name, path, accept string
		wantStatus         int
		wantType           string
		wantBody           string
	}{
		{"default json", "/api/v1/resolve", "", http.StatusOK, "application/json", `"id":"global/style/go"`},
		{"format md", "/api/v1/resolve?format=md", "", http.StatusOK, md, "# Context\n\n## style/go (global)\n\nUse gofmt.\n"},
		{"accept md", "/api/v1/resolve", "text/markdown", http.StatusOK, md, "## style/go (global)"},
		{"accept list md", "/api/v1/resolve", "text/plain, text/markdown;q=0.9", http.StatusOK, md, "## style/go"},
		{"format json wins", "/api/v1/resolve?format=json", "text/markdown", http.StatusOK, "application/json", `"id"`},
		{"invalid format", "/api/v1/resolve?format=xml", "", http.StatusBadRequest, "application/json", `invalid format \"xml\": use json or md`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := get(tt.path, tt.accept)
			expectStatus(t, rec, tt.wantStatus)
			if got := rec.Header().Get("Content-Type"); got != tt.wantType {
				t.Errorf("Content-Type = %q, want %q", got, tt.wantType)
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("body = %q, want contains %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

func httpRecorder() *httptest.ResponseRecorder { return httptest.NewRecorder() }
