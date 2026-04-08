package sse

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

// writeSSEEvent writes a single SSE event with event type "TEST" to the response writer and flushes.
func writeSSEEvent(w http.ResponseWriter, id, data string) {
	fmt.Fprintf(w, "id: %s\nevent: TEST\ndata: %s\n\n", id, data)
	w.(http.Flusher).Flush()
}

func newTestClient(url string) *client.Client {
	c := client.NewClient(client.WithBaseURL(url))
	c.SetMockCookies(map[string]string{"csid": "a", "cstoken": "b", "AZACSRF": "c"})
	return c
}

func TestReceivesEvents(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		writeSSEEvent(w, "e1", `{"foo":"bar"}`)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := New(ctx, Config{
		Client:   c,
		Endpoint: "/events",
		Referer:  "https://example.com",
	})

	select {
	case e := <-sub.Events():
		if e.Event != "TEST" {
			t.Errorf("event type = %q, want TEST", e.Event)
		}
		if e.ID != "e1" {
			t.Errorf("event ID = %q, want e1", e.ID)
		}
		if string(e.Data) != `{"foo":"bar"}` {
			t.Errorf("data = %q, want {\"foo\":\"bar\"}", string(e.Data))
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for event")
	}

	sub.Close()
}

func TestReconnectsAfterStreamDrop(t *testing.T) {
	var connCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := connCount.Add(1)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		writeSSEEvent(w, fmt.Sprintf("evt-%d", n), `{"n":`+fmt.Sprintf("%d", n)+`}`)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := &Subscription{
		cfg: Config{
			Client:   c,
			Endpoint: "/events",
			Referer:  "https://example.com",
		},
		ctx:           ctx,
		cancel:        cancel,
		events:        make(chan RawEvent, 100),
		errors:        make(chan error, 10),
		retryInterval: 10 * time.Millisecond,
	}
	go sub.start()

	var events []RawEvent
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
		writeSSEEvent(w, "my-event-42", `{}`)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := &Subscription{
		cfg: Config{
			Client:   c,
			Endpoint: "/events",
			Referer:  "https://example.com",
		},
		ctx:           ctx,
		cancel:        cancel,
		events:        make(chan RawEvent, 100),
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

func TestRespectsServerRetryField(t *testing.T) {
	var connCount atomic.Int32
	connTimes := make(chan time.Time, 10)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := connCount.Add(1)
		connTimes <- time.Now()

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		if n == 1 {
			fmt.Fprintf(w, "retry: 200\nid: e1\nevent: TEST\ndata: {}\n\n")
			w.(http.Flusher).Flush()
			return
		}
		writeSSEEvent(w, "e2", `{}`)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := &Subscription{
		cfg: Config{
			Client:   c,
			Endpoint: "/events",
			Referer:  "https://example.com",
		},
		ctx:           ctx,
		cancel:        cancel,
		events:        make(chan RawEvent, 100),
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
			t.Fatalf("timed out, got %d events", eventsReceived)
		}
	}

	cancel()
	sub.wg.Wait()

	var times []time.Time
	close(connTimes)
	for ct := range connTimes {
		times = append(times, ct)
	}

	if len(times) < 2 {
		t.Fatalf("expected at least 2 connections, got %d", len(times))
	}

	gap := times[1].Sub(times[0])
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

	c := newTestClient(srv.URL)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := &Subscription{
		cfg: Config{
			Client:   c,
			Endpoint: "/events",
			Referer:  "https://example.com",
		},
		ctx:    ctx,
		cancel: cancel,
		events: make(chan RawEvent, 100),
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

func TestCloseDuringReconnectWait(t *testing.T) {
	var connCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connCount.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "server error")
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	ctx, cancel := context.WithCancel(context.Background())

	sub := &Subscription{
		cfg: Config{
			Client:   c,
			Endpoint: "/events",
			Referer:  "https://example.com",
		},
		ctx:           ctx,
		cancel:        cancel,
		events:        make(chan RawEvent, 100),
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
	case <-time.After(2 * time.Second):
		t.Fatal("Close() hung during reconnect wait")
	}
}

func TestChannelsCloseOnNonRecoverableError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "forbidden")
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := New(ctx, Config{
		Client:   c,
		Endpoint: "/events",
		Referer:  "https://example.com",
	})

	// Drain the error that gets sent before channels close.
	select {
	case <-sub.Errors():
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for error")
	}

	// Both channels should now be closed.
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
		{4, 30 * time.Second},
		{5, 30 * time.Second},
		{6, 30 * time.Second},
		{100, 30 * time.Second},
	}

	for _, tt := range tests {
		got := ExponentialBackoff(base, tt.attempt)
		if got != tt.want {
			t.Errorf("ExponentialBackoff(%v, %d) = %v, want %v", base, tt.attempt, got, tt.want)
		}
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
			got := IsRecoverable(tt.err)
			if got != tt.want {
				t.Errorf("IsRecoverable(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
