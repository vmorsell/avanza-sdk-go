package trading

import (
	"context"
	"encoding/json"
	"fmt"

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

// OrderSide represents the side of an order (buy or sell).
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// OrderCondition represents the condition type for an order.
type OrderCondition string

const (
	OrderConditionNormal     OrderCondition = "NORMAL"
	OrderConditionFillOrKill OrderCondition = "FILL_OR_KILL"
)

// OrderMetadata contains additional metadata about the order.
type OrderMetadata struct {
	OrderEntryMode  string `json:"orderEntryMode"`
	HasTouchedPrice string `json:"hasTouchedPrice"`
}

// PlaceOrderRequest represents a request to place a new order.
type PlaceOrderRequest struct {
	IsDividendReinvestment bool           `json:"isDividendReinvestment"`
	RequestID              string         `json:"requestId"`
	OrderRequestParameters interface{}    `json:"orderRequestParameters"`
	Price                  float64        `json:"price"`
	Volume                 int            `json:"volume"`
	OpenVolume             interface{}    `json:"openVolume"`
	AccountID              string         `json:"accountId"`
	Side                   OrderSide      `json:"side"`
	OrderbookID            string         `json:"orderbookId"`
	ValidUntil             interface{}    `json:"validUntil"`
	Metadata               OrderMetadata  `json:"metadata"`
	Condition              OrderCondition `json:"condition"`
}

// PlaceOrderResponse represents the response from placing an order.
type PlaceOrderResponse struct {
	OrderRequestStatus string   `json:"orderRequestStatus"`
	Message            string   `json:"message"`
	Parameters         []string `json:"parameters"`
	OrderID            string   `json:"orderId"`
}

// PlaceOrder places a new order.
func (s *Service) PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*PlaceOrderResponse, error) {
	httpResp, err := s.client.Post(ctx, "/_api/trading-critical/rest/order/new", req)
	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", httpResp.StatusCode)
	}

	var resp PlaceOrderResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.OrderRequestStatus != "SUCCESS" {
		return &resp, fmt.Errorf("order request failed: %s", resp.Message)
	}

	return &resp, nil
}
