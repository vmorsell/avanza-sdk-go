// Package market provides market data functionality for the Avanza API.
package market

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/vmorsell/avanza-sdk-go/client"
)

// OrderDepthSubscription represents an active order depth subscription.
type OrderDepthSubscription struct {
	orderbookID string
	client      *client.Client
	ctx         context.Context
	cancel      context.CancelFunc
	events      chan OrderDepthEvent
	errors      chan error
	wg          sync.WaitGroup
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

// start begins the SSE stream processing.
func (s *OrderDepthSubscription) start() {
	s.wg.Add(1)
	defer s.wg.Done()

	defer func() {
		if r := recover(); r != nil {
			s.trySendError(fmt.Errorf("subscription panic: %v", r))
		}
	}()

	endpoint := fmt.Sprintf("/_push/order-depth-web-push/%s", url.PathEscape(s.orderbookID))

	req, err := http.NewRequestWithContext(s.ctx, "GET", s.client.BaseURL()+endpoint, nil)
	if err != nil {
		s.trySendError(fmt.Errorf("create request: %w", err))
		return
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
		s.trySendError(fmt.Errorf("request failed: %w", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.trySendError(client.NewHTTPError(resp))
		return
	}

	s.processSSEStream(resp)
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
func (s *OrderDepthSubscription) processSSEStream(resp *http.Response) {
	scanner := bufio.NewScanner(resp.Body)

	var event OrderDepthEvent

	for scanner.Scan() {
		select {
		case <-s.ctx.Done():
			return
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
		case "retry":
			if retry, err := json.Number(value).Int64(); err == nil {
				event.Retry = int(retry)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		s.trySendError(fmt.Errorf("stream error: %w", err))
	}
}
