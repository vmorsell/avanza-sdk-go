package client

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestSimpleRateLimiter_EnforcesInterval(t *testing.T) {
	limiter := &SimpleRateLimiter{Interval: 50 * time.Millisecond}
	ctx := context.Background()

	start := time.Now()

	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("first wait: %v", err)
	}

	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("second wait: %v", err)
	}

	elapsed := time.Since(start)

	// Second call should have waited ~50ms
	if elapsed < 40*time.Millisecond {
		t.Errorf("expected at least 40ms between calls, got %v", elapsed)
	}
}

func TestSimpleRateLimiter_FirstCallImmediate(t *testing.T) {
	limiter := &SimpleRateLimiter{Interval: 1 * time.Second}
	ctx := context.Background()

	start := time.Now()
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("wait: %v", err)
	}
	elapsed := time.Since(start)

	// First call should not wait. Use generous bound for CI under load.
	if elapsed > 50*time.Millisecond {
		t.Errorf("first call should be immediate, took %v", elapsed)
	}
}

func TestSimpleRateLimiter_ContextCancellation(t *testing.T) {
	limiter := &SimpleRateLimiter{Interval: 1 * time.Second}
	ctx := context.Background()

	// First call to set lastCall
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("first wait: %v", err)
	}

	// Cancel context before second call can complete
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()

	err := limiter.Wait(cancelCtx)
	if err == nil {
		t.Fatal("expected context cancelled error, got nil")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestSimpleRateLimiter_ConcurrentAccess(t *testing.T) {
	limiter := &SimpleRateLimiter{Interval: 10 * time.Millisecond}
	ctx := context.Background()

	var wg sync.WaitGroup
	const goroutines = 10

	start := time.Now()

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := limiter.Wait(ctx); err != nil {
				t.Errorf("wait: %v", err)
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	// With 10 goroutines and 10ms interval, should take at least ~90ms
	// (first is immediate, 9 more need to wait).
	// Use a generous lower bound to avoid flakiness.
	if elapsed < 50*time.Millisecond {
		t.Errorf("concurrent calls should be spread over time, total was only %v", elapsed)
	}
}

func TestSimpleRateLimiter_NoRaceBetweenCalls(t *testing.T) {
	// This test primarily exercises the race detector.
	limiter := &SimpleRateLimiter{Interval: 1 * time.Millisecond}
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = limiter.Wait(ctx)
		}()
	}
	wg.Wait()
}

func TestSimpleRateLimiter_ContextTimeout(t *testing.T) {
	limiter := &SimpleRateLimiter{Interval: 5 * time.Second}
	ctx := context.Background()

	// First call to set lastCall
	if err := limiter.Wait(ctx); err != nil {
		t.Fatalf("first wait: %v", err)
	}

	// Second call with short timeout should fail
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := limiter.Wait(timeoutCtx)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}
