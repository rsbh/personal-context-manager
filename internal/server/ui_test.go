package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func uiFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html":    {Data: []byte("<!doctype html><title>pcm</title>")},
		"assets/app.js": {Data: []byte("console.log(1)")},
	}
}

func get(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestServeUI(t *testing.T) {
	st := newTestStore(t)
	h := NewRouter(st, WithUI(uiFS()))

	tests := []struct {
		path, wantType, wantBody string
	}{
		{"/", "text/html; charset=utf-8", "<title>pcm</title>"},
		{"/contexts/global/git/commits", "text/html; charset=utf-8", "<title>pcm</title>"},
		{"/assets/app.js", "text/javascript; charset=utf-8", "console.log(1)"},
	}
	for _, tt := range tests {
		rec := get(t, h, tt.path)
		expectStatus(t, rec, http.StatusOK)
		if got := rec.Header().Get("Content-Type"); got != tt.wantType {
			t.Errorf("%s Content-Type = %q, want %q", tt.path, got, tt.wantType)
		}
		if !strings.Contains(rec.Body.String(), tt.wantBody) {
			t.Errorf("%s body = %q", tt.path, rec.Body.String())
		}
	}

	// API stays JSON.
	for _, path := range []string{"/api/v1/nope", "/api/other"} {
		rec := get(t, h, path)
		expectStatus(t, rec, http.StatusNotFound)
		if got := rec.Header().Get("Content-Type"); got != "application/json" {
			t.Errorf("%s Content-Type = %q", path, got)
		}
	}
	expectStatus(t, get(t, h, "/health"), http.StatusOK)
}

func TestServeUIPlaceholder(t *testing.T) {
	h := NewRouter(newTestStore(t), WithUI(fstest.MapFS{}))
	rec := get(t, h, "/")
	expectStatus(t, rec, http.StatusOK)
	if !strings.Contains(rec.Body.String(), "UI not built") || rec.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("placeholder = %q", rec.Body.String())
	}
}

func TestNoUIIsJSON404(t *testing.T) {
	h := NewRouter(newTestStore(t))
	rec := get(t, h, "/")
	expectStatus(t, rec, http.StatusNotFound)
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q", rec.Header().Get("Content-Type"))
	}
}

func TestListenAddr(t *testing.T) {
	if got := ListenAddr(8080); got != "127.0.0.1:8080" {
		t.Errorf("ListenAddr = %q", got)
	}
}
