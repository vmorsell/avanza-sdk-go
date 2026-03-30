package trading

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

func TestSubscribeToOrders_RequiresAuth(t *testing.T) {
	c := client.NewClient()
	svc := NewService(c)

	_, err := svc.SubscribeToOrders(context.Background())
	if err == nil {
		t.Fatal("expected error for unauthenticated request")
	}
	if got := err.Error(); got != "subscribe to orders: no authentication cookies found - please authenticate first" {
		t.Errorf("unexpected error: %s", got)
	}
}

func TestSubscribeToOrders_RequiresEssentialCookies(t *testing.T) {
	c := client.NewClient()
	c.SetMockCookies(map[string]string{"csid": "a"}) // missing cstoken and AZACSRF
	svc := NewService(c)

	_, err := svc.SubscribeToOrders(context.Background())
	if err == nil {
		t.Fatal("expected error for missing cookies")
	}
}

func TestOrdersSubscription_ReceivesEvents(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		data := `{"id":"123","accountId":"456","orderbook":{"id":"5240","name":"Ericsson B","tickerSymbol":"ERIC B","marketplaceName":"XSTO","countryCode":"SE","instrumentType":"STOCK","tradable":true,"volumeFactor":1,"currencyCode":"SEK","flagCode":"SE"},"currentVolume":100,"originalVolume":100,"openVolume":null,"price":90,"validDate":null,"type":"BUY","state":{"value":"Väntande","description":"Din order skickas iväg när marknaden öppnar.","name":"ACTIVE_PENDING"},"action":"NEW","modifiable":true,"deletable":true,"sum":9000,"visibleDate":null,"orderDateTime":1769636379557,"eventTimeStamp":1769636379587,"uniqueId":"123_NEW_1769636379587","additionalParameters":null,"detailedCancelStatus":null,"condition":"NORMAL"}`
		writeSSEEvent(w, "123_NEW_1769636379587", "ORDER", data)
	}))
	defer srv.Close()

	c := client.NewClient(client.WithBaseURL(srv.URL))
	c.SetMockCookies(map[string]string{"csid": "a", "cstoken": "b", "AZACSRF": "c"})
	svc := NewService(c)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, err := svc.SubscribeToOrders(ctx)
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	select {
	case e := <-sub.Events():
		if e.Event != "ORDER" {
			t.Errorf("event type = %q, want ORDER", e.Event)
		}
		if e.Data.ID != "123" {
			t.Errorf("order ID = %q, want 123", e.Data.ID)
		}
		if e.Data.Action != OrderActionNew {
			t.Errorf("action = %q, want NEW", e.Data.Action)
		}
		if e.Data.Type != OrderSideBuy {
			t.Errorf("type = %q, want BUY", e.Data.Type)
		}
		if e.Data.Price != 90 {
			t.Errorf("price = %f, want 90", e.Data.Price)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for event")
	}

	sub.Close()
}
