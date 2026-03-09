package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"bookcabin/flight/internal/ratelimit"
	"bookcabin/flight/internal/retry"
)

// EnhancedHTTPTransport implements the Transport interface with rate limiting and retry logic
type EnhancedHTTPTransport struct {
	BaseURL     string
	APIKey      string
	HTTPClient  *http.Client
	Timeout     time.Duration
	RateLimiter *ratelimit.TokenBucket
	RetryConfig retry.Config
}

// NewEnhancedHTTPTransport creates a new enhanced HTTP transport with rate limiting and retry
func NewEnhancedHTTPTransport(baseURL, apiKey string, timeout time.Duration, rateLimit float64, burst int) *EnhancedHTTPTransport {
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	return &EnhancedHTTPTransport{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		Timeout:     timeout,
		RateLimiter: ratelimit.NewTokenBucket(rateLimit, burst),
		RetryConfig: retry.Config{
			MaxRetries:          3,
			InitialInterval:     500 * time.Millisecond,
			MaxInterval:         5 * time.Second,
			Multiplier:          2.0,
			RandomizationFactor: 0.1,
		},
	}
}

// Fetch makes an HTTP request with rate limiting and retry logic
func (t *EnhancedHTTPTransport) Fetch(ctx context.Context, req Request) (*Response, error) {
	// Wait for rate limiter
	if err := t.RateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Execute with retry
	var response *Response
	var lastErr error

	retryErr := retry.Do(
		ctx,
		func() error {
			resp, err := t.doRequest(ctx, req)
			if err != nil {
				lastErr = err
				return err
			}
			response = resp
			return nil
		},
		func(err error) bool {
			// Retry on network errors and 5xx server errors
			if err == nil {
				return false
			}
			// Check if it's a server error (5xx)
			if respErr, ok := err.(*httpResponseError); ok && respErr.statusCode >= 500 {
				return true
			}
			// Retry on context errors (except context.Canceled)
			if err == context.DeadlineExceeded {
				return true
			}
			return false
		},
		t.RetryConfig,
	)

	if retryErr != nil {
		return nil, fmt.Errorf("request failed after retries: %w", lastErr)
	}

	return response, nil
}

// httpResponseError represents an HTTP response error
type httpResponseError struct {
	statusCode int
	message    string
}

func (e *httpResponseError) Error() string {
	return fmt.Sprintf("HTTP error %d: %s", e.statusCode, e.message)
}

// doRequest makes a single HTTP request
func (t *EnhancedHTTPTransport) doRequest(ctx context.Context, req Request) (*Response, error) {
	// Build request URL
	url := fmt.Sprintf("%s/api/flights/search", t.BaseURL)

	// Build request body
	reqBodyMap := map[string]interface{}{
		"origin":         req.Origin,
		"destination":    req.Destination,
		"departure_date": req.DepartureDate,
		"passengers":     req.Passengers,
		"cabin_class":    req.CabinClass,
	}

	// Add return_date if present (round-trip support)
	if req.ReturnDate != nil && *req.ReturnDate != "" {
		reqBodyMap["return_date"] = *req.ReturnDate
	}

	reqBody, err := json.Marshal(reqBodyMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	if t.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+t.APIKey)
	}

	// Make the request
	resp, err := t.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP error status
	if resp.StatusCode >= 400 {
		return nil, &httpResponseError{
			statusCode: resp.StatusCode,
			message:    string(body),
		}
	}

	return &Response{
		Body:       body,
		StatusCode: resp.StatusCode,
		Headers:    getHeaders(resp.Header),
	}, nil
}

// GetTimeout returns the configured timeout
func (t *EnhancedHTTPTransport) GetTimeout() time.Duration {
	return t.Timeout
}

// SetRetryConfig allows updating retry configuration
func (t *EnhancedHTTPTransport) SetRetryConfig(config retry.Config) {
	t.RetryConfig = config
}
