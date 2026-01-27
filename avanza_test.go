package avanza

import (
	"net/http"
	"testing"
	"time"

	"github.com/vmorsell/avanza-sdk-go/client"
)

func TestNew_DefaultClient(t *testing.T) {
	a := New()

	if a.client == nil {
		t.Fatal("expected client to be non-nil")
	}
	if a.Auth == nil {
		t.Fatal("expected Auth to be non-nil")
	}
	if a.Accounts == nil {
		t.Fatal("expected Accounts to be non-nil")
	}
	if a.Trading == nil {
		t.Fatal("expected Trading to be non-nil")
	}
	if a.Market == nil {
		t.Fatal("expected Market to be non-nil")
	}
}

func TestNew_OptionsAreComposable(t *testing.T) {
	customURL := "http://localhost:9999"
	customUA := "TestAgent/2.0"

	a := New(
		WithBaseURL(customURL),
		WithUserAgent(customUA),
	)

	if got := a.client.BaseURL(); got != customURL {
		t.Errorf("BaseURL = %q, want %q", got, customURL)
	}
	if got := a.client.UserAgent(); got != customUA {
		t.Errorf("UserAgent = %q, want %q", got, customUA)
	}
}

func TestNew_AllOptionsTogether(t *testing.T) {
	customURL := "http://localhost:8888"
	customUA := "AllOpts/1.0"
	customHTTP := &http.Client{Timeout: 99 * time.Second}
	customLimiter := &client.SimpleRateLimiter{Interval: 500 * time.Millisecond}

	a := New(
		WithBaseURL(customURL),
		WithHTTPClient(customHTTP),
		WithUserAgent(customUA),
		WithRateLimiter(customLimiter),
	)

	if got := a.client.BaseURL(); got != customURL {
		t.Errorf("BaseURL = %q, want %q", got, customURL)
	}
	if got := a.client.UserAgent(); got != customUA {
		t.Errorf("UserAgent = %q, want %q", got, customUA)
	}
	if got := a.client.HTTPClient(); got != customHTTP {
		t.Error("expected custom HTTP client to be set")
	}
	if got := a.client.HTTPClient().Timeout; got != 99*time.Second {
		t.Errorf("HTTPClient timeout = %v, want 99s", got)
	}
}

func TestNew_SingleOption(t *testing.T) {
	a := New(WithBaseURL("http://example.com"))

	if got := a.client.BaseURL(); got != "http://example.com" {
		t.Errorf("BaseURL = %q, want http://example.com", got)
	}
	// Other defaults should still be set
	if got := a.client.UserAgent(); got != client.DefaultUserAgent {
		t.Errorf("UserAgent should be default, got %q", got)
	}
}
