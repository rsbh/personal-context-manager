package client

import (
	"context"
	"net/http"
	"net/url"
)

type section struct {
	Scope   string `json:"scope"`
	Project string `json:"project,omitempty"`
	Name    string `json:"name"`
}

// scopeQuery returns query values for a scope. project is empty for global.
func scopeQuery(scope, project string) url.Values {
	q := url.Values{"scope": {scope}}
	if project != "" {
		q.Set("project", project)
	}
	return q
}

// ListSections returns the section names in a scope.
func (c *Client) ListSections(ctx context.Context, scope, project string) ([]string, error) {
	var out []string
	path := "/api/v1/sections?" + scopeQuery(scope, project).Encode()
	err := c.do(ctx, http.MethodGet, path, nil, &out)
	return out, err
}

// CreateSection creates a section in a scope.
func (c *Client) CreateSection(ctx context.Context, scope, project, name string) error {
	return c.do(ctx, http.MethodPost, "/api/v1/sections", section{Scope: scope, Project: project, Name: name}, nil)
}

// DeleteSection deletes a section. Without force, the section must be empty.
func (c *Client) DeleteSection(ctx context.Context, scope, project, name string, force bool) error {
	q := scopeQuery(scope, project)
	if force {
		q.Set("force", "true")
	}
	return c.do(ctx, http.MethodDelete, "/api/v1/sections/"+url.PathEscape(name)+"?"+q.Encode(), nil, nil)
}
