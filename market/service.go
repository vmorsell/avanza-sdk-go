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
func (s *Service) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
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
