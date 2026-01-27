// Package trading provides trading functionality for the Avanza API.
package trading

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vmorsell/avanza-sdk-go/client"
)

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
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	httpResp, err := s.client.Post(ctx, "/_api/trading-critical/rest/order/new", req)
	if err != nil {
		return nil, fmt.Errorf("post: %w", err)
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

// GetOrders returns all current orders.
func (s *Service) GetOrders(ctx context.Context) (*GetOrdersResponse, error) {
	httpResp, err := s.client.Get(ctx, "/_api/trading/rest/orders")
	if err != nil {
		return nil, fmt.Errorf("get: %w", err)
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
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	httpResp, err := s.client.Post(ctx, "/_api/trading-critical/rest/order/validation/validate", req)
	if err != nil {
		return nil, fmt.Errorf("post: %w", err)
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
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	httpResp, err := s.client.Post(ctx, "/_api/trading/preliminary-fee/preliminaryfee", req)
	if err != nil {
		return nil, fmt.Errorf("post: %w", err)
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
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	httpResp, err := s.client.Post(ctx, "/_api/trading/stoploss/new", req)
	if err != nil {
		return nil, fmt.Errorf("post: %w", err)
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
		return nil, fmt.Errorf("get: %w", err)
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
