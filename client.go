package typesafe

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Option func(*Client)

func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
		if baseURL == "" {
			c.baseURL = DefaultBaseURL
			return
		}
		c.baseURL = baseURL
	}
}

func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		if h != nil {
			c.httpClient = h
		}
	}
}

func WithMaxRetries(n int) Option {
	return func(c *Client) {
		if n >= 0 {
			c.maxRetries = n
		}
	}
}

func WithRetryBackoff(d time.Duration) Option {
	return func(c *Client) {
		if d >= 0 {
			c.retryBackoff = d
		}
	}
}

type Client struct {
	apiKey       string
	baseURL      string
	httpClient   *http.Client
	maxRetries   int
	retryBackoff time.Duration
}

func NewClient(apiKey string, options ...Option) *Client {
	c := &Client{
		apiKey:       apiKey,
		baseURL:      DefaultBaseURL,
		httpClient:   http.DefaultClient,
		maxRetries:   2,
		retryBackoff: 250 * time.Millisecond,
	}
	for _, option := range options {
		if option != nil {
			option(c)
		}
	}
	return c
}

func (c *Client) Evaluate(ctx context.Context, req EvaluationRequest) (*EvaluationResponse, error) {
	if c == nil {
		return nil, errors.New("typesafe: nil client")
	}
	if req.Model == "" {
		req.Model = JevLatest
	}
	if err := validateRequest(req); err != nil {
		return nil, err
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("typesafe: encode request: %w", err)
	}
	for attempt := 0; ; attempt++ {
		result, retry, err := c.do(ctx, body)
		if err == nil {
			return result, nil
		}
		if !retry || attempt >= c.maxRetries {
			return nil, err
		}
		delay := c.retryBackoff * time.Duration(1<<attempt)
		if waitErr := wait(ctx, delay); waitErr != nil {
			return nil, waitErr
		}
	}
}

func (c *Client) do(ctx context.Context, body []byte) (*EvaluationResponse, bool, error) {
	baseURL := c.baseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/systemone", bytes.NewReader(body))
	if err != nil {
		return nil, false, fmt.Errorf("typesafe: create request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+c.apiKey)
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, false, fmt.Errorf("typesafe: request: %w", err)
	}
	defer response.Body.Close()
	responseBody, readErr := io.ReadAll(response.Body)
	if readErr != nil {
		return nil, false, fmt.Errorf("typesafe: read response: %w", readErr)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		retry := response.StatusCode == http.StatusTooManyRequests || response.StatusCode == 529
		return nil, retry, &APIError{
			StatusCode: response.StatusCode,
			Status:     response.Status,
			Body:       string(responseBody),
			RetryAfter: response.Header.Get("Retry-After"),
		}
	}
	var result EvaluationResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, false, fmt.Errorf("typesafe: decode response: %w", err)
	}
	return &result, false, nil
}

func wait(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
