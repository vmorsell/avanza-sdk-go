// Package accounts provides account management functionality for the Avanza API.
package accounts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"time"

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
	if urlParameterID == "" {
		return nil, fmt.Errorf("urlParameterID is required")
	}

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

// GetTransactions returns transactions for the authenticated user.
// From and To dates are required (format: YYYY-MM-DD).
//
// Note: This method does not include aggregated result values. Use
// GetAggregatedValues for total account values on specific dates.
func (s *Service) GetTransactions(ctx context.Context, req *TransactionsRequest) (*TransactionsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	if req.From == "" {
		return nil, fmt.Errorf("from date is required")
	}
	if req.To == "" {
		return nil, fmt.Errorf("to date is required")
	}
	fromDate, err := time.Parse("2006-01-02", req.From)
	if err != nil {
		return nil, fmt.Errorf("from date must be in YYYY-MM-DD format")
	}
	toDate, err := time.Parse("2006-01-02", req.To)
	if err != nil {
		return nil, fmt.Errorf("to date must be in YYYY-MM-DD format")
	}
	if fromDate.After(toDate) {
		return nil, fmt.Errorf("from date must not be after to date")
	}

	params := url.Values{}
	params.Set("from", req.From)
	params.Set("to", req.To)
	params.Set("includeResult", "false")

	endpoint := "/_api/transactions/list?" + params.Encode()

	resp, err := s.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(resp)
	}

	var transactions TransactionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&transactions); err != nil {
		return nil, fmt.Errorf("get transactions: failed to decode response: %w", err)
	}

	return &transactions, nil
}

// GetAggregatedValues returns the total value of accounts on specific dates.
// EncryptedAccountIDs and Dates are required.
func (s *Service) GetAggregatedValues(ctx context.Context, req *AggregatedValuesRequest) (AggregatedValuesResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	if len(req.EncryptedAccountIDs) == 0 {
		return nil, fmt.Errorf("at least one encrypted account ID is required")
	}
	if slices.Contains(req.EncryptedAccountIDs, "") {
		return nil, fmt.Errorf("encrypted account ID cannot be empty")
	}
	if len(req.Dates) == 0 {
		return nil, fmt.Errorf("at least one date is required")
	}
	for _, d := range req.Dates {
		if _, err := time.Parse("2006-01-02", d); err != nil {
			return nil, fmt.Errorf("date %q must be in YYYY-MM-DD format", d)
		}
	}

	resp, err := s.client.Post(ctx, "/_api/account-performance/aggregatedAccountsValues", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(resp)
	}

	var values AggregatedValuesResponse
	if err := json.NewDecoder(resp.Body).Decode(&values); err != nil {
		return nil, fmt.Errorf("get aggregated values: failed to decode response: %w", err)
	}

	return values, nil
}
