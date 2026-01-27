// Package accounts provides account management functionality for the Avanza API.
package accounts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/vmorsell/avanza-sdk-go/client"
)

// Service handles account operations.
type Service struct {
	client *client.Client
}

// NewService creates a new accounts service.
func NewService(client *client.Client) *Service {
	return &Service{
		client: client,
	}
}

// GetOverview returns the account overview with categories, accounts, and loans.
func (s *Service) GetOverview(ctx context.Context) (*AccountOverview, error) {
	resp, err := s.client.Get(ctx, "/_api/account-overview/overview/categorizedAccounts")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(resp)
	}

	var overview AccountOverview
	if err := json.NewDecoder(resp.Body).Decode(&overview); err != nil {
		return nil, fmt.Errorf("get account overview: failed to decode response: %w", err)
	}

	return &overview, nil
}

// GetTradingAccounts returns all trading accounts.
func (s *Service) GetTradingAccounts(ctx context.Context) ([]TradingAccount, error) {
	resp, err := s.client.Get(ctx, "/_api/trading-critical/rest/accounts")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(resp)
	}

	var accounts []TradingAccount
	if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
		return nil, fmt.Errorf("get trading accounts: failed to decode response: %w", err)
	}

	return accounts, nil
}

// GetPositions returns positions for an account by its URL parameter ID.
func (s *Service) GetPositions(ctx context.Context, urlParameterID string) (*AccountPositions, error) {
	endpoint := fmt.Sprintf("/_api/position-data/positions/%s", url.PathEscape(urlParameterID))

	resp, err := s.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(resp)
	}

	var positions AccountPositions
	if err := json.NewDecoder(resp.Body).Decode(&positions); err != nil {
		return nil, fmt.Errorf("get account positions: failed to decode response: %w", err)
	}

	return &positions, nil
}
