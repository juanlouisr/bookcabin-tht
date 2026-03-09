package models

import "context"

// Provider defines the interface for flight data providers
type Provider interface {
	// Name returns the provider name
	Name() string

	// Search fetches flights from the provider
	Search(ctx context.Context, request SearchRequest) ([]UnifiedFlight, error)
}

// ProviderResult wraps the result from a provider call
type ProviderResult struct {
	Provider string
	Flights  []UnifiedFlight
	Error    error
}

// ProviderConfig holds configuration for a provider
type ProviderConfig struct {
	Name        string
	MinDelayMs  int
	MaxDelayMs  int
	SuccessRate float64
}
