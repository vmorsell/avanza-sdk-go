package accounts

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vmorsell/avanza-sdk-go/client"
)

func newTestClient(baseURL string) *client.Client {
	return client.NewClient(client.WithBaseURL(baseURL))
}

func TestGetOverview_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/account-overview/overview/categorizedAccounts" {
			t.Errorf("expected path /_api/account-overview/overview/categorizedAccounts, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(AccountOverview{
			Categories: []Category{
				{
					ID:   "cat-1",
					Name: "Sparande",
					TotalValue: Money{
						Value:            100000.50,
						Unit:             "SEK",
						UnitType:         "CURRENCY",
						DecimalPrecision: 2,
					},
				},
			},
			Accounts: []Account{
				{
					ID:         "acc-1",
					CategoryID: "cat-1",
					Type:       "ISK",
					Status:     "ACTIVE",
				},
			},
			Loans: []Loan{},
		})
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	overview, err := svc.GetOverview(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(overview.Categories) != 1 {
		t.Fatalf("expected 1 category, got %d", len(overview.Categories))
	}
	if got, want := overview.Categories[0].Name, "Sparande"; got != want {
		t.Errorf("category name = %q, want %q", got, want)
	}
	if len(overview.Accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(overview.Accounts))
	}
	if got, want := overview.Accounts[0].Type, "ISK"; got != want {
		t.Errorf("account type = %q, want %q", got, want)
	}
}

func TestGetOverview_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	_, err := svc.GetOverview(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetOverview_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.GetOverview(ctx)
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}

func TestGetTradingAccounts_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/trading-critical/rest/accounts" {
			t.Errorf("expected path /_api/trading-critical/rest/accounts, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]TradingAccount{
			{
				Name:                 "ISK",
				AccountID:            "acc-1",
				AccountTypeName:      "Investeringssparkonto",
				AccountType:          "ISK",
				AvailableForPurchase: 50000.00,
				IsTradable:           true,
				URLParameterID:       "abc123",
			},
			{
				Name:                 "KF",
				AccountID:            "acc-2",
				AccountTypeName:      "Kapitalförsäkring",
				AccountType:          "KF",
				AvailableForPurchase: 25000.00,
				IsTradable:           true,
				URLParameterID:       "def456",
			},
		})
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	accounts, err := svc.GetTradingAccounts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(accounts))
	}
	if got, want := accounts[0].AccountID, "acc-1"; got != want {
		t.Errorf("accounts[0].AccountID = %q, want %q", got, want)
	}
	if got, want := accounts[1].AccountType, "KF"; got != want {
		t.Errorf("accounts[1].AccountType = %q, want %q", got, want)
	}
	if !accounts[0].IsTradable {
		t.Error("expected accounts[0].IsTradable to be true")
	}
}

func TestGetTradingAccounts_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	_, err := svc.GetTradingAccounts(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetTradingAccounts_EmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	accounts, err := svc.GetTradingAccounts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(accounts) != 0 {
		t.Errorf("expected 0 accounts, got %d", len(accounts))
	}
}

func TestGetPositions_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/position-data/positions/abc123" {
			t.Errorf("expected path /_api/position-data/positions/abc123, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(AccountPositions{
			WithOrderbook: []AccountPosition{
				{
					ID: "pos-1",
					Instrument: Instrument{
						ID:       "inst-1",
						Name:     "Volvo B",
						Currency: "SEK",
						ISIN:     "SE0000115446",
					},
					Value: Money{
						Value:            5000.00,
						Unit:             "SEK",
						DecimalPrecision: 2,
					},
				},
			},
			CashPositions: []CashPosition{
				{
					TotalBalance: Money{Value: 10000.00, Unit: "SEK"},
					ID:           "cash-1",
				},
			},
		})
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	positions, err := svc.GetPositions(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(positions.WithOrderbook) != 1 {
		t.Fatalf("expected 1 position, got %d", len(positions.WithOrderbook))
	}
	if got, want := positions.WithOrderbook[0].Instrument.Name, "Volvo B"; got != want {
		t.Errorf("instrument name = %q, want %q", got, want)
	}
	if len(positions.CashPositions) != 1 {
		t.Fatalf("expected 1 cash position, got %d", len(positions.CashPositions))
	}
	if got, want := positions.CashPositions[0].TotalBalance.Value, 10000.00; got != want {
		t.Errorf("cash balance = %v, want %v", got, want)
	}
}

func TestGetPositions_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	_, err := svc.GetPositions(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPositions_URLEscaping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The path-escaped form of "a/b" is "a%2Fb"
		if r.URL.RawPath != "/_api/position-data/positions/a%2Fb" {
			t.Errorf("expected escaped path, got raw=%q path=%q", r.URL.RawPath, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(AccountPositions{})
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	_, err := svc.GetPositions(context.Background(), "a/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetPositions_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.GetPositions(ctx, "abc123")
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}
