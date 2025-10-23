// Package avanza provides a Go client library for the Avanza trading platform API.
package avanza

import (
	"net/http"

	"github.com/vmorsell/avanza/internal/accounts"
	"github.com/vmorsell/avanza/internal/auth"
	"github.com/vmorsell/avanza/internal/client"
)

// Avanza is the main client for interacting with the Avanza API.
// It provides access to various services like authentication and account management.
type Avanza struct {
	client *client.Client
	// Auth provides BankID authentication functionality.
	Auth *auth.AuthService
	// Accounts provides account overview and management functionality.
	Accounts *accounts.AccountsService
}

// Option is a functional option for configuring the Avanza client.
type Option func(*Avanza)

// WithBaseURL sets a custom base URL for the client.
// This is primarily used for testing against mock servers.
//
// Example:
//
//	client := avanza.New(avanza.WithBaseURL("https://test.example.com"))
func WithBaseURL(url string) Option {
	return func(a *Avanza) {
		a.client = client.NewClient(client.WithBaseURL(url))
		a.Auth = auth.NewAuthService(a.client)
		a.Accounts = accounts.NewAccountsService(a.client)
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
		a.Accounts = accounts.NewAccountsService(a.client)
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
//	client := avanza.New(avanza.WithBaseURL("https://test.example.com"))
//
//	// With custom HTTP client
//	httpClient := &http.Client{Timeout: 60 * time.Second}
//	client := avanza.New(avanza.WithHTTPClient(httpClient))
func New(opts ...Option) *Avanza {
	a := &Avanza{
		client: client.NewClient(),
	}
	a.Auth = auth.NewAuthService(a.client)
	a.Accounts = accounts.NewAccountsService(a.client)

	for _, opt := range opts {
		opt(a)
	}

	return a
}
