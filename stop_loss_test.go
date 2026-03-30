package avanza

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vmorsell/avanza-sdk-go/trading"
)

func TestPlaceStopLoss_Success(t *testing.T) {
	const (
		testOrderbookID          = "orderbookID"
		testAccountID            = "accountID"
		testStopLossTriggerValue = 200.0
		testStopLossOrderPrice   = 200.0
		testStopLossOrderVolume  = 3
		testStopLossValidDays    = 8
		testStopLossValidUntil   = "2025-11-23"
		testStopLossOrderID      = "A4^1758088943198^1705191"
		testParentStopLossID     = "0"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/trading/stoploss/new" {
			t.Errorf("expected path /_api/trading/stoploss/new, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// Verify request body
		var req trading.PlaceStopLossRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if got, want := req.AccountID, testAccountID; got != want {
			t.Errorf("req.AccountID = %v, want %v", got, want)
		}

		if got, want := req.OrderbookID, testOrderbookID; got != want {
			t.Errorf("req.OrderbookID = %v, want %v", got, want)
		}

		if got, want := req.StopLossTrigger.Type, trading.StopLossTriggerLessOrEqual; got != want {
			t.Errorf("req.StopLossTrigger.Type = %v, want %v", got, want)
		}

		if got, want := req.StopLossTrigger.Value, testStopLossTriggerValue; got != want {
			t.Errorf("req.StopLossTrigger.Value = %v, want %v", got, want)
		}

		if got, want := req.StopLossOrderEvent.Type, trading.StopLossOrderEventBuy; got != want {
			t.Errorf("req.StopLossOrderEvent.Type = %v, want %v", got, want)
		}

		if got, want := req.StopLossOrderEvent.Price, testStopLossOrderPrice; got != want {
			t.Errorf("req.StopLossOrderEvent.Price = %v, want %v", got, want)
		}

		if got, want := req.StopLossOrderEvent.Volume, testStopLossOrderVolume; got != want {
			t.Errorf("req.StopLossOrderEvent.Volume = %v, want %v", got, want)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(trading.PlaceStopLossResponse{
			Status:          trading.StopLossStatusSuccess,
			StopLossOrderID: testStopLossOrderID,
		})
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	req := &trading.PlaceStopLossRequest{
		ParentStopLossID: testParentStopLossID,
		AccountID:        testAccountID,
		OrderbookID:      testOrderbookID,
		StopLossTrigger: trading.StopLossTrigger{
			Type:                      trading.StopLossTriggerLessOrEqual,
			Value:                     testStopLossTriggerValue,
			ValueType:                 trading.StopLossValueMonetary,
			ValidUntil:                testStopLossValidUntil,
			TriggerOnMarketMakerQuote: false,
		},
		StopLossOrderEvent: trading.StopLossOrderEvent{
			Type:                trading.StopLossOrderEventBuy,
			Price:               testStopLossOrderPrice,
			Volume:              testStopLossOrderVolume,
			ValidDays:           testStopLossValidDays,
			PriceType:           trading.StopLossPriceMonetary,
			ShortSellingAllowed: false,
		},
	}

	resp, err := avanza.Trading.PlaceStopLoss(context.Background(), req)
	if err != nil {
		t.Fatalf("PlaceStopLoss failed: %v", err)
	}

	if got, want := resp.Status, trading.StopLossStatusSuccess; got != want {
		t.Errorf("resp.Status = %v, want %v", got, want)
	}

	if got, want := resp.StopLossOrderID, testStopLossOrderID; got != want {
		t.Errorf("resp.StopLossOrderID = %v, want %v", got, want)
	}
}

func TestPlaceStopLoss_FailedStatus(t *testing.T) {
	const (
		testOrderbookID          = "orderbookID"
		testAccountID            = "accountID"
		testStopLossTriggerValue = 200.0
		testStopLossOrderPrice   = 200.0
		testStopLossOrderVolume  = 3
		testStopLossValidDays    = 8
		testParentStopLossID     = "0"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(trading.PlaceStopLossResponse{
			Status:          trading.StopLossStatusError,
			StopLossOrderID: "",
		})
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	req := &trading.PlaceStopLossRequest{
		ParentStopLossID: testParentStopLossID,
		AccountID:        testAccountID,
		OrderbookID:      testOrderbookID,
		StopLossTrigger: trading.StopLossTrigger{
			Type:      trading.StopLossTriggerLessOrEqual,
			Value:     testStopLossTriggerValue,
			ValueType: trading.StopLossValueMonetary,
		},
		StopLossOrderEvent: trading.StopLossOrderEvent{
			Type:      trading.StopLossOrderEventBuy,
			Price:     testStopLossOrderPrice,
			Volume:    testStopLossOrderVolume,
			ValidDays: testStopLossValidDays,
			PriceType: trading.StopLossPriceMonetary,
		},
	}

	resp, err := avanza.Trading.PlaceStopLoss(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if got, want := resp.Status, trading.StopLossStatusError; got != want {
		t.Errorf("resp.Status = %v, want %v", got, want)
	}
}

func TestPlaceStopLoss_HTTPError(t *testing.T) {
	const (
		testOrderbookID          = "orderbookID"
		testAccountID            = "accountID"
		testStopLossTriggerValue = 200.0
		testStopLossOrderPrice   = 200.0
		testStopLossOrderVolume  = 3
		testStopLossValidDays    = 8
		testParentStopLossID     = "0"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	req := &trading.PlaceStopLossRequest{
		ParentStopLossID: testParentStopLossID,
		AccountID:        testAccountID,
		OrderbookID:      testOrderbookID,
		StopLossTrigger: trading.StopLossTrigger{
			Type:      trading.StopLossTriggerLessOrEqual,
			Value:     testStopLossTriggerValue,
			ValueType: trading.StopLossValueMonetary,
		},
		StopLossOrderEvent: trading.StopLossOrderEvent{
			Type:      trading.StopLossOrderEventBuy,
			Price:     testStopLossOrderPrice,
			Volume:    testStopLossOrderVolume,
			ValidDays: testStopLossValidDays,
			PriceType: trading.StopLossPriceMonetary,
		},
	}

	_, err := avanza.Trading.PlaceStopLoss(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPlaceStopLoss_ContextCancellation(t *testing.T) {
	const (
		testOrderbookID          = "orderbookID"
		testAccountID            = "accountID"
		testStopLossTriggerValue = 200.0
		testStopLossOrderPrice   = 200.0
		testStopLossOrderVolume  = 3
		testStopLossValidDays    = 8
		testParentStopLossID     = "0"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay to allow context cancellation
		<-r.Context().Done()
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := &trading.PlaceStopLossRequest{
		ParentStopLossID: testParentStopLossID,
		AccountID:        testAccountID,
		OrderbookID:      testOrderbookID,
		StopLossTrigger: trading.StopLossTrigger{
			Type:      trading.StopLossTriggerLessOrEqual,
			Value:     testStopLossTriggerValue,
			ValueType: trading.StopLossValueMonetary,
		},
		StopLossOrderEvent: trading.StopLossOrderEvent{
			Type:      trading.StopLossOrderEventBuy,
			Price:     testStopLossOrderPrice,
			Volume:    testStopLossOrderVolume,
			ValidDays: testStopLossValidDays,
			PriceType: trading.StopLossPriceMonetary,
		},
	}

	_, err := avanza.Trading.PlaceStopLoss(ctx, req)
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}

func TestGetStopLoss_Success(t *testing.T) {
	const (
		testAccountURLParamID = "OtKL0XM2WURBPEL5e1yc5w"
		testStopLossOrderID   = "A4^1773297345776^844590"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		wantPath := "/_api/trading/stoploss/" + testAccountURLParamID + "/A4%5E1773297345776%5E844590"
		if r.URL.RawPath != "" {
			if got := r.URL.RawPath; got != wantPath {
				t.Errorf("path = %s, want %s", got, wantPath)
			}
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(trading.StopLossOrder{
			ID:     testStopLossOrderID,
			Status: trading.StopLossStatusSuccess,
			Account: trading.StopLossAccount{
				ID:             "84039",
				Name:           "1 Bas",
				URLParameterID: testAccountURLParamID,
			},
			Orderbook: trading.StopLossOrderbook{
				ID:   "5246",
				Name: "Investor A",
			},
			Trigger: trading.StopLossTriggerResponse{
				Value:     310,
				Type:      trading.StopLossTriggerLessOrEqual,
				ValidUntil: "2026-04-29",
				ValueType: trading.StopLossValueMonetary,
			},
			Order: trading.StopLossOrderDetails{
				Type:      trading.StopLossOrderEventBuy,
				Price:     308,
				Volume:    10,
				ValidDays: 8,
				PriceType: trading.StopLossPriceMonetary,
			},
			Editable:  true,
			Deletable: true,
		})
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	resp, err := avanza.Trading.GetStopLoss(context.Background(), &trading.GetStopLossRequest{
		AccountURLParameterID: testAccountURLParamID,
		StopLossOrderID:       testStopLossOrderID,
	})
	if err != nil {
		t.Fatalf("GetStopLoss failed: %v", err)
	}

	if got, want := resp.ID, testStopLossOrderID; got != want {
		t.Errorf("resp.ID = %v, want %v", got, want)
	}
	if got, want := resp.Trigger.Value, 310.0; got != want {
		t.Errorf("resp.Trigger.Value = %v, want %v", got, want)
	}
	if !resp.Editable {
		t.Error("expected Editable to be true")
	}
}

func TestGetStopLoss_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	_, err := avanza.Trading.GetStopLoss(context.Background(), &trading.GetStopLossRequest{
		AccountURLParameterID: "abc",
		StopLossOrderID:       "xyz",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestModifyStopLoss_Success(t *testing.T) {
	const (
		testStopLossOrderID = "A4^1773297345776^844590"
		testAccountID       = "84039"
		testOrderbookID     = "5246"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/trading/stoploss/modify" {
			t.Errorf("expected path /_api/trading/stoploss/modify, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req trading.ModifyStopLossRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if got, want := req.StopLossOrderID, testStopLossOrderID; got != want {
			t.Errorf("req.StopLossOrderID = %v, want %v", got, want)
		}
		if got, want := req.AccountID, testAccountID; got != want {
			t.Errorf("req.AccountID = %v, want %v", got, want)
		}
		if got, want := req.OrderbookID, testOrderbookID; got != want {
			t.Errorf("req.OrderbookID = %v, want %v", got, want)
		}
		if req.TriggerAllChildren {
			t.Error("expected TriggerAllChildren to be false")
		}
		if got, want := req.StopLossOrderEvent.Price, 309.0; got != want {
			t.Errorf("req.StopLossOrderEvent.Price = %v, want %v", got, want)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(trading.PlaceStopLossResponse{
			Status:          trading.StopLossStatusSuccess,
			StopLossOrderID: testStopLossOrderID,
		})
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	resp, err := avanza.Trading.ModifyStopLoss(context.Background(), &trading.ModifyStopLossRequest{
		ParentStopLossID: "0",
		StopLossOrderID:  testStopLossOrderID,
		AccountID:        testAccountID,
		OrderbookID:      testOrderbookID,
		StopLossTrigger: trading.StopLossTrigger{
			Type:      trading.StopLossTriggerLessOrEqual,
			Value:     310,
			ValueType: trading.StopLossValueMonetary,
			ValidUntil: "2026-04-29",
		},
		StopLossOrderEvent: trading.StopLossOrderEvent{
			Type:      trading.StopLossOrderEventBuy,
			Price:     309,
			Volume:    10,
			ValidDays: 8,
			PriceType: trading.StopLossPriceMonetary,
		},
	})
	if err != nil {
		t.Fatalf("ModifyStopLoss failed: %v", err)
	}

	if got, want := resp.Status, trading.StopLossStatusSuccess; got != want {
		t.Errorf("resp.Status = %v, want %v", got, want)
	}
	if got, want := resp.StopLossOrderID, testStopLossOrderID; got != want {
		t.Errorf("resp.StopLossOrderID = %v, want %v", got, want)
	}
}

func TestModifyStopLoss_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	_, err := avanza.Trading.ModifyStopLoss(context.Background(), &trading.ModifyStopLossRequest{
		StopLossOrderID: "orderID",
		AccountID:       "accountID",
		OrderbookID:     "orderbookID",
		StopLossTrigger: trading.StopLossTrigger{
			Type:      trading.StopLossTriggerLessOrEqual,
			Value:     100,
			ValueType: trading.StopLossValueMonetary,
		},
		StopLossOrderEvent: trading.StopLossOrderEvent{
			Type:      trading.StopLossOrderEventSell,
			Price:     99,
			Volume:    5,
			ValidDays: 3,
			PriceType: trading.StopLossPriceMonetary,
		},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteStopLoss_Success(t *testing.T) {
	const (
		testAccountID       = "84039"
		testStopLossOrderID = "A4^1773297345776^844590"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}

		wantPath := "/_api/trading/stoploss/" + testAccountID + "/A4%5E1773297345776%5E844590"
		if r.URL.RawPath != "" {
			if got := r.URL.RawPath; got != wantPath {
				t.Errorf("path = %s, want %s", got, wantPath)
			}
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	err := avanza.Trading.DeleteStopLoss(context.Background(), &trading.DeleteStopLossRequest{
		AccountID:       testAccountID,
		StopLossOrderID: testStopLossOrderID,
	})
	if err != nil {
		t.Fatalf("DeleteStopLoss failed: %v", err)
	}
}

func TestDeleteStopLoss_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	err := avanza.Trading.DeleteStopLoss(context.Background(), &trading.DeleteStopLossRequest{
		AccountID:       "accountID",
		StopLossOrderID: "orderID",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
