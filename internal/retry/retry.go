package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RetryableFunc is the function to retry
type RetryableFunc func() error

// IsRetryableFunc determines if an error is retryable
type IsRetryableFunc func(error) bool

// Config holds retry configuration
type Config struct {
	MaxRetries          int
	InitialInterval     time.Duration
	MaxInterval         time.Duration
	Multiplier          float64
	RandomizationFactor float64
}

// DefaultConfig returns default retry configuration
func DefaultConfig() Config {
	return Config{
		MaxRetries:          3,
		InitialInterval:     100 * time.Millisecond,
		MaxInterval:         30 * time.Second,
		Multiplier:          2.0,
		RandomizationFactor: 0.1,
	}
}

// Do executes the function with exponential backoff retry
func Do(ctx context.Context, fn RetryableFunc, isRetryable IsRetryableFunc, config Config) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Check context before attempting
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is retryable
		if isRetryable != nil && !isRetryable(err) {
			return err // Not retryable, return immediately
		}

		// Don't retry after last attempt
		if attempt >= config.MaxRetries {
			break
		}

		// Calculate backoff duration
		backoff := calculateBackoff(attempt, config)

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("max retries exceeded (%d): %w", config.MaxRetries, lastErr)
}

// calculateBackoff calculates the backoff duration with exponential backoff and jitter
func calculateBackoff(attempt int, config Config) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	// Calculate exponential backoff
	backoff := float64(config.InitialInterval) * math.Pow(config.Multiplier, float64(attempt))

	// Apply max interval cap
	if backoff > float64(config.MaxInterval) {
		backoff = float64(config.MaxInterval)
	}

	// Add randomization (jitter) to avoid thundering herd
	if config.RandomizationFactor > 0 {
		delta := backoff * config.RandomizationFactor
		min := backoff - delta
		max := backoff + delta
		backoff = min + rand.Float64()*(max-min)
	}

	return time.Duration(backoff)
}

// DefaultIsRetryable returns true for common retryable errors
func DefaultIsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for context errors (not retryable)
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}

	// Retry on any other error (network errors, timeouts, etc.)
	return true
}

// WithRetry wraps a function to add retry logic
func WithRetry(fn RetryableFunc, config Config) RetryableFunc {
	return func() error {
		return Do(context.Background(), fn, DefaultIsRetryable, config)
	}
}
