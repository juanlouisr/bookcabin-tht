package transport

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// MockTransport simulates a transport layer for testing
// It adds configurable delays and failure rates
type MockTransport struct {
	DataLoader  DataLoader
	MinDelayMs  int
	MaxDelayMs  int
	SuccessRate float64 // 0.0 to 1.0
}

// NewMockTransport creates a new mock transport with a data loader
func NewMockTransport(dataLoader DataLoader, minDelayMs, maxDelayMs int, successRate float64) *MockTransport {
	return &MockTransport{
		DataLoader:  dataLoader,
		MinDelayMs:  minDelayMs,
		MaxDelayMs:  maxDelayMs,
		SuccessRate: successRate,
	}
}

// NewMockTransportFromFile creates a new mock transport that loads data from a file
func NewMockTransportFromFile(filePath string, minDelayMs, maxDelayMs int, successRate float64) *MockTransport {
	return NewMockTransport(
		NewFileDataLoader(filePath),
		minDelayMs,
		maxDelayMs,
		successRate,
	)
}

// NewMockTransportFromBytes creates a new mock transport with in-memory data
func NewMockTransportFromBytes(data []byte, minDelayMs, maxDelayMs int, successRate float64) *MockTransport {
	return NewMockTransport(
		NewBytesDataLoader(data),
		minDelayMs,
		maxDelayMs,
		successRate,
	)
}

// Fetch retrieves data using the data loader and simulates network behavior
func (t *MockTransport) Fetch(ctx context.Context, req Request) (*Response, error) {
	// Simulate network delay
	delay := t.MinDelayMs + rand.Intn(t.MaxDelayMs-t.MinDelayMs+1)

	select {
	case <-time.After(time.Duration(delay) * time.Millisecond):
		// Continue after delay
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Simulate random failure
	if rand.Float64() > t.SuccessRate {
		return nil, fmt.Errorf("simulated transport failure")
	}

	// Load data using the data loader
	data, err := t.DataLoader.Load()
	if err != nil {
		return nil, err
	}

	return &Response{
		Body:       data,
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

// GetTimeout returns the configured timeout
func (t *MockTransport) GetTimeout() time.Duration {
	return 5 * time.Second
}
