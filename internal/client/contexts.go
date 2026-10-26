package client

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/rsbh/personal-context-manager/internal/store"
)

// ContextID returns the API id: global/<section>/<name> or projects/<project>/<section>/<name>.
func ContextID(scope, project, section, name string) string {
	if scope == "global" {
		return strings.Join([]string{"global", section, name}, "/")
	}
	return strings.Join([]string{"projects", project, section, name}, "/")
}

type createContextRequest struct {
	Scope       string            `json:"scope"`
	Project     string            `json:"project,omitempty"`
	Section     string            `json:"section"`
	Name        string            `json:"name"`
	Frontmatter store.Frontmatter `json:"frontmatter"`
	Content     string            `json:"content"`
}

type updateContextRequest struct {
	Frontmatter store.Frontmatter `json:"frontmatter"`
	Content     string            `json:"content"`
}

// ListContexts returns contexts without content. section is optional.
func (c *Client) ListContexts(ctx context.Context, scope, project, section string) ([]store.Context, error) {
	q := scopeQuery(scope, project)
	if section != "" {
		q.Set("section", section)
	}
	var out []store.Context
	err := c.do(ctx, http.MethodGet, "/api/v1/contexts?"+q.Encode(), nil, &out)
	return out, err
}

func (c *Client) GetContext(ctx context.Context, id string) (store.Context, error) {
	var out store.Context
	err := c.do(ctx, http.MethodGet, "/api/v1/contexts/"+id, nil, &out)
	return out, err
}

func (c *Client) CreateContext(ctx context.Context, in store.Context) (store.Context, error) {
	req := createContextRequest{
		Scope:       in.Scope,
		Project:     in.Project,
		Section:     in.Section,
		Name:        in.Name,
		Frontmatter: in.Frontmatter,
		Content:     in.Content,
	}
	var out store.Context
	err := c.do(ctx, http.MethodPost, "/api/v1/contexts", req, &out)
	return out, err
}

// UpdateContext replaces the frontmatter and content of context id.
func (c *Client) UpdateContext(ctx context.Context, id string, fm store.Frontmatter, content string) (store.Context, error) {
	var out store.Context
	err := c.do(ctx, http.MethodPut, "/api/v1/contexts/"+id, updateContextRequest{Frontmatter: fm, Content: content}, &out)
	return out, err
}

func (c *Client) DeleteContext(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/contexts/"+id, nil, nil)
}

// Resolve returns merged global + project contexts with content.
// Empty project means global only. section is optional.
func (c *Client) Resolve(ctx context.Context, project, section string) ([]store.Context, error) {
	q := url.Values{}
	if project != "" {
		q.Set("project", project)
	}
	if section != "" {
		q.Set("section", section)
	}
	path := "/api/v1/resolve"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	var out []store.Context
	err := c.do(ctx, http.MethodGet, path, nil, &out)
	return out, err
}

// ResolveMarkdown returns merged global + project contexts as markdown, rendered by the server.
func (c *Client) ResolveMarkdown(ctx context.Context, project, section string) (string, error) {
	q := url.Values{"format": {"md"}}
	if project != "" {
		q.Set("project", project)
	}
	if section != "" {
		q.Set("section", section)
	}
	resp, err := c.send(ctx, http.MethodGet, "/api/v1/resolve?"+q.Encode(), nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return string(b), err
}
