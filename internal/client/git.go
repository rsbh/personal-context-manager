package client

import (
	"context"
	"net/http"
	"strconv"

	"github.com/rsbh/personal-context-manager/internal/gitsync"
)

func (c *Client) GitStatus(ctx context.Context) (gitsync.Status, error) {
	var out gitsync.Status
	err := c.do(ctx, http.MethodGet, "/api/v1/git/status", nil, &out)
	return out, err
}

// Sync commits, pulls and pushes. A conflict is an *APIError with status 409 and Files.
func (c *Client) Sync(ctx context.Context) (gitsync.SyncResult, error) {
	var out gitsync.SyncResult
	err := c.doTimeout(ctx, c.syncTimeout, http.MethodPost, "/api/v1/git/sync", nil, &out)
	return out, err
}

func (c *Client) GitLog(ctx context.Context, n int) ([]gitsync.Commit, error) {
	var out []gitsync.Commit
	err := c.do(ctx, http.MethodGet, "/api/v1/git/log?n="+strconv.Itoa(n), nil, &out)
	return out, err
}
