// Package market provides market data functionality for the Avanza API.
package market

// OrderDepthLevel contains bid/ask prices and volumes at a single price level.
type OrderDepthLevel struct {
	BuyPrice   float64 `json:"buyPrice"`
	BuyVolume  float64 `json:"buyVolume"`
	SellPrice  float64 `json:"sellPrice"`
	SellVolume float64 `json:"sellVolume"`
}

// OrderDepthData contains the complete order book snapshot.
type OrderDepthData struct {
	OrderbookID           string            `json:"orderbookId"`
	Levels                []OrderDepthLevel `json:"levels"`
	MarketMakerLevelInAsk int               `json:"marketMakerLevelInAsk"`
	MarketMakerLevelInBid int               `json:"marketMakerLevelInBid"`
}

// OrderDepthEvent is a single event from the order depth subscription stream.
type OrderDepthEvent struct {
	Event string         `json:"event"`
	Data  OrderDepthData `json:"data"`
	ID    string         `json:"id"`
	Retry int            `json:"retry"`
}

// SearchRequest configures an instrument search.
type SearchRequest struct {
	// Query is the search string (required).
	Query string

	// Types filters results by instrument type (e.g. "STOCK", "FUND", "CERTIFICATE",
	// "EXCHANGE_TRADED_FUND", "WARRANT", "OPTION", "INDEX"). Empty means all types.
	Types []string

	// From is the pagination offset. Default 0.
	From int

	// Size is the number of results per page. Default 30.
	Size int
}

// searchAPIRequest is the internal request body sent to the search endpoint.
type searchAPIRequest struct {
	Query           string              `json:"query"`
	SearchFilter    searchFilter        `json:"searchFilter"`
	ScreenSize      string              `json:"screenSize"`
	OriginPath      string              `json:"originPath"`
	OriginPlatform  string              `json:"originPlatform"`
	SearchSessionID string              `json:"searchSessionId"`
	Pagination      searchAPIPagination `json:"pagination"`
}

type searchFilter struct {
	Types []string `json:"types"`
}

type searchAPIPagination struct {
	From int `json:"from"`
	Size int `json:"size"`
}

// SearchResponse contains the search results.
type SearchResponse struct {
	TotalNumberOfHits int              `json:"totalNumberOfHits"`
	Hits              []SearchHit      `json:"hits"`
	SearchQuery       string           `json:"searchQuery"`
	Facets            SearchFacets     `json:"facets"`
	Pagination        SearchPagination `json:"pagination"`
}

// SearchHit is a single search result.
type SearchHit struct {
	Type             string         `json:"type"`
	Title            string         `json:"title"`
	Description      string         `json:"description"`
	FlagCode         string         `json:"flagCode"`
	OrderbookID      string         `json:"orderBookId"`
	URLSlugName      string         `json:"urlSlugName"`
	Tradable         bool           `json:"tradeable"`
	Sellable         bool           `json:"sellable"`
	Buyable          bool           `json:"buyable"`
	Price            SearchHitPrice `json:"price"`
	StockSectors     []StockSector  `json:"stockSectors"`
	FundTags         []FundTag      `json:"fundTags"`
	MarketPlaceName  string         `json:"marketPlaceName"`
	SubType          *string        `json:"subType"`
}

// SearchHitPrice contains price and change data for a search hit.
type SearchHitPrice struct {
	Last                           string  `json:"last"`
	Currency                       string  `json:"currency"`
	TodayChangePercent             string  `json:"todayChangePercent"`
	TodayChangeValue               string  `json:"todayChangeValue"`
	TodayChangeDirection           int     `json:"todayChangeDirection"`
	ThreeMonthsAgoChangePercent    *string `json:"threeMonthsAgoChangePercent"`
	ThreeMonthsAgoChangeDirection  int     `json:"threeMonthsAgoChangeDirection"`
	Spread                         *string `json:"spread"`
}

// StockSector categorizes a stock by industry sector.
type StockSector struct {
	ID          int    `json:"id"`
	Level       int    `json:"level"`
	Name        string `json:"name"`
	EnglishName string `json:"englishName"`
}

// FundTag describes a fund classification.
type FundTag struct {
	Title       string `json:"title"`
	Category    string `json:"category"`
	TagCategory string `json:"tagCategory"`
}

// SearchFacets contains result counts grouped by instrument type.
type SearchFacets struct {
	Types []TypeFacet `json:"types"`
}

// TypeFacet is the hit count for a single instrument type.
type TypeFacet struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// SearchPagination describes the pagination state of search results.
type SearchPagination struct {
	Size int `json:"size"`
	From int `json:"from"`
}
