// Package market provides market data functionality for the Avanza API.
package market

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/vmorsell/avanza-sdk-go/client"
	"github.com/vmorsell/avanza-sdk-go/internal/sse"
)

// Service handles market data and real-time subscriptions.
type Service struct {
	client *client.Client
}

// NewService creates a new market service.
func NewService(client *client.Client) *Service {
	return &Service{
		client: client,
	}
}

// Search searches for instruments by name, ticker, or other text.
//
// This returns public market data and does not require an authenticated session.
func (s *Service) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	if req.Query == "" {
		return nil, fmt.Errorf("query is required")
	}

	size := req.Size
	if size == 0 {
		size = 30
	}

	types := req.Types
	if types == nil {
		types = []string{}
	}

	apiReq := searchAPIRequest{
		Query:           req.Query,
		SearchFilter:    searchFilter{Types: types},
		ScreenSize:      "DESKTOP",
		OriginPath:      "/hem/hem.html",
		OriginPlatform:  "PWA",
		SearchSessionID: uuid.New().String(),
		Pagination:      searchAPIPagination{From: req.From, Size: size},
	}

	httpResp, err := s.client.Post(ctx, "/_api/search/filtered-search", apiReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp SearchResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// GetStock returns detailed market data for a stock.
//
// This returns public market data and does not require an authenticated session.
func (s *Service) GetStock(ctx context.Context, orderbookID string) (*Stock, error) {
	if orderbookID == "" {
		return nil, fmt.Errorf("orderbookID is required")
	}

	endpoint := fmt.Sprintf("/_api/market-guide/stock/%s", url.PathEscape(orderbookID))

	httpResp, err := s.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp Stock
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// GetCertificate returns detailed market data for a certificate.
func (s *Service) GetCertificate(ctx context.Context, orderbookID string) (*Certificate, error) {
	if orderbookID == "" {
		return nil, fmt.Errorf("orderbookID is required")
	}

	endpoint := fmt.Sprintf("/_api/market-guide/certificate/%s", url.PathEscape(orderbookID))

	httpResp, err := s.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp Certificate
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// GetWarrant returns detailed market data for a warrant.
func (s *Service) GetWarrant(ctx context.Context, orderbookID string) (*Warrant, error) {
	if orderbookID == "" {
		return nil, fmt.Errorf("orderbookID is required")
	}

	endpoint := fmt.Sprintf("/_api/market-guide/warrant/%s", url.PathEscape(orderbookID))

	httpResp, err := s.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp Warrant
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// GetOrderbook returns trading parameters for an instrument (tick sizes, feature support, validity dates).
func (s *Service) GetOrderbook(ctx context.Context, orderbookID string) (*Orderbook, error) {
	if orderbookID == "" {
		return nil, fmt.Errorf("orderbookID is required")
	}

	endpoint := fmt.Sprintf("/_api/trading-critical/rest/orderbook/%s", url.PathEscape(orderbookID))

	httpResp, err := s.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp Orderbook
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// GetMarketData returns real-time quote, order depth, and trades for an instrument.
func (s *Service) GetMarketData(ctx context.Context, orderbookID string) (*MarketData, error) {
	if orderbookID == "" {
		return nil, fmt.Errorf("orderbookID is required")
	}

	endpoint := fmt.Sprintf("/_api/trading-critical/rest/marketdata/%s", url.PathEscape(orderbookID))

	httpResp, err := s.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp MarketData
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// GetMarketMakerPriceChart returns OHLC bars and market-maker quotes for an instrument
// over the given time period. The server selects an appropriate bar resolution; available
// alternatives are reported in the response metadata.
func (s *Service) GetMarketMakerPriceChart(ctx context.Context, orderbookID string, timePeriod TimePeriod) (*MarketMakerPriceChart, error) {
	if orderbookID == "" {
		return nil, fmt.Errorf("orderbookID is required")
	}
	if timePeriod == "" {
		return nil, fmt.Errorf("timePeriod is required")
	}

	endpoint := fmt.Sprintf("/_api/price-chart/marketmaker/%s?timePeriod=%s",
		url.PathEscape(orderbookID), url.QueryEscape(string(timePeriod)))

	httpResp, err := s.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp MarketMakerPriceChart
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// GetStockDetails returns extended company and instrument data for a stock:
// ownership, corporate events, dividends, holdings, fund/ETF exposure, and more.
//
// This returns public market data and does not require an authenticated session.
func (s *Service) GetStockDetails(ctx context.Context, orderbookID string) (*StockDetails, error) {
	if orderbookID == "" {
		return nil, fmt.Errorf("orderbookID is required")
	}

	endpoint := fmt.Sprintf("/_api/market-guide/stock/%s/details", url.PathEscape(orderbookID))

	httpResp, err := s.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp StockDetails
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// GetStockQuote returns the latest quote for a stock. This is the same quote
// carried on the full Stock response, fetched on its own for cheap polling.
//
// This endpoint exists only for stocks. This returns public market data and
// does not require an authenticated session.
func (s *Service) GetStockQuote(ctx context.Context, orderbookID string) (*Quote, error) {
	if orderbookID == "" {
		return nil, fmt.Errorf("orderbookID is required")
	}

	endpoint := fmt.Sprintf("/_api/market-guide/stock/%s/quote", url.PathEscape(orderbookID))

	httpResp, err := s.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp Quote
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// GetStockOrderDepth returns the current order book snapshot for a stock.
//
// This endpoint exists only for stocks. This returns public market data and
// does not require an authenticated session.
func (s *Service) GetStockOrderDepth(ctx context.Context, orderbookID string) (*MarketDataOrderDepth, error) {
	if orderbookID == "" {
		return nil, fmt.Errorf("orderbookID is required")
	}

	endpoint := fmt.Sprintf("/_api/market-guide/stock/%s/orderdepth", url.PathEscape(orderbookID))

	httpResp, err := s.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp MarketDataOrderDepth
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// GetStockMarketPlace returns the trading venue status and schedule for a stock.
//
// This endpoint exists only for stocks. This returns public market data and
// does not require an authenticated session.
func (s *Service) GetStockMarketPlace(ctx context.Context, orderbookID string) (*MarketPlace, error) {
	if orderbookID == "" {
		return nil, fmt.Errorf("orderbookID is required")
	}

	endpoint := fmt.Sprintf("/_api/market-guide/stock/%s/marketplace", url.PathEscape(orderbookID))

	httpResp, err := s.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp MarketPlace
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// GetOffHoursPrice returns the latest pre- or post-market price for an instrument.
// The returned OffHoursPrice has a nil Quote when there is no off-hours session
// for the instrument (typical for derivatives, or outside pre/post-market hours).
//
// This returns public market data and does not require an authenticated session.
func (s *Service) GetOffHoursPrice(ctx context.Context, orderbookID string) (*OffHoursPrice, error) {
	if orderbookID == "" {
		return nil, fmt.Errorf("orderbookID is required")
	}

	endpoint := fmt.Sprintf("/_push/market-offhours-price/latest/%s", url.PathEscape(orderbookID))

	httpResp, err := s.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp OffHoursPrice
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// GetCertificateDetails returns extended data for a certificate: issuer, leverage,
// regulatory documents, order book, and collateral terms.
//
// This returns public market data and does not require an authenticated session.
func (s *Service) GetCertificateDetails(ctx context.Context, orderbookID string) (*CertificateDetails, error) {
	if orderbookID == "" {
		return nil, fmt.Errorf("orderbookID is required")
	}

	endpoint := fmt.Sprintf("/_api/market-guide/certificate/%s/details", url.PathEscape(orderbookID))

	httpResp, err := s.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp CertificateDetails
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// GetWarrantDetails returns extended data for a warrant: issuer, regulatory
// documents, order book, and trading terms.
//
// This returns public market data and does not require an authenticated session.
func (s *Service) GetWarrantDetails(ctx context.Context, orderbookID string) (*WarrantDetails, error) {
	if orderbookID == "" {
		return nil, fmt.Errorf("orderbookID is required")
	}

	endpoint := fmt.Sprintf("/_api/market-guide/warrant/%s/details", url.PathEscape(orderbookID))

	httpResp, err := s.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp WarrantDetails
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// GetStockPriceChart returns OHLC bars for a stock over the given time period,
// along with the previous closing price. The server selects an appropriate bar
// resolution; available alternatives are reported in the response metadata.
//
// This returns public market data and does not require an authenticated session.
func (s *Service) GetStockPriceChart(ctx context.Context, orderbookID string, timePeriod TimePeriod) (*StockPriceChart, error) {
	if orderbookID == "" {
		return nil, fmt.Errorf("orderbookID is required")
	}
	if timePeriod == "" {
		return nil, fmt.Errorf("timePeriod is required")
	}

	endpoint := fmt.Sprintf("/_api/price-chart/stock/%s?timePeriod=%s",
		url.PathEscape(orderbookID), url.QueryEscape(string(timePeriod)))

	httpResp, err := s.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp StockPriceChart
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// GetStockPriceChartComparison returns the OHLC series for orderbookID rebased for
// comparison against compareOrderbookID over the given time period. The returned
// StockPriceChart carries no PreviousClosingPrice for comparison requests.
//
// This returns public market data and does not require an authenticated session.
func (s *Service) GetStockPriceChartComparison(ctx context.Context, orderbookID, compareOrderbookID string, timePeriod TimePeriod) (*StockPriceChart, error) {
	if orderbookID == "" {
		return nil, fmt.Errorf("orderbookID is required")
	}
	if compareOrderbookID == "" {
		return nil, fmt.Errorf("compareOrderbookID is required")
	}
	if timePeriod == "" {
		return nil, fmt.Errorf("timePeriod is required")
	}

	endpoint := fmt.Sprintf("/_api/price-chart/stock/%s/compare/%s?timePeriod=%s",
		url.PathEscape(orderbookID), url.PathEscape(compareOrderbookID), url.QueryEscape(string(timePeriod)))

	httpResp, err := s.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp StockPriceChart
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// GetNews returns the news feed for an instrument: press releases, analyst
// notes, and articles from external Swedish financial media.
//
// This returns public market data and does not require an authenticated session.
func (s *Service) GetNews(ctx context.Context, orderbookID string) (*News, error) {
	if orderbookID == "" {
		return nil, fmt.Errorf("orderbookID is required")
	}

	endpoint := fmt.Sprintf("/_api/market-guide/news/%s", url.PathEscape(orderbookID))

	httpResp, err := s.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp News
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// GetForum returns recent community forum posts for an instrument.
//
// This returns public market data and does not require an authenticated session.
func (s *Service) GetForum(ctx context.Context, orderbookID string) (*Forum, error) {
	if orderbookID == "" {
		return nil, fmt.Errorf("orderbookID is required")
	}

	endpoint := fmt.Sprintf("/_api/market-guide/forum/%s", url.PathEscape(orderbookID))

	httpResp, err := s.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, client.NewHTTPError(httpResp)
	}

	var resp Forum
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &resp, nil
}

// SubscribeToOrderDepth subscribes to order depth updates. Call Close() when done.
func (s *Service) SubscribeToOrderDepth(ctx context.Context, orderbookID string) (*OrderDepthSubscription, error) {
	if orderbookID == "" {
		return nil, fmt.Errorf("orderbookID is required")
	}

	cookies := s.client.Cookies()
	if len(cookies) == 0 {
		return nil, fmt.Errorf("subscribe to order depth: no authentication cookies found - please authenticate first")
	}

	essentialCookies := []string{"csid", "cstoken", "AZACSRF"}
	for _, cookie := range essentialCookies {
		if _, exists := cookies[cookie]; !exists {
			return nil, fmt.Errorf("subscribe to order depth: missing essential cookie: %s - please authenticate first", cookie)
		}
	}

	escapedID := url.PathEscape(orderbookID)
	sub := sse.New(ctx, sse.Config{
		Client:   s.client,
		Endpoint: fmt.Sprintf("/_push/order-depth-web-push/%s", escapedID),
		Referer:  fmt.Sprintf("https://www.avanza.se/handla/order.html/kop/%s", escapedID),
	})

	return newOrderDepthSubscription(sub), nil
}
