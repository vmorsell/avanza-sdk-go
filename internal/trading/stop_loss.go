package trading

import (
	"context"
	"encoding/json"
	"fmt"
)

// StopLossTriggerType represents the type of stop loss trigger.
type StopLossTriggerType string

const (
	StopLossTriggerLessOrEqual    StopLossTriggerType = "LESS_OR_EQUAL"
	StopLossTriggerGreaterOrEqual StopLossTriggerType = "GREATER_OR_EQUAL"
)

// StopLossValueType represents the type of stop loss value.
type StopLossValueType string

const (
	StopLossValueMonetary   StopLossValueType = "MONETARY"
	StopLossValuePercentage StopLossValueType = "PERCENTAGE"
)

// StopLossOrderEventType represents the type of stop loss order event.
type StopLossOrderEventType string

const (
	StopLossOrderEventBuy  StopLossOrderEventType = "BUY"
	StopLossOrderEventSell StopLossOrderEventType = "SELL"
)

// StopLossPriceType represents the type of stop loss price.
type StopLossPriceType string

const (
	StopLossPriceMonetary   StopLossPriceType = "MONETARY"
	StopLossPricePercentage StopLossPriceType = "PERCENTAGE"
)

// StopLossTrigger represents the trigger conditions for a stop loss order.
type StopLossTrigger struct {
	Type                      StopLossTriggerType `json:"type"`
	Value                     float64             `json:"value"`
	ValueType                 StopLossValueType   `json:"valueType"`
	ValidUntil                string              `json:"validUntil"`
	TriggerOnMarketMakerQuote bool                `json:"triggerOnMarketMakerQuote"`
}

// StopLossOrderEvent represents the order event that will be triggered.
type StopLossOrderEvent struct {
	Type                StopLossOrderEventType `json:"type"`
	Price               float64                `json:"price"`
	Volume              int                    `json:"volume"`
	ValidDays           int                    `json:"validDays"`
	PriceType           StopLossPriceType      `json:"priceType"`
	ShortSellingAllowed bool                   `json:"shortSellingAllowed"`
}

// PlaceStopLossRequest represents a request to place a stop loss order.
type PlaceStopLossRequest struct {
	ParentStopLossID   string             `json:"parentStopLossId"`
	AccountID          string             `json:"accountId"`
	OrderBookID        string             `json:"orderBookId"`
	StopLossTrigger    StopLossTrigger    `json:"stopLossTrigger"`
	StopLossOrderEvent StopLossOrderEvent `json:"stopLossOrderEvent"`
}

// PlaceStopLossResponse represents the response from placing a stop loss order.
type PlaceStopLossResponse struct {
	Status          string `json:"status"`
	StopLossOrderID string `json:"stoplossOrderId"`
}

// PlaceStopLoss places a new stop loss order.
func (s *Service) PlaceStopLoss(ctx context.Context, req *PlaceStopLossRequest) (*PlaceStopLossResponse, error) {
	httpResp, err := s.client.Post(ctx, "/_api/trading/stoploss/new", req)
	if err != nil {
		return nil, fmt.Errorf("failed to place stop loss order: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", httpResp.StatusCode)
	}

	var resp PlaceStopLossResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.Status != "SUCCESS" {
		return &resp, fmt.Errorf("stop loss order request failed: %s", resp.Status)
	}

	return &resp, nil
}
