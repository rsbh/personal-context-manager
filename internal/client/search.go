package client

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/rsbh/personal-context-manager/internal/search"
)

// SearchParams selects hits. Empty fields match all. Global with Project
// matches global contexts and that project. Limit 0 is the server default.
type SearchParams struct {
	Q       string
	Scope   string
	Project string
	Global  bool
	Section string
	Limit   int
}

// Search runs a full-text search over contexts.
func (c *Client) Search(ctx context.Context, p SearchParams) ([]search.Hit, error) {
	q := url.Values{"q": {p.Q}}
	if p.Scope != "" {
		q.Set("scope", p.Scope)
	}
	if p.Project != "" {
		q.Set("project", p.Project)
	}
	if p.Global {
		q.Set("global", "true")
	}
	if p.Section != "" {
		q.Set("section", p.Section)
	}
	if p.Limit != 0 {
		q.Set("limit", strconv.Itoa(p.Limit))
	}
	var out []search.Hit
	err := c.do(ctx, http.MethodGet, "/api/v1/search?"+q.Encode(), nil, &out)
	return out, err
}

// RebuildIndex rebuilds the search index from md files and returns the count.
func (c *Client) RebuildIndex(ctx context.Context) (int, error) {
	var out struct {
		Indexed int `json:"indexed"`
	}
	err := c.do(ctx, http.MethodPost, "/api/v1/index/rebuild", nil, &out)
	return out.Indexed, err
}
