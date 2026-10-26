package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/rsbh/personal-context-manager/internal/gitsync"
	"github.com/rsbh/personal-context-manager/internal/store"
)

type server struct {
	store *store.Store
	git   *gitsync.Repo
}

// Option configures the router.
type Option func(*server)

// WithGit enables the /git endpoints.
func WithGit(g *gitsync.Repo) Option {
	return func(s *server) { s.git = g }
}

// NewRouter returns the HTTP router with all routes registered.
func NewRouter(st *store.Store, opts ...Option) http.Handler {
	s := &server{store: st}
	for _, o := range opts {
		o(s)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", healthHandler)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/projects", s.listProjects)
		r.Post("/projects", s.createProject)
		r.Get("/projects/{project}", s.getProject)
		r.Delete("/projects/{project}", s.deleteProject)

		r.Get("/sections", s.listSections)
		r.Post("/sections", s.createSection)
		r.Delete("/sections/{section}", s.deleteSection)

		r.Get("/contexts", s.listContexts)
		r.Post("/contexts", s.createContext)
		r.Get("/contexts/*", s.getContext)
		r.Put("/contexts/*", s.updateContext)
		r.Delete("/contexts/*", s.deleteContext)

		r.Get("/resolve", s.resolve)

		r.Get("/git/status", s.gitStatus)
		r.Post("/git/sync", s.gitSync)
		r.Get("/git/log", s.gitLog)
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
