package accounts

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vmorsell/avanza/internal/client"
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

	if overview.Categories[0].ID != "SPARANDE" {
		t.Errorf("expected category ID SPARANDE, got %s", overview.Categories[0].ID)
	}

	if len(overview.Accounts) != 1 {
		t.Errorf("expected 1 account, got %d", len(overview.Accounts))
	}

	if overview.Accounts[0].ID != "FOO" {
		t.Errorf("expected account ID FOO, got %s", overview.Accounts[0].ID)
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
