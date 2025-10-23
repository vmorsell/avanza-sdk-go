package accounts

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vmorsell/avanza-sdk-go/internal/client"
)

func TestNewAccountsService(t *testing.T) {
	c := client.NewClient()
	service := NewAccountsService(c)

	if service == nil {
		t.Fatal("expected service to be non-nil")
	}

	if service.client == nil {
		t.Error("expected client to be non-nil")
	}
}

func TestGetAccountOverview_Success(t *testing.T) {
	const (
		expectedCategoryID    = "SPARANDE"
		expectedCategoryName  = "Sparande"
		expectedAccountID     = "FOO"
		expectedAccountName   = "Test Account"
		expectedCategoryValue = 100000.00
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}

		if r.URL.Path != "/_api/account-overview/overview/categorizedAccounts" {
			t.Errorf("expected path /_api/account-overview/overview/categorizedAccounts, got %s", r.URL.Path)
		}

		response := AccountOverview{
			Categories: []Category{
				{
					ID:   expectedCategoryID,
					Name: expectedCategoryName,
					TotalValue: Money{
						Value:            expectedCategoryValue,
						Unit:             "SEK",
						UnitType:         "MONETARY",
						DecimalPrecision: 2,
					},
				},
			},
			Accounts: []Account{
				{
					ID:         expectedAccountID,
					CategoryID: expectedCategoryID,
					Type:       "INVESTERINGSSPARKONTO",
					Name: AccountName{
						DefaultName:     expectedAccountID,
						UserDefinedName: expectedAccountName,
					},
					Status: "ACTIVE",
				},
			},
			Loans: []Loan{},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := client.NewClient(client.WithBaseURL(server.URL))
	service := NewAccountsService(c)

	ctx := context.Background()
	overview, err := service.GetAccountOverview(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(overview.Categories) != 1 {
		t.Errorf("expected 1 category, got %d", len(overview.Categories))
	}

	if overview.Categories[0].ID != expectedCategoryID {
		t.Errorf("expected category ID %q, got %q", expectedCategoryID, overview.Categories[0].ID)
	}

	if len(overview.Accounts) != 1 {
		t.Errorf("expected 1 account, got %d", len(overview.Accounts))
	}

	if overview.Accounts[0].ID != expectedAccountID {
		t.Errorf("expected account ID %q, got %q", expectedAccountID, overview.Accounts[0].ID)
	}
}

func TestGetAccountOverview_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer server.Close()

	c := client.NewClient(client.WithBaseURL(server.URL))
	service := NewAccountsService(c)

	ctx := context.Background()
	_, err := service.GetAccountOverview(ctx)
	if err == nil {
		t.Error("expected error for HTTP error status, got nil")
	}
}

func TestGetAccountOverview_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	c := client.NewClient(client.WithBaseURL(server.URL))
	service := NewAccountsService(c)

	ctx := context.Background()
	_, err := service.GetAccountOverview(ctx)
	if err == nil {
		t.Error("expected error for malformed JSON, got nil")
	}
}

func TestAccountOverview_JSONMarshaling(t *testing.T) {
	overview := AccountOverview{
		Categories: []Category{
			{
				ID:   "SPARANDE",
				Name: "Sparande",
				TotalValue: Money{
					Value:            100000.00,
					Unit:             "SEK",
					UnitType:         "MONETARY",
					DecimalPrecision: 2,
				},
			},
		},
		Accounts: []Account{
			{
				ID:         "FOO",
				CategoryID: "SPARANDE",
				Type:       "INVESTERINGSSPARKONTO",
				Name: AccountName{
					DefaultName:     "FOO",
					UserDefinedName: "Test Account",
				},
				Status: "ACTIVE",
			},
		},
		Loans: []Loan{},
	}

	data, err := json.Marshal(overview)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded AccountOverview
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Categories[0].ID != overview.Categories[0].ID {
		t.Errorf("expected category ID %s, got %s", overview.Categories[0].ID, decoded.Categories[0].ID)
	}

	if decoded.Accounts[0].ID != overview.Accounts[0].ID {
		t.Errorf("expected account ID %s, got %s", overview.Accounts[0].ID, decoded.Accounts[0].ID)
	}
}

func TestGetTradingAccounts_Success(t *testing.T) {
	const (
		expectedAccountName = "Kontonamn"
		expectedAccountID   = "12345"
		expectedURLParamID  = "id"
		expectedBalance     = 10000.00
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}

		if r.URL.Path != "/_api/trading-critical/rest/accounts" {
			t.Errorf("expected path /_api/trading-critical/rest/accounts, got %s", r.URL.Path)
		}

		response := []TradingAccount{
			{
				Name:                 expectedAccountName,
				AccountID:            expectedAccountID,
				AccountTypeName:      "Investeringssparkonto",
				AccountType:          "INVESTERINGSSPARKONTO",
				AvailableForPurchase: 20000.00,
				HasCredit:            true,
				IsTradable:           true,
				URLParameterID:       expectedURLParamID,
				CurrencyBalances: []CurrencyBalance{
					{
						Currency:    "SEK",
						CountryCode: "SE",
						Balance:     expectedBalance,
					},
				},
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := client.NewClient(client.WithBaseURL(server.URL))
	service := NewAccountsService(c)

	ctx := context.Background()
	accounts, err := service.GetTradingAccounts(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(accounts) != 1 {
		t.Errorf("expected 1 account, got %d", len(accounts))
	}

	if accounts[0].Name != expectedAccountName {
		t.Errorf("expected account name %q, got %q", expectedAccountName, accounts[0].Name)
	}

	if accounts[0].URLParameterID != expectedURLParamID {
		t.Errorf("expected URL parameter ID %q, got %q", expectedURLParamID, accounts[0].URLParameterID)
	}
}

func TestGetAccountPositions_Success(t *testing.T) {
	const (
		expectedURLParamID     = "id"
		expectedAccountID      = "12345"
		expectedAccountName    = "Kontonamn"
		expectedInstrumentID   = "1234"
		expectedInstrumentName = "Stock A"
		expectedPositionID     = "12345-1234"
		expectedCashID         = "12345-SEK"
		expectedVolume         = 46.0
		expectedValue          = 10000.0
		expectedCashBalance    = 10000.0
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET method, got %s", r.Method)
		}

		expectedPath := "/_api/position-data/positions/" + expectedURLParamID
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		response := AccountPositions{
			WithOrderbook: []AccountPosition{
				{
					Account: AccountInfo{
						ID:             expectedAccountID,
						Type:           "INVESTERINGSSPARKONTO",
						Name:           expectedAccountName,
						URLParameterID: expectedURLParamID,
						HasCredit:      true,
					},
					Instrument: Instrument{
						ID:       expectedInstrumentID,
						Type:     "STOCK",
						Name:     expectedInstrumentName,
						Currency: "SEK",
						ISIN:     "SE0343203421",
					},
					ID: expectedPositionID,
					Volume: Money{
						Value: expectedVolume,
						Unit:  "",
					},
					Value: Money{
						Value: expectedValue,
						Unit:  "SEK",
					},
				},
			},
			CashPositions: []CashPosition{
				{
					Account: AccountInfo{
						ID: expectedAccountID,
					},
					TotalBalance: Money{
						Value: expectedCashBalance,
						Unit:  "SEK",
					},
					ID: expectedCashID,
				},
			},
			WithCreditAccount: true,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	c := client.NewClient(client.WithBaseURL(server.URL))
	service := NewAccountsService(c)

	ctx := context.Background()
	positions, err := service.GetAccountPositions(ctx, expectedURLParamID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(positions.WithOrderbook) != 1 {
		t.Errorf("expected 1 position with orderbook, got %d", len(positions.WithOrderbook))
	}

	if positions.WithOrderbook[0].Instrument.Name != expectedInstrumentName {
		t.Errorf("expected instrument name %q, got %q", expectedInstrumentName, positions.WithOrderbook[0].Instrument.Name)
	}

	if len(positions.CashPositions) != 1 {
		t.Errorf("expected 1 cash position, got %d", len(positions.CashPositions))
	}

	if !positions.WithCreditAccount {
		t.Error("expected WithCreditAccount to be true")
	}
}
