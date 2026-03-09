package transport

import (
	"context"
	"time"
)

// Transport defines the interface for raw data transport
// This abstraction allows supporting HTTP, SOAP, gRPC, or any other protocol
// The transport layer only returns raw bytes, parsing is done by the provider
type Transport interface {
	// Fetch retrieves raw data from the provider's endpoint
	// Returns raw bytes that the provider will parse
	Fetch(ctx context.Context, request Request) (*Response, error)

	// GetTimeout returns the configured timeout for this transport
	GetTimeout() time.Duration
}

// Request represents a transport request
type Request struct {
	// Origin airport code
	Origin string
	// Destination airport code
	Destination string
	// DepartureDate in YYYY-MM-DD format
	DepartureDate string
	// ReturnDate in YYYY-MM-DD format (optional)
	ReturnDate *string
	// Number of passengers
	Passengers int
	// Cabin class (economy, business, first)
	CabinClass string
}

// Response represents a transport response
type Response struct {
	// Raw response body bytes
	Body []byte
	// Status code (if applicable)
	StatusCode int
	// Headers (if applicable)
	Headers map[string]string
}
