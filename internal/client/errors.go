// Package client provides HTTP client functionality for the Avanza API.
package client

import (
	"fmt"
	"io"
	"net/http"
)

const (
	maxErrorBodySize = 1024 // 1KB
)

// HTTPError represents an HTTP error response from the Avanza API.
// It includes the status code and response body for debugging.
//
// Users can check for HTTPError using errors.As:
//
//	var httpErr *client.HTTPError
//	if errors.As(err, &httpErr) {
//	    fmt.Printf("Status: %d, Body: %s\n", httpErr.StatusCode, httpErr.Body)
//	}
type HTTPError struct {
	StatusCode int
	Body       string
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	if e.Body != "" {
		return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
	}
	return fmt.Sprintf("HTTP %d", e.StatusCode)
}

// NewHTTPError creates an HTTPError from an HTTP response.
// It reads the response body to include in the error message, limiting
// the size to prevent memory exhaustion.
func NewHTTPError(resp *http.Response) *HTTPError {
	limitedReader := io.LimitReader(resp.Body, maxErrorBodySize)
	body, _ := io.ReadAll(limitedReader)
	return &HTTPError{
		StatusCode: resp.StatusCode,
		Body:       string(body),
	}
}
