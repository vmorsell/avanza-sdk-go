package avanza

import (
	"context"
	"encoding/json"
	"fmt"
)

// PreliminaryFeeRequest represents a request to get preliminary fees for an order.
type PreliminaryFeeRequest struct {
	AccountID   string `json:"accountId"`
	OrderbookID string `json:"orderbookId"`
	Price       string `json:"price"`
	Volume      string `json:"volume"`
	Side        string `json:"side"`
}

// PreliminaryFeeResponse represents the response from getting preliminary fees.
type PreliminaryFeeResponse struct {
	Commission          string              `json:"commission"`
	MarketFees          string              `json:"marketFees"`
	TotalFees           string              `json:"totalFees"`
	TotalSum            string              `json:"totalSum"`
	TotalSumWithoutFees string              `json:"totalSumWithoutFees"`
	OrderbookCurrency   string              `json:"orderbookCurrency"`
	TransactionTax      *string             `json:"transactionTax"`
	CurrencyExchangeFee CurrencyExchangeFee `json:"currencyExchangeFee"`
	Campaign            *string             `json:"campaign"`
}

// CurrencyExchangeFee represents currency exchange fee information.
type CurrencyExchangeFee struct {
	Rate string `json:"rate"`
	Sum  string `json:"sum"`
}

// GetPreliminaryFee gets the preliminary fees for an order before placing it.
func (a *Avanza) GetPreliminaryFee(ctx context.Context, req *PreliminaryFeeRequest) (*PreliminaryFeeResponse, error) {
	httpResp, err := a.client.Post(ctx, "/_api/trading/preliminary-fee/preliminaryfee", req)
	if err != nil {
		return nil, fmt.Errorf("failed to get preliminary fee: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", httpResp.StatusCode)
	}

	var resp PreliminaryFeeResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &resp, nil
}
