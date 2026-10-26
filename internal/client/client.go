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

const (
	healthTimeout      = 2 * time.Second
	defaultTimeout     = 30 * time.Second
	defaultSyncTimeout = 5 * time.Minute // git fetch and push can be slow
)

// Client calls the pcm server API.
type Client struct {
	baseURL     string
	http        *http.Client
	timeout     time.Duration
	syncTimeout time.Duration
}

// Option configures a Client.
type Option func(*Client)

// WithTimeout sets the timeout for normal API calls.
func WithTimeout(d time.Duration) Option { return func(c *Client) { c.timeout = d } }

// WithSyncTimeout sets the timeout for git sync calls.
func WithSyncTimeout(d time.Duration) Option { return func(c *Client) { c.syncTimeout = d } }

// New returns a Client for the server at baseURL.
func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL:     baseURL,
		http:        &http.Client{},
		timeout:     defaultTimeout,
		syncTimeout: defaultSyncTimeout,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// LocalURL returns the server URL for a local port.
func LocalURL(port int) string {
	return fmt.Sprintf("http://127.0.0.1:%d", port)
}

// BaseURL returns the server URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}

var (
	// ErrNotReachable means the client could not connect to the server.
	ErrNotReachable = errors.New("server not reachable")
	// ErrTimeout means the server did not answer in time.
	ErrTimeout = errors.New("request timed out")
)

// APIError is an error response from the server.
type APIError struct {
	Status  int
	Message string
	Files   []string // sync conflict files
}

func (e *APIError) Error() string { return e.Message }

// do sends a JSON request to path and decodes the response into out (if not nil).
func (c *Client) do(ctx context.Context, method, path string, in, out any) error {
	return c.doTimeout(ctx, c.timeout, method, path, in, out)
}

func (c *Client) doTimeout(ctx context.Context, d time.Duration, method, path string, in, out any) error {
	resp, err := c.sendTimeout(ctx, d, method, path, in)
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

// send sends a request with the normal timeout.
func (c *Client) send(ctx context.Context, method, path string, in any) (*http.Response, error) {
	return c.sendTimeout(ctx, c.timeout, method, path, in)
}

// cancelBody cancels the request context when the body is closed.
type cancelBody struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (b cancelBody) Close() error {
	defer b.cancel()
	return b.ReadCloser.Close()
}

// sendTimeout sends a request. It returns an *APIError for status >= 400,
// ErrTimeout after d, and ErrNotReachable if it cannot connect.
// The caller closes the body of a good response.
func (c *Client) sendTimeout(ctx context.Context, d time.Duration, method, path string, in any) (*http.Response, error) {
	var body io.Reader
	if in != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(in); err != nil {
			return nil, err
		}
		body = &buf
	}
	ctx, cancel := context.WithTimeout(ctx, d)
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		cancel()
		return nil, err
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		cancel()
		return nil, fmt.Errorf("%w after %s", ErrTimeout, d)
	case err != nil:
		cancel()
		return nil, fmt.Errorf("%w: %w", ErrNotReachable, err)
	}
	resp.Body = cancelBody{resp.Body, cancel}
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
	ctx, cancel := context.WithTimeout(ctx, healthTimeout)
	defer cancel()
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
