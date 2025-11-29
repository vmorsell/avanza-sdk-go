package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client == nil {
		t.Fatal("expected client to be non-nil")
	}

	if client.baseURL != BaseURL {
		t.Errorf("expected baseURL to be %s, got %s", BaseURL, client.baseURL)
	}

	if client.cookies == nil {
		t.Error("expected cookies map to be initialized")
	}

	if client.httpClient == nil {
		t.Error("expected httpClient to be initialized")
	}

	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("expected timeout to be 30s, got %v", client.httpClient.Timeout)
	}
}

func TestWithBaseURL(t *testing.T) {
	newURL := "https://test.example.com"
	client := NewClient(WithBaseURL(newURL))

	if client.baseURL != newURL {
		t.Errorf("expected baseURL to be %s, got %s", newURL, client.baseURL)
	}
}

func TestWithHTTPClient(t *testing.T) {
	customClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	client := NewClient(WithHTTPClient(customClient))

	if client.httpClient != customClient {
		t.Error("expected custom HTTP client to be set")
	}

	if client.httpClient.Timeout != 60*time.Second {
		t.Errorf("expected timeout to be 60s, got %v", client.httpClient.Timeout)
	}
}

func TestMultipleOptions(t *testing.T) {
	customURL := "https://test.example.com"
	customClient := &http.Client{
		Timeout: 45 * time.Second,
	}

	client := NewClient(
		WithBaseURL(customURL),
		WithHTTPClient(customClient),
	)

	if client.baseURL != customURL {
		t.Errorf("expected baseURL to be %s, got %s", customURL, client.baseURL)
	}

	if client.httpClient.Timeout != 45*time.Second {
		t.Errorf("expected timeout to be 45s, got %v", client.httpClient.Timeout)
	}
}

func TestPost_Success(t *testing.T) {
	tests := []struct {
		name           string
		endpoint       string
		body           interface{}
		serverResponse int
		responseBody   string
	}{
		{
			name:           "successful post with body",
			endpoint:       "/test",
			body:           map[string]string{"foo": "bar"},
			serverResponse: http.StatusOK,
			responseBody:   `{"status":"ok"}`,
		},
		{
			name:           "post with nil body",
			endpoint:       "/test",
			body:           nil,
			serverResponse: http.StatusOK,
			responseBody:   `{"status":"ok"}`,
		},
		{
			name:           "post with accepted status",
			endpoint:       "/test",
			body:           map[string]string{"foo": "bar"},
			serverResponse: http.StatusAccepted,
			responseBody:   `{"status":"accepted"}`,
		},
		{
			name:           "post with created status",
			endpoint:       "/api/create",
			body:           map[string]int{"value": 42},
			serverResponse: http.StatusCreated,
			responseBody:   `{"id":"FOO"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("expected POST method, got %s", r.Method)
				}

				if !strings.HasSuffix(r.URL.Path, tt.endpoint) {
					t.Errorf("expected endpoint %s, got %s", tt.endpoint, r.URL.Path)
				}

				w.WriteHeader(tt.serverResponse)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := NewClient(WithBaseURL(server.URL))

			ctx := context.Background()
			resp, err := client.Post(ctx, tt.endpoint, tt.body)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.serverResponse {
				t.Errorf("expected status %d, got %d", tt.serverResponse, resp.StatusCode)
			}
		})
	}
}

func TestPost_MarshalError(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Functions cannot be marshaled to JSON
	invalidBody := map[string]interface{}{
		"func": func() {},
	}

	_, err := client.Post(ctx, "/test", invalidBody)
	if err == nil {
		t.Error("expected marshal error, got nil")
	}
}

func TestPost_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.Post(ctx, "/test", nil)
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}

func TestPost_ImmediateCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.Post(ctx, "/test", nil)
	if err == nil {
		t.Error("expected context canceled error, got nil")
	}
}

func TestExtractCookies_AZACSRF(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "AZACSRF",
			Value: "FOO",
		})
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	ctx := context.Background()
	resp, err := client.Post(ctx, "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if client.securityToken != "FOO" {
		t.Errorf("expected security token to be FOO, got %s", client.securityToken)
	}

	if client.cookies["AZACSRF"] != "FOO" {
		t.Errorf("expected AZACSRF cookie to be FOO, got %s", client.cookies["AZACSRF"])
	}
}

func TestExtractCookies_MultipleCookies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "AZACSRF", Value: "FOO"})
		http.SetCookie(w, &http.Cookie{Name: "BAR", Value: "42"})
		http.SetCookie(w, &http.Cookie{Name: "sessionId", Value: "FOO"})
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	ctx := context.Background()
	resp, err := client.Post(ctx, "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	expectedCookies := map[string]string{
		"AZACSRF":   "FOO",
		"BAR":       "42",
		"sessionId": "FOO",
	}

	for name, expectedValue := range expectedCookies {
		if got := client.cookies[name]; got != expectedValue {
			t.Errorf("expected cookie %s to be %s, got %s", name, expectedValue, got)
		}
	}
}

func TestExtractCookies_EmptyCookiesIgnored(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "", Value: "FOO"})
		http.SetCookie(w, &http.Cookie{Name: "BAR", Value: ""})
		http.SetCookie(w, &http.Cookie{Name: "valid", Value: "42"})
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	ctx := context.Background()
	resp, err := client.Post(ctx, "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if len(client.cookies) != 1 {
		t.Errorf("expected 1 cookie, got %d", len(client.cookies))
	}

	if client.cookies["valid"] != "42" {
		t.Errorf("expected valid cookie to be 42, got %s", client.cookies["valid"])
	}
}

func TestCookiesPersistAcrossRequests(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		if callCount == 1 {
			// First request sets cookies
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "FOO"})
			http.SetCookie(w, &http.Cookie{Name: "AZACSRF", Value: "BAR"})
		} else {
			// Second request verifies cookies were sent
			cookie := r.Header.Get("Cookie")
			if !strings.Contains(cookie, "session=FOO") {
				t.Error("expected session cookie to be sent")
			}
			if !strings.Contains(cookie, "AZACSRF=BAR") {
				t.Error("expected AZACSRF cookie to be sent")
			}
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	ctx := context.Background()

	// First request
	resp1, err := client.Post(ctx, "/first", nil)
	if err != nil {
		t.Fatalf("first request failed: %v", err)
	}
	_ = resp1.Body.Close()

	// Second request
	resp2, err := client.Post(ctx, "/second", nil)
	if err != nil {
		t.Fatalf("second request failed: %v", err)
	}
	_ = resp2.Body.Close()

	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}

func TestSetHeaders_AllHeadersSet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedHeaders := map[string]string{
			"Accept":          "application/json, text/plain, */*",
			"Accept-Language": "en-US,en;q=0.8",
			"Cache-Control":   "no-cache",
			"Content-Type":    "application/json;charset=UTF-8",
			"Origin":          "https://www.avanza.se",
			"Pragma":          "no-cache",
			"Referer":         "https://www.avanza.se/logga-in.html",
		}

		for key, expected := range expectedHeaders {
			if got := r.Header.Get(key); got != expected {
				t.Errorf("expected header %s to be %q, got %q", key, expected, got)
			}
		}

		userAgent := r.Header.Get("User-Agent")
		if !strings.Contains(userAgent, "Mozilla") {
			t.Errorf("expected User-Agent to contain Mozilla, got %q", userAgent)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	ctx := context.Background()
	resp, err := client.Post(ctx, "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
}

func TestSetHeaders_WithSecurityToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		securityToken := r.Header.Get("X-SecurityToken")
		if securityToken != "FOO" {
			t.Errorf("expected X-SecurityToken to be FOO, got %s", securityToken)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	// Manually set security token to simulate authenticated state
	client.securityToken = "FOO"

	ctx := context.Background()
	resp, err := client.Post(ctx, "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
}

func TestSetHeaders_WithCookies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie := r.Header.Get("Cookie")

		if !strings.Contains(cookie, "foo=bar") {
			t.Error("expected Cookie header to contain foo=bar")
		}
		if !strings.Contains(cookie, "baz=42") {
			t.Error("expected Cookie header to contain baz=42")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.cookies["foo"] = "bar"
	client.cookies["baz"] = "42"

	ctx := context.Background()
	resp, err := client.Post(ctx, "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
}

func TestNewHTTPError_SizeLimit(t *testing.T) {
	// Create a large error response body (larger than maxErrorBodySize)
	largeBody := make([]byte, 2048) // 2KB
	for i := range largeBody {
		largeBody[i] = 'A'
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(largeBody)
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	httpErr := NewHTTPError(resp)

	// Verify the error body is limited to maxErrorBodySize (1024 bytes)
	if len(httpErr.Body) > 1024 {
		t.Errorf("expected error body to be limited to 1024 bytes, got %d bytes", len(httpErr.Body))
	}

	// Verify the body contains the first part of the response
	if len(httpErr.Body) != 1024 {
		t.Errorf("expected error body to be exactly 1024 bytes (truncated), got %d bytes", len(httpErr.Body))
	}

	// Verify all characters are 'A' (the first 1024 bytes)
	for i, b := range []byte(httpErr.Body) {
		if b != 'A' {
			t.Errorf("expected byte at index %d to be 'A', got %c", i, b)
		}
	}
}
