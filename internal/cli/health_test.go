package cli

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rsbh/personal-context-manager/internal/client"
)

func runHealth(t *testing.T, url string) (string, error) {
	t.Helper()
	a := &app{client: client.New(url)}
	cmd := newHealthCmd(a)
	cmd.SilenceUsage, cmd.SilenceErrors = true, true
	var out bytes.Buffer
	cmd.SetOut(&out)
	err := cmd.Execute()
	return out.String(), err
}

func TestHealthOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	out, err := runHealth(t, srv.URL)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	want := "server: ok\nurl:    " + srv.URL + "\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}

func TestHealthNotReachable(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()

	out, err := runHealth(t, url)
	if !errors.Is(err, errSilent) {
		t.Fatalf("err = %v, want errSilent", err)
	}
	want := "server: not reachable\nurl:    " + url + "\nhint:   run `pcm server start`\n"
	if out != want {
		t.Errorf("out = %q, want %q", out, want)
	}
}
