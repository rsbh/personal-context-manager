package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func slowServer(t *testing.T, d time.Duration) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(d):
		case <-r.Context().Done():
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestTimeoutIsNotNotReachable(t *testing.T) {
	c := New(slowServer(t, time.Second), WithTimeout(50*time.Millisecond))
	_, err := c.ListProjects(context.Background())
	if !errors.Is(err, ErrTimeout) || errors.Is(err, ErrNotReachable) {
		t.Fatalf("err = %v, want ErrTimeout only", err)
	}
	if err.Error() != "request timed out after 50ms" {
		t.Errorf("message = %q", err.Error())
	}
}

func TestSyncUsesSyncTimeout(t *testing.T) {
	// Normal timeout is too short for the slow server, sync timeout is long enough.
	c := New(slowServer(t, 200*time.Millisecond), WithTimeout(50*time.Millisecond), WithSyncTimeout(2*time.Second))
	if _, err := c.Sync(context.Background()); err != nil {
		t.Errorf("sync err = %v", err)
	}
}

func TestDefaultTimeouts(t *testing.T) {
	c := New("http://x")
	if c.timeout != 30*time.Second || c.syncTimeout != 5*time.Minute {
		t.Errorf("timeouts = %s, %s", c.timeout, c.syncTimeout)
	}
}
