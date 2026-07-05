// Package sse provides a shared Server-Sent Events client with automatic reconnection.
package sse

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/vmorsell/avanza-sdk-go/client"
)

const (
	defaultRetryInterval = 3 * time.Second
	maxRetryInterval     = 30 * time.Second

	// maxEventSize is the largest single SSE line we accept. The scanner's
	// default 64KB limit is too small for large order-depth snapshots.
	maxEventSize = 1024 * 1024
)

// Config holds the parameters that differ between subscription types.
type Config struct {
	Client   *client.Client
	Endpoint string // e.g. "/_push/trading/orders/"
	Referer  string // e.g. "https://www.avanza.se/min-ekonomi/ordrar.html"
}

// RawEvent is a single SSE event with unparsed data.
type RawEvent struct {
	Event string
	Data  json.RawMessage
	ID    string
	Retry int
}

// Subscription manages an SSE connection with automatic reconnection.
type Subscription struct {
	cfg           Config
	ctx           context.Context
	cancel        context.CancelFunc
	events        chan RawEvent
	errors        chan error
	wg            sync.WaitGroup
	lastEventID   string
	retryInterval time.Duration
}

// New creates and starts a Subscription.
func New(ctx context.Context, cfg Config) *Subscription {
	subCtx, cancel := context.WithCancel(ctx)
	s := &Subscription{
		cfg:    cfg,
		ctx:    subCtx,
		cancel: cancel,
		events: make(chan RawEvent, 100),
		errors: make(chan error, 10),
	}
	s.wg.Add(1)
	go s.start()
	return s
}

// Events returns a channel that receives raw SSE events.
func (s *Subscription) Events() <-chan RawEvent {
	return s.events
}

// Errors returns a channel that receives any errors from the subscription.
func (s *Subscription) Errors() <-chan error {
	return s.errors
}

// Close stops the subscription and cleans up resources.
func (s *Subscription) Close() {
	s.cancel()
	s.wg.Wait()
}

func (s *Subscription) trySendError(err error) {
	select {
	case s.errors <- err:
	case <-s.ctx.Done():
	}
}

func (s *Subscription) trySendEvent(event RawEvent) {
	select {
	case s.events <- event:
	case <-s.ctx.Done():
	}
}

// start begins the SSE stream processing with automatic reconnection.
// The caller must have incremented s.wg before starting the goroutine.
func (s *Subscription) start() {
	defer s.wg.Done()
	defer close(s.events)
	defer close(s.errors)

	defer func() {
		if r := recover(); r != nil {
			s.trySendError(fmt.Errorf("subscription panic: %v", r))
		}
	}()

	s.retryInterval = defaultRetryInterval

	for attempt := 0; ; attempt++ {
		connected, err := s.connectAndStream()

		if s.ctx.Err() != nil {
			return
		}
		if err != nil && !IsRecoverable(err) {
			s.trySendError(err)
			return
		}
		if connected {
			attempt = 0
		}

		wait := s.retryInterval
		if attempt > 0 {
			wait = ExponentialBackoff(s.retryInterval, attempt)
		}

		select {
		case <-s.ctx.Done():
			return
		case <-time.After(wait):
		}
	}
}

func (s *Subscription) connectAndStream() (bool, error) {
	req, err := http.NewRequestWithContext(s.ctx, "GET", s.cfg.Client.BaseURL()+s.cfg.Endpoint, nil)
	if err != nil {
		return false, fmt.Errorf("create request: %w", err)
	}

	s.setSSEHeaders(req)

	baseClient := s.cfg.Client.HTTPClient()
	httpClient := &http.Client{
		Transport: baseClient.Transport,
		Timeout:   0,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, client.NewHTTPError(resp)
	}

	err = s.processSSEStream(resp)
	return true, err
}

func (s *Subscription) setSSEHeaders(req *http.Request) {
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Accept-Language", "en-US,en;q=0.6")
	req.Header.Set("aza-do-not-touch-session", "true")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Priority", "u=1, i")
	req.Header.Set("Referer", s.cfg.Referer)
	req.Header.Set("Sec-Ch-Ua", `"Not)A;Brand";v="8", "Chromium";v="138", "Google Chrome";v="138"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", s.cfg.Client.UserAgent())

	if s.lastEventID != "" {
		req.Header.Set("Last-Event-ID", s.lastEventID)
	}

	if token := s.cfg.Client.SecurityToken(); token != "" {
		req.Header.Set("X-SecurityToken", token)
	}

	if cookie := s.cfg.Client.CookieHeader(); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
}

func (s *Subscription) processSSEStream(resp *http.Response) error {
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), maxEventSize)

	var event RawEvent
	var dataLines []string

	for scanner.Scan() {
		select {
		case <-s.ctx.Done():
			return nil
		default:
		}

		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			if event.Event != "" || len(dataLines) > 0 {
				if event.Event == "" {
					event.Event = "message"
				}
				if len(dataLines) > 0 {
					event.Data = json.RawMessage(strings.Join(dataLines, "\n"))
				}
				s.trySendEvent(event)
			}
			event = RawEvent{}
			dataLines = nil
			continue
		}

		field, value, found := strings.Cut(line, ":")
		if !found {
			// A line with no colon is a field with an empty value.
			field = line
			value = ""
		}
		field = strings.TrimSpace(field)
		value = strings.TrimSpace(value)

		switch field {
		case "event":
			event.Event = value
		case "data":
			dataLines = append(dataLines, value)
		case "id":
			event.ID = value
			s.lastEventID = value
		case "retry":
			if retry, err := json.Number(value).Int64(); err == nil {
				event.Retry = int(retry)
				s.retryInterval = time.Duration(retry) * time.Millisecond
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("stream error: %w", err)
	}
	return nil
}

// IsRecoverable reports whether the error is transient and the connection should be retried.
func IsRecoverable(err error) bool {
	if err == nil {
		return true
	}

	var httpErr *client.HTTPError
	if errors.As(err, &httpErr) {
		switch {
		case httpErr.StatusCode == http.StatusRequestTimeout,
			httpErr.StatusCode == http.StatusTooManyRequests:
			return true
		case httpErr.StatusCode >= 400 && httpErr.StatusCode < 500:
			return false
		case httpErr.StatusCode >= 500:
			return true
		}
	}

	if errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	return true
}

// ExponentialBackoff returns a wait duration using exponential backoff.
// The formula is base * 2^min(attempt, 5), capped at maxRetryInterval.
func ExponentialBackoff(base time.Duration, attempt int) time.Duration {
	wait := base << min(max(attempt, 0), 5)
	return min(wait, maxRetryInterval)
}
