// Package client provides HTTP client functionality for the Avanza API.
package client

import (
	"context"
	"sync"
	"time"
)

const (
	// DefaultRateLimitInterval is the default minimum interval between requests.
	DefaultRateLimitInterval = 100 * time.Millisecond
)

// RateLimiter controls request rate. Implementations block until the next request is allowed.
type RateLimiter interface {
	Wait(ctx context.Context) error
}

// SimpleRateLimiter enforces a minimum interval between requests.
// It is safe for concurrent use.
type SimpleRateLimiter struct {
	// Interval is the minimum time between requests.
	Interval time.Duration
	mu       sync.Mutex
	lastCall time.Time
}

// Wait blocks until the rate limiter allows a request to proceed.
func (r *SimpleRateLimiter) Wait(ctx context.Context) error {
	r.mu.Lock()
	now := time.Now()
	elapsed := now.Sub(r.lastCall)

	if elapsed >= r.Interval {
		r.lastCall = now
		r.mu.Unlock()
		return nil
	}

	waitTime := r.Interval - elapsed
	r.lastCall = now.Add(waitTime)
	r.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(waitTime):
		return nil
	}
}
