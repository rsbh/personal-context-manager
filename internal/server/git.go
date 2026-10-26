package server

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/rsbh/personal-context-manager/internal/gitsync"
)

const defaultLogN = 10

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

func (s *server) gitSync(w http.ResponseWriter, r *http.Request) {
	if !s.requireGit(w) {
		return
	}
	res, err := s.git.Sync(r.Context())
	if err != nil {
		writeGitError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

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
