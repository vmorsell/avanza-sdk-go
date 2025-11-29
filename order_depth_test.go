package avanza

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vmorsell/avanza-sdk-go/client"
	"github.com/vmorsell/avanza-sdk-go/market"
)

func TestOrderDepthSubscription(t *testing.T) {
	// Create a mock SSE server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request is for order depth subscription
		expectedPath := "/_push/order-depth-web-push/2185403"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Verify SSE headers
		if r.Header.Get("Accept") != "text/event-stream" {
			t.Errorf("Expected Accept: text/event-stream, got %s", r.Header.Get("Accept"))
		}

		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Send SSE events
		events := []string{
			"event: info\ndata: connected\nid: test-1\nretry: 1000\n\n",
			"event: ORDER_DEPTH\ndata: {\"orderbookId\":\"2185403\",\"levels\":[{\"buyPrice\":10.00,\"buyVolume\":500,\"sellPrice\":10.15,\"sellVolume\":25000}],\"marketMakerLevelInAsk\":0,\"marketMakerLevelInBid\":1}\nid: test-2\nretry: 1000\n\n",
			"event: info\ndata: heartbeat\nid: test-3\nretry: 1000\n\n",
		}

		for _, event := range events {
			w.Write([]byte(event))
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer server.Close()

	// Create Avanza client with mock server and mock cookies
	avanzaClient := New(WithBaseURL(server.URL))

	// Mock authentication cookies
	avanzaClient.client.SetMockCookies(map[string]string{
		"AZAPERSISTENCE": "test-persistence",
		"csid":           "test-csid",
		"cstoken":        "test-cstoken",
		"AZACSRF":        "test-csrf",
	})

	// Test subscription
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	subscription, err := avanzaClient.Market.SubscribeToOrderDepth(ctx, "2185403")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer subscription.Close()

	// Collect events
	var events []market.OrderDepthEvent
	var errors []error

	done := make(chan bool)
	go func() {
		defer close(done)
		for {
			select {
			case event := <-subscription.Events():
				events = append(events, event)
			case err := <-subscription.Errors():
				errors = append(errors, err)
			case <-ctx.Done():
				return
			}
		}
	}()

	<-done

	// Verify we received events
	if len(events) == 0 {
		t.Error("Expected to receive events, but got none")
	}

	// Check for ORDER_DEPTH event
	var orderDepthEvent *market.OrderDepthEvent
	for _, event := range events {
		if event.Event == "ORDER_DEPTH" {
			orderDepthEvent = &event
			break
		}
	}

	if orderDepthEvent == nil {
		t.Error("Expected to receive ORDER_DEPTH event")
	} else {
		if orderDepthEvent.Data.OrderbookID != "2185403" {
			t.Errorf("Expected orderbook ID 2185403, got %s", orderDepthEvent.Data.OrderbookID)
		}
		if len(orderDepthEvent.Data.Levels) == 0 {
			t.Error("Expected at least one price level")
		} else {
			level := orderDepthEvent.Data.Levels[0]
			if level.BuyPrice != 10.00 {
				t.Errorf("Expected buy price 10.00, got %f", level.BuyPrice)
			}
			if level.BuyVolume != 500 {
				t.Errorf("Expected buy volume 500, got %f", level.BuyVolume)
			}
		}
	}
}

func TestOrderDepthDataUnmarshal(t *testing.T) {
	jsonData := `{
		"orderbookId": "2185403",
		"levels": [
			{
				"buyPrice": 10.00,
				"buyVolume": 500,
				"sellPrice": 10.15,
				"sellVolume": 25000
			},
			{
				"buyPrice": 9.96,
				"buyVolume": 25000,
				"sellPrice": 11.14,
				"sellVolume": 4761
			}
		],
		"marketMakerLevelInAsk": 0,
		"marketMakerLevelInBid": 1
	}`

	var data market.OrderDepthData
	err := json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		t.Fatalf("Failed to unmarshal order depth data: %v", err)
	}

	if data.OrderbookID != "2185403" {
		t.Errorf("Expected orderbook ID 2185403, got %s", data.OrderbookID)
	}

	if len(data.Levels) != 2 {
		t.Errorf("Expected 2 levels, got %d", len(data.Levels))
	}

	if data.MarketMakerLevelInAsk != 0 {
		t.Errorf("Expected market maker level in ask 0, got %d", data.MarketMakerLevelInAsk)
	}

	if data.MarketMakerLevelInBid != 1 {
		t.Errorf("Expected market maker level in bid 1, got %d", data.MarketMakerLevelInBid)
	}

	// Check first level
	level := data.Levels[0]
	if level.BuyPrice != 10.00 {
		t.Errorf("Expected buy price 10.00, got %f", level.BuyPrice)
	}
	if level.BuyVolume != 500 {
		t.Errorf("Expected buy volume 500, got %f", level.BuyVolume)
	}
	if level.SellPrice != 10.15 {
		t.Errorf("Expected sell price 10.15, got %f", level.SellPrice)
	}
	if level.SellVolume != 25000 {
		t.Errorf("Expected sell volume 25000, got %f", level.SellVolume)
	}
}

func TestClientMethods(t *testing.T) {
	// Test the new client methods
	c := client.NewClient()

	if c.BaseURL() != client.BaseURL {
		t.Errorf("Expected base URL %s, got %s", client.BaseURL, c.BaseURL())
	}

	if c.HTTPClient() == nil {
		t.Error("Expected HTTP client to be non-nil")
	}

	if c.SecurityToken() != "" {
		t.Error("Expected security token to be empty initially")
	}

	cookies := c.Cookies()
	if len(cookies) != 0 {
		t.Error("Expected cookies to be empty initially")
	}
}
