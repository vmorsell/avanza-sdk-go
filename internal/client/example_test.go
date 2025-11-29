package client_test

import (
	"errors"
	"fmt"
	"time"

	"github.com/vmorsell/avanza-sdk-go/internal/client"
)

func ExampleNewClient() {
	c := client.NewClient()
	_ = c
}

func ExampleWithRateLimiter() {
	limiter := &client.SimpleRateLimiter{
		Interval: 200 * time.Millisecond,
	}
	c := client.NewClient(client.WithRateLimiter(limiter))
	_ = c
}

func ExampleWithBaseURL() {
	c := client.NewClient(client.WithBaseURL("http://localhost:8080"))
	_ = c
}

func ExampleWithUserAgent() {
	c := client.NewClient(client.WithUserAgent("MyApp/1.0"))
	_ = c
}

func ExampleHTTPError() {
	// In real usage, this would come from an HTTP response
	// This example shows how to check for HTTPError types

	var err error
	// ... some operation that returns an error ...

	var httpErr *client.HTTPError
	if errors.As(err, &httpErr) {
		fmt.Printf("HTTP %d: %s\n", httpErr.StatusCode, httpErr.Body)
	}
}
