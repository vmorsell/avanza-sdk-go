package market

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

func TestSearch_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/search/filtered-search" {
			t.Errorf("expected path /_api/search/filtered-search, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req searchAPIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Query != "investor" {
			t.Errorf("req.Query = %q, want %q", req.Query, "investor")
		}
		if req.ScreenSize != "DESKTOP" {
			t.Errorf("req.ScreenSize = %q, want %q", req.ScreenSize, "DESKTOP")
		}
		if req.SearchSessionID == "" {
			t.Error("expected non-empty searchSessionId")
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SearchResponse{
			TotalNumberOfHits: 2,
			Hits: []SearchHit{
				{
					Type:        "STOCK",
					Title:       "Investor B (INVE B)",
					OrderbookID: "5247",
					Tradable:   true,
					Buyable:     true,
					Sellable:    true,
					Price: SearchHitPrice{
						Last:                 "345,45",
						Currency:             "SEK",
						TodayChangeDirection: -1,
					},
					StockSectors: []StockSector{
						{ID: 21, Level: 1, Name: "Finans", EnglishName: "Financials"},
					},
					MarketPlaceName: "Stockholmsbörsen",
				},
				{
					Type:            "FUND",
					Title:           "Spiltan Aktiefond Investmentbolag",
					OrderbookID:     "325406",
					Tradable:       true,
					MarketPlaceName: "Fondmarknaden",
					FundTags: []FundTag{
						{Title: "Aktiefond", Category: "fund-type", TagCategory: "TYPE"},
					},
				},
			},
			SearchQuery: "investor",
			Facets: SearchFacets{
				Types: []TypeFacet{
					{Type: "STOCK", Count: 607},
					{Type: "FUND", Count: 151},
				},
			},
			Pagination: SearchPagination{Size: 30, From: 0},
		})
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	resp, err := svc.Search(context.Background(), &SearchRequest{Query: "investor"})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if resp.TotalNumberOfHits != 2 {
		t.Errorf("TotalNumberOfHits = %d, want 2", resp.TotalNumberOfHits)
	}
	if len(resp.Hits) != 2 {
		t.Fatalf("len(Hits) = %d, want 2", len(resp.Hits))
	}

	hit := resp.Hits[0]
	if hit.Type != "STOCK" {
		t.Errorf("Hits[0].Type = %q, want %q", hit.Type, "STOCK")
	}
	if hit.OrderbookID != "5247" {
		t.Errorf("Hits[0].OrderbookID = %q, want %q", hit.OrderbookID, "5247")
	}
	if hit.Price.Currency != "SEK" {
		t.Errorf("Hits[0].Price.Currency = %q, want %q", hit.Price.Currency, "SEK")
	}
	if len(hit.StockSectors) != 1 {
		t.Fatalf("len(Hits[0].StockSectors) = %d, want 1", len(hit.StockSectors))
	}
	if hit.StockSectors[0].EnglishName != "Financials" {
		t.Errorf("StockSectors[0].EnglishName = %q, want %q", hit.StockSectors[0].EnglishName, "Financials")
	}

	fund := resp.Hits[1]
	if len(fund.FundTags) != 1 {
		t.Fatalf("len(Hits[1].FundTags) = %d, want 1", len(fund.FundTags))
	}
	if fund.FundTags[0].TagCategory != "TYPE" {
		t.Errorf("FundTags[0].TagCategory = %q, want %q", fund.FundTags[0].TagCategory, "TYPE")
	}

	if len(resp.Facets.Types) != 2 {
		t.Errorf("len(Facets.Types) = %d, want 2", len(resp.Facets.Types))
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	svc := NewService(newTestClient("http://localhost"))
	_, err := svc.Search(context.Background(), &SearchRequest{Query: ""})
	if err == nil {
		t.Fatal("expected error for empty query, got nil")
	}
}

func TestSearch_WithTypeFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req searchAPIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if len(req.SearchFilter.Types) != 1 || req.SearchFilter.Types[0] != "STOCK" {
			t.Errorf("SearchFilter.Types = %v, want [STOCK]", req.SearchFilter.Types)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SearchResponse{
			TotalNumberOfHits: 0,
			Hits:              []SearchHit{},
			Pagination:        SearchPagination{Size: 30, From: 0},
		})
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	_, err := svc.Search(context.Background(), &SearchRequest{
		Query: "test",
		Types: []string{"STOCK"},
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
}

func TestSearch_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req searchAPIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Pagination.From != 30 {
			t.Errorf("Pagination.From = %d, want 30", req.Pagination.From)
		}
		if req.Pagination.Size != 10 {
			t.Errorf("Pagination.Size = %d, want 10", req.Pagination.Size)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SearchResponse{
			Pagination: SearchPagination{Size: 10, From: 30},
		})
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	resp, err := svc.Search(context.Background(), &SearchRequest{
		Query: "test",
		From:  30,
		Size:  10,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if resp.Pagination.From != 30 {
		t.Errorf("resp.Pagination.From = %d, want 30", resp.Pagination.From)
	}
}

func TestSearch_DefaultSize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req searchAPIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.Pagination.Size != 30 {
			t.Errorf("Pagination.Size = %d, want default 30", req.Pagination.Size)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SearchResponse{})
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	_, err := svc.Search(context.Background(), &SearchRequest{Query: "test"})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
}

func TestSearch_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	_, err := svc.Search(context.Background(), &SearchRequest{Query: "test"})
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

func TestSearch_NilTypesBecomesEmptyArray(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req searchAPIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if req.SearchFilter.Types == nil {
			t.Error("expected types to be empty array, got nil")
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SearchResponse{})
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	_, err := svc.Search(context.Background(), &SearchRequest{Query: "test"})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
}
