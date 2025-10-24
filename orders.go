package avanza

import (
	"context"
	"encoding/json"
	"fmt"
)

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
func (a *Avanza) PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*PlaceOrderResponse, error) {
	httpResp, err := a.client.Post(ctx, "/_api/trading-critical/rest/order/new", req)
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

// OrderAccount represents account information for an order.
type OrderAccount struct {
	AccountID string `json:"accountId"`
	Name      struct {
		Value string `json:"value"`
	} `json:"name"`
	Type struct {
		AccountType string `json:"accountType"`
	} `json:"type"`
	URLParameterID string `json:"urlParameterId"`
}

// OrderOrderbook represents orderbook information for an order.
type OrderOrderbook struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	CountryCode    string `json:"countryCode"`
	Currency       string `json:"currency"`
	InstrumentType string `json:"instrumentType"`
	VolumeFactor   string `json:"volumeFactor"`
	ISIN           string `json:"isin"`
	MIC            string `json:"mic"`
}

// Order represents a single order.
type Order struct {
	Account              OrderAccount           `json:"account"`
	OrderID              string                 `json:"orderId"`
	Volume               int                    `json:"volume"`
	OriginalVolume       int                    `json:"originalVolume"`
	Price                float64                `json:"price"`
	Amount               float64                `json:"amount"`
	OrderbookID          string                 `json:"orderbookId"`
	Side                 string                 `json:"side"`
	ValidUntil           string                 `json:"validUntil"`
	Created              string                 `json:"created"`
	Deletable            bool                   `json:"deletable"`
	Modifiable           bool                   `json:"modifiable"`
	Message              string                 `json:"message"`
	State                string                 `json:"state"`
	StateText            string                 `json:"stateText"`
	StateMessage         string                 `json:"stateMessage"`
	Orderbook            OrderOrderbook         `json:"orderbook"`
	AdditionalParameters map[string]interface{} `json:"additionalParameters"`
	Condition            string                 `json:"condition"`
}

// GetOrdersResponse represents the response from getting orders.
type GetOrdersResponse struct {
	Orders          []Order       `json:"orders"`
	FundOrders      []interface{} `json:"fundOrders"`
	CancelledOrders []interface{} `json:"cancelledOrders"`
}

// GetOrders retrieves all current orders.
func (a *Avanza) GetOrders(ctx context.Context) (*GetOrdersResponse, error) {
	httpResp, err := a.client.Get(ctx, "/_api/trading/rest/orders")
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", httpResp.StatusCode)
	}

	var resp GetOrdersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &resp, nil
}
