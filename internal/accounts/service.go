// Package accounts provides account management functionality for the Avanza API.
package accounts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/vmorsell/avanza-sdk-go/internal/client"
)

// Service handles account-related operations.
type Service struct {
	client *client.Client
}

// NewService creates a new accounts service with the given HTTP client.
func NewService(client *client.Client) *Service {
	return &Service{
		client: client,
	}
}

// GetOverview retrieves the complete account overview including categories, accounts, and loans.
func (s *Service) GetOverview(ctx context.Context) (*AccountOverview, error) {
	resp, err := s.client.Get(ctx, "/_api/account-overview/overview/categorizedAccounts")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get account overview: unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var overview AccountOverview
	if err := json.NewDecoder(resp.Body).Decode(&overview); err != nil {
		return nil, fmt.Errorf("get account overview: failed to decode response: %w", err)
	}

	return &overview, nil
}

// GetTradingAccounts retrieves all trading accounts for the authenticated user.
func (s *Service) GetTradingAccounts(ctx context.Context) ([]TradingAccount, error) {
	resp, err := s.client.Get(ctx, "/_api/trading-critical/rest/accounts")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get trading accounts: unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var accounts []TradingAccount
	if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
		return nil, fmt.Errorf("get trading accounts: failed to decode response: %w", err)
	}

	return accounts, nil
}

// GetPositions retrieves positions for a specific account using its URL parameter ID.
func (s *Service) GetPositions(ctx context.Context, urlParameterID string) (*AccountPositions, error) {
	endpoint := fmt.Sprintf("/_api/position-data/positions/%s", urlParameterID)

	resp, err := s.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get account positions: unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var positions AccountPositions
	if err := json.NewDecoder(resp.Body).Decode(&positions); err != nil {
		return nil, fmt.Errorf("get account positions: failed to decode response: %w", err)
	}

	return &positions, nil
}

