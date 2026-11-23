package dify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	URL    string
	APIKey string
	HTTP   *http.Client
}

func New(url, apiKey string) *Client {
	return &Client{URL: url, APIKey: apiKey, HTTP: http.DefaultClient}
}

func (c *Client) Stream(ctx context.Context, query string, userID int64, inputs map[string]string) (*http.Response, error) {
	body := map[string]any{
		"inputs":        inputs,
		"query":         query,
		"response_mode": "streaming",
		"user":          fmt.Sprintf("%d", userID),
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("Dify API %d: %s", resp.StatusCode, string(data))
	}
	return resp, nil
}
