package client

import (
	"context"
	"net/http"
	"net/url"
)

type project struct {
	Name string `json:"name"`
}

func (c *Client) ListProjects(ctx context.Context) ([]string, error) {
	var out []string
	err := c.do(ctx, http.MethodGet, "/api/v1/projects", nil, &out)
	return out, err
}

func (c *Client) GetProject(ctx context.Context, name string) (string, error) {
	var out project
	err := c.do(ctx, http.MethodGet, "/api/v1/projects/"+url.PathEscape(name), nil, &out)
	return out.Name, err
}

func (c *Client) CreateProject(ctx context.Context, name string) error {
	return c.do(ctx, http.MethodPost, "/api/v1/projects", project{Name: name}, nil)
}

func (c *Client) DeleteProject(ctx context.Context, name string, force bool) error {
	path := "/api/v1/projects/" + url.PathEscape(name)
	if force {
		path += "?force=true"
	}
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}
