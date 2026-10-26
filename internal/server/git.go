package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/rsbh/personal-context-manager/internal/gitsync"
)

const (
	defaultLogN = 10
	syncTimeout = 5 * time.Minute
)

type conflictResponse struct {
	Error string   `json:"error"`
	Files []string `json:"files"`
}

// requireGit writes 501 when the server has no git repo.
func (s *server) requireGit(w http.ResponseWriter) bool {
	if s.git == nil {
		writeError(w, http.StatusNotImplemented, "git sync not configured")
		return false
	}
	return true
}

// writeGitError maps a conflict to 409 with files. Other git errors are 500 with the git message.
func writeGitError(w http.ResponseWriter, err error) {
	var ce *gitsync.ConflictError
	if errors.As(err, &ce) {
		writeJSON(w, http.StatusConflict, conflictResponse{Error: ce.Error(), Files: ce.Files})
		return
	}
	log.Printf("git error: %v", err)
	writeError(w, http.StatusInternalServerError, err.Error())
}

// gitStatus godoc
//
//	@Summary	Git status of the context repo
//	@Tags		git
//	@Produce	json
//	@Success	200	{object}	gitsync.Status
//	@Failure	500	{object}	errorResponse
//	@Failure	501	{object}	errorResponse
//	@Router		/api/v1/git/status [get]
func (s *server) gitStatus(w http.ResponseWriter, r *http.Request) {
	if !s.requireGit(w) {
		return
	}
	st, err := s.git.Status(r.Context())
	if err != nil {
		writeGitError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

// gitSync godoc
//
//	@Summary	Commit, pull and push the context repo
//	@Tags		git
//	@Produce	json
//	@Success	200	{object}	gitsync.SyncResult
//	@Failure	409	{object}	conflictResponse
//	@Failure	500	{object}	errorResponse
//	@Failure	501	{object}	errorResponse
//	@Router		/api/v1/git/sync [post]
func (s *server) gitSync(w http.ResponseWriter, r *http.Request) {
	if !s.requireGit(w) {
		return
	}
	// A client disconnect must not kill git in the middle of a push.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), syncTimeout)
	defer cancel()
	res, err := s.git.Sync(ctx)
	if err != nil {
		writeGitError(w, err)
		return
	}
	// A pull can add, change or remove any file.
	s.reindexAll()
	writeJSON(w, http.StatusOK, res)
}

// gitLog godoc
//
//	@Summary	Recent commits of the context repo
//	@Tags		git
//	@Produce	json
//	@Param		n	query		int	false	"Number of commits"	default(10)
//	@Success	200	{array}		gitsync.Commit
//	@Failure	400	{object}	errorResponse
//	@Failure	500	{object}	errorResponse
//	@Failure	501	{object}	errorResponse
//	@Router		/api/v1/git/log [get]
func (s *server) gitLog(w http.ResponseWriter, r *http.Request) {
	if !s.requireGit(w) {
		return
	}
	n := defaultLogN
	if v := r.URL.Query().Get("n"); v != "" {
		var err error
		if n, err = strconv.Atoi(v); err != nil || n < 1 {
			writeError(w, http.StatusBadRequest, "invalid n: use a number >= 1")
			return
		}
	}
	commits, err := s.git.Log(r.Context(), n)
	if err != nil {
		writeGitError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, commits)
}
