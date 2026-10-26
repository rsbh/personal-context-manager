package cli

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rsbh/personal-context-manager/internal/client"
)

func TestSlowServerPrintsTimeoutNotHealth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.Write([]byte(`{"status":"ok"}`))
			return
		}
		select {
		case <-time.After(time.Second):
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(srv.Close)

	a := &app{client: client.New(srv.URL, client.WithTimeout(50*time.Millisecond)), cwd: t.TempDir()}
	out, _, err := runCLI(t, a, "get", "projects")
	if err == nil || errors.Is(err, errSilent) || !strings.Contains(err.Error(), "timed out") {
		t.Errorf("err = %v, want timeout error", err)
	}
	if strings.Contains(out, "server:") {
		t.Errorf("health lines printed: %q", out)
	}
}

func TestSyncStatusPushHint(t *testing.T) {
	a := newGitTestApp(t)
	mustRun(t, a, "create", "section", "style", "-g")
	mustRun(t, a, "sync")
	mustRun(t, a, "config", "set", "remote", bareRemote(t))
	out := mustRun(t, a, "sync", "status")
	if !strings.Contains(out, "ahead:   1\n") || !strings.Contains(out, "hint:    run pcm sync to push\n") {
		t.Errorf("out = %q", out)
	}
}
