package server

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"github.com/rsbh/personal-context-manager/internal/store"
)

const maxBodyBytes = 1 << 20

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error())
		return false
	}
	return true
}

// contextID reads the wildcard id. It accepts raw slashes and %2F.
func contextID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id, err := url.PathUnescape(chi.URLParam(r, "*"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return "", false
	}
	return id, true
}

func force(r *http.Request) bool {
	return r.URL.Query().Get("force") == "true"
}

// Projects

func (s *server) listProjects(w http.ResponseWriter, _ *http.Request) {
	projects, err := s.store.ListProjects()
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

type project struct {
	Name string `json:"name"`
}

func (s *server) getProject(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "project")
	if err := s.store.GetProject(name); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, project{Name: name})
}

type createProjectRequest = project

func (s *server) createProject(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if !decode(w, r, &req) {
		return
	}
	if err := s.store.CreateProject(req.Name); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, req)
}

func (s *server) deleteProject(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteProject(chi.URLParam(r, "project"), force(r)); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Sections

func (s *server) listSections(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	sections, err := s.store.ListSections(q.Get("scope"), q.Get("project"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sections)
}

type createSectionRequest struct {
	Scope   string `json:"scope"`
	Project string `json:"project,omitempty"`
	Name    string `json:"name"`
}

func (s *server) createSection(w http.ResponseWriter, r *http.Request) {
	var req createSectionRequest
	if !decode(w, r, &req) {
		return
	}
	if err := s.store.CreateSection(req.Scope, req.Project, req.Name); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, req)
}

func (s *server) deleteSection(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	err := s.store.DeleteSection(q.Get("scope"), q.Get("project"), chi.URLParam(r, "section"), force(r))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Contexts

func (s *server) listContexts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	contexts, err := s.store.List(store.Filter{
		Scope:   q.Get("scope"),
		Project: q.Get("project"),
		Section: q.Get("section"),
	})
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, contexts)
}

func (s *server) getContext(w http.ResponseWriter, r *http.Request) {
	id, ok := contextID(w, r)
	if !ok {
		return
	}
	c, err := s.store.Get(id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

type createContextRequest struct {
	Scope       string            `json:"scope"`
	Project     string            `json:"project,omitempty"`
	Section     string            `json:"section"`
	Name        string            `json:"name"`
	Frontmatter store.Frontmatter `json:"frontmatter"`
	Content     string            `json:"content"`
}

func (s *server) createContext(w http.ResponseWriter, r *http.Request) {
	var req createContextRequest
	if !decode(w, r, &req) {
		return
	}
	c, err := s.store.Create(store.Context{
		Scope:       req.Scope,
		Project:     req.Project,
		Section:     req.Section,
		Name:        req.Name,
		Frontmatter: req.Frontmatter,
		Content:     req.Content,
	})
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

type updateContextRequest struct {
	Frontmatter store.Frontmatter `json:"frontmatter"`
	Content     string            `json:"content"`
}

func (s *server) updateContext(w http.ResponseWriter, r *http.Request) {
	id, ok := contextID(w, r)
	if !ok {
		return
	}
	var req updateContextRequest
	if !decode(w, r, &req) {
		return
	}
	c, err := s.store.Update(id, req.Frontmatter, req.Content)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *server) deleteContext(w http.ResponseWriter, r *http.Request) {
	id, ok := contextID(w, r)
	if !ok {
		return
	}
	if err := s.store.Delete(id); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
