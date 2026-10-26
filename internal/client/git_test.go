package client

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"testing"
)

func TestGitStatus(t *testing.T) {
	f := &fakeAPI{status: http.StatusOK, resp: `{"branch":"main","remote":"r","changes":["?? a"],"ahead":1,"behind":2}`}
	st, err := f.start(t).GitStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	f.expect(t, http.MethodGet, "/api/v1/git/status", "", "")
	if st.Branch != "main" || st.Ahead != 1 || st.Behind != 2 || len(st.Changes) != 1 {
		t.Errorf("status = %+v", st)
	}
}

func TestSync(t *testing.T) {
	f := &fakeAPI{status: http.StatusOK, resp: `{"committed":2,"pulled":1,"pushed":true,"remote":"r"}`}
	res, err := f.start(t).Sync(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	f.expect(t, http.MethodPost, "/api/v1/git/sync", "", "")
	if res.Committed != 2 || res.Pulled != 1 || !res.Pushed {
		t.Errorf("result = %+v", res)
	}
}

func TestSyncConflict(t *testing.T) {
	f := &fakeAPI{status: http.StatusConflict, resp: `{"error":"sync conflict, rebase aborted","files":["a.md"]}`}
	_, err := f.start(t).Sync(context.Background())
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != http.StatusConflict || !slices.Equal(apiErr.Files, []string{"a.md"}) {
		t.Errorf("err = %#v", err)
	}
}

func TestGitLog(t *testing.T) {
	f := &fakeAPI{status: http.StatusOK, resp: `[{"hash":"abc1234","date":"2026-10-26 18:00","message":"m"}]`}
	commits, err := f.start(t).GitLog(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	f.expect(t, http.MethodGet, "/api/v1/git/log", "n=5", "")
	if len(commits) != 1 || commits[0].Hash != "abc1234" {
		t.Errorf("log = %+v", commits)
	}
}
