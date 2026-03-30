// Package trading provides trading functionality for the Avanza API.
package trading

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/vmorsell/avanza-sdk-go/internal/sse"
)

// OrdersSubscription represents an active orders subscription.
type OrdersSubscription struct {
	sub    *sse.Subscription
	events chan OrderEvent
	errors chan error
	wg     sync.WaitGroup
}

// Events returns a channel that receives order events.
func (s *OrdersSubscription) Events() <-chan OrderEvent {
	return s.events
}

// Errors returns a channel that receives any errors from the subscription.
func (s *OrdersSubscription) Errors() <-chan error {
	return s.errors
}

// Close stops the subscription and cleans up resources.
// Always call Close() when done with the subscription to prevent resource leaks.
func (s *OrdersSubscription) Close() {
	s.sub.Close()
	s.wg.Wait()
	close(s.events)
	close(s.errors)
}

func newOrdersSubscription(sub *sse.Subscription) *OrdersSubscription {
	s := &OrdersSubscription{
		sub:    sub,
		events: make(chan OrderEvent, 100),
		errors: make(chan error, 10),
	}
	s.wg.Add(1)
	go s.run()
	return s
}

func (s *OrdersSubscription) run() {
	defer s.wg.Done()

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

func (s *OrdersSubscription) forwardEvent(raw sse.RawEvent) {
	if raw.Event != "ORDER" {
		s.events <- OrderEvent{Event: raw.Event, ID: raw.ID, Retry: raw.Retry}
		return
	}

	var data OrderEventData
	if err := json.Unmarshal(raw.Data, &data); err != nil {
		select {
		case s.errors <- fmt.Errorf("parse order data: %w", err):
		default:
		}
		return
	}

	s.events <- OrderEvent{
		Event: raw.Event,
		Data:  data,
		ID:    raw.ID,
		Retry: raw.Retry,
	}
}
