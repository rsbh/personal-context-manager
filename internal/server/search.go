package server

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/rsbh/personal-context-manager/internal/search"
	"github.com/rsbh/personal-context-manager/internal/store"
)

type rebuildResponse struct {
	Indexed int `json:"indexed"`
}

// WithSearch enables /search and /index, and keeps idx in step with context writes.
func WithSearch(idx *search.Index) Option {
	return func(s *server) { s.search = idx }
}

// requireSearch writes 501 when the server has no search index.
func (s *server) requireSearch(w http.ResponseWriter) bool {
	if s.search == nil {
		writeError(w, http.StatusNotImplemented, "search not configured")
		return false
	}
	return true
}

func (s *server) searchContexts(w http.ResponseWriter, r *http.Request) {
	if !s.requireSearch(w) {
		return
	}
	q := r.URL.Query()
	text := q.Get("q")
	if strings.TrimSpace(text) == "" {
		writeError(w, http.StatusBadRequest, "missing q")
		return
	}
	limit := 0
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			writeError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = n
	}
	hits, err := s.search.Search(search.Query{
		Q:       text,
		Scope:   q.Get("scope"),
		Project: q.Get("project"),
		Global:  q.Get("global") == "true",
		Section: q.Get("section"),
		Limit:   limit,
	})
	switch {
	case errors.Is(err, search.ErrEmptyQuery):
		writeError(w, http.StatusBadRequest, "missing q")
	case err != nil:
		log.Printf("search error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	default:
		writeJSON(w, http.StatusOK, hits)
	}
}

func (s *server) rebuildIndex(w http.ResponseWriter, _ *http.Request) {
	if !s.requireSearch(w) {
		return
	}
	n, err := RebuildIndex(s.store, s.search)
	if err != nil {
		log.Printf("index rebuild error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, rebuildResponse{Indexed: n})
}

// RebuildIndex replaces the index with all contexts in the store.
func RebuildIndex(st *store.Store, idx *search.Index) (int, error) {
	all, err := st.All()
	if err != nil {
		return 0, err
	}
	docs := make([]search.Doc, len(all))
	for i, c := range all {
		docs[i] = toDoc(c)
	}
	return idx.Rebuild(docs)
}

func toDoc(c store.Context) search.Doc {
	return search.Doc{
		ID:          c.ID,
		Scope:       c.Scope,
		Project:     c.Project,
		Section:     c.Section,
		Name:        c.Name,
		Title:       c.Frontmatter.Title,
		Description: c.Frontmatter.Description,
		Content:     c.Content,
		Tags:        c.Frontmatter.Tags,
	}
}

// The md file is the source of truth: an index error after a store write
// is logged, not returned. The next rebuild repairs the index.

func (s *server) indexContext(c store.Context) {
	if s.search == nil {
		return
	}
	if err := s.search.Index(toDoc(c)); err != nil {
		log.Printf("index %s: %v", c.ID, err)
	}
}

// reindexAll rebuilds after any delete or a git sync.
func (s *server) reindexAll() {
	if s.search == nil {
		return
	}
	if _, err := RebuildIndex(s.store, s.search); err != nil {
		log.Printf("index rebuild: %v", err)
	}
}
