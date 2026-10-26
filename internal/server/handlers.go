package server

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/rsbh/personal-context-manager/internal/markdown"
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

// listProjects godoc
//
//	@Summary	List projects
//	@Tags		projects
//	@Produce	json
//	@Success	200	{array}		string
//	@Failure	500	{object}	errorResponse
//	@Router		/api/v1/projects [get]
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

// getProject godoc
//
//	@Summary	Get a project
//	@Tags		projects
//	@Produce	json
//	@Param		project	path		string	true	"Project name"
//	@Success	200		{object}	project
//	@Failure	400		{object}	errorResponse
//	@Failure	404		{object}	errorResponse
//	@Router		/api/v1/projects/{project} [get]
func (s *server) getProject(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "project")
	if err := s.store.GetProject(name); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, project{Name: name})
}

type createProjectRequest = project

// createProject godoc
//
//	@Summary	Create a project
//	@Tags		projects
//	@Accept		json
//	@Produce	json
//	@Param		body	body		createProjectRequest	true	"Project"
//	@Success	201		{object}	project
//	@Failure	400		{object}	errorResponse
//	@Failure	409		{object}	errorResponse
//	@Router		/api/v1/projects [post]
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

// deleteProject godoc
//
//	@Summary	Delete a project
//	@Tags		projects
//	@Produce	json
//	@Param		project	path	string	true	"Project name"
//	@Param		force	query	bool	false	"Delete when not empty"
//	@Success	204
//	@Failure	404	{object}	errorResponse
//	@Failure	409	{object}	errorResponse
//	@Router		/api/v1/projects/{project} [delete]
func (s *server) deleteProject(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteProject(chi.URLParam(r, "project"), force(r)); err != nil {
		writeStoreError(w, err)
		return
	}
	s.reindexAll()
	w.WriteHeader(http.StatusNoContent)
}

// Sections

// listSections godoc
//
//	@Summary	List sections
//	@Tags		sections
//	@Produce	json
//	@Param		scope	query		string	false	"global or project"
//	@Param		project	query		string	false	"Project name, for scope project"
//	@Success	200		{array}		string
//	@Failure	400		{object}	errorResponse
//	@Failure	404		{object}	errorResponse
//	@Router		/api/v1/sections [get]
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

// createSection godoc
//
//	@Summary	Create a section
//	@Tags		sections
//	@Accept		json
//	@Produce	json
//	@Param		body	body		createSectionRequest	true	"Section"
//	@Success	201		{object}	createSectionRequest
//	@Failure	400		{object}	errorResponse
//	@Failure	404		{object}	errorResponse
//	@Failure	409		{object}	errorResponse
//	@Router		/api/v1/sections [post]
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

// deleteSection godoc
//
//	@Summary	Delete a section
//	@Tags		sections
//	@Produce	json
//	@Param		section	path	string	true	"Section name"
//	@Param		scope	query	string	false	"global or project"
//	@Param		project	query	string	false	"Project name, for scope project"
//	@Param		force	query	bool	false	"Delete when not empty"
//	@Success	204
//	@Failure	400	{object}	errorResponse
//	@Failure	404	{object}	errorResponse
//	@Failure	409	{object}	errorResponse
//	@Router		/api/v1/sections/{section} [delete]
func (s *server) deleteSection(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	err := s.store.DeleteSection(q.Get("scope"), q.Get("project"), chi.URLParam(r, "section"), force(r))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	s.reindexAll()
	w.WriteHeader(http.StatusNoContent)
}

// Contexts

// listContexts godoc
//
//	@Summary	List contexts
//	@Tags		contexts
//	@Produce	json
//	@Param		scope	query		string	false	"global or project"
//	@Param		project	query		string	false	"Project name"
//	@Param		section	query		string	false	"Section name"
//	@Success	200		{array}		store.Context
//	@Failure	400		{object}	errorResponse
//	@Failure	404		{object}	errorResponse
//	@Router		/api/v1/contexts [get]
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

// getContext godoc
//
//	@Summary		Get a context
//	@Description	The id can contain "/" (raw or %2F), for example global/style/go.
//	@Tags			contexts
//	@Produce		json
//	@Param			id	path		string	true	"Context id"
//	@Success		200	{object}	store.Context
//	@Failure		400	{object}	errorResponse
//	@Failure		404	{object}	errorResponse
//	@Router			/api/v1/contexts/{id} [get]
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

// createContext godoc
//
//	@Summary	Create a context
//	@Tags		contexts
//	@Accept		json
//	@Produce	json
//	@Param		body	body		createContextRequest	true	"Context"
//	@Success	201		{object}	store.Context
//	@Failure	400		{object}	errorResponse
//	@Failure	404		{object}	errorResponse
//	@Failure	409		{object}	errorResponse
//	@Router		/api/v1/contexts [post]
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
	s.indexContext(c)
	writeJSON(w, http.StatusCreated, c)
}

type updateContextRequest struct {
	Frontmatter store.Frontmatter `json:"frontmatter"`
	Content     string            `json:"content"`
}

// updateContext godoc
//
//	@Summary		Update a context
//	@Description	The id can contain "/" (raw or %2F), for example global/style/go.
//	@Tags			contexts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Context id"
//	@Param			body	body		updateContextRequest	true	"New frontmatter and content"
//	@Success		200		{object}	store.Context
//	@Failure		400		{object}	errorResponse
//	@Failure		404		{object}	errorResponse
//	@Router			/api/v1/contexts/{id} [put]
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
	s.indexContext(c)
	writeJSON(w, http.StatusOK, c)
}

// deleteContext godoc
//
//	@Summary		Delete a context
//	@Description	The id can contain "/" (raw or %2F), for example global/style/go.
//	@Tags			contexts
//	@Produce		json
//	@Param			id	path	string	true	"Context id"
//	@Success		204
//	@Failure		400	{object}	errorResponse
//	@Failure		404	{object}	errorResponse
//	@Router			/api/v1/contexts/{id} [delete]
func (s *server) deleteContext(w http.ResponseWriter, r *http.Request) {
	id, ok := contextID(w, r)
	if !ok {
		return
	}
	if err := s.store.Delete(id); err != nil {
		writeStoreError(w, err)
		return
	}
	s.reindexAll()
	w.WriteHeader(http.StatusNoContent)
}

// Resolve

// resolve returns JSON by default, markdown with ?format=md or Accept: text/markdown.
// The format param wins over Accept.
//
//	@Summary	Resolve contexts for a project
//	@Tags		resolve
//	@Produce	json
//	@Produce	text/markdown
//	@Param		project	query		string	false	"Project name"
//	@Param		section	query		string	false	"Section name"
//	@Param		format	query		string	false	"Response format"	Enums(json, md)
//	@Success	200		{array}		store.Context
//	@Failure	400		{object}	errorResponse
//	@Failure	404		{object}	errorResponse
//	@Router		/api/v1/resolve [get]
func (s *server) resolve(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	format := q.Get("format")
	switch format {
	case "":
		format = "json"
		if acceptsMarkdown(r) {
			format = "md"
		}
	case "json", "md":
	default:
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid format %q: use json or md", format))
		return
	}

	contexts, err := s.store.Resolve(q.Get("project"), q.Get("section"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	if format == "json" {
		writeJSON(w, http.StatusOK, contexts)
		return
	}
	w.Header().Set("Content-Type", markdown.ContentType)
	w.WriteHeader(http.StatusOK)
	markdown.Write(w, contexts)
}

func acceptsMarkdown(r *http.Request) bool {
	for _, part := range strings.Split(r.Header.Get("Accept"), ",") {
		if mt, _, err := mime.ParseMediaType(strings.TrimSpace(part)); err == nil && mt == "text/markdown" {
			return true
		}
	}
	return false
}
