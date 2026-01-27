// Package client provides HTTP client functionality for the Avanza API.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	// BaseURL is the base URL for the Avanza API.
	BaseURL = "https://www.avanza.se"

	// DefaultUserAgent mimics a browser to avoid detection.
	DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"
)

// Client manages sessions, cookies, and security tokens for Avanza API requests.
// It is safe for concurrent use.
type Client struct {
	httpClient    *http.Client
	baseURL       string
	mu            sync.RWMutex
	cookies       map[string]string
	securityToken string
	userAgent     string
	rateLimiter   RateLimiter
}

// BaseURL returns the base URL configured for the client.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// HTTPClient returns the underlying HTTP client.
func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}

// SecurityToken returns the current CSRF security token.
func (c *Client) SecurityToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.securityToken
}

// Cookies returns a copy of the current session cookies.
func (c *Client) Cookies() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cookies := make(map[string]string, len(c.cookies))
	for k, v := range c.cookies {
		cookies[k] = v
	}
	return cookies
}

// UserAgent returns the current User-Agent string.
func (c *Client) UserAgent() string {
	return c.userAgent
}

// SetMockCookies sets cookies for testing. AZACSRF is also set as the security token.
func (c *Client) SetMockCookies(cookies map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cookies = make(map[string]string, len(cookies))
	for k, v := range cookies {
		c.cookies[k] = v
		if k == "AZACSRF" {
			c.securityToken = v
		}
	}
}

// Option is a functional option for configuring the Client.
type Option func(*Client)

// WithBaseURL sets a custom base URL for the client.
// This is primarily used for testing against mock servers.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient sets a custom HTTP client.
// This is useful for configuring custom timeouts or transport settings.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithUserAgent sets a custom User-Agent string.
//
//	client := NewClient(WithUserAgent("MyApp/1.0"))
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// WithRateLimiter sets a rate limiter. Defaults to 100ms interval.
// Pass nil to disable (not recommended).
//
//	limiter := &SimpleRateLimiter{Interval: 200 * time.Millisecond}
//	client := NewClient(WithRateLimiter(limiter))
func WithRateLimiter(limiter RateLimiter) Option {
	return func(c *Client) {
		c.rateLimiter = limiter
	}
}

// NewClient creates a new HTTP client. Manages cookies and security tokens automatically.
// Rate limiting is enabled by default (100ms interval).
//
//	client := NewClient()
//	client := NewClient(WithBaseURL("http://localhost:8080"))
func NewClient(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:     BaseURL,
		cookies:     make(map[string]string),
		userAgent:   DefaultUserAgent,
		rateLimiter: &SimpleRateLimiter{Interval: DefaultRateLimitInterval},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Post sends a POST request. Body is marshaled to JSON.
// Cookies, security tokens, and rate limiting are handled automatically.
func (c *Client) Post(ctx context.Context, endpoint string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	var jsonBody []byte
	var err error
	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	c.setHeaders(req)

	if c.rateLimiter != nil {
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}

	c.extractCookies(resp)
	return resp, nil
}

// Get sends a GET request. Cookies, security tokens, and rate limiting are handled automatically.
func (c *Client) Get(ctx context.Context, endpoint string) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	c.setHeaders(req)

	if c.rateLimiter != nil {
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}

	c.extractCookies(resp)
	return resp, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("Origin", "https://www.avanza.se")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", "https://www.avanza.se/logga-in.html")
	req.Header.Set("User-Agent", c.userAgent)

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.securityToken != "" {
		req.Header.Set("X-SecurityToken", c.securityToken)
	}

	if len(c.cookies) > 0 {
		var cookiePairs []string
		for name, value := range c.cookies {
			if name != "" && value != "" {
				cookiePairs = append(cookiePairs, fmt.Sprintf("%s=%s", name, value))
			}
		}
		if len(cookiePairs) > 0 {
			req.Header.Set("Cookie", strings.Join(cookiePairs, "; "))
		}
	}
}

func (c *Client) extractCookies(resp *http.Response) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, cookie := range resp.Cookies() {
		if cookie.Name != "" && cookie.Value != "" {
			c.cookies[cookie.Name] = cookie.Value
			if cookie.Name == "AZACSRF" {
				c.securityToken = cookie.Value
			}
		}
	}
}
