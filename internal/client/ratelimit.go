// Package client provides HTTP client functionality for the Avanza API.
package client

import (
	"context"
	"sync"
	"time"
)

const (
	// DefaultRateLimitInterval is the default minimum interval between HTTP requests.
	// This helps prevent overwhelming the API and reduces the risk of being blocked.
	DefaultRateLimitInterval = 100 * time.Millisecond
)

// RateLimiter controls the rate of HTTP requests.
// Implementations should block until the next request is allowed.
type RateLimiter interface {
	// Wait blocks until the rate limiter allows a request to proceed.
	// Returns an error if the context is cancelled.
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
	r.mu.Unlock()

	if elapsed < r.Interval {
		waitTime := r.Interval - elapsed
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
		}
	}

	r.mu.Lock()
	r.lastCall = time.Now()
	r.mu.Unlock()
	return nil
}
