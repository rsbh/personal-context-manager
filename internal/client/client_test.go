package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		wantErr bool
	}{
		{"ok", http.StatusOK, `{"status":"ok"}`, false},
		{"bad status", http.StatusInternalServerError, `{"status":"ok"}`, true},
		{"bad body", http.StatusOK, `{"status":"down"}`, true},
		{"not json", http.StatusOK, `x`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/health" {
					t.Errorf("path = %s, want /health", r.URL.Path)
				}
				w.WriteHeader(tt.status)
				w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			err := New(srv.URL).Health(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHealthNotReachable(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL
	srv.Close()

	if err := New(url).Health(context.Background()); err == nil {
		t.Error("want error for closed server")
	}
}
