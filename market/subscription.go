// Package market provides market data functionality for the Avanza API.
package market

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/vmorsell/avanza-sdk-go/internal/sse"
)

// OrderDepthSubscription represents an active order depth subscription.
type OrderDepthSubscription struct {
	sub       *sse.Subscription
	events    chan OrderDepthEvent
	errors    chan error
	done      chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup
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
// Always call Close() when done with the subscription to prevent resource leaks.
func (s *OrderDepthSubscription) Close() {
	s.closeOnce.Do(func() { close(s.done) })
	s.sub.Close()
	s.wg.Wait()
}

func newOrderDepthSubscription(sub *sse.Subscription) *OrderDepthSubscription {
	s := &OrderDepthSubscription{
		sub:    sub,
		events: make(chan OrderDepthEvent, 100),
		errors: make(chan error, 10),
		done:   make(chan struct{}),
	}
	s.wg.Add(1)
	go s.run()
	return s
}

func (s *OrderDepthSubscription) run() {
	defer s.wg.Done()
	defer close(s.events)
	defer close(s.errors)

	rawEvents := s.sub.Events()
	rawErrors := s.sub.Errors()

	for rawEvents != nil || rawErrors != nil {
		select {
		case raw, ok := <-rawEvents:
			if !ok {
				rawEvents = nil
				continue
			}
			s.forwardEvent(raw)
		case err, ok := <-rawErrors:
			if !ok {
				rawErrors = nil
				continue
			}
			select {
			case s.errors <- err:
			default:
			}
		}
	}
}

func (s *OrderDepthSubscription) forwardEvent(raw sse.RawEvent) {
	if raw.Event != "ORDER_DEPTH" {
		s.trySendEvent(OrderDepthEvent{Event: raw.Event, ID: raw.ID, Retry: raw.Retry})
		return
	}

	var data OrderDepthData
	if err := json.Unmarshal(raw.Data, &data); err != nil {
		select {
		case s.errors <- fmt.Errorf("parse order depth data: %w", err):
		default:
		}
		return
	}

	s.trySendEvent(OrderDepthEvent{
		Event: raw.Event,
		Data:  data,
		ID:    raw.ID,
		Retry: raw.Retry,
	})
}

// trySendEvent sends without blocking Close: if the consumer has stopped
// reading and the subscription is being closed, the event is dropped.
func (s *OrderDepthSubscription) trySendEvent(event OrderDepthEvent) {
	select {
	case s.events <- event:
	case <-s.done:
	}
}
