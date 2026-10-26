package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewRouter returns the HTTP router with all routes registered.
func NewRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/health", healthHandler)
	return r
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
