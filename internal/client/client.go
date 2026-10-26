// Package client calls the pcm-server HTTP API.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const timeout = 2 * time.Second

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: timeout}}
}

// LocalURL returns the server URL for a local port.
func LocalURL(port int) string {
	return fmt.Sprintf("http://localhost:%d", port)
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

// ErrNotReachable means the server did not answer.
var ErrNotReachable = errors.New("server not reachable")

// APIError is an error response from the server.
type APIError struct {
	Status  int
	Message string
	Files   []string // sync conflict files
}

func (e *APIError) Error() string { return e.Message }

// do sends a JSON request to path and decodes the response into out (if not nil).
func (c *Client) do(ctx context.Context, method, path string, in, out any) error {
	resp, err := c.send(ctx, method, path, in)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// send sends a request. It returns an *APIError for status >= 400.
// The caller closes the body of a good response.
func (c *Client) send(ctx context.Context, method, path string, in any) (*http.Response, error) {
	var body io.Reader
	if in != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(in); err != nil {
			return nil, err
		}
		body = &buf
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotReachable, err)
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		var e struct {
			Error string   `json:"error"`
			Files []string `json:"files"`
		}
		if json.NewDecoder(resp.Body).Decode(&e) != nil || e.Error == "" {
			e.Error = http.StatusText(resp.StatusCode)
		}
		return nil, &APIError{Status: resp.StatusCode, Message: e.Error, Files: e.Files}
	}
	return resp, nil
}

// Health returns nil if the server answers GET /health with status ok.
func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("decode health: %w", err)
	}
	if body.Status != "ok" {
		return fmt.Errorf("status %q", body.Status)
	}
	return nil
}
