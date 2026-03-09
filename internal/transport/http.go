package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPTransport implements the Transport interface for HTTP/REST APIs
type HTTPTransport struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	Timeout    time.Duration
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(baseURL, apiKey string, timeout time.Duration) *HTTPTransport {
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	return &HTTPTransport{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		Timeout: timeout,
	}
}

// Fetch makes an HTTP request and returns raw bytes
func (t *HTTPTransport) Fetch(ctx context.Context, req Request) (*Response, error) {
	// Build request URL
	url := fmt.Sprintf("%s/api/flights/search", t.BaseURL)

	// Build request body
	reqBody, err := json.Marshal(map[string]interface{}{
		"origin":         req.Origin,
		"destination":    req.Destination,
		"departure_date": req.DepartureDate,
		"passengers":     req.Passengers,
		"cabin_class":    req.CabinClass,
	})
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

	return &Response{
		Body:       body,
		StatusCode: resp.StatusCode,
		Headers:    getHeaders(resp.Header),
	}, nil
}

// GetTimeout returns the configured timeout
func (t *HTTPTransport) GetTimeout() time.Duration {
	return t.Timeout
}

func getHeaders(header http.Header) map[string]string {
	headers := make(map[string]string)
	for key, values := range header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return headers
}
