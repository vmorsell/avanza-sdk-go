// Package client provides HTTP client functionality for the Avanza API.
package client

import (
	"context"
	"sync"
	"time"
)

const (
	DefaultRateLimitInterval = 100 * time.Millisecond
)

// RateLimiter is an interface for rate limiting HTTP requests.
// Implementations should block until the request is allowed to proceed.
type RateLimiter interface {
	// Wait blocks until the rate limiter allows a request to proceed.
	// It should respect context cancellation.
	Wait(ctx context.Context) error
}

// SimpleRateLimiter is a simple rate limiter that enforces a minimum interval between requests.
// It is safe for concurrent use.
type SimpleRateLimiter struct {
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
