package accounts

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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
	if got, want := positions.CashPositions[0].TotalBalance.Value, 1000.00; got != want {
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

func TestGetTransactions_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/transactions/list" {
			t.Errorf("expected path /_api/transactions/list, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if got := r.URL.Query().Get("from"); got != "2025-08-01" {
			t.Errorf("expected from=2025-08-01, got %s", got)
		}
		if got := r.URL.Query().Get("to"); got != "2025-10-31" {
			t.Errorf("expected to=2025-10-31, got %s", got)
		}
		if got := r.URL.Query().Get("includeResult"); got != "false" {
			t.Errorf("expected includeResult=false, got %s", got)
		}

		instrumentName := "Test Instrument AB"
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(TransactionsResponse{
			Transactions: []Transaction{
				{
					ID:             "TXN-12345-001",
					Date:           "2025-10-28T00:00:00",
					SettlementDate: "2025-10-30",
					TradeDate:      "2025-10-28",
					Account: TransactionAccount{
						ID:             "12345",
						Name:           "Test Account",
						Type:           "INVESTERINGSSPARKONTO",
						URLParameterID: "test-url-id",
					},
					Orderbook: &TransactionOrderbook{
						ID:          "99999",
						FlagCode:    "SE",
						Name:        "Test Instrument AB",
						Marketplace: "First North Stockholm",
						Type:        "CERTIFICATE",
						Currency:    "SEK",
						ISIN:        "SE0000000001",
					},
					InstrumentName:     &instrumentName,
					Type:               "SELL",
					BackofficeType:     "SELL",
					BackofficeTypeText: "Sälj",
					Amount: &Money{
						Value:            1234.56,
						Unit:             "SEK",
						UnitType:         "MONETARY",
						DecimalPrecision: 2,
					},
					VerificationNumber: "0000000001",
				},
			},
			TransactionsAfterFiltering: 1,
			FirstTransactionDate:       "2020-01-01",
		})
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	resp, err := svc.GetTransactions(context.Background(), &TransactionsRequest{
		From: "2025-08-01",
		To:   "2025-10-31",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Transactions) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(resp.Transactions))
	}
	if got, want := resp.Transactions[0].ID, "TXN-12345-001"; got != want {
		t.Errorf("transaction ID = %q, want %q", got, want)
	}
	if got, want := resp.Transactions[0].Type, "SELL"; got != want {
		t.Errorf("transaction type = %q, want %q", got, want)
	}
	if resp.Transactions[0].Amount == nil {
		t.Fatal("expected amount to be set")
	}
	if got, want := resp.Transactions[0].Amount.Value, 123.456; got != want {
		t.Errorf("amount = %v, want %v", got, want)
	}
	if got, want := resp.FirstTransactionDate, "2020-01-01"; got != want {
		t.Errorf("firstTransactionDate = %q, want %q", got, want)
	}
}

func TestGetTransactions_MissingFromDate(t *testing.T) {
	c := newTestClient("http://unused")
	svc := NewService(c)

	_, err := svc.GetTransactions(context.Background(), &TransactionsRequest{
		To: "2025-10-31",
	})
	if err == nil {
		t.Fatal("expected error for missing from date, got nil")
	}
	if got, want := err.Error(), "from date is required"; got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestGetTransactions_MissingToDate(t *testing.T) {
	c := newTestClient("http://unused")
	svc := NewService(c)

	_, err := svc.GetTransactions(context.Background(), &TransactionsRequest{
		From: "2025-08-01",
	})
	if err == nil {
		t.Fatal("expected error for missing to date, got nil")
	}
	if got, want := err.Error(), "to date is required"; got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestGetTransactions_NilRequest(t *testing.T) {
	c := newTestClient("http://unused")
	svc := NewService(c)

	_, err := svc.GetTransactions(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request, got nil")
	}
	if got, want := err.Error(), "request is required"; got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestGetTransactions_InvalidFromDateFormat(t *testing.T) {
	c := newTestClient("http://unused")
	svc := NewService(c)

	_, err := svc.GetTransactions(context.Background(), &TransactionsRequest{
		From: "01-28-2025",
		To:   "2025-10-31",
	})
	if err == nil {
		t.Fatal("expected error for invalid from date format, got nil")
	}
	if got, want := err.Error(), "from date must be in YYYY-MM-DD format"; got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestGetTransactions_InvalidToDateFormat(t *testing.T) {
	c := newTestClient("http://unused")
	svc := NewService(c)

	_, err := svc.GetTransactions(context.Background(), &TransactionsRequest{
		From: "2025-08-01",
		To:   "31/10/2025",
	})
	if err == nil {
		t.Fatal("expected error for invalid to date format, got nil")
	}
	if got, want := err.Error(), "to date must be in YYYY-MM-DD format"; got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestGetTransactions_FromAfterTo(t *testing.T) {
	c := newTestClient("http://unused")
	svc := NewService(c)

	_, err := svc.GetTransactions(context.Background(), &TransactionsRequest{
		From: "2025-10-31",
		To:   "2025-08-01",
	})
	if err == nil {
		t.Fatal("expected error for from after to, got nil")
	}
	if got, want := err.Error(), "from date must not be after to date"; got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestGetTransactions_SameDayRange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("from"); got != "2025-10-15" {
			t.Errorf("expected from=2025-10-15, got %s", got)
		}
		if got := r.URL.Query().Get("to"); got != "2025-10-15" {
			t.Errorf("expected to=2025-10-15, got %s", got)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(TransactionsResponse{
			Transactions:               []Transaction{},
			TransactionsAfterFiltering: 0,
		})
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	_, err := svc.GetTransactions(context.Background(), &TransactionsRequest{
		From: "2025-10-15",
		To:   "2025-10-15",
	})
	if err != nil {
		t.Fatalf("same-day range should be valid, got error: %v", err)
	}
}

func TestGetTransactions_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	_, err := svc.GetTransactions(context.Background(), &TransactionsRequest{
		From: "2025-08-01",
		To:   "2025-10-31",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var httpErr *client.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *client.HTTPError, got %T", err)
	}
	if httpErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", httpErr.StatusCode, http.StatusUnauthorized)
	}
}

func TestGetTransactions_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	_, err := svc.GetTransactions(context.Background(), &TransactionsRequest{
		From: "2025-08-01",
		To:   "2025-10-31",
	})
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode response") {
		t.Errorf("error = %q, want to contain 'failed to decode response'", err.Error())
	}
}

func TestGetTransactions_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(TransactionsResponse{
			Transactions:               []Transaction{},
			TransactionsAfterFiltering: 0,
			FirstTransactionDate:       "2020-01-01",
		})
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	resp, err := svc.GetTransactions(context.Background(), &TransactionsRequest{
		From: "2025-08-01",
		To:   "2025-10-31",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Transactions) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(resp.Transactions))
	}
}

func TestGetTransactions_NilOptionalFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(TransactionsResponse{
			Transactions: []Transaction{
				{
					ID:   "TXN-001",
					Type: "DEPOSIT",
					Account: TransactionAccount{
						ID:   "12345",
						Name: "Test Account",
					},
					// Orderbook, Amount, InstrumentName are nil
				},
			},
			TransactionsAfterFiltering: 1,
		})
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	resp, err := svc.GetTransactions(context.Background(), &TransactionsRequest{
		From: "2025-08-01",
		To:   "2025-10-31",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Transactions) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(resp.Transactions))
	}
	tx := resp.Transactions[0]
	if tx.Orderbook != nil {
		t.Error("expected nil Orderbook")
	}
	if tx.Amount != nil {
		t.Error("expected nil Amount")
	}
	if tx.InstrumentName != nil {
		t.Error("expected nil InstrumentName")
	}
}

func TestGetTransactions_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.GetTransactions(ctx, &TransactionsRequest{
		From: "2025-08-01",
		To:   "2025-10-31",
	})
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}

func TestGetAggregatedValues_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/account-performance/aggregatedAccountsValues" {
			t.Errorf("expected path /_api/account-performance/aggregatedAccountsValues, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req AggregatedValuesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if len(req.EncryptedAccountIDs) != 2 {
			t.Errorf("expected 2 account IDs, got %d", len(req.EncryptedAccountIDs))
		}
		if len(req.Dates) != 2 {
			t.Errorf("expected 2 dates, got %d", len(req.Dates))
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(AggregatedValuesResponse{
			{
				Date: "2026-01-25",
				Value: Money{
					Value:            2963043.66,
					Unit:             "SEK",
					UnitType:         "MONETARY",
					DecimalPrecision: 2,
				},
			},
			{
				Date: "2026-01-28",
				Value: Money{
					Value:            2984827.19,
					Unit:             "SEK",
					UnitType:         "MONETARY",
					DecimalPrecision: 2,
				},
			},
		})
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	resp, err := svc.GetAggregatedValues(context.Background(), &AggregatedValuesRequest{
		EncryptedAccountIDs: []string{"abc123", "def456"},
		Dates:               []string{"2026-01-25", "2026-01-28"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp) != 2 {
		t.Fatalf("expected 2 values, got %d", len(resp))
	}
	if got, want := resp[0].Date, "2026-01-25"; got != want {
		t.Errorf("date = %q, want %q", got, want)
	}
	if got, want := resp[0].Value.Value, 296304.366; got != want {
		t.Errorf("value = %v, want %v", got, want)
	}
	if got, want := resp[1].Date, "2026-01-28"; got != want {
		t.Errorf("date = %q, want %q", got, want)
	}
}

func TestGetAggregatedValues_NilRequest(t *testing.T) {
	c := newTestClient("http://unused")
	svc := NewService(c)

	_, err := svc.GetAggregatedValues(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request, got nil")
	}
	if got, want := err.Error(), "request is required"; got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestGetAggregatedValues_EmptyAccountIDs(t *testing.T) {
	c := newTestClient("http://unused")
	svc := NewService(c)

	_, err := svc.GetAggregatedValues(context.Background(), &AggregatedValuesRequest{
		EncryptedAccountIDs: []string{},
		Dates:               []string{"2026-01-25"},
	})
	if err == nil {
		t.Fatal("expected error for empty account IDs, got nil")
	}
	if got, want := err.Error(), "at least one encrypted account ID is required"; got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestGetAggregatedValues_EmptyDates(t *testing.T) {
	c := newTestClient("http://unused")
	svc := NewService(c)

	_, err := svc.GetAggregatedValues(context.Background(), &AggregatedValuesRequest{
		EncryptedAccountIDs: []string{"abc123"},
		Dates:               []string{},
	})
	if err == nil {
		t.Fatal("expected error for empty dates, got nil")
	}
	if got, want := err.Error(), "at least one date is required"; got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestGetAggregatedValues_InvalidDateFormat(t *testing.T) {
	c := newTestClient("http://unused")
	svc := NewService(c)

	_, err := svc.GetAggregatedValues(context.Background(), &AggregatedValuesRequest{
		EncryptedAccountIDs: []string{"abc123"},
		Dates:               []string{"2026-01-25", "01-28-2026"},
	})
	if err == nil {
		t.Fatal("expected error for invalid date format, got nil")
	}
	if !strings.Contains(err.Error(), "must be in YYYY-MM-DD format") {
		t.Errorf("error = %q, want to contain 'must be in YYYY-MM-DD format'", err.Error())
	}
}

func TestGetAggregatedValues_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	_, err := svc.GetAggregatedValues(context.Background(), &AggregatedValuesRequest{
		EncryptedAccountIDs: []string{"abc123"},
		Dates:               []string{"2026-01-25"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var httpErr *client.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *client.HTTPError, got %T", err)
	}
	if httpErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("status code = %d, want %d", httpErr.StatusCode, http.StatusUnauthorized)
	}
}

func TestGetAggregatedValues_EmptyAccountID(t *testing.T) {
	c := newTestClient("http://unused")
	svc := NewService(c)

	_, err := svc.GetAggregatedValues(context.Background(), &AggregatedValuesRequest{
		EncryptedAccountIDs: []string{"abc123", ""},
		Dates:               []string{"2026-01-25"},
	})
	if err == nil {
		t.Fatal("expected error for empty account ID, got nil")
	}
	if got, want := err.Error(), "encrypted account ID cannot be empty"; got != want {
		t.Errorf("error = %q, want %q", got, want)
	}
}

func TestGetAggregatedValues_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	_, err := svc.GetAggregatedValues(context.Background(), &AggregatedValuesRequest{
		EncryptedAccountIDs: []string{"abc123"},
		Dates:               []string{"2026-01-25"},
	})
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode response") {
		t.Errorf("error = %q, want to contain 'failed to decode response'", err.Error())
	}
}

func TestGetAggregatedValues_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	svc := NewService(c)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.GetAggregatedValues(ctx, &AggregatedValuesRequest{
		EncryptedAccountIDs: []string{"abc123"},
		Dates:               []string{"2026-01-25"},
	})
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}
