package market

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vmorsell/avanza-sdk-go/client"
)

func writeSSEEvent(w http.ResponseWriter, id, event, data string) {
	fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", id, event, data)
	w.(http.Flusher).Flush()
}

func TestOrderDepthSubscription_ReceivesEvents(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		data := `{"orderbookId":"12345","levels":[{"buyPrice":100.5,"buyVolume":200,"sellPrice":101.0,"sellVolume":150}],"marketMakerLevelInAsk":0,"marketMakerLevelInBid":0}`
		writeSSEEvent(w, "e1", "ORDER_DEPTH", data)
	}))
	defer srv.Close()

	c := client.NewClient(client.WithBaseURL(srv.URL))
	c.SetMockCookies(map[string]string{"csid": "a", "cstoken": "b", "AZACSRF": "c"})
	svc := NewService(c)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, err := svc.SubscribeToOrderDepth(ctx, "12345")
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	select {
	case e := <-sub.Events():
		if e.Event != "ORDER_DEPTH" {
			t.Errorf("event type = %q, want ORDER_DEPTH", e.Event)
		}
		if e.Data.OrderbookID != "12345" {
			t.Errorf("orderbookID = %q, want 12345", e.Data.OrderbookID)
		}
		if len(e.Data.Levels) != 1 {
			t.Fatalf("levels count = %d, want 1", len(e.Data.Levels))
		}
		if e.Data.Levels[0].BuyPrice != 100.5 {
			t.Errorf("buy price = %f, want 100.5", e.Data.Levels[0].BuyPrice)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for event")
	}

	sub.Close()
}

func TestOrderDepthChannelsCloseOnSSEDeath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "forbidden")
	}))
	defer srv.Close()

	c := client.NewClient(client.WithBaseURL(srv.URL))
	c.SetMockCookies(map[string]string{"csid": "a", "cstoken": "b", "AZACSRF": "c"})
	svc := NewService(c)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, err := svc.SubscribeToOrderDepth(ctx, "12345")
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	// Drain the error forwarded from the SSE layer.
	select {
	case <-sub.Errors():
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for error")
	}

	// Both channels should close once the SSE sub dies.
	select {
	case _, ok := <-sub.Events():
		if ok {
			t.Fatal("events channel should be closed")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("events channel not closed")
	}

	select {
	case _, ok := <-sub.Errors():
		if ok {
			t.Fatal("errors channel should be closed")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("errors channel not closed")
	}
}
