package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/rsbh/personal-context-manager/internal/store"
)

type server struct {
	store *store.Store
}

// NewRouter returns the HTTP router with all routes registered.
func NewRouter(st *store.Store) http.Handler {
	s := &server{store: st}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", healthHandler)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/projects", s.listProjects)
		r.Post("/projects", s.createProject)
		r.Delete("/projects/{project}", s.deleteProject)

		r.Get("/sections", s.listSections)
		r.Post("/sections", s.createSection)
		r.Delete("/sections/{section}", s.deleteSection)

		r.Get("/contexts", s.listContexts)
		r.Post("/contexts", s.createContext)
		r.Get("/contexts/*", s.getContext)
		r.Put("/contexts/*", s.updateContext)
		r.Delete("/contexts/*", s.deleteContext)
	})
	return r
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
