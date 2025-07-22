package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const BaseURL = "https://www.avanza.se"

type Client struct {
	httpClient    *http.Client
	baseURL       string
	cookies       map[string]string
	securityToken string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: BaseURL,
		cookies: make(map[string]string),
	}
}

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
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36")

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
	for _, cookie := range resp.Cookies() {
		if cookie.Name != "" && cookie.Value != "" {
			c.cookies[cookie.Name] = cookie.Value
			if cookie.Name == "AZACSRF" {
				c.securityToken = cookie.Value
			}
		}
	}
}
