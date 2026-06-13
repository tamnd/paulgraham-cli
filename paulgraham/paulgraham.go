// Package paulgraham is the library behind the pg command: the HTTP client,
// request shaping, and the typed data models for Paul Graham's essays at
// paulgraham.com. No API key required — all data is publicly available HTML.
package paulgraham

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL   = "http://paulgraham.com"
	defaultUserAgent = "pg/dev (+https://github.com/tamnd/paulgraham-cli)"
)

// Config holds constructor parameters for Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Retries   int
	Timeout   time.Duration
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   defaultBaseURL,
		UserAgent: defaultUserAgent,
		Rate:      200 * time.Millisecond,
		Retries:   5,
		Timeout:   30 * time.Second,
	}
}

// Client talks to paulgraham.com over HTTP.
type Client struct {
	baseURL   string
	http      *http.Client
	userAgent string
	rate      time.Duration
	retries   int

	last time.Time
}

// NewClient returns a Client configured from cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		baseURL:   strings.TrimRight(cfg.BaseURL, "/"),
		http:      &http.Client{Timeout: cfg.Timeout},
		userAgent: cfg.UserAgent,
		rate:      cfg.Rate,
		retries:   cfg.Retries,
	}
}

// ListEssays fetches /articles.html and returns all essay links.
func (c *Client) ListEssays(ctx context.Context) ([]Essay, error) {
	body, err := c.get(ctx, c.baseURL+"/articles.html")
	if err != nil {
		return nil, fmt.Errorf("list essays: %w", err)
	}
	return parseArticlesHTML(string(body), c.baseURL), nil
}

// GetEssay fetches an essay by slug and returns its content.
func (c *Client) GetEssay(ctx context.Context, slug string) (EssayContent, error) {
	url := c.baseURL + "/" + slug + ".html"
	body, err := c.get(ctx, url)
	if err != nil {
		return EssayContent{}, fmt.Errorf("get essay %q: %w", slug, err)
	}
	return parseEssayHTML(string(body), slug, url), nil
}

// get fetches a URL with pacing and retries.
func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	if c.rate <= 0 {
		return
	}
	if wait := c.rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}
