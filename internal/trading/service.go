// Package trading provides trading functionality for the Avanza API.
package trading

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vmorsell/avanza-sdk-go/internal/client"
)

// Service handles trading-related operations including orders, stop loss orders, validation, and fees.
type Service struct {
	client *client.Client
}

// NewService creates a new trading service with the given HTTP client.
func NewService(client *client.Client) *Service {
	return &Service{
		client: client,
	}
}

// PlaceOrder places a new order.
func (s *Service) PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*PlaceOrderResponse, error) {
	httpResp, err := s.client.Post(ctx, "/_api/trading-critical/rest/order/new", req)
	if err != nil {
		return nil, fmt.Errorf("place order: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("place order: %w", client.NewHTTPError(httpResp))
	}

	var resp PlaceOrderResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("place order: failed to decode response: %w", err)
	}

	if resp.OrderRequestStatus != OrderRequestStatusSuccess {
		return &resp, fmt.Errorf("place order: order request failed: %s", resp.Message)
	}

	return &resp, nil
}

// GetOrders retrieves all current orders.
func (s *Service) GetOrders(ctx context.Context) (*GetOrdersResponse, error) {
	httpResp, err := s.client.Get(ctx, "/_api/trading/rest/orders")
	if err != nil {
		return nil, fmt.Errorf("get orders: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get orders: %w", client.NewHTTPError(httpResp))
	}

	var resp GetOrdersResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("get orders: failed to decode response: %w", err)
	}

	return &resp, nil
}

// ValidateOrder validates an order before placing it.
func (s *Service) ValidateOrder(ctx context.Context, req *ValidateOrderRequest) (*ValidateOrderResponse, error) {
	httpResp, err := s.client.Post(ctx, "/_api/trading-critical/rest/order/validation/validate", req)
	if err != nil {
		return nil, fmt.Errorf("validate order: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("validate order: %w", client.NewHTTPError(httpResp))
	}

	var resp ValidateOrderResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("validate order: failed to decode response: %w", err)
	}

	return &resp, nil
}

// GetPreliminaryFee gets the preliminary fees for an order before placing it.
func (s *Service) GetPreliminaryFee(ctx context.Context, req *PreliminaryFeeRequest) (*PreliminaryFeeResponse, error) {
	httpResp, err := s.client.Post(ctx, "/_api/trading/preliminary-fee/preliminaryfee", req)
	if err != nil {
		return nil, fmt.Errorf("get preliminary fee: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get preliminary fee: %w", client.NewHTTPError(httpResp))
	}

	var resp PreliminaryFeeResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("get preliminary fee: failed to decode response: %w", err)
	}

	return &resp, nil
}

// PlaceStopLoss places a new stop loss order.
func (s *Service) PlaceStopLoss(ctx context.Context, req *PlaceStopLossRequest) (*PlaceStopLossResponse, error) {
	httpResp, err := s.client.Post(ctx, "/_api/trading/stoploss/new", req)
	if err != nil {
		return nil, fmt.Errorf("place stop loss order: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("place stop loss order: %w", client.NewHTTPError(httpResp))
	}

	var resp PlaceStopLossResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("place stop loss order: failed to decode response: %w", err)
	}

	if resp.Status != StopLossStatusSuccess {
		return &resp, fmt.Errorf("place stop loss order: stop loss order request failed: %s", resp.Status)
	}

	return &resp, nil
}

// GetStopLossOrders retrieves all active stop loss orders.
func (s *Service) GetStopLossOrders(ctx context.Context) ([]StopLossOrder, error) {
	httpResp, err := s.client.Get(ctx, "/_api/trading/stoploss/")
	if err != nil {
		return nil, fmt.Errorf("get stop loss orders: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get stop loss orders: %w", client.NewHTTPError(httpResp))
	}

	var orders []StopLossOrder
	if err := json.NewDecoder(httpResp.Body).Decode(&orders); err != nil {
		return nil, fmt.Errorf("get stop loss orders: failed to decode response: %w", err)
	}

	return orders, nil
}
