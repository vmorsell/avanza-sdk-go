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
// Error Handling:
//
// All methods return errors that should be checked. Use errors.As to check for
// HTTPError types:
//
//	var httpErr *client.HTTPError
//	if errors.As(err, &httpErr) {
//		fmt.Printf("HTTP %d: %s\n", httpErr.StatusCode, httpErr.Body)
//	}
//
// Context Usage:
//
// All methods accept context.Context for cancellation and timeouts. Always
// provide a context with appropriate timeout for production code:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
package avanza

import (
	"net/http"

	"github.com/vmorsell/avanza-sdk-go/internal/accounts"
	"github.com/vmorsell/avanza-sdk-go/internal/auth"
	"github.com/vmorsell/avanza-sdk-go/internal/client"
	"github.com/vmorsell/avanza-sdk-go/internal/market"
	"github.com/vmorsell/avanza-sdk-go/internal/trading"
)

// Avanza is the main client for interacting with the Avanza API.
type Avanza struct {
	client *client.Client
	// Auth handles BankID authentication and session management.
	Auth *auth.AuthService
	// Accounts provides account overview, positions, and trading accounts.
	Accounts *accounts.Service
	// Trading handles order placement, validation, fees, and stop loss orders.
	Trading *trading.Service
	// Market provides real-time market data subscriptions.
	Market *market.Service
}

// Option is a functional option for configuring the Avanza client.
type Option func(*Avanza)

// WithBaseURL sets a custom base URL for the client.
// This is primarily used for testing against mock servers.
//
// Example:
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
// This is useful for configuring custom timeouts or transport settings.
//
// Example:
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

// WithUserAgent sets a custom User-Agent string for HTTP requests.
// If not set, a default User-Agent is used.
//
// Example:
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

// New creates a new Avanza client with optional configuration.
//
// Example:
//
//	// Default configuration
//	client := avanza.New()
//
//	// With custom base URL for testing
//	client := avanza.New(avanza.WithBaseURL("http://localhost:8080"))
//
//	// With custom HTTP client
//	httpClient := &http.Client{Timeout: 60 * time.Second}
//	client := avanza.New(avanza.WithHTTPClient(httpClient))
//
//	// With custom User-Agent
//	client := avanza.New(avanza.WithUserAgent("MyApp/1.0"))
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
