package trading

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vmorsell/avanza-sdk-go/client"
)

func newTestClient(baseURL string) *client.Client {
	return client.NewClient(client.WithBaseURL(baseURL))
}

// --- PlaceOrder ---

func TestPlaceOrder_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/_api/trading-critical/rest/order/new" {
			t.Errorf("path = %s, want /_api/trading-critical/rest/order/new", r.URL.Path)
		}

		var req PlaceOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.AccountID != "12345" {
			t.Errorf("accountId = %q, want %q", req.AccountID, "12345")
		}
		if req.Side != OrderSideBuy {
			t.Errorf("side = %q, want %q", req.Side, OrderSideBuy)
		}
		if req.Price != 100.0 {
			t.Errorf("price = %f, want 100.0", req.Price)
		}
		if req.Volume != 10 {
			t.Errorf("volume = %d, want 10", req.Volume)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(PlaceOrderResponse{
			OrderRequestStatus: OrderRequestStatusSuccess,
			OrderID:            "999",
		})
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	resp, err := svc.PlaceOrder(context.Background(), &PlaceOrderRequest{
		AccountID:   "12345",
		OrderbookID: "5246",
		Price:       100.0,
		Volume:      10,
		Side:        OrderSideBuy,
		Condition:   OrderConditionNormal,
	})
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}
	if resp.OrderID != "999" {
		t.Errorf("OrderID = %q, want %q", resp.OrderID, "999")
	}
}

func TestPlaceOrder_ValidationErrors(t *testing.T) {
	svc := NewService(newTestClient("http://localhost"))

	tests := []struct {
		name string
		req  *PlaceOrderRequest
	}{
		{"empty accountId", &PlaceOrderRequest{OrderbookID: "1", Price: 1, Volume: 1, Side: OrderSideBuy, Condition: OrderConditionNormal}},
		{"empty orderbookId", &PlaceOrderRequest{AccountID: "1", Price: 1, Volume: 1, Side: OrderSideBuy, Condition: OrderConditionNormal}},
		{"zero price", &PlaceOrderRequest{AccountID: "1", OrderbookID: "1", Price: 0, Volume: 1, Side: OrderSideBuy, Condition: OrderConditionNormal}},
		{"zero volume", &PlaceOrderRequest{AccountID: "1", OrderbookID: "1", Price: 1, Volume: 0, Side: OrderSideBuy, Condition: OrderConditionNormal}},
		{"invalid side", &PlaceOrderRequest{AccountID: "1", OrderbookID: "1", Price: 1, Volume: 1, Side: "INVALID", Condition: OrderConditionNormal}},
		{"invalid condition", &PlaceOrderRequest{AccountID: "1", OrderbookID: "1", Price: 1, Volume: 1, Side: OrderSideBuy, Condition: "INVALID"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.PlaceOrder(context.Background(), tt.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestPlaceOrder_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	_, err := svc.PlaceOrder(context.Background(), &PlaceOrderRequest{
		AccountID: "1", OrderbookID: "1", Price: 1, Volume: 1, Side: OrderSideBuy, Condition: OrderConditionNormal,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var httpErr *client.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *client.HTTPError, got %T", err)
	}
	if httpErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", httpErr.StatusCode, http.StatusBadRequest)
	}
}

func TestPlaceOrder_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(PlaceOrderResponse{
			OrderRequestStatus: OrderRequestStatusError,
			Message:            "insufficient funds",
		})
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	_, err := svc.PlaceOrder(context.Background(), &PlaceOrderRequest{
		AccountID: "1", OrderbookID: "1", Price: 1, Volume: 1, Side: OrderSideBuy, Condition: OrderConditionNormal,
	})
	if err == nil {
		t.Fatal("expected error for ERROR status, got nil")
	}
}

// --- ModifyOrder ---

func TestModifyOrder_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/_api/trading-critical/rest/order/modify" {
			t.Errorf("path = %s, want /_api/trading-critical/rest/order/modify", r.URL.Path)
		}

		var req ModifyOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.OrderID != "999" {
			t.Errorf("orderId = %q, want %q", req.OrderID, "999")
		}
		if req.Price != 335.0 {
			t.Errorf("price = %f, want 335.0", req.Price)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ModifyOrderResponse{
			OrderRequestStatus: OrderRequestStatusSuccess,
			OrderID:            "999",
		})
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	resp, err := svc.ModifyOrder(context.Background(), &ModifyOrderRequest{
		OrderID:   "999",
		AccountID: "12345",
		Price:     335.0,
		Volume:    10,
	})
	if err != nil {
		t.Fatalf("ModifyOrder failed: %v", err)
	}
	if resp.OrderID != "999" {
		t.Errorf("OrderID = %q, want %q", resp.OrderID, "999")
	}
}

func TestModifyOrder_ValidationErrors(t *testing.T) {
	svc := NewService(newTestClient("http://localhost"))

	tests := []struct {
		name string
		req  *ModifyOrderRequest
	}{
		{"empty orderId", &ModifyOrderRequest{AccountID: "1", Price: 1, Volume: 1}},
		{"empty accountId", &ModifyOrderRequest{OrderID: "1", Price: 1, Volume: 1}},
		{"zero price", &ModifyOrderRequest{OrderID: "1", AccountID: "1", Price: 0, Volume: 1}},
		{"zero volume", &ModifyOrderRequest{OrderID: "1", AccountID: "1", Price: 1, Volume: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ModifyOrder(context.Background(), tt.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestModifyOrder_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ModifyOrderResponse{
			OrderRequestStatus: OrderRequestStatusError,
			Message:            "order not found",
		})
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	_, err := svc.ModifyOrder(context.Background(), &ModifyOrderRequest{
		OrderID: "1", AccountID: "1", Price: 1, Volume: 1,
	})
	if err == nil {
		t.Fatal("expected error for ERROR status, got nil")
	}
}

// --- DeleteOrder ---

func TestDeleteOrder_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/_api/trading-critical/rest/order/delete" {
			t.Errorf("path = %s, want /_api/trading-critical/rest/order/delete", r.URL.Path)
		}

		var req DeleteOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.OrderID != "999" {
			t.Errorf("orderId = %q, want %q", req.OrderID, "999")
		}
		if req.AccountID != "12345" {
			t.Errorf("accountId = %q, want %q", req.AccountID, "12345")
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(DeleteOrderResponse{
			OrderRequestStatus: OrderRequestStatusSuccess,
			OrderID:            "999",
		})
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	resp, err := svc.DeleteOrder(context.Background(), &DeleteOrderRequest{
		AccountID: "12345",
		OrderID:   "999",
	})
	if err != nil {
		t.Fatalf("DeleteOrder failed: %v", err)
	}
	if resp.OrderID != "999" {
		t.Errorf("OrderID = %q, want %q", resp.OrderID, "999")
	}
}

func TestDeleteOrder_ValidationErrors(t *testing.T) {
	svc := NewService(newTestClient("http://localhost"))

	tests := []struct {
		name string
		req  *DeleteOrderRequest
	}{
		{"empty accountId", &DeleteOrderRequest{OrderID: "1"}},
		{"empty orderId", &DeleteOrderRequest{AccountID: "1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.DeleteOrder(context.Background(), tt.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestDeleteOrder_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(DeleteOrderResponse{
			OrderRequestStatus: OrderRequestStatusError,
			Message:            "order already deleted",
		})
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	_, err := svc.DeleteOrder(context.Background(), &DeleteOrderRequest{
		AccountID: "1", OrderID: "1",
	})
	if err == nil {
		t.Fatal("expected error for ERROR status, got nil")
	}
}

// --- GetOrder ---

func TestGetOrder_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/_api/trading-critical/rest/order/find" {
			t.Errorf("path = %s, want /_api/trading-critical/rest/order/find", r.URL.Path)
		}
		if r.URL.Query().Get("orderId") != "867695279" {
			t.Errorf("orderId = %q, want %q", r.URL.Query().Get("orderId"), "867695279")
		}
		if r.URL.Query().Get("cAccountId") != "84039" {
			t.Errorf("cAccountId = %q, want %q", r.URL.Query().Get("cAccountId"), "84039")
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(GetOrderResponse{
			OrderID:         "867695279",
			OrderbookID:     "5246",
			Side:            OrderSideBuy,
			State:           "ACTIVE",
			MarketReference: "1009",
			Price:           330.0,
			Volume:          10,
			OriginalVolume:  10,
			AccountID:       "84039",
			Condition:       OrderConditionNormal,
			ValidUntil:      "2026-03-30",
			Modifiable:      true,
			Deletable:       true,
		})
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	resp, err := svc.GetOrder(context.Background(), &GetOrderRequest{
		OrderID:   "867695279",
		AccountID: "84039",
	})
	if err != nil {
		t.Fatalf("GetOrder failed: %v", err)
	}
	if resp.OrderID != "867695279" {
		t.Errorf("OrderID = %q, want %q", resp.OrderID, "867695279")
	}
	if resp.Side != OrderSideBuy {
		t.Errorf("Side = %q, want %q", resp.Side, OrderSideBuy)
	}
	if resp.State != "ACTIVE" {
		t.Errorf("State = %q, want %q", resp.State, "ACTIVE")
	}
	if resp.Price != 330.0 {
		t.Errorf("Price = %f, want 330.0", resp.Price)
	}
	if !resp.Modifiable {
		t.Error("Modifiable = false, want true")
	}
}

func TestGetOrder_ValidationErrors(t *testing.T) {
	svc := NewService(newTestClient("http://localhost"))

	tests := []struct {
		name string
		req  *GetOrderRequest
	}{
		{"empty orderId", &GetOrderRequest{AccountID: "1"}},
		{"empty accountId", &GetOrderRequest{OrderID: "1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GetOrder(context.Background(), tt.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestGetOrder_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	_, err := svc.GetOrder(context.Background(), &GetOrderRequest{
		OrderID: "999", AccountID: "1",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var httpErr *client.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *client.HTTPError, got %T", err)
	}
	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", httpErr.StatusCode, http.StatusNotFound)
	}
}

// --- GetOrders ---

func TestGetOrders_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/_api/trading/rest/orders" {
			t.Errorf("path = %s, want /_api/trading/rest/orders", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(GetOrdersResponse{
			Orders: []Order{
				{OrderID: "111", Side: OrderSideBuy, Price: 100.0},
				{OrderID: "222", Side: OrderSideSell, Price: 200.0},
			},
		})
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	resp, err := svc.GetOrders(context.Background())
	if err != nil {
		t.Fatalf("GetOrders failed: %v", err)
	}
	if len(resp.Orders) != 2 {
		t.Fatalf("len(Orders) = %d, want 2", len(resp.Orders))
	}
	if resp.Orders[0].OrderID != "111" {
		t.Errorf("Orders[0].OrderID = %q, want %q", resp.Orders[0].OrderID, "111")
	}
}

func TestGetOrders_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	_, err := svc.GetOrders(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var httpErr *client.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *client.HTTPError, got %T", err)
	}
}

// --- ValidateOrder ---

func TestValidateOrder_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/_api/trading-critical/rest/order/validation/validate" {
			t.Errorf("path = %s, want /_api/trading-critical/rest/order/validation/validate", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ValidateOrderResponse{
			CommissionWarning:  ValidationResult{Valid: true},
			EmployeeValidation: ValidationResult{Valid: true},
		})
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	resp, err := svc.ValidateOrder(context.Background(), &ValidateOrderRequest{
		AccountID:   "1",
		OrderbookID: "5246",
		Price:       100.0,
		Volume:      10,
		Side:        OrderSideBuy,
		Condition:   OrderConditionNormal,
		ISIN:        "SE0000107419",
		Currency:    "SEK",
		MarketPlace: "XSTO",
	})
	if err != nil {
		t.Fatalf("ValidateOrder failed: %v", err)
	}
	if !resp.CommissionWarning.Valid {
		t.Error("CommissionWarning.Valid = false, want true")
	}
}

func TestValidateOrder_ValidationErrors(t *testing.T) {
	svc := NewService(newTestClient("http://localhost"))
	base := ValidateOrderRequest{
		AccountID: "1", OrderbookID: "1", Price: 1, Volume: 1,
		Side: OrderSideBuy, Condition: OrderConditionNormal,
		ISIN: "X", Currency: "SEK", MarketPlace: "XSTO",
	}

	tests := []struct {
		name   string
		modify func(*ValidateOrderRequest)
	}{
		{"empty accountId", func(r *ValidateOrderRequest) { r.AccountID = "" }},
		{"empty orderbookId", func(r *ValidateOrderRequest) { r.OrderbookID = "" }},
		{"zero price", func(r *ValidateOrderRequest) { r.Price = 0 }},
		{"zero volume", func(r *ValidateOrderRequest) { r.Volume = 0 }},
		{"invalid side", func(r *ValidateOrderRequest) { r.Side = "INVALID" }},
		{"invalid condition", func(r *ValidateOrderRequest) { r.Condition = "INVALID" }},
		{"empty isin", func(r *ValidateOrderRequest) { r.ISIN = "" }},
		{"empty currency", func(r *ValidateOrderRequest) { r.Currency = "" }},
		{"empty marketPlace", func(r *ValidateOrderRequest) { r.MarketPlace = "" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := base
			tt.modify(&req)
			_, err := svc.ValidateOrder(context.Background(), &req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

// --- GetPreliminaryFee ---

func TestGetPreliminaryFee_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/_api/trading/preliminary-fee/preliminaryfee" {
			t.Errorf("path = %s, want /_api/trading/preliminary-fee/preliminaryfee", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(PreliminaryFeeResponse{
			Commission: "39",
			TotalFees:  "39",
			TotalSum:   "3039",
		})
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	resp, err := svc.GetPreliminaryFee(context.Background(), &PreliminaryFeeRequest{
		AccountID:   "1",
		OrderbookID: "5246",
		Price:       "300.0",
		Volume:      "10",
		Side:        OrderSideBuy,
	})
	if err != nil {
		t.Fatalf("GetPreliminaryFee failed: %v", err)
	}
	if resp.TotalFees != "39" {
		t.Errorf("TotalFees = %q, want %q", resp.TotalFees, "39")
	}
}

func TestGetPreliminaryFee_ValidationErrors(t *testing.T) {
	svc := NewService(newTestClient("http://localhost"))

	tests := []struct {
		name string
		req  *PreliminaryFeeRequest
	}{
		{"empty accountId", &PreliminaryFeeRequest{OrderbookID: "1", Price: "1", Volume: "1", Side: OrderSideBuy}},
		{"empty orderbookId", &PreliminaryFeeRequest{AccountID: "1", Price: "1", Volume: "1", Side: OrderSideBuy}},
		{"empty price", &PreliminaryFeeRequest{AccountID: "1", OrderbookID: "1", Price: "", Volume: "1", Side: OrderSideBuy}},
		{"invalid price", &PreliminaryFeeRequest{AccountID: "1", OrderbookID: "1", Price: "abc", Volume: "1", Side: OrderSideBuy}},
		{"zero price", &PreliminaryFeeRequest{AccountID: "1", OrderbookID: "1", Price: "0", Volume: "1", Side: OrderSideBuy}},
		{"empty volume", &PreliminaryFeeRequest{AccountID: "1", OrderbookID: "1", Price: "1", Volume: "", Side: OrderSideBuy}},
		{"invalid volume", &PreliminaryFeeRequest{AccountID: "1", OrderbookID: "1", Price: "1", Volume: "abc", Side: OrderSideBuy}},
		{"zero volume", &PreliminaryFeeRequest{AccountID: "1", OrderbookID: "1", Price: "1", Volume: "0", Side: OrderSideBuy}},
		{"invalid side", &PreliminaryFeeRequest{AccountID: "1", OrderbookID: "1", Price: "1", Volume: "1", Side: "INVALID"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GetPreliminaryFee(context.Background(), tt.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}
