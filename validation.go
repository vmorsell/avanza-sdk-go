package avanza

import (
	"context"
	"encoding/json"
	"fmt"
)

// ValidateOrderRequest represents a request to validate an order before placing it.
type ValidateOrderRequest struct {
	IsDividendReinvestment bool           `json:"isDividendReinvestment"`
	RequestID              *string        `json:"requestId"`
	OrderRequestParameters interface{}    `json:"orderRequestParameters"`
	Price                  float64        `json:"price"`
	Volume                 int            `json:"volume"`
	OpenVolume             interface{}    `json:"openVolume"`
	AccountID              string         `json:"accountId"`
	Side                   OrderSide      `json:"side"`
	OrderbookID            string         `json:"orderbookId"`
	ValidUntil             interface{}    `json:"validUntil"`
	Metadata               interface{}    `json:"metadata"`
	Condition              OrderCondition `json:"condition"`
	ISIN                   string         `json:"isin"`
	Currency               string         `json:"currency"`
	MarketPlace            string         `json:"marketPlace"`
}

// ValidateOrderResponse represents the response from order validation.
type ValidateOrderResponse struct {
	CommissionWarning      ValidationResult `json:"commissionWarning"`
	EmployeeValidation     ValidationResult `json:"employeeValidation"`
	LargeInScaleWarning    ValidationResult `json:"largeInScaleWarning"`
	OrderValueLimitWarning ValidationResult `json:"orderValueLimitWarning"`
	PriceRampingWarning    ValidationResult `json:"priceRampingWarning"`
	CanadaOddLotWarning    ValidationResult `json:"canadaOddLotWarning"`
}

// ValidationResult represents the result of a validation check.
type ValidationResult struct {
	Valid bool `json:"valid"`
}

// ValidateOrder validates an order before placing it.
func (a *Avanza) ValidateOrder(ctx context.Context, req *ValidateOrderRequest) (*ValidateOrderResponse, error) {
	httpResp, err := a.client.Post(ctx, "/_api/trading-critical/rest/order/validation/validate", req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate order: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", httpResp.StatusCode)
	}

	var resp ValidateOrderResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &resp, nil
}
