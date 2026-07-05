// Package avanza provides a Go client library for the Avanza trading platform API.
//
// This is an unofficial, reverse-engineered SDK. Use at your own risk.
//
// Quick Start:
//
//	client := avanza.New()
//	ctx := context.Background()
//
//	// Authenticate with BankID
//	startResp, err := client.Auth.StartBankID(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	client.Auth.DisplayQRCode(startResp.QRToken)
//	collectResp, err := client.Auth.PollBankIDWithQRUpdates(ctx)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Establish session before making API calls
//	if err := client.Auth.EstablishSession(ctx, collectResp); err != nil {
//		log.Fatal(err)
//	}
//
//	// Use the client
//	overview, err := client.Accounts.GetOverview(ctx)
//
// Error handling:
//
//	var httpErr *client.HTTPError
//	if errors.As(err, &httpErr) {
//		fmt.Printf("HTTP %d: %s\n", httpErr.StatusCode, httpErr.Body)
//	}
//
// Session lifetime:
//
// Session cookies captured at login are held for the life of the process and
// are never refreshed. When Avanza expires the session server-side, API calls
// return *client.HTTPError with StatusCode 401 or 403, and active SSE
// subscriptions stop reconnecting (4xx is treated as non-recoverable). There
// is no automatic re-authentication — since login requires a human to scan a
// BankID QR code, callers must detect the expired session and run the auth
// flow again.
package avanza

import (
	"net/http"

	"github.com/vmorsell/avanza-sdk-go/accounts"
	"github.com/vmorsell/avanza-sdk-go/auth"
	"github.com/vmorsell/avanza-sdk-go/client"
	"github.com/vmorsell/avanza-sdk-go/market"
	"github.com/vmorsell/avanza-sdk-go/trading"
)

// Avanza is the main client for the Avanza API.
type Avanza struct {
	client   *client.Client
	Auth     *auth.AuthService
	Accounts *accounts.Service
	Trading  *trading.Service
	Market   *market.Service
}

// Option is a functional option for configuring the Avanza client.
type Option func(*config)

// config collects all options before building the client.
type config struct {
	clientOpts []client.Option
}

// WithBaseURL sets a custom base URL. Useful for testing.
//
//	client := avanza.New(avanza.WithBaseURL("http://localhost:8080"))
func WithBaseURL(url string) Option {
	return func(c *config) {
		c.clientOpts = append(c.clientOpts, client.WithBaseURL(url))
	}
}

// WithHTTPClient sets a custom HTTP client.
//
//	httpClient := &http.Client{Timeout: 60 * time.Second}
//	client := avanza.New(avanza.WithHTTPClient(httpClient))
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *config) {
		c.clientOpts = append(c.clientOpts, client.WithHTTPClient(httpClient))
	}
}

// WithUserAgent sets a custom User-Agent string.
//
//	client := avanza.New(avanza.WithUserAgent("MyApp/1.0"))
func WithUserAgent(userAgent string) Option {
	return func(c *config) {
		c.clientOpts = append(c.clientOpts, client.WithUserAgent(userAgent))
	}
}

// WithRateLimiter sets a rate limiter. Defaults to 100ms interval.
// Pass nil to disable (not recommended).
//
//	limiter := &client.SimpleRateLimiter{Interval: 200 * time.Millisecond}
//	client := avanza.New(avanza.WithRateLimiter(limiter))
func WithRateLimiter(limiter client.RateLimiter) Option {
	return func(c *config) {
		c.clientOpts = append(c.clientOpts, client.WithRateLimiter(limiter))
	}
}

// New creates a new Avanza client.
//
//	client := avanza.New()
//	client := avanza.New(avanza.WithBaseURL("http://localhost:8080"))
func New(opts ...Option) *Avanza {
	cfg := &config{}
	for _, opt := range opts {
		opt(cfg)
	}

	c := client.NewClient(cfg.clientOpts...)

	return &Avanza{
		client:   c,
		Auth:     auth.NewAuthService(c),
		Accounts: accounts.NewService(c),
		Trading:  trading.NewService(c),
		Market:   market.NewService(c),
	}
}
