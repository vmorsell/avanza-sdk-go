package trading

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vmorsell/avanza-sdk-go/client"
)

// writeSSEEvent writes a single SSE event to the response writer and flushes.
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

	cancel()
	sub.wg.Wait()
}

func TestOrdersSubscription_ReconnectsAfterDrop(t *testing.T) {
	var connCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := connCount.Add(1)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		data := `{"id":"123","accountId":"456","orderbook":{"id":"5240","name":"Test","tickerSymbol":"TST","marketplaceName":"XSTO","countryCode":"SE","instrumentType":"STOCK","tradable":true,"volumeFactor":1,"currencyCode":"SEK","flagCode":"SE"},"currentVolume":100,"originalVolume":100,"openVolume":null,"price":90,"validDate":null,"type":"BUY","state":{"value":"Test","description":"Test","name":"ACTIVE_PENDING"},"action":"NEW","modifiable":true,"deletable":true,"sum":9000,"visibleDate":null,"orderDateTime":1769636379557,"eventTimeStamp":1769636379587,"uniqueId":"evt","additionalParameters":null,"detailedCancelStatus":null,"condition":"NORMAL"}`
		writeSSEEvent(w, fmt.Sprintf("evt-%d", n), "ORDER", data)
		// Drop connection
	}))
	defer srv.Close()

	c := client.NewClient(client.WithBaseURL(srv.URL))
	c.SetMockCookies(map[string]string{"csid": "a", "cstoken": "b", "AZACSRF": "c"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := &OrdersSubscription{
		client:        c,
		ctx:           ctx,
		cancel:        cancel,
		events:        make(chan OrderEvent, 100),
		errors:        make(chan error, 10),
		retryInterval: 10 * time.Millisecond,
	}
	go sub.start()

	var events []OrderEvent
	timeout := time.After(5 * time.Second)
	for len(events) < 2 {
		select {
		case e := <-sub.events:
			events = append(events, e)
		case <-timeout:
			t.Fatalf("timed out waiting for events, got %d", len(events))
		}
	}

	cancel()
	sub.wg.Wait()

	if events[0].ID != "evt-1" {
		t.Errorf("first event ID = %q, want evt-1", events[0].ID)
	}
	if events[1].ID != "evt-2" {
		t.Errorf("second event ID = %q, want evt-2", events[1].ID)
	}
	if connCount.Load() < 2 {
		t.Errorf("connection count = %d, want >= 2", connCount.Load())
	}
}

func TestOrdersSubscription_SendsLastEventID(t *testing.T) {
	var connCount atomic.Int32
	var secondRequestLastEventID atomic.Value

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := connCount.Add(1)

		if n == 2 {
			secondRequestLastEventID.Store(r.Header.Get("Last-Event-ID"))
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		data := `{"id":"123","accountId":"456","orderbook":{"id":"5240","name":"Test","tickerSymbol":"TST","marketplaceName":"XSTO","countryCode":"SE","instrumentType":"STOCK","tradable":true,"volumeFactor":1,"currencyCode":"SEK","flagCode":"SE"},"currentVolume":100,"originalVolume":100,"openVolume":null,"price":90,"validDate":null,"type":"BUY","state":{"value":"Test","description":"Test","name":"ACTIVE_PENDING"},"action":"NEW","modifiable":true,"deletable":true,"sum":9000,"visibleDate":null,"orderDateTime":1769636379557,"eventTimeStamp":1769636379587,"uniqueId":"evt","additionalParameters":null,"detailedCancelStatus":null,"condition":"NORMAL"}`
		writeSSEEvent(w, "my-event-42", "ORDER", data)
	}))
	defer srv.Close()

	c := client.NewClient(client.WithBaseURL(srv.URL))
	c.SetMockCookies(map[string]string{"csid": "a", "cstoken": "b", "AZACSRF": "c"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := &OrdersSubscription{
		client:        c,
		ctx:           ctx,
		cancel:        cancel,
		events:        make(chan OrderEvent, 100),
		errors:        make(chan error, 10),
		retryInterval: 10 * time.Millisecond,
	}
	go sub.start()

	timeout := time.After(5 * time.Second)
	eventsReceived := 0
	for eventsReceived < 2 {
		select {
		case <-sub.events:
			eventsReceived++
		case <-timeout:
			t.Fatalf("timed out waiting for reconnection, got %d events", eventsReceived)
		}
	}

	cancel()
	sub.wg.Wait()

	got, ok := secondRequestLastEventID.Load().(string)
	if !ok || got != "my-event-42" {
		t.Errorf("Last-Event-ID on reconnect = %q, want my-event-42", got)
	}
}

func TestOrdersSubscription_StopsOn4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "forbidden")
	}))
	defer srv.Close()

	c := client.NewClient(client.WithBaseURL(srv.URL))
	c.SetMockCookies(map[string]string{"csid": "a", "cstoken": "b", "AZACSRF": "c"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := &OrdersSubscription{
		client: c,
		ctx:    ctx,
		cancel: cancel,
		events: make(chan OrderEvent, 100),
		errors: make(chan error, 10),
	}
	go sub.start()

	select {
	case err := <-sub.errors:
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for error")
	}

	sub.wg.Wait()
}

func TestOrdersSubscription_CloseDuringWait(t *testing.T) {
	var connCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connCount.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "server error")
	}))
	defer srv.Close()

	c := client.NewClient(client.WithBaseURL(srv.URL))
	c.SetMockCookies(map[string]string{"csid": "a", "cstoken": "b", "AZACSRF": "c"})

	ctx, cancel := context.WithCancel(context.Background())

	sub := &OrdersSubscription{
		client:        c,
		ctx:           ctx,
		cancel:        cancel,
		events:        make(chan OrderEvent, 100),
		errors:        make(chan error, 10),
		retryInterval: 10 * time.Second,
	}
	go sub.start()

	deadline := time.After(5 * time.Second)
	for connCount.Load() < 1 {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for first connection attempt")
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}

	done := make(chan struct{})
	go func() {
		sub.Close()
		close(done)
	}()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("Close() hung during reconnect wait")
	}
}
