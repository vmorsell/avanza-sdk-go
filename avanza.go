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
type Option func(*Avanza)

// WithBaseURL sets a custom base URL. Useful for testing.
//
//	client := avanza.New(avanza.WithBaseURL("http://localhost:8080"))
func WithBaseURL(url string) Option {
	return func(a *Avanza) {
		a.client = client.NewClient(client.WithBaseURL(url))
		a.Auth = auth.NewAuthService(a.client)
		a.Accounts = accounts.NewService(a.client)
		a.Trading = trading.NewService(a.client)
		a.Market = market.NewService(a.client)
	}
}

// WithHTTPClient sets a custom HTTP client.
//
//	httpClient := &http.Client{Timeout: 60 * time.Second}
//	client := avanza.New(avanza.WithHTTPClient(httpClient))
func WithHTTPClient(httpClient *http.Client) Option {
	return func(a *Avanza) {
		a.client = client.NewClient(client.WithHTTPClient(httpClient))
		a.Auth = auth.NewAuthService(a.client)
		a.Accounts = accounts.NewService(a.client)
		a.Trading = trading.NewService(a.client)
		a.Market = market.NewService(a.client)
	}
}

// WithUserAgent sets a custom User-Agent string.
//
//	client := avanza.New(avanza.WithUserAgent("MyApp/1.0"))
func WithUserAgent(userAgent string) Option {
	return func(a *Avanza) {
		a.client = client.NewClient(client.WithUserAgent(userAgent))
		a.Auth = auth.NewAuthService(a.client)
		a.Accounts = accounts.NewService(a.client)
		a.Trading = trading.NewService(a.client)
		a.Market = market.NewService(a.client)
	}
}

// WithRateLimiter sets a rate limiter. Defaults to 100ms interval.
// Pass nil to disable (not recommended).
//
//	limiter := &client.SimpleRateLimiter{Interval: 200 * time.Millisecond}
//	client := avanza.New(avanza.WithRateLimiter(limiter))
func WithRateLimiter(limiter client.RateLimiter) Option {
	return func(a *Avanza) {
		a.client = client.NewClient(client.WithRateLimiter(limiter))
		a.Auth = auth.NewAuthService(a.client)
		a.Accounts = accounts.NewService(a.client)
		a.Trading = trading.NewService(a.client)
		a.Market = market.NewService(a.client)
	}
}

// New creates a new Avanza client.
//
//	client := avanza.New()
//	client := avanza.New(avanza.WithBaseURL("http://localhost:8080"))
func New(opts ...Option) *Avanza {
	a := &Avanza{
		client: client.NewClient(),
	}
	a.Auth = auth.NewAuthService(a.client)
	a.Accounts = accounts.NewService(a.client)
	a.Trading = trading.NewService(a.client)
	a.Market = market.NewService(a.client)

	for _, opt := range opts {
		opt(a)
	}

	return a
}
