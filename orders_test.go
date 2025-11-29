package avanza

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
		var req PlaceOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if got, want := req.Side, OrderSideBuy; got != want {
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
		_ = json.NewEncoder(w).Encode(PlaceOrderResponse{
			OrderRequestStatus: "SUCCESS",
			Message:            "",
			Parameters:         []string{""},
			OrderID:            testOrderID,
		})
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	req := &PlaceOrderRequest{
		IsDividendReinvestment: false,
		RequestID:              testRequestID,
		Price:                  testPrice,
		Volume:                 testVolume,
		AccountID:              testAccountID,
		Side:                   OrderSideBuy,
		OrderbookID:            testOrderbookID,
		Metadata: OrderMetadata{
			OrderEntryMode:  "ADVANCED",
			HasTouchedPrice: "true",
		},
		Condition: OrderConditionNormal,
	}

	resp, err := avanza.Trading.PlaceOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}

	if got, want := resp.OrderRequestStatus, OrderRequestStatusSuccess; got != want {
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
		_ = json.NewEncoder(w).Encode(PlaceOrderResponse{
			OrderRequestStatus: "ERROR",
			Message:            "Insufficient funds",
			Parameters:         []string{},
			OrderID:            "",
		})
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	req := &PlaceOrderRequest{
		RequestID:   testRequestID,
		Price:       testPrice,
		Volume:      testVolume,
		AccountID:   testAccountID,
		Side:        OrderSideBuy,
		OrderbookID: testOrderbookID,
		Condition:   OrderConditionNormal,
	}

	resp, err := avanza.Trading.PlaceOrder(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if got, want := resp.OrderRequestStatus, OrderRequestStatusError; got != want {
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

	req := &PlaceOrderRequest{
		RequestID:   testRequestID,
		Price:       testPrice,
		Volume:      testVolume,
		AccountID:   testAccountID,
		Side:        OrderSideBuy,
		OrderbookID: testOrderbookID,
		Condition:   OrderConditionNormal,
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
		var req PlaceOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if got, want := req.Side, OrderSideSell; got != want {
			t.Errorf("req.Side = %v, want %v", got, want)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(PlaceOrderResponse{
			OrderRequestStatus: "SUCCESS",
			OrderID:            testOrderID2,
		})
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	req := &PlaceOrderRequest{
		RequestID:   testRequestID,
		Price:       testPrice,
		Volume:      testVolume,
		AccountID:   testAccountID,
		Side:        OrderSideSell,
		OrderbookID: testOrderbookID,
		Condition:   OrderConditionNormal,
	}

	resp, err := avanza.Trading.PlaceOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}

	if got, want := resp.OrderRequestStatus, OrderRequestStatusSuccess; got != want {
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

	req := &PlaceOrderRequest{
		RequestID:   testRequestID,
		Price:       testPrice,
		Volume:      testVolume,
		AccountID:   testAccountID,
		Side:        OrderSideBuy,
		OrderbookID: testOrderbookID,
		Condition:   OrderConditionNormal,
	}

	_, err := avanza.Trading.PlaceOrder(ctx, req)
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}
