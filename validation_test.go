package avanza

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateOrder_Success(t *testing.T) {
	const (
		testOrderbookID = "orderbookID"
		testAccountID   = "accountID"
		testPrice       = 2.0
		testVolume      = 1
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/trading-critical/rest/order/validation/validate" {
			t.Errorf("expected path /_api/trading-critical/rest/order/validation/validate, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// Verify request body
		var req ValidateOrderRequest
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

		if got, want := req.AccountID, testAccountID; got != want {
			t.Errorf("req.AccountID = %v, want %v", got, want)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ValidateOrderResponse{
			CommissionWarning: ValidationResult{
				Valid: false,
			},
			EmployeeValidation: ValidationResult{
				Valid: true,
			},
			LargeInScaleWarning: ValidationResult{
				Valid: true,
			},
			OrderValueLimitWarning: ValidationResult{
				Valid: true,
			},
			PriceRampingWarning: ValidationResult{
				Valid: true,
			},
			CanadaOddLotWarning: ValidationResult{
				Valid: true,
			},
		})
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	req := &ValidateOrderRequest{
		IsDividendReinvestment: false,
		RequestID:              nil,
		OrderRequestParameters: nil,
		Price:                  testPrice,
		Volume:                 testVolume,
		OpenVolume:             nil,
		AccountID:              testAccountID,
		Side:                   OrderSideBuy,
		OrderbookID:            testOrderbookID,
		ValidUntil:             nil,
		Metadata:               nil,
		Condition:              OrderConditionNormal,
		ISIN:                   "SE0015811963",
		Currency:               "SEK",
		MarketPlace:            "XSTO",
	}

	resp, err := avanza.ValidateOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("ValidateOrder failed: %v", err)
	}

	if got, want := resp.CommissionWarning.Valid, false; got != want {
		t.Errorf("resp.CommissionWarning.Valid = %v, want %v", got, want)
	}

	if got, want := resp.EmployeeValidation.Valid, true; got != want {
		t.Errorf("resp.EmployeeValidation.Valid = %v, want %v", got, want)
	}

	if got, want := resp.LargeInScaleWarning.Valid, true; got != want {
		t.Errorf("resp.LargeInScaleWarning.Valid = %v, want %v", got, want)
	}

	if got, want := resp.OrderValueLimitWarning.Valid, true; got != want {
		t.Errorf("resp.OrderValueLimitWarning.Valid = %v, want %v", got, want)
	}

	if got, want := resp.PriceRampingWarning.Valid, true; got != want {
		t.Errorf("resp.PriceRampingWarning.Valid = %v, want %v", got, want)
	}

	if got, want := resp.CanadaOddLotWarning.Valid, true; got != want {
		t.Errorf("resp.CanadaOddLotWarning.Valid = %v, want %v", got, want)
	}
}

func TestValidateOrder_HTTPError(t *testing.T) {
	const (
		testOrderbookID = "orderbookID"
		testAccountID   = "accountID"
		testPrice       = 2.0
		testVolume      = 1
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	req := &ValidateOrderRequest{
		Price:       testPrice,
		Volume:      testVolume,
		AccountID:   testAccountID,
		Side:        OrderSideBuy,
		OrderbookID: testOrderbookID,
		Condition:   OrderConditionNormal,
		ISIN:        "SE0015811963",
		Currency:    "SEK",
		MarketPlace: "XSTO",
	}

	_, err := avanza.ValidateOrder(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestValidateOrder_ContextCancellation(t *testing.T) {
	const (
		testOrderbookID = "orderbookID"
		testAccountID   = "accountID"
		testPrice       = 2.0
		testVolume      = 1
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay to allow context cancellation
		<-r.Context().Done()
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := &ValidateOrderRequest{
		Price:       testPrice,
		Volume:      testVolume,
		AccountID:   testAccountID,
		Side:        OrderSideBuy,
		OrderbookID: testOrderbookID,
		Condition:   OrderConditionNormal,
		ISIN:        "SE0015811963",
		Currency:    "SEK",
		MarketPlace: "XSTO",
	}

	_, err := avanza.ValidateOrder(ctx, req)
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}
