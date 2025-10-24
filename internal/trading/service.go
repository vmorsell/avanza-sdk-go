package trading

import (
	"github.com/vmorsell/avanza-sdk-go/internal/client"
)

// Service provides trading-related operations.
type Service struct {
	client *client.Client
}

// NewService creates a new trading service.
func NewService(c *client.Client) *Service {
	return &Service{
		client: c,
	}
}
