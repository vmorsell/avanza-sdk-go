// Package market provides market data functionality for the Avanza API.
package market

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/vmorsell/avanza-sdk-go/client"
)

const (
	defaultRetryInterval = 3 * time.Second
	maxRetryInterval     = 30 * time.Second
)

// OrderDepthSubscription represents an active order depth subscription.
type OrderDepthSubscription struct {
	orderbookID   string
	client        *client.Client
	ctx           context.Context
	cancel        context.CancelFunc
	events        chan OrderDepthEvent
	errors        chan error
	wg            sync.WaitGroup
	lastEventID   string
	retryInterval time.Duration
}

// Events returns a channel that receives order depth events.
func (s *OrderDepthSubscription) Events() <-chan OrderDepthEvent {
	return s.events
}

// Errors returns a channel that receives any errors from the subscription.
func (s *OrderDepthSubscription) Errors() <-chan error {
	return s.errors
}

// Close stops the subscription and cleans up resources.
// It waits for the background goroutine to finish before closing channels.
//
// Always call Close() when done with the subscription to prevent resource leaks.
func (s *OrderDepthSubscription) Close() {
	s.cancel()
	s.wg.Wait()
	close(s.events)
	close(s.errors)
}

// trySendError sends an error without blocking if the context is cancelled.
func (s *OrderDepthSubscription) trySendError(err error) {
	select {
	case s.errors <- err:
	case <-s.ctx.Done():
	}
}

// trySendEvent sends an event without blocking if the context is cancelled.
func (s *OrderDepthSubscription) trySendEvent(event OrderDepthEvent) {
	select {
	case s.events <- event:
	case <-s.ctx.Done():
	}
}

// start begins the SSE stream processing with automatic reconnection.
func (s *OrderDepthSubscription) start() {
	s.wg.Add(1)
	defer s.wg.Done()

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
		if err != nil && !isRecoverable(err) {
			s.trySendError(err)
			return
		}
		if connected {
			attempt = 0
		}

		wait := s.retryInterval
		if attempt > 0 {
			wait = exponentialBackoff(s.retryInterval, attempt)
		}

		select {
		case <-s.ctx.Done():
			return
		case <-time.After(wait):
		}
	}
}

// connectAndStream establishes an SSE connection and processes the stream.
// It returns (true, err) if it connected and streamed before failing,
// or (false, err) if it couldn't connect at all.
func (s *OrderDepthSubscription) connectAndStream() (bool, error) {
	endpoint := fmt.Sprintf("/_push/order-depth-web-push/%s", url.PathEscape(s.orderbookID))

	req, err := http.NewRequestWithContext(s.ctx, "GET", s.client.BaseURL()+endpoint, nil)
	if err != nil {
		return false, fmt.Errorf("create request: %w", err)
	}

	s.setSSEHeaders(req)

	// Reuse transport for connection pooling, disable timeout for long-lived SSE
	baseClient := s.client.HTTPClient()
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

// setSSEHeaders sets the appropriate headers for Server-Sent Events.
func (s *OrderDepthSubscription) setSSEHeaders(req *http.Request) {
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Accept-Language", "en-US,en;q=0.6")
	req.Header.Set("aza-do-not-touch-session", "true")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Priority", "u=1, i")
	req.Header.Set("Referer", fmt.Sprintf("https://www.avanza.se/handla/order.html/kop/%s", url.PathEscape(s.orderbookID)))
	req.Header.Set("Sec-Ch-Ua", `"Not)A;Brand";v="8", "Chromium";v="138", "Brave";v="138"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Gpc", "1")
	req.Header.Set("User-Agent", s.client.UserAgent())

	if s.lastEventID != "" {
		req.Header.Set("Last-Event-ID", s.lastEventID)
	}

	if token := s.client.SecurityToken(); token != "" {
		req.Header.Set("X-Securitytoken", token)
	}

	if cookies := s.client.Cookies(); len(cookies) > 0 {
		var cookiePairs []string
		for name, value := range cookies {
			if name != "" && value != "" {
				cookiePairs = append(cookiePairs, fmt.Sprintf("%s=%s", name, value))
			}
		}
		if len(cookiePairs) > 0 {
			req.Header.Set("Cookie", strings.Join(cookiePairs, "; "))
		}
	}
}

// processSSEStream processes the Server-Sent Events stream.
// It returns an error if the stream ends unexpectedly, or nil if it ends cleanly.
func (s *OrderDepthSubscription) processSSEStream(resp *http.Response) error {
	scanner := bufio.NewScanner(resp.Body)

	var event OrderDepthEvent

	for scanner.Scan() {
		select {
		case <-s.ctx.Done():
			return nil
		default:
		}

		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			// SSE protocol: empty line marks end of event
			if event.Event != "" {
				s.trySendEvent(event)
				event = OrderDepthEvent{}
			}
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		field := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch field {
		case "event":
			event.Event = value
		case "data":
			if event.Event == "ORDER_DEPTH" {
				var orderDepthData OrderDepthData
				if err := json.Unmarshal([]byte(value), &orderDepthData); err != nil {
					s.trySendError(fmt.Errorf("parse order depth data: %w", err))
					continue
				}
				event.Data = orderDepthData
			}
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

// isRecoverable reports whether the error is transient and the connection should be retried.
func isRecoverable(err error) bool {
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

	// Network/IO errors are recoverable
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	return true
}

// exponentialBackoff returns a wait duration using exponential backoff.
// The formula is base * 2^min(attempt, 5), capped at maxRetryInterval.
func exponentialBackoff(base time.Duration, attempt int) time.Duration {
	wait := base << uint(min(attempt, 5))
	return min(wait, maxRetryInterval)
}
