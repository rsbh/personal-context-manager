package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// fakeAPI answers one request and records it.
type fakeAPI struct {
	method, path, query, body string
	status                    int
	resp                      string
}

func (f *fakeAPI) start(t *testing.T) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		f.method, f.path, f.query, f.body = r.Method, r.URL.Path, r.URL.RawQuery, string(b)
		w.WriteHeader(f.status)
		_, _ = io.WriteString(w, f.resp)
	}))
	t.Cleanup(srv.Close)
	return New(srv.URL)
}

func (f *fakeAPI) expect(t *testing.T, method, path, query, body string) {
	t.Helper()
	if f.method != method || f.path != path || f.query != query || f.body != body {
		t.Errorf("request = %s %s?%s %q, want %s %s?%s %q", f.method, f.path, f.query, f.body, method, path, query, body)
	}
}

func TestListProjects(t *testing.T) {
	f := &fakeAPI{status: http.StatusOK, resp: `["api","web"]`}
	got, err := f.start(t).ListProjects(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	f.expect(t, http.MethodGet, "/api/v1/projects", "", "")
	if want := []string{"api", "web"}; !reflect.DeepEqual(got, want) {
		t.Errorf("projects = %v, want %v", got, want)
	}
}

func TestGetProject(t *testing.T) {
	f := &fakeAPI{status: http.StatusOK, resp: `{"name":"api"}`}
	got, err := f.start(t).GetProject(context.Background(), "api")
	if err != nil {
		t.Fatal(err)
	}
	f.expect(t, http.MethodGet, "/api/v1/projects/api", "", "")
	if got != "api" {
		t.Errorf("project = %q, want api", got)
	}
}

func TestCreateProject(t *testing.T) {
	f := &fakeAPI{status: http.StatusCreated, resp: `{"name":"api"}`}
	if err := f.start(t).CreateProject(context.Background(), "api"); err != nil {
		t.Fatal(err)
	}
	f.expect(t, http.MethodPost, "/api/v1/projects", "", `{"name":"api"}`+"\n")
}

func TestDeleteProject(t *testing.T) {
	for _, force := range []bool{false, true} {
		f := &fakeAPI{status: http.StatusNoContent}
		if err := f.start(t).DeleteProject(context.Background(), "api", force); err != nil {
			t.Fatal(err)
		}
		query := ""
		if force {
			query = "force=true"
		}
		f.expect(t, http.MethodDelete, "/api/v1/projects/api", query, "")
	}
}

func TestAPIError(t *testing.T) {
	f := &fakeAPI{status: http.StatusConflict, resp: `{"error":"project \"api\": already exists"}`}
	err := f.start(t).CreateProject(context.Background(), "api")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *APIError", err)
	}
	if apiErr.Status != http.StatusConflict || apiErr.Message != `project "api": already exists` {
		t.Errorf("APIError = %+v", apiErr)
	}
	if err.Error() != `project "api": already exists` {
		t.Errorf("Error() = %q", err.Error())
	}
}

func TestNotReachable(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()

	_, err := New(url).ListProjects(context.Background())
	if !errors.Is(err, ErrNotReachable) {
		t.Errorf("err = %v, want ErrNotReachable", err)
	}
}
