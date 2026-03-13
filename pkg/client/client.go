package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/wbhemingway/go-cartographer/internal/models"
)

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func New(url string, apiKey string) *Client {
	return &Client{
		BaseURL:    url,
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) RequestMap(ctx context.Context, world models.World) (io.ReadCloser, error) {
	payload, err := json.Marshal(world)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal world data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/render", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("server returned error code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
