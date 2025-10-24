package trading

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vmorsell/avanza-sdk-go/internal/client"
)

// Stop loss test constants
const (
	testStopLossTriggerValue = 200.0
	testStopLossOrderPrice   = 200.0
	testStopLossOrderVolume  = 3
	testStopLossValidDays    = 8
	testStopLossValidUntil   = "2025-11-23"
	testStopLossOrderID      = "A4^1758088943198^1705191"
	testParentStopLossID     = "0"
)

func TestPlaceStopLoss_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/trading/stoploss/new" {
			t.Errorf("expected path /_api/trading/stoploss/new, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// Verify request body
		var req PlaceStopLossRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if got, want := req.AccountID, testAccountID; got != want {
			t.Errorf("req.AccountID = %v, want %v", got, want)
		}

		if got, want := req.OrderBookID, testOrderbookID; got != want {
			t.Errorf("req.OrderBookID = %v, want %v", got, want)
		}

		if got, want := req.StopLossTrigger.Type, StopLossTriggerLessOrEqual; got != want {
			t.Errorf("req.StopLossTrigger.Type = %v, want %v", got, want)
		}

		if got, want := req.StopLossTrigger.Value, testStopLossTriggerValue; got != want {
			t.Errorf("req.StopLossTrigger.Value = %v, want %v", got, want)
		}

		if got, want := req.StopLossOrderEvent.Type, StopLossOrderEventBuy; got != want {
			t.Errorf("req.StopLossOrderEvent.Type = %v, want %v", got, want)
		}

		if got, want := req.StopLossOrderEvent.Price, testStopLossOrderPrice; got != want {
			t.Errorf("req.StopLossOrderEvent.Price = %v, want %v", got, want)
		}

		if got, want := req.StopLossOrderEvent.Volume, testStopLossOrderVolume; got != want {
			t.Errorf("req.StopLossOrderEvent.Volume = %v, want %v", got, want)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(PlaceStopLossResponse{
			Status:          "SUCCESS",
			StopLossOrderID: testStopLossOrderID,
		})
	}))
	defer server.Close()

	c := client.NewClient(client.WithBaseURL(server.URL))
	s := NewService(c)

	req := &PlaceStopLossRequest{
		ParentStopLossID: testParentStopLossID,
		AccountID:        testAccountID,
		OrderBookID:      testOrderbookID,
		StopLossTrigger: StopLossTrigger{
			Type:                      StopLossTriggerLessOrEqual,
			Value:                     testStopLossTriggerValue,
			ValueType:                 StopLossValueMonetary,
			ValidUntil:                testStopLossValidUntil,
			TriggerOnMarketMakerQuote: false,
		},
		StopLossOrderEvent: StopLossOrderEvent{
			Type:                StopLossOrderEventBuy,
			Price:               testStopLossOrderPrice,
			Volume:              testStopLossOrderVolume,
			ValidDays:           testStopLossValidDays,
			PriceType:           StopLossPriceMonetary,
			ShortSellingAllowed: false,
		},
	}

	resp, err := s.PlaceStopLoss(context.Background(), req)
	if err != nil {
		t.Fatalf("PlaceStopLoss failed: %v", err)
	}

	if got, want := resp.Status, "SUCCESS"; got != want {
		t.Errorf("resp.Status = %v, want %v", got, want)
	}

	if got, want := resp.StopLossOrderID, testStopLossOrderID; got != want {
		t.Errorf("resp.StopLossOrderID = %v, want %v", got, want)
	}
}

func TestPlaceStopLoss_FailedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(PlaceStopLossResponse{
			Status:          "ERROR",
			StopLossOrderID: "",
		})
	}))
	defer server.Close()

	c := client.NewClient(client.WithBaseURL(server.URL))
	s := NewService(c)

	req := &PlaceStopLossRequest{
		ParentStopLossID: testParentStopLossID,
		AccountID:        testAccountID,
		OrderBookID:      testOrderbookID,
		StopLossTrigger: StopLossTrigger{
			Type:      StopLossTriggerLessOrEqual,
			Value:     testStopLossTriggerValue,
			ValueType: StopLossValueMonetary,
		},
		StopLossOrderEvent: StopLossOrderEvent{
			Type:      StopLossOrderEventBuy,
			Price:     testStopLossOrderPrice,
			Volume:    testStopLossOrderVolume,
			ValidDays: testStopLossValidDays,
		},
	}

	resp, err := s.PlaceStopLoss(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if got, want := resp.Status, "ERROR"; got != want {
		t.Errorf("resp.Status = %v, want %v", got, want)
	}
}

func TestPlaceStopLoss_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer server.Close()

	c := client.NewClient(client.WithBaseURL(server.URL))
	s := NewService(c)

	req := &PlaceStopLossRequest{
		ParentStopLossID: testParentStopLossID,
		AccountID:        testAccountID,
		OrderBookID:      testOrderbookID,
		StopLossTrigger: StopLossTrigger{
			Type:      StopLossTriggerLessOrEqual,
			Value:     testStopLossTriggerValue,
			ValueType: StopLossValueMonetary,
		},
		StopLossOrderEvent: StopLossOrderEvent{
			Type:      StopLossOrderEventBuy,
			Price:     testStopLossOrderPrice,
			Volume:    testStopLossOrderVolume,
			ValidDays: testStopLossValidDays,
		},
	}

	_, err := s.PlaceStopLoss(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPlaceStopLoss_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay to allow context cancellation
		<-r.Context().Done()
	}))
	defer server.Close()

	c := client.NewClient(client.WithBaseURL(server.URL))
	s := NewService(c)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := &PlaceStopLossRequest{
		ParentStopLossID: testParentStopLossID,
		AccountID:        testAccountID,
		OrderBookID:      testOrderbookID,
		StopLossTrigger: StopLossTrigger{
			Type:      StopLossTriggerLessOrEqual,
			Value:     testStopLossTriggerValue,
			ValueType: StopLossValueMonetary,
		},
		StopLossOrderEvent: StopLossOrderEvent{
			Type:      StopLossOrderEventBuy,
			Price:     testStopLossOrderPrice,
			Volume:    testStopLossOrderVolume,
			ValidDays: testStopLossValidDays,
		},
	}

	_, err := s.PlaceStopLoss(ctx, req)
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}
