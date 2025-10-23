// Package avanza provides a Go client library for the Avanza trading platform API.
package avanza

import (
	"github.com/vmorsell/avanza/internal/auth"
	"github.com/vmorsell/avanza/internal/client"
)

// Avanza is the main client for interacting with the Avanza API.
// It provides access to various services like authentication.
type Avanza struct {
	client *client.Client
	// Auth provides BankID authentication functionality.
	Auth *auth.AuthService
}

// New creates a new Avanza client with default configuration.
func New() *Avanza {
	client := client.NewClient()

	return &Avanza{
		client: client,
		Auth:   auth.NewAuthService(client),
	}
}
