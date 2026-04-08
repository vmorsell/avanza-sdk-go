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

func TestOrdersChannelsCloseOnSSEDeath(t *testing.T) {
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

	sub, err := svc.SubscribeToOrders(ctx)
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	// Drain the error forwarded from the SSE layer.
	select {
	case <-sub.Errors():
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for error")
	}

	// Both trading-level channels should close once the SSE sub dies.
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

func TestStopLossChannelsCloseOnSSEDeath(t *testing.T) {
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

	sub, err := svc.SubscribeToStopLoss(ctx)
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	// Drain the error forwarded from the SSE layer.
	select {
	case <-sub.Errors():
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for error")
	}

	// Both trading-level channels should close once the SSE sub dies.
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

func TestSubscribeToStopLoss_RequiresAuth(t *testing.T) {
	c := client.NewClient()
	svc := NewService(c)

	_, err := svc.SubscribeToStopLoss(context.Background())
	if err == nil {
		t.Fatal("expected error for unauthenticated request")
	}
	if got := err.Error(); got != "subscribe to stop loss: no authentication cookies found - please authenticate first" {
		t.Errorf("unexpected error: %s", got)
	}
}

func TestSubscribeToStopLoss_RequiresEssentialCookies(t *testing.T) {
	c := client.NewClient()
	c.SetMockCookies(map[string]string{"csid": "a"}) // missing cstoken and AZACSRF
	svc := NewService(c)

	_, err := svc.SubscribeToStopLoss(context.Background())
	if err == nil {
		t.Fatal("expected error for missing cookies")
	}
}

func TestStopLossSubscription_ReceivesUpdatedEvent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		data := `{"id":"A4^1773297345776^844478","uniqueId":"A4^1773297345776^844478_UPDATED_1774857773425","status":"ACTIVE","accountId":"84039","orderbook":{"id":"5246","name":"Investor A","countryCode":"SE","currency":"SEK","shortName":"INVE A","type":"STOCK"},"order":{"type":"SELL","price":361.000000,"volume":10.000000,"shortSellingAllowed":false,"validDays":8,"priceType":"MONETARY","priceDecimalPrecision":6},"trigger":{"value":360.000000,"type":"MORE_OR_EQUAL","validUntil":"2026-04-29","validDays":null,"valueType":"MONETARY","extremePrice":null},"editable":true,"deletable":true,"message":null,"pushAction":"UPDATED"}`
		writeSSEEvent(w, "A4^1773297345776^844478_UPDATED_1774857773425", "STOPLOSS", data)
	}))
	defer srv.Close()

	c := client.NewClient(client.WithBaseURL(srv.URL))
	c.SetMockCookies(map[string]string{"csid": "a", "cstoken": "b", "AZACSRF": "c"})
	svc := NewService(c)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, err := svc.SubscribeToStopLoss(ctx)
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	select {
	case e := <-sub.Events():
		if e.Event != "STOPLOSS" {
			t.Errorf("event type = %q, want STOPLOSS", e.Event)
		}
		if e.Data.ID != "A4^1773297345776^844478" {
			t.Errorf("ID = %q, want A4^1773297345776^844478", e.Data.ID)
		}
		if e.Data.Status != StopLossEventStatusActive {
			t.Errorf("status = %q, want ACTIVE", e.Data.Status)
		}
		if e.Data.PushAction != StopLossPushActionUpdated {
			t.Errorf("pushAction = %q, want UPDATED", e.Data.PushAction)
		}
		if e.Data.Order == nil {
			t.Fatal("order should not be nil for UPDATED event")
		}
		if e.Data.Order.Price != 361 {
			t.Errorf("order price = %f, want 361", e.Data.Order.Price)
		}
		if e.Data.Trigger == nil {
			t.Fatal("trigger should not be nil for UPDATED event")
		}
		if e.Data.Trigger.Value != 360 {
			t.Errorf("trigger value = %f, want 360", e.Data.Trigger.Value)
		}
		if e.Data.Trigger.Type != StopLossTriggerMoreOrEqual {
			t.Errorf("trigger type = %q, want MORE_OR_EQUAL", e.Data.Trigger.Type)
		}
		if e.Data.Editable != true {
			t.Error("editable should be true")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for event")
	}

	sub.Close()
}

func TestStopLossSubscription_ReceivesDeletedEvent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		data := `{"id":"A4^1773297345776^844478","uniqueId":"A4^1773297345776^844478_DELETED_1774857787422","status":"DELETED","accountId":"84039","orderbook":{"id":"5246","name":"Investor A","countryCode":"SE","currency":"SEK","shortName":"INVE A","type":"STOCK"},"order":null,"trigger":null,"editable":false,"deletable":false,"message":null,"pushAction":"DELETED"}`
		writeSSEEvent(w, "A4^1773297345776^844478_DELETED_1774857787422", "STOPLOSS", data)
	}))
	defer srv.Close()

	c := client.NewClient(client.WithBaseURL(srv.URL))
	c.SetMockCookies(map[string]string{"csid": "a", "cstoken": "b", "AZACSRF": "c"})
	svc := NewService(c)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub, err := svc.SubscribeToStopLoss(ctx)
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}

	select {
	case e := <-sub.Events():
		if e.Data.Status != StopLossEventStatusDeleted {
			t.Errorf("status = %q, want DELETED", e.Data.Status)
		}
		if e.Data.PushAction != StopLossPushActionDeleted {
			t.Errorf("pushAction = %q, want DELETED", e.Data.PushAction)
		}
		if e.Data.Order != nil {
			t.Error("order should be nil for DELETED event")
		}
		if e.Data.Trigger != nil {
			t.Error("trigger should be nil for DELETED event")
		}
		if e.Data.Editable != false {
			t.Error("editable should be false")
		}
		if e.Data.Deletable != false {
			t.Error("deletable should be false")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for event")
	}

	sub.Close()
}
