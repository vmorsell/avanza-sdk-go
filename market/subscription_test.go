package market

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

func TestReconnectsAfterStreamDrop(t *testing.T) {
	var connCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := connCount.Add(1)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		data := `{"orderbookId":"12345","levels":[],"marketMakerLevelInAsk":0,"marketMakerLevelInBid":0}`
		writeSSEEvent(w, fmt.Sprintf("evt-%d", n), "ORDER_DEPTH", data)

		if n == 1 {
			// First connection: send one event, then drop
			return
		}
		// Second connection: send one event, then drop
	}))
	defer srv.Close()

	c := client.NewClient(client.WithBaseURL(srv.URL))
	c.SetMockCookies(map[string]string{"csid": "a", "cstoken": "b", "AZACSRF": "c"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := &OrderDepthSubscription{
		orderbookID:   "12345",
		client:        c,
		ctx:           ctx,
		cancel:        cancel,
		events:        make(chan OrderDepthEvent, 100),
		errors:        make(chan error, 10),
		retryInterval: 10 * time.Millisecond, // fast for testing
	}
	go sub.start()

	// Collect events from both connections
	var events []OrderDepthEvent
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
		t.Errorf("first event ID = %q, want %q", events[0].ID, "evt-1")
	}
	if events[1].ID != "evt-2" {
		t.Errorf("second event ID = %q, want %q", events[1].ID, "evt-2")
	}
	if connCount.Load() < 2 {
		t.Errorf("connection count = %d, want >= 2", connCount.Load())
	}
}

func TestSendsLastEventIDOnReconnect(t *testing.T) {
	var connCount atomic.Int32
	var secondRequestLastEventID atomic.Value

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := connCount.Add(1)

		if n == 2 {
			secondRequestLastEventID.Store(r.Header.Get("Last-Event-ID"))
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		data := `{"orderbookId":"12345","levels":[],"marketMakerLevelInAsk":0,"marketMakerLevelInBid":0}`
		writeSSEEvent(w, "my-event-42", "ORDER_DEPTH", data)
		// Close connection to trigger reconnect
	}))
	defer srv.Close()

	c := client.NewClient(client.WithBaseURL(srv.URL))
	c.SetMockCookies(map[string]string{"csid": "a", "cstoken": "b", "AZACSRF": "c"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := &OrderDepthSubscription{
		orderbookID:   "12345",
		client:        c,
		ctx:           ctx,
		cancel:        cancel,
		events:        make(chan OrderDepthEvent, 100),
		errors:        make(chan error, 10),
		retryInterval: 10 * time.Millisecond,
	}
	go sub.start()

	// Wait for two connections
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
		t.Errorf("Last-Event-ID on reconnect = %q, want %q", got, "my-event-42")
	}
}

func TestRespectsServerRetryField(t *testing.T) {
	var connCount atomic.Int32
	connTimes := make(chan time.Time, 10)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := connCount.Add(1)
		connTimes <- time.Now()

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		if n == 1 {
			// Send a retry field with 200ms interval, then an event, then drop
			data := `{"orderbookId":"12345","levels":[],"marketMakerLevelInAsk":0,"marketMakerLevelInBid":0}`
			fmt.Fprintf(w, "retry: 200\nid: e1\nevent: ORDER_DEPTH\ndata: %s\n\n", data)
			w.(http.Flusher).Flush()
			return
		}
		// Second connection: send event
		data := `{"orderbookId":"12345","levels":[],"marketMakerLevelInAsk":0,"marketMakerLevelInBid":0}`
		writeSSEEvent(w, "e2", "ORDER_DEPTH", data)
	}))
	defer srv.Close()

	c := client.NewClient(client.WithBaseURL(srv.URL))
	c.SetMockCookies(map[string]string{"csid": "a", "cstoken": "b", "AZACSRF": "c"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := &OrderDepthSubscription{
		orderbookID:   "12345",
		client:        c,
		ctx:           ctx,
		cancel:        cancel,
		events:        make(chan OrderDepthEvent, 100),
		errors:        make(chan error, 10),
		retryInterval: 10 * time.Millisecond,
	}
	go sub.start()

	// Wait for two events (one per connection)
	timeout := time.After(5 * time.Second)
	eventsReceived := 0
	for eventsReceived < 2 {
		select {
		case <-sub.events:
			eventsReceived++
		case <-timeout:
			t.Fatalf("timed out, got %d events", eventsReceived)
		}
	}

	cancel()
	sub.wg.Wait()

	// Verify the retry interval was respected
	var times []time.Time
	close(connTimes)
	for ct := range connTimes {
		times = append(times, ct)
	}

	if len(times) < 2 {
		t.Fatalf("expected at least 2 connections, got %d", len(times))
	}

	gap := times[1].Sub(times[0])
	// The server set retry to 200ms. Allow some tolerance.
	if gap < 150*time.Millisecond {
		t.Errorf("reconnect gap = %v, want >= 150ms (server set retry: 200ms)", gap)
	}
}

func TestStopsOn4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "forbidden")
	}))
	defer srv.Close()

	c := client.NewClient(client.WithBaseURL(srv.URL))
	c.SetMockCookies(map[string]string{"csid": "a", "cstoken": "b", "AZACSRF": "c"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := &OrderDepthSubscription{
		orderbookID: "12345",
		client:      c,
		ctx:         ctx,
		cancel:      cancel,
		events:      make(chan OrderDepthEvent, 100),
		errors:      make(chan error, 10),
	}
	go sub.start()

	// Should receive a fatal error
	select {
	case err := <-sub.errors:
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		// Should contain status code info
		got := err.Error()
		if got == "" {
			t.Error("expected non-empty error message")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for error")
	}

	// The goroutine should exit (not retry)
	sub.wg.Wait()
}

func TestExponentialBackoff(t *testing.T) {
	base := 3 * time.Second

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 3 * time.Second},
		{1, 6 * time.Second},
		{2, 12 * time.Second},
		{3, 24 * time.Second},
		{4, 30 * time.Second}, // capped at maxRetryInterval
		{5, 30 * time.Second}, // capped
		{6, 30 * time.Second}, // capped, attempt clamped to 5
		{100, 30 * time.Second},
	}

	for _, tt := range tests {
		got := exponentialBackoff(base, tt.attempt)
		if got != tt.want {
			t.Errorf("exponentialBackoff(%v, %d) = %v, want %v", base, tt.attempt, got, tt.want)
		}
	}
}

func TestCloseDuringReconnectWait(t *testing.T) {
	var connCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connCount.Add(1)
		// Always fail with 500 to trigger reconnect with backoff
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "server error")
	}))
	defer srv.Close()

	c := client.NewClient(client.WithBaseURL(srv.URL))
	c.SetMockCookies(map[string]string{"csid": "a", "cstoken": "b", "AZACSRF": "c"})

	ctx, cancel := context.WithCancel(context.Background())

	sub := &OrderDepthSubscription{
		orderbookID:   "12345",
		client:        c,
		ctx:           ctx,
		cancel:        cancel,
		events:        make(chan OrderDepthEvent, 100),
		errors:        make(chan error, 10),
		retryInterval: 10 * time.Second, // long wait to ensure we interrupt it
	}
	go sub.start()

	// Wait for at least one connection attempt
	deadline := time.After(5 * time.Second)
	for connCount.Load() < 1 {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for first connection attempt")
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}

	// Close while it's waiting in the backoff sleep
	done := make(chan struct{})
	go func() {
		sub.Close()
		close(done)
	}()

	select {
	case <-done:
		// success: Close() returned promptly
	case <-time.After(2 * time.Second):
		t.Fatal("Close() hung during reconnect wait")
	}
}

func TestIsRecoverable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, true},
		{"408 Request Timeout", &client.HTTPError{StatusCode: 408}, true},
		{"429 Too Many Requests", &client.HTTPError{StatusCode: 429}, true},
		{"403 Forbidden", &client.HTTPError{StatusCode: 403}, false},
		{"401 Unauthorized", &client.HTTPError{StatusCode: 401}, false},
		{"404 Not Found", &client.HTTPError{StatusCode: 404}, false},
		{"500 Internal Server Error", &client.HTTPError{StatusCode: 500}, true},
		{"502 Bad Gateway", &client.HTTPError{StatusCode: 502}, true},
		{"503 Service Unavailable", &client.HTTPError{StatusCode: 503}, true},
		{"generic error", fmt.Errorf("network down"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRecoverable(tt.err)
			if got != tt.want {
				t.Errorf("isRecoverable(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
