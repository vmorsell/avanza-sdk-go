// Package market provides market data functionality for the Avanza API.
package market

import (
	"context"
	"fmt"

	"github.com/vmorsell/avanza-sdk-go/internal/client"
)

// Service handles market data operations including real-time subscriptions.
type Service struct {
	client *client.Client
}

// NewService creates a new market service with the given HTTP client.
func NewService(client *client.Client) *Service {
	return &Service{
		client: client,
	}
}

// SubscribeToOrderDepth subscribes to order depth updates for a specific orderbook.
// Returns a subscription that can be used to receive events and handle errors.
func (s *Service) SubscribeToOrderDepth(ctx context.Context, orderbookID string) (*OrderDepthSubscription, error) {
	// Verify we have authentication cookies
	cookies := s.client.Cookies()
	if len(cookies) == 0 {
		return nil, fmt.Errorf("subscribe to order depth: no authentication cookies found - please authenticate first")
	}

	// Check for essential authentication cookies
	essentialCookies := []string{"csid", "cstoken", "AZACSRF"}
	for _, cookie := range essentialCookies {
		if _, exists := cookies[cookie]; !exists {
			return nil, fmt.Errorf("subscribe to order depth: missing essential cookie: %s - please authenticate first", cookie)
		}
	}

	subscriptionCtx, cancel := context.WithCancel(ctx)

	subscription := &OrderDepthSubscription{
		orderbookID: orderbookID,
		client:      s.client,
		ctx:         subscriptionCtx,
		cancel:      cancel,
		events:      make(chan OrderDepthEvent, 100),
		errors:      make(chan error, 10),
	}

	go subscription.start()

	return subscription, nil
}

