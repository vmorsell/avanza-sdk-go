package avanza

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vmorsell/avanza-sdk-go/trading"
)

func TestPlaceOrder_Success(t *testing.T) {
	const (
		testOrderbookID = "orderbookID"
		testAccountID   = "accountID"
		testPrice       = 2.0
		testVolume      = 1
		testOrderID     = "orderID1"
		testRequestID   = "reqID"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/trading-critical/rest/order/new" {
			t.Errorf("expected path /_api/trading-critical/rest/order/new, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// Verify request body
		var req trading.PlaceOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if got, want := req.Side, trading.OrderSideBuy; got != want {
			t.Errorf("req.Side = %v, want %v", got, want)
		}

		if got, want := req.OrderbookID, testOrderbookID; got != want {
			t.Errorf("req.OrderbookID = %v, want %v", got, want)
		}

		if got, want := req.Price, testPrice; got != want {
			t.Errorf("req.Price = %v, want %v", got, want)
		}

		if got, want := req.Volume, testVolume; got != want {
			t.Errorf("req.Volume = %v, want %v", got, want)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(trading.PlaceOrderResponse{
			OrderRequestStatus: trading.OrderRequestStatusSuccess,
			Message:            "",
			Parameters:         []string{""},
			OrderID:            testOrderID,
		})
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	req := &trading.PlaceOrderRequest{
		IsDividendReinvestment: false,
		RequestID:              testRequestID,
		Price:                  testPrice,
		Volume:                 testVolume,
		AccountID:              testAccountID,
		Side:                   trading.OrderSideBuy,
		OrderbookID:            testOrderbookID,
		Metadata: trading.OrderMetadata{
			OrderEntryMode:  "ADVANCED",
			HasTouchedPrice: "true",
		},
		Condition: trading.OrderConditionNormal,
	}

	resp, err := avanza.Trading.PlaceOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}

	if got, want := resp.OrderRequestStatus, trading.OrderRequestStatusSuccess; got != want {
		t.Errorf("resp.OrderRequestStatus = %v, want %v", got, want)
	}

	if got, want := resp.OrderID, testOrderID; got != want {
		t.Errorf("resp.OrderID = %v, want %v", got, want)
	}
}

func TestPlaceOrder_FailedStatus(t *testing.T) {
	const (
		testRequestID   = "reqID"
		testPrice       = 2.0
		testVolume      = 1
		testAccountID   = "accountID"
		testOrderbookID = "orderbookID"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(trading.PlaceOrderResponse{
			OrderRequestStatus: trading.OrderRequestStatusError,
			Message:            "Insufficient funds",
			Parameters:         []string{},
			OrderID:            "",
		})
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	req := &trading.PlaceOrderRequest{
		RequestID:   testRequestID,
		Price:       testPrice,
		Volume:      testVolume,
		AccountID:   testAccountID,
		Side:        trading.OrderSideBuy,
		OrderbookID: testOrderbookID,
		Condition:   trading.OrderConditionNormal,
	}

	resp, err := avanza.Trading.PlaceOrder(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if got, want := resp.OrderRequestStatus, trading.OrderRequestStatusError; got != want {
		t.Errorf("resp.OrderRequestStatus = %v, want %v", got, want)
	}

	if got, want := resp.Message, "Insufficient funds"; got != want {
		t.Errorf("resp.Message = %v, want %v", got, want)
	}
}

func TestPlaceOrder_HTTPError(t *testing.T) {
	const (
		testRequestID   = "reqID"
		testPrice       = 2.0
		testVolume      = 1
		testAccountID   = "accountID"
		testOrderbookID = "orderbookID"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	req := &trading.PlaceOrderRequest{
		RequestID:   testRequestID,
		Price:       testPrice,
		Volume:      testVolume,
		AccountID:   testAccountID,
		Side:        trading.OrderSideBuy,
		OrderbookID: testOrderbookID,
		Condition:   trading.OrderConditionNormal,
	}

	_, err := avanza.Trading.PlaceOrder(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPlaceOrder_SellOrder(t *testing.T) {
	const (
		testRequestID   = "reqID"
		testPrice       = 2.0
		testVolume      = 1
		testAccountID   = "accountID"
		testOrderbookID = "orderbookID"
		testOrderID2    = "orderID2"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req trading.PlaceOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if got, want := req.Side, trading.OrderSideSell; got != want {
			t.Errorf("req.Side = %v, want %v", got, want)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(trading.PlaceOrderResponse{
			OrderRequestStatus: trading.OrderRequestStatusSuccess,
			OrderID:            testOrderID2,
		})
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	req := &trading.PlaceOrderRequest{
		RequestID:   testRequestID,
		Price:       testPrice,
		Volume:      testVolume,
		AccountID:   testAccountID,
		Side:        trading.OrderSideSell,
		OrderbookID: testOrderbookID,
		Condition:   trading.OrderConditionNormal,
	}

	resp, err := avanza.Trading.PlaceOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}

	if got, want := resp.OrderRequestStatus, trading.OrderRequestStatusSuccess; got != want {
		t.Errorf("resp.OrderRequestStatus = %v, want %v", got, want)
	}
}

func TestPlaceOrder_ContextCancellation(t *testing.T) {
	const (
		testRequestID   = "reqID"
		testPrice       = 2.0
		testVolume      = 1
		testAccountID   = "accountID"
		testOrderbookID = "orderbookID"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay to allow context cancellation
		<-r.Context().Done()
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := &trading.PlaceOrderRequest{
		RequestID:   testRequestID,
		Price:       testPrice,
		Volume:      testVolume,
		AccountID:   testAccountID,
		Side:        trading.OrderSideBuy,
		OrderbookID: testOrderbookID,
		Condition:   trading.OrderConditionNormal,
	}

	_, err := avanza.Trading.PlaceOrder(ctx, req)
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}

func TestDeleteOrder_Success(t *testing.T) {
	const (
		testAccountID = "accountID"
		testOrderID   = "orderID123"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/trading-critical/rest/order/delete" {
			t.Errorf("expected path /_api/trading-critical/rest/order/delete, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req trading.DeleteOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if got, want := req.AccountID, testAccountID; got != want {
			t.Errorf("req.AccountID = %v, want %v", got, want)
		}

		if got, want := req.OrderID, testOrderID; got != want {
			t.Errorf("req.OrderID = %v, want %v", got, want)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(trading.DeleteOrderResponse{
			OrderRequestStatus: trading.OrderRequestStatusSuccess,
			Message:            "",
			Parameters:         []string{""},
			OrderID:            testOrderID,
		})
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	req := &trading.DeleteOrderRequest{
		AccountID: testAccountID,
		OrderID:   testOrderID,
	}

	resp, err := avanza.Trading.DeleteOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("DeleteOrder failed: %v", err)
	}

	if got, want := resp.OrderRequestStatus, trading.OrderRequestStatusSuccess; got != want {
		t.Errorf("resp.OrderRequestStatus = %v, want %v", got, want)
	}

	if got, want := resp.OrderID, testOrderID; got != want {
		t.Errorf("resp.OrderID = %v, want %v", got, want)
	}
}

func TestDeleteOrder_FailedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(trading.DeleteOrderResponse{
			OrderRequestStatus: trading.OrderRequestStatusError,
			Message:            "Order not found",
			Parameters:         []string{},
			OrderID:            "",
		})
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	req := &trading.DeleteOrderRequest{
		AccountID: "account123",
		OrderID:   "order456",
	}

	resp, err := avanza.Trading.DeleteOrder(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if got, want := resp.OrderRequestStatus, trading.OrderRequestStatusError; got != want {
		t.Errorf("resp.OrderRequestStatus = %v, want %v", got, want)
	}
}

func TestDeleteOrder_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	req := &trading.DeleteOrderRequest{
		AccountID: "account123",
		OrderID:   "order456",
	}

	_, err := avanza.Trading.DeleteOrder(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestModifyOrder_Success(t *testing.T) {
	const (
		testAccountID = "accountID"
		testOrderID   = "orderID456"
		testPrice     = 360.0
		testVolume    = 10
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/trading-critical/rest/order/modify" {
			t.Errorf("expected path /_api/trading-critical/rest/order/modify, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req trading.ModifyOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if got, want := req.OrderID, testOrderID; got != want {
			t.Errorf("req.OrderID = %v, want %v", got, want)
		}

		if got, want := req.AccountID, testAccountID; got != want {
			t.Errorf("req.AccountID = %v, want %v", got, want)
		}

		if got, want := req.Price, testPrice; got != want {
			t.Errorf("req.Price = %v, want %v", got, want)
		}

		if got, want := req.Volume, testVolume; got != want {
			t.Errorf("req.Volume = %v, want %v", got, want)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(trading.ModifyOrderResponse{
			OrderRequestStatus: trading.OrderRequestStatusSuccess,
			Message:            "",
			Parameters:         []string{""},
			OrderID:            testOrderID,
		})
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	req := &trading.ModifyOrderRequest{
		OrderID:    testOrderID,
		AccountID:  testAccountID,
		Price:      testPrice,
		Volume:     testVolume,
		ValidUntil: "2026-01-29",
	}

	resp, err := avanza.Trading.ModifyOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("ModifyOrder failed: %v", err)
	}

	if got, want := resp.OrderRequestStatus, trading.OrderRequestStatusSuccess; got != want {
		t.Errorf("resp.OrderRequestStatus = %v, want %v", got, want)
	}

	if got, want := resp.OrderID, testOrderID; got != want {
		t.Errorf("resp.OrderID = %v, want %v", got, want)
	}
}

func TestModifyOrder_FailedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(trading.ModifyOrderResponse{
			OrderRequestStatus: trading.OrderRequestStatusError,
			Message:            "Order not found",
			Parameters:         []string{},
			OrderID:            "",
		})
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	req := &trading.ModifyOrderRequest{
		OrderID:   "order456",
		AccountID: "account123",
		Price:     100.0,
		Volume:    10,
	}

	resp, err := avanza.Trading.ModifyOrder(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if got, want := resp.OrderRequestStatus, trading.OrderRequestStatusError; got != want {
		t.Errorf("resp.OrderRequestStatus = %v, want %v", got, want)
	}
}

func TestModifyOrder_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	req := &trading.ModifyOrderRequest{
		OrderID:   "order456",
		AccountID: "account123",
		Price:     100.0,
		Volume:    10,
	}

	_, err := avanza.Trading.ModifyOrder(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
