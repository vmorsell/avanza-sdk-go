// Package trading provides trading functionality for the Avanza API.
package trading

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/vmorsell/avanza-sdk-go/client"
)

// SubscribeToOrders subscribes to real-time order updates. Call Close() when done.
func (s *Service) SubscribeToOrders(ctx context.Context) (*OrdersSubscription, error) {
	cookies := s.client.Cookies()
	if len(cookies) == 0 {
		return nil, fmt.Errorf("subscribe to orders: no authentication cookies found - please authenticate first")
	}

	essentialCookies := []string{"csid", "cstoken", "AZACSRF"}
	for _, cookie := range essentialCookies {
		if _, exists := cookies[cookie]; !exists {
			return nil, fmt.Errorf("subscribe to orders: missing essential cookie: %s - please authenticate first", cookie)
		}
	}

	subscriptionCtx, cancel := context.WithCancel(ctx)

	subscription := &OrdersSubscription{
		client: s.client,
		ctx:    subscriptionCtx,
		cancel: cancel,
		events: make(chan OrderEvent, 100),
		errors: make(chan error, 10),
	}

	go subscription.start()

	return subscription, nil
}

// Service handles trading operations: orders, stop loss, validation, and fees.
type Service struct {
	client *client.Client
}

// NewService creates a new trading service.
func NewService(client *client.Client) *Service {
	return &Service{
		client: client,
	}
}

// PlaceOrder places a new order. Consider validating first with ValidateOrder
// and checking fees with GetPreliminaryFee.
func (s *Service) PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*PlaceOrderResponse, error) {
	if req.AccountID == "" {
		return nil, fmt.Errorf("accountId is required")
	}
	if req.OrderbookID == "" {
		return nil, fmt.Errorf("orderbookId is required")
	}
	if req.Price <= 0 {
		return nil, fmt.Errorf("price must be greater than 0")
	}
	if req.Volume <= 0 {
		return nil, fmt.Errorf("volume must be greater than 0")
	}
	if req.Side != OrderSideBuy && req.Side != OrderSideSell {
		return nil, fmt.Errorf("side must be %s or %s", OrderSideBuy, OrderSideSell)
	}
	if req.Condition != OrderConditionNormal && req.Condition != OrderConditionFillOrKill {
		return nil, fmt.Errorf("condition must be %s or %s", OrderConditionNormal, OrderConditionFillOrKill)
	}

	httpResp, err := s.client.Post(ctx, "/_api/trading-critical/rest/order/new", req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp PlaceOrderResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if resp.OrderRequestStatus != OrderRequestStatusSuccess {
		return &resp, fmt.Errorf("order request failed: %s", resp.Message)
	}

	return &resp, nil
}

// DeleteOrder deletes an existing order.
func (s *Service) DeleteOrder(ctx context.Context, req *DeleteOrderRequest) (*DeleteOrderResponse, error) {
	if req.AccountID == "" {
		return nil, fmt.Errorf("accountId is required")
	}
	if req.OrderID == "" {
		return nil, fmt.Errorf("orderId is required")
	}

	httpResp, err := s.client.Post(ctx, "/_api/trading-critical/rest/order/delete", req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp DeleteOrderResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if resp.OrderRequestStatus != OrderRequestStatusSuccess {
		return &resp, fmt.Errorf("delete order request failed: %s", resp.Message)
	}

	return &resp, nil
}

// ModifyOrder modifies an existing order.
func (s *Service) ModifyOrder(ctx context.Context, req *ModifyOrderRequest) (*ModifyOrderResponse, error) {
	if req.OrderID == "" {
		return nil, fmt.Errorf("orderId is required")
	}
	if req.AccountID == "" {
		return nil, fmt.Errorf("accountId is required")
	}
	if req.Price <= 0 {
		return nil, fmt.Errorf("price must be greater than 0")
	}
	if req.Volume <= 0 {
		return nil, fmt.Errorf("volume must be greater than 0")
	}

	httpResp, err := s.client.Post(ctx, "/_api/trading-critical/rest/order/modify", req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp ModifyOrderResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if resp.OrderRequestStatus != OrderRequestStatusSuccess {
		return &resp, fmt.Errorf("modify order request failed: %s", resp.Message)
	}

	return &resp, nil
}

// GetOrders returns all current orders.
func (s *Service) GetOrders(ctx context.Context) (*GetOrdersResponse, error) {
	httpResp, err := s.client.Get(ctx, "/_api/trading/rest/orders")
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp GetOrdersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// ValidateOrder validates an order before placing it.
func (s *Service) ValidateOrder(ctx context.Context, req *ValidateOrderRequest) (*ValidateOrderResponse, error) {
	if req.AccountID == "" {
		return nil, fmt.Errorf("accountId is required")
	}
	if req.OrderbookID == "" {
		return nil, fmt.Errorf("orderbookId is required")
	}
	if req.Price <= 0 {
		return nil, fmt.Errorf("price must be greater than 0")
	}
	if req.Volume <= 0 {
		return nil, fmt.Errorf("volume must be greater than 0")
	}
	if req.Side != OrderSideBuy && req.Side != OrderSideSell {
		return nil, fmt.Errorf("side must be %s or %s", OrderSideBuy, OrderSideSell)
	}
	if req.Condition != OrderConditionNormal && req.Condition != OrderConditionFillOrKill {
		return nil, fmt.Errorf("condition must be %s or %s", OrderConditionNormal, OrderConditionFillOrKill)
	}
	if req.ISIN == "" {
		return nil, fmt.Errorf("isin is required")
	}
	if req.Currency == "" {
		return nil, fmt.Errorf("currency is required")
	}
	if req.MarketPlace == "" {
		return nil, fmt.Errorf("marketPlace is required")
	}

	httpResp, err := s.client.Post(ctx, "/_api/trading-critical/rest/order/validation/validate", req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp ValidateOrderResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// GetPreliminaryFee estimates fees for an order.
func (s *Service) GetPreliminaryFee(ctx context.Context, req *PreliminaryFeeRequest) (*PreliminaryFeeResponse, error) {
	if req.AccountID == "" {
		return nil, fmt.Errorf("accountId is required")
	}
	if req.OrderbookID == "" {
		return nil, fmt.Errorf("orderbookId is required")
	}
	if req.Price == "" {
		return nil, fmt.Errorf("price is required")
	}
	price, err := strconv.ParseFloat(req.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("price must be a valid number: %w", err)
	}
	if price <= 0 {
		return nil, fmt.Errorf("price must be greater than 0")
	}
	if req.Volume == "" {
		return nil, fmt.Errorf("volume is required")
	}
	volume, err := strconv.Atoi(req.Volume)
	if err != nil {
		return nil, fmt.Errorf("volume must be a valid integer: %w", err)
	}
	if volume <= 0 {
		return nil, fmt.Errorf("volume must be greater than 0")
	}
	if req.Side != OrderSideBuy && req.Side != OrderSideSell {
		return nil, fmt.Errorf("side must be %s or %s", OrderSideBuy, OrderSideSell)
	}

	httpResp, err := s.client.Post(ctx, "/_api/trading/preliminary-fee/preliminaryfee", req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp PreliminaryFeeResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// PlaceStopLoss places a new stop loss order.
func (s *Service) PlaceStopLoss(ctx context.Context, req *PlaceStopLossRequest) (*PlaceStopLossResponse, error) {
	if req.AccountID == "" {
		return nil, fmt.Errorf("accountId is required")
	}
	if req.OrderbookID == "" {
		return nil, fmt.Errorf("orderbookId is required")
	}
	// Validate trigger
	if req.StopLossTrigger.Type != StopLossTriggerLessOrEqual && req.StopLossTrigger.Type != StopLossTriggerGreaterOrEqual {
		return nil, fmt.Errorf("stopLossTrigger.type must be %s or %s", StopLossTriggerLessOrEqual, StopLossTriggerGreaterOrEqual)
	}
	if req.StopLossTrigger.Value <= 0 {
		return nil, fmt.Errorf("stopLossTrigger.value must be greater than 0")
	}
	if req.StopLossTrigger.ValueType != StopLossValueMonetary && req.StopLossTrigger.ValueType != StopLossValuePercentage {
		return nil, fmt.Errorf("stopLossTrigger.valueType must be %s or %s", StopLossValueMonetary, StopLossValuePercentage)
	}
	// Validate order event
	if req.StopLossOrderEvent.Type != StopLossOrderEventBuy && req.StopLossOrderEvent.Type != StopLossOrderEventSell {
		return nil, fmt.Errorf("stopLossOrderEvent.type must be %s or %s", StopLossOrderEventBuy, StopLossOrderEventSell)
	}
	if req.StopLossOrderEvent.Price <= 0 {
		return nil, fmt.Errorf("stopLossOrderEvent.price must be greater than 0")
	}
	if req.StopLossOrderEvent.Volume <= 0 {
		return nil, fmt.Errorf("stopLossOrderEvent.volume must be greater than 0")
	}
	if req.StopLossOrderEvent.ValidDays <= 0 {
		return nil, fmt.Errorf("stopLossOrderEvent.validDays must be greater than 0")
	}
	if req.StopLossOrderEvent.PriceType != StopLossPriceMonetary && req.StopLossOrderEvent.PriceType != StopLossPricePercentage {
		return nil, fmt.Errorf("stopLossOrderEvent.priceType must be %s or %s", StopLossPriceMonetary, StopLossPricePercentage)
	}

	httpResp, err := s.client.Post(ctx, "/_api/trading/stoploss/new", req)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp PlaceStopLossResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if resp.Status != StopLossStatusSuccess {
		return &resp, fmt.Errorf("stop loss order request failed: %s", resp.Status)
	}

	return &resp, nil
}

// GetStopLossOrders returns all active stop loss orders.
func (s *Service) GetStopLossOrders(ctx context.Context) ([]StopLossOrder, error) {
	httpResp, err := s.client.Get(ctx, "/_api/trading/stoploss/")
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var orders []StopLossOrder
	if err := json.NewDecoder(httpResp.Body).Decode(&orders); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return orders, nil
}
