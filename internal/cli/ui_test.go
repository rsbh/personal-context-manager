package cli

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rsbh/personal-context-manager/internal/client"
)

func stubBrowser(t *testing.T) *[]string {
	t.Helper()
	var opened []string
	old := openBrowser
	openBrowser = func(url string) error { opened = append(opened, url); return nil }
	t.Cleanup(func() { openBrowser = old })
	return &opened
}

func TestUIOpensBrowser(t *testing.T) {
	opened := stubBrowser(t)
	t.Setenv("PCM_NO_BROWSER", "")
	a := newTestApp(t)
	out := mustRun(t, a, "ui")
	if out != "ui: "+a.client.BaseURL()+"\n" {
		t.Errorf("out = %q", out)
	}
	if len(*opened) != 1 || (*opened)[0] != a.client.BaseURL() {
		t.Errorf("opened = %v", *opened)
	}
}

func TestUINoBrowserEnv(t *testing.T) {
	opened := stubBrowser(t)
	t.Setenv("PCM_NO_BROWSER", "1")
	mustRun(t, newTestApp(t), "ui")
	if len(*opened) != 0 {
		t.Errorf("opened = %v", *opened)
	}
}

func TestUIServerDown(t *testing.T) {
	opened := stubBrowser(t)
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()

	out, _, err := runCLI(t, &app{client: client.New(url)}, "ui")
	if !errors.Is(err, errSilent) || !strings.Contains(out, "hint:   run `pcm server start`") {
		t.Errorf("out = %q, err = %v", out, err)
	}
	if len(*opened) != 0 {
		t.Errorf("opened = %v", *opened)
	}
}
