package server

import (
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

const placeholderHTML = `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><title>pcm</title></head>
<body style="font-family: system-ui, sans-serif; padding: 40px">
<h1>pcm</h1>
<p>UI not built: run <code>make web</code>, then rebuild pcm-server.</p>
<p>The API is at <code>/api/v1</code>.</p>
</body></html>
`

// WithUI serves the web app from fsys at /. Unknown paths get index.html (SPA routing).
func WithUI(fsys fs.FS) Option {
	return func(s *server) { s.ui = fsys }
}

// ListenAddr is the server address: local machine only, the UI and API have no auth.
func ListenAddr(port int) string {
	return fmt.Sprintf("127.0.0.1:%d", port)
}

// notFound serves the UI for unknown non-API GET paths. Everything else is a JSON 404.
func (s *server) notFound(w http.ResponseWriter, r *http.Request) {
	isAPI := r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/")
	if s.ui == nil || isAPI || (r.Method != http.MethodGet && r.Method != http.MethodHead) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if info, err := fs.Stat(s.ui, name); name != "" && err == nil && !info.IsDir() {
		http.FileServerFS(s.ui).ServeHTTP(w, r)
		return
	}
	index, err := fs.ReadFile(s.ui, "index.html")
	if err != nil {
		index = []byte(placeholderHTML)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(index) // the client is gone if this fails
}
