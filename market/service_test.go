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

// --- GetStock tests ---

func TestGetStock_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/market-guide/stock/5247" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/_api/market-guide/stock/5247")
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want %q", r.Method, http.MethodGet)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"orderbookId": "5247",
			"name": "Investor B",
			"isin": "SE0015811963",
			"instrumentId": "5407",
			"sectors": [{"sectorId": "51", "sectorName": "Investmentbolag"}],
			"tradable": "BUYABLE_AND_SELLABLE",
			"listing": {
				"shortName": "INVE B",
				"tickerSymbol": "INVE B",
				"countryCode": "SE",
				"currency": "SEK",
				"marketPlaceCode": "XSTO",
				"marketPlaceName": "Stockholmsbörsen",
				"marketListName": "Large Cap Stockholm",
				"tickSizeListId": "2349",
				"marketTradesAvailable": true
			},
			"marketPlace": {
				"marketOpen": true,
				"timeLeftMs": 12475078,
				"openingTime": "09:00:00",
				"todayClosingTime": "17:30:00",
				"normalClosingTime": "17:30:00"
			},
			"historicalClosingPrices": {
				"oneDay": 345.45,
				"oneMonth": 377.75,
				"start": 2.83,
				"startDate": "1984-09-18"
			},
			"keyIndicators": {
				"numberOfOwners": 505118,
				"priceEarningsRatio": 6.75,
				"beta": 1.00,
				"marketCapital": {"value": 1062973451084.00, "currency": "SEK"},
				"earningsPerShare": {"value": 51.42, "currency": "SEK"},
				"dividend": {
					"exDate": "2026-05-08",
					"paymentDate": "2026-05-15",
					"amount": 4,
					"currencyCode": "SEK",
					"exDateStatus": "FUTURE"
				},
				"nextReport": {"date": "2026-04-21", "reportType": "INTERIM"}
			},
			"quote": {
				"buy": 346.85,
				"sell": 347.00,
				"last": 346.85,
				"change": 1.40,
				"changePercent": 0.41,
				"timeOfLast": 1774872123000,
				"updated": 1774872124091,
				"isRealTime": true
			},
			"type": "STOCK"
		}`))
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	resp, err := svc.GetStock(context.Background(), "5247")
	if err != nil {
		t.Fatalf("GetStock failed: %v", err)
	}

	if resp.OrderbookID != "5247" {
		t.Errorf("OrderbookID = %q, want %q", resp.OrderbookID, "5247")
	}
	if resp.Name != "Investor B" {
		t.Errorf("Name = %q, want %q", resp.Name, "Investor B")
	}
	if resp.Type != "STOCK" {
		t.Errorf("Type = %q, want %q", resp.Type, "STOCK")
	}
	if resp.Listing.TickerSymbol != "INVE B" {
		t.Errorf("Listing.TickerSymbol = %q, want %q", resp.Listing.TickerSymbol, "INVE B")
	}
	if resp.Listing.MarketListName != "Large Cap Stockholm" {
		t.Errorf("Listing.MarketListName = %q, want %q", resp.Listing.MarketListName, "Large Cap Stockholm")
	}
	if !resp.MarketPlace.MarketOpen {
		t.Error("MarketPlace.MarketOpen = false, want true")
	}
	if len(resp.Sectors) != 1 || resp.Sectors[0].SectorName != "Investmentbolag" {
		t.Errorf("Sectors = %v, want [{51, Investmentbolag}]", resp.Sectors)
	}
	if resp.HistoricalClosingPrices.OneDay == nil || *resp.HistoricalClosingPrices.OneDay != 345.45 {
		t.Errorf("HistoricalClosingPrices.OneDay = %v, want 345.45", resp.HistoricalClosingPrices.OneDay)
	}
	if resp.HistoricalClosingPrices.ThreeMonths != nil {
		t.Errorf("HistoricalClosingPrices.ThreeMonths = %v, want nil", resp.HistoricalClosingPrices.ThreeMonths)
	}
	if resp.KeyIndicators.NumberOfOwners != 505118 {
		t.Errorf("KeyIndicators.NumberOfOwners = %d, want 505118", resp.KeyIndicators.NumberOfOwners)
	}
	if resp.KeyIndicators.PriceEarningsRatio != 6.75 {
		t.Errorf("KeyIndicators.PriceEarningsRatio = %f, want 6.75", resp.KeyIndicators.PriceEarningsRatio)
	}
	if resp.KeyIndicators.MarketCapital.Currency != "SEK" {
		t.Errorf("KeyIndicators.MarketCapital.Currency = %q, want %q", resp.KeyIndicators.MarketCapital.Currency, "SEK")
	}
	if resp.KeyIndicators.Dividend == nil {
		t.Fatal("KeyIndicators.Dividend = nil, want non-nil")
	}
	if resp.KeyIndicators.Dividend.Amount != 4 {
		t.Errorf("KeyIndicators.Dividend.Amount = %f, want 4", resp.KeyIndicators.Dividend.Amount)
	}
	if resp.KeyIndicators.NextReport == nil || resp.KeyIndicators.NextReport.ReportType != "INTERIM" {
		t.Errorf("KeyIndicators.NextReport = %v, want INTERIM report", resp.KeyIndicators.NextReport)
	}
	if resp.Quote.Last != 346.85 {
		t.Errorf("Quote.Last = %f, want 346.85", resp.Quote.Last)
	}
}

func TestGetStock_EmptyOrderbookID(t *testing.T) {
	svc := NewService(newTestClient("http://localhost"))
	_, err := svc.GetStock(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty orderbookID, got nil")
	}
}

func TestGetStock_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	_, err := svc.GetStock(context.Background(), "999")
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

// --- GetCertificate tests ---

func TestGetCertificate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/market-guide/certificate/2321838" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/_api/market-guide/certificate/2321838")
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want %q", r.Method, http.MethodGet)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"orderbookId": "2321838",
			"name": "BEAR OMX X10 AVA 42",
			"isin": "GB00BVQ17X50",
			"tradable": "BUYABLE_AND_SELLABLE",
			"listing": {
				"shortName": "BEAR OMX X10 AVA 42",
				"tickerSymbol": "BEAR OMX X10 AVA 42",
				"countryCode": "SE",
				"currency": "SEK",
				"marketPlaceCode": "NMTF",
				"marketPlaceName": "Nordic MTF",
				"tickSizeListId": "2437",
				"marketTradesAvailable": true
			},
			"historicalClosingPrices": {
				"oneDay": 34.81,
				"oneWeek": 38.25,
				"start": 14.89,
				"startDate": "2026-02-26"
			},
			"keyIndicators": {
				"leverage": 1E+1,
				"isAza": true,
				"productLink": "https://etp.morganstanley.com/SE/SV/product-details/",
				"numberOfOwners": 2
			},
			"quote": {
				"buy": 29.62,
				"sell": 29.68,
				"last": 29.26,
				"change": -5.55,
				"changePercent": -15.94,
				"spread": 0.20,
				"timeOfLast": 1774872051000,
				"isRealTime": true
			},
			"type": "CERTIFICATE",
			"underlying": {
				"orderbookId": "19002",
				"name": "OMX Stockholm 30",
				"instrumentType": "INDEX",
				"instrumentSubType": "SECTOR",
				"quote": {
					"last": 2886.62,
					"change": 22.70,
					"changePercent": 0.79,
					"isRealTime": true
				},
				"listing": {
					"shortName": "OMXS30",
					"tickerSymbol": "OMXS30",
					"countryCode": "SE",
					"currency": "SEK"
				},
				"previousClosingPrice": 2863.92,
				"reference": false
			},
			"assetCategory": "Aktier",
			"category": "Aktieindex",
			"subCategory": "OMX Stockholm 30 Index"
		}`))
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	resp, err := svc.GetCertificate(context.Background(), "2321838")
	if err != nil {
		t.Fatalf("GetCertificate failed: %v", err)
	}

	if resp.OrderbookID != "2321838" {
		t.Errorf("OrderbookID = %q, want %q", resp.OrderbookID, "2321838")
	}
	if resp.Name != "BEAR OMX X10 AVA 42" {
		t.Errorf("Name = %q, want %q", resp.Name, "BEAR OMX X10 AVA 42")
	}
	if resp.Type != "CERTIFICATE" {
		t.Errorf("Type = %q, want %q", resp.Type, "CERTIFICATE")
	}
	if resp.KeyIndicators.Leverage != 10 {
		t.Errorf("KeyIndicators.Leverage = %f, want 10", resp.KeyIndicators.Leverage)
	}
	if !resp.KeyIndicators.IsAza {
		t.Error("KeyIndicators.IsAza = false, want true")
	}
	if resp.Underlying.OrderbookID != "19002" {
		t.Errorf("Underlying.OrderbookID = %q, want %q", resp.Underlying.OrderbookID, "19002")
	}
	if resp.Underlying.InstrumentType != "INDEX" {
		t.Errorf("Underlying.InstrumentType = %q, want %q", resp.Underlying.InstrumentType, "INDEX")
	}
	if resp.Underlying.PreviousClosingPrice != 2863.92 {
		t.Errorf("Underlying.PreviousClosingPrice = %f, want 2863.92", resp.Underlying.PreviousClosingPrice)
	}
	if resp.AssetCategory != "Aktier" {
		t.Errorf("AssetCategory = %q, want %q", resp.AssetCategory, "Aktier")
	}
	if resp.Quote.ChangePercent != -15.94 {
		t.Errorf("Quote.ChangePercent = %f, want -15.94", resp.Quote.ChangePercent)
	}
}

func TestGetCertificate_EmptyOrderbookID(t *testing.T) {
	svc := NewService(newTestClient("http://localhost"))
	_, err := svc.GetCertificate(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty orderbookID, got nil")
	}
}

func TestGetCertificate_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	_, err := svc.GetCertificate(context.Background(), "999")
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

// --- GetWarrant tests ---

func TestGetWarrant_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_api/market-guide/warrant/564075" {
			t.Errorf("path = %q, want %q", r.URL.Path, "/_api/market-guide/warrant/564075")
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want %q", r.Method, http.MethodGet)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"orderbookId": "564075",
			"name": "MINI L OMX AVA 9",
			"isin": "GB00BVZVJC97",
			"tradable": "BUYABLE_AND_SELLABLE",
			"listing": {
				"shortName": "MINI L OMX AVA 9",
				"tickerSymbol": "MINI L OMX AVA 9",
				"countryCode": "SE",
				"currency": "SEK",
				"marketPlaceCode": "NMTF",
				"marketPlaceName": "Nordic MTF",
				"tickSizeListId": "2437",
				"marketTradesAvailable": true
			},
			"historicalClosingPrices": {
				"oneDay": 1817.29,
				"threeYears": 1128.16,
				"fiveYears": 867.51,
				"start": 604.00,
				"startDate": "2015-05-21"
			},
			"keyIndicators": {
				"parity": 1,
				"barrierLevel": 1037.83,
				"financingLevel": 1017.38,
				"direction": "Lång",
				"leverage": 1.54,
				"isAza": true,
				"numberOfOwners": 16,
				"subType": "MINI_FUTURE"
			},
			"quote": {
				"buy": 1867.19,
				"sell": 1867.44,
				"last": 1859.14,
				"change": 41.85,
				"changePercent": 2.30,
				"spread": 0.01,
				"timeOfLast": 1774872161000,
				"isRealTime": true
			},
			"type": "WARRANT",
			"underlying": {
				"orderbookId": "19002",
				"name": "OMX Stockholm 30",
				"instrumentType": "INDEX",
				"instrumentSubType": "SECTOR",
				"quote": {
					"last": 2885.37,
					"isRealTime": true
				},
				"listing": {
					"shortName": "OMXS30",
					"tickerSymbol": "OMXS30",
					"countryCode": "SE",
					"currency": "SEK"
				},
				"previousClosingPrice": 2863.92,
				"reference": false
			},
			"assetCategory": "Aktier",
			"category": "Aktieindex",
			"subCategory": "OMX Stockholm 30 Index"
		}`))
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	resp, err := svc.GetWarrant(context.Background(), "564075")
	if err != nil {
		t.Fatalf("GetWarrant failed: %v", err)
	}

	if resp.OrderbookID != "564075" {
		t.Errorf("OrderbookID = %q, want %q", resp.OrderbookID, "564075")
	}
	if resp.Name != "MINI L OMX AVA 9" {
		t.Errorf("Name = %q, want %q", resp.Name, "MINI L OMX AVA 9")
	}
	if resp.Type != "WARRANT" {
		t.Errorf("Type = %q, want %q", resp.Type, "WARRANT")
	}
	if resp.KeyIndicators.Parity != 1 {
		t.Errorf("KeyIndicators.Parity = %f, want 1", resp.KeyIndicators.Parity)
	}
	if resp.KeyIndicators.BarrierLevel != 1037.83 {
		t.Errorf("KeyIndicators.BarrierLevel = %f, want 1037.83", resp.KeyIndicators.BarrierLevel)
	}
	if resp.KeyIndicators.Direction != "Lång" {
		t.Errorf("KeyIndicators.Direction = %q, want %q", resp.KeyIndicators.Direction, "Lång")
	}
	if resp.KeyIndicators.SubType != "MINI_FUTURE" {
		t.Errorf("KeyIndicators.SubType = %q, want %q", resp.KeyIndicators.SubType, "MINI_FUTURE")
	}
	if !resp.KeyIndicators.IsAza {
		t.Error("KeyIndicators.IsAza = false, want true")
	}
	if resp.Underlying.Name != "OMX Stockholm 30" {
		t.Errorf("Underlying.Name = %q, want %q", resp.Underlying.Name, "OMX Stockholm 30")
	}
	if resp.SubCategory != "OMX Stockholm 30 Index" {
		t.Errorf("SubCategory = %q, want %q", resp.SubCategory, "OMX Stockholm 30 Index")
	}
	if resp.HistoricalClosingPrices.ThreeYears == nil || *resp.HistoricalClosingPrices.ThreeYears != 1128.16 {
		t.Errorf("HistoricalClosingPrices.ThreeYears = %v, want 1128.16", resp.HistoricalClosingPrices.ThreeYears)
	}
}

func TestGetWarrant_EmptyOrderbookID(t *testing.T) {
	svc := NewService(newTestClient("http://localhost"))
	_, err := svc.GetWarrant(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty orderbookID, got nil")
	}
}

func TestGetWarrant_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	svc := NewService(newTestClient(server.URL))
	_, err := svc.GetWarrant(context.Background(), "999")
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
