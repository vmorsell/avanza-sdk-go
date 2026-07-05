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
	sub       *sse.Subscription
	events    chan OrderEvent
	errors    chan error
	done      chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup
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
	s.closeOnce.Do(func() { close(s.done) })
	s.sub.Close()
	s.wg.Wait()
}

func newOrdersSubscription(sub *sse.Subscription) *OrdersSubscription {
	s := &OrdersSubscription{
		sub:    sub,
		events: make(chan OrderEvent, 100),
		errors: make(chan error, 10),
		done:   make(chan struct{}),
	}
	s.wg.Add(1)
	go s.run()
	return s
}

func (s *OrdersSubscription) run() {
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

func (s *OrdersSubscription) forwardEvent(raw sse.RawEvent) {
	if raw.Event != "ORDER" {
		s.trySendEvent(OrderEvent{Event: raw.Event, ID: raw.ID, Retry: raw.Retry})
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

	s.trySendEvent(OrderEvent{
		Event: raw.Event,
		Data:  data,
		ID:    raw.ID,
		Retry: raw.Retry,
	})
}

// trySendEvent sends without blocking Close: if the consumer has stopped
// reading and the subscription is being closed, the event is dropped.
func (s *OrdersSubscription) trySendEvent(event OrderEvent) {
	select {
	case s.events <- event:
	case <-s.done:
	}
}

// StopLossSubscription represents an active stop loss subscription.
type StopLossSubscription struct {
	sub       *sse.Subscription
	events    chan StopLossEvent
	errors    chan error
	done      chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup
}

// Events returns a channel that receives stop loss events.
func (s *StopLossSubscription) Events() <-chan StopLossEvent {
	return s.events
}

// Errors returns a channel that receives any errors from the subscription.
func (s *StopLossSubscription) Errors() <-chan error {
	return s.errors
}

// Close stops the subscription and cleans up resources.
// Always call Close() when done with the subscription to prevent resource leaks.
func (s *StopLossSubscription) Close() {
	s.closeOnce.Do(func() { close(s.done) })
	s.sub.Close()
	s.wg.Wait()
}

func newStopLossSubscription(sub *sse.Subscription) *StopLossSubscription {
	s := &StopLossSubscription{
		sub:    sub,
		events: make(chan StopLossEvent, 100),
		errors: make(chan error, 10),
		done:   make(chan struct{}),
	}
	s.wg.Add(1)
	go s.run()
	return s
}

func (s *StopLossSubscription) run() {
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

func (s *StopLossSubscription) forwardEvent(raw sse.RawEvent) {
	if raw.Event != "STOPLOSS" {
		s.trySendEvent(StopLossEvent{Event: raw.Event, ID: raw.ID, Retry: raw.Retry})
		return
	}

	var data StopLossEventData
	if err := json.Unmarshal(raw.Data, &data); err != nil {
		select {
		case s.errors <- fmt.Errorf("parse stop loss data: %w", err):
		default:
		}
		return
	}

	s.trySendEvent(StopLossEvent{
		Event: raw.Event,
		Data:  data,
		ID:    raw.ID,
		Retry: raw.Retry,
	})
}

// trySendEvent sends without blocking Close: if the consumer has stopped
// reading and the subscription is being closed, the event is dropped.
func (s *StopLossSubscription) trySendEvent(event StopLossEvent) {
	select {
	case s.events <- event:
	case <-s.done:
	}
}
