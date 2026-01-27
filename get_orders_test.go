package avanza

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vmorsell/avanza-sdk-go/trading"
)

func TestGetOrders_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/trading/rest/orders" {
			t.Errorf("expected path /_api/trading/rest/orders, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(trading.GetOrdersResponse{
			Orders: []trading.Order{
				{
					OrderID:     "order-1",
					Volume:      100,
					Price:       10.50,
					Side:        trading.OrderSideBuy,
					OrderbookID: "5247",
					State:       "ACTIVE",
					Account: trading.OrderAccount{
						AccountID: "acc-1",
					},
					Orderbook: trading.OrderOrderbook{
						ID:   "5247",
						Name: "Volvo B",
					},
				},
				{
					OrderID:     "order-2",
					Volume:      50,
					Price:       200.00,
					Side:        trading.OrderSideSell,
					OrderbookID: "1234",
					State:       "ACTIVE",
				},
			},
			FundOrders:      []interface{}{},
			CancelledOrders: []interface{}{},
		})
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	resp, err := avanza.Trading.GetOrders(context.Background())
	if err != nil {
		t.Fatalf("GetOrders failed: %v", err)
	}

	if len(resp.Orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(resp.Orders))
	}

	if got, want := resp.Orders[0].OrderID, "order-1"; got != want {
		t.Errorf("orders[0].OrderID = %q, want %q", got, want)
	}
	if got, want := resp.Orders[0].Side, trading.OrderSideBuy; got != want {
		t.Errorf("orders[0].Side = %v, want %v", got, want)
	}
	if got, want := resp.Orders[0].Orderbook.Name, "Volvo B"; got != want {
		t.Errorf("orders[0].Orderbook.Name = %q, want %q", got, want)
	}
	if got, want := resp.Orders[1].Side, trading.OrderSideSell; got != want {
		t.Errorf("orders[1].Side = %v, want %v", got, want)
	}
}

func TestGetOrders_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	_, err := avanza.Trading.GetOrders(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetOrders_EmptyOrders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(trading.GetOrdersResponse{
			Orders:          []trading.Order{},
			FundOrders:      []interface{}{},
			CancelledOrders: []interface{}{},
		})
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	resp, err := avanza.Trading.GetOrders(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Orders) != 0 {
		t.Errorf("expected 0 orders, got %d", len(resp.Orders))
	}
}

func TestGetOrders_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := avanza.Trading.GetOrders(ctx)
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}

func TestGetStopLossOrders_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/trading/stoploss/" {
			t.Errorf("expected path /_api/trading/stoploss/, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]trading.StopLossOrder{
			{
				ID:     "sl-1",
				Status: trading.StopLossStatusSuccess,
				Account: trading.StopLossAccount{
					ID:   "acc-1",
					Name: "ISK",
				},
				Orderbook: trading.StopLossOrderbook{
					ID:   "5247",
					Name: "Volvo B",
				},
				Trigger: trading.StopLossTriggerResponse{
					Value:     90.0,
					Type:      trading.StopLossTriggerLessOrEqual,
					ValueType: trading.StopLossValueMonetary,
				},
				Order: trading.StopLossOrderDetails{
					Type:      trading.StopLossOrderEventSell,
					Price:     89.0,
					Volume:    100,
					ValidDays: 30,
					PriceType: trading.StopLossPriceMonetary,
				},
				Editable:  true,
				Deletable: true,
			},
		})
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	orders, err := avanza.Trading.GetStopLossOrders(context.Background())
	if err != nil {
		t.Fatalf("GetStopLossOrders failed: %v", err)
	}

	if len(orders) != 1 {
		t.Fatalf("expected 1 stop loss order, got %d", len(orders))
	}
	if got, want := orders[0].ID, "sl-1"; got != want {
		t.Errorf("orders[0].ID = %q, want %q", got, want)
	}
	if got, want := orders[0].Trigger.Value, 90.0; got != want {
		t.Errorf("orders[0].Trigger.Value = %v, want %v", got, want)
	}
	if got, want := orders[0].Order.Type, trading.StopLossOrderEventSell; got != want {
		t.Errorf("orders[0].Order.Type = %v, want %v", got, want)
	}
	if !orders[0].Editable {
		t.Error("expected orders[0].Editable to be true")
	}
}

func TestGetStopLossOrders_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	_, err := avanza.Trading.GetStopLossOrders(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetStopLossOrders_EmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	orders, err := avanza.Trading.GetStopLossOrders(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(orders) != 0 {
		t.Errorf("expected 0 orders, got %d", len(orders))
	}
}

func TestGetStopLossOrders_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := avanza.Trading.GetStopLossOrders(ctx)
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}
