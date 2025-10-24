package trading

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vmorsell/avanza-sdk-go/internal/client"
)

func TestGetPreliminaryFee_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/trading/preliminary-fee/preliminaryfee" {
			t.Errorf("expected path /_api/trading/preliminary-fee/preliminaryfee, got %s", r.URL.Path)
		}

		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		// Verify request body
		var req PreliminaryFeeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if got, want := req.AccountID, testAccountID; got != want {
			t.Errorf("req.AccountID = %v, want %v", got, want)
		}

		if got, want := req.OrderbookID, testOrderbookID; got != want {
			t.Errorf("req.OrderbookID = %v, want %v", got, want)
		}

		if got, want := req.Price, fmt.Sprintf("%.4f", testPrice); got != want {
			t.Errorf("req.Price = %v, want %v", got, want)
		}

		if got, want := req.Volume, fmt.Sprintf("%d", testVolume); got != want {
			t.Errorf("req.Volume = %v, want %v", got, want)
		}

		if got, want := req.Side, "BUY"; got != want {
			t.Errorf("req.Side = %v, want %v", got, want)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(PreliminaryFeeResponse{
			Commission:          testCommission,
			MarketFees:          testMarketFees,
			TotalFees:           testTotalFees,
			TotalSum:            testTotalSum,
			TotalSumWithoutFees: testTotalSumWithoutFees,
			OrderbookCurrency:   testOrderbookCurrency,
			TransactionTax:      nil,
			CurrencyExchangeFee: CurrencyExchangeFee{
				Rate: "",
				Sum:  "",
			},
			Campaign: nil,
		})
	}))
	defer server.Close()

	c := client.NewClient(client.WithBaseURL(server.URL))
	s := NewService(c)

	req := &PreliminaryFeeRequest{
		AccountID:   testAccountID,
		OrderbookID: testOrderbookID,
		Price:       fmt.Sprintf("%.4f", testPrice),
		Volume:      fmt.Sprintf("%d", testVolume),
		Side:        "BUY",
	}

	resp, err := s.GetPreliminaryFee(context.Background(), req)
	if err != nil {
		t.Fatalf("GetPreliminaryFee failed: %v", err)
	}

	if got, want := resp.Commission, testCommission; got != want {
		t.Errorf("resp.Commission = %v, want %v", got, want)
	}

	if got, want := resp.TotalFees, testTotalFees; got != want {
		t.Errorf("resp.TotalFees = %v, want %v", got, want)
	}

	if got, want := resp.TotalSum, testTotalSum; got != want {
		t.Errorf("resp.TotalSum = %v, want %v", got, want)
	}

	if got, want := resp.OrderbookCurrency, testOrderbookCurrency; got != want {
		t.Errorf("resp.OrderbookCurrency = %v, want %v", got, want)
	}
}

func TestGetPreliminaryFee_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer server.Close()

	c := client.NewClient(client.WithBaseURL(server.URL))
	s := NewService(c)

	req := &PreliminaryFeeRequest{
		AccountID:   testAccountID,
		OrderbookID: testOrderbookID,
		Price:       fmt.Sprintf("%.4f", testPrice),
		Volume:      fmt.Sprintf("%d", testVolume),
		Side:        "BUY",
	}

	_, err := s.GetPreliminaryFee(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPreliminaryFee_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay to allow context cancellation
		<-r.Context().Done()
	}))
	defer server.Close()

	c := client.NewClient(client.WithBaseURL(server.URL))
	s := NewService(c)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := &PreliminaryFeeRequest{
		AccountID:   testAccountID,
		OrderbookID: testOrderbookID,
		Price:       fmt.Sprintf("%.4f", testPrice),
		Volume:      fmt.Sprintf("%d", testVolume),
		Side:        "BUY",
	}

	_, err := s.GetPreliminaryFee(ctx, req)
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}
