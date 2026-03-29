package avanza

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vmorsell/avanza-sdk-go/market"
)

func TestSearch_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/search/filtered-search" {
			t.Errorf("expected path /_api/search/filtered-search, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(market.SearchResponse{
			TotalNumberOfHits: 87,
			Hits: []market.SearchHit{
				{
					Type:            "CERTIFICATE",
					Title:           "BULL OMX X18 AVA 39",
					OrderbookID:     "2017182",
					Tradable:       true,
					Buyable:         true,
					Sellable:        true,
					MarketPlaceName: "Nordic MTF",
					Price: market.SearchHitPrice{
						Last:                 "0,329",
						Currency:             "SEK",
						TodayChangeDirection: -1,
					},
				},
			},
			SearchQuery: "bull omx ava",
			Facets: market.SearchFacets{
				Types: []market.TypeFacet{
					{Type: "CERTIFICATE", Count: 87},
				},
			},
			Pagination: market.SearchPagination{Size: 30, From: 0},
		})
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))
	resp, err := avanza.Market.Search(context.Background(), &market.SearchRequest{
		Query: "bull omx ava",
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if resp.TotalNumberOfHits != 87 {
		t.Errorf("TotalNumberOfHits = %d, want 87", resp.TotalNumberOfHits)
	}
	if len(resp.Hits) != 1 {
		t.Fatalf("len(Hits) = %d, want 1", len(resp.Hits))
	}
	if resp.Hits[0].Type != "CERTIFICATE" {
		t.Errorf("Hits[0].Type = %q, want %q", resp.Hits[0].Type, "CERTIFICATE")
	}
	if resp.Hits[0].OrderbookID != "2017182" {
		t.Errorf("Hits[0].OrderbookID = %q, want %q", resp.Hits[0].OrderbookID, "2017182")
	}
}

func TestSearch_Pagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Pagination struct {
				From int `json:"from"`
				Size int `json:"size"`
			} `json:"pagination"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if body.Pagination.From != 30 {
			t.Errorf("pagination.from = %d, want 30", body.Pagination.From)
		}
		if body.Pagination.Size != 30 {
			t.Errorf("pagination.size = %d, want 30", body.Pagination.Size)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(market.SearchResponse{
			TotalNumberOfHits: 1778,
			Hits:              []market.SearchHit{},
			Pagination:        market.SearchPagination{Size: 30, From: 30},
		})
	}))
	defer server.Close()

	avanza := New(WithBaseURL(server.URL))
	resp, err := avanza.Market.Search(context.Background(), &market.SearchRequest{
		Query: "inv",
		From:  30,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if resp.Pagination.From != 30 {
		t.Errorf("Pagination.From = %d, want 30", resp.Pagination.From)
	}
}
