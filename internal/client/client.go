// Package client calls the pcm-server HTTP API.
package client

import (
	"context"
	"encoding/json"
	"fmt"
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
