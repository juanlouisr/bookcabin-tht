package ratelimit

import (
	"context"
	"sync"
	"time"
)

// RateLimiter defines the interface for rate limiting
type RateLimiter interface {
	// Wait blocks until the rate limiter allows the next request
	Wait(ctx context.Context) error
	// Allow returns true if the request is allowed immediately
	Allow() bool
}

// TokenBucket implements a token bucket rate limiter
type TokenBucket struct {
	rate       float64   // tokens per second
	burst      int       // maximum tokens in bucket
	tokens     float64   // current tokens
	lastUpdate time.Time // last time tokens were updated
	mu         sync.Mutex
}

// NewTokenBucket creates a new token bucket rate limiter
// rate: tokens per second (requests per second)
// burst: maximum number of tokens (burst capacity)
func NewTokenBucket(rate float64, burst int) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		burst:      burst,
		tokens:     float64(burst),
		lastUpdate: time.Now(),
	}
}

// Wait blocks until a token is available or context is cancelled
func (tb *TokenBucket) Wait(ctx context.Context) error {
	for {
		tb.mu.Lock()
		tb.updateTokens()

		if tb.tokens >= 1 {
			tb.tokens--
			tb.mu.Unlock()
			return nil
		}

		// Calculate wait time for next token
		waitTime := time.Duration((1 - tb.tokens) / tb.rate * float64(time.Second))
		tb.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Try again
		}
	}
}

// Allow returns true if a token is available immediately
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.updateTokens()

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

// updateTokens adds tokens based on time elapsed
func (tb *TokenBucket) updateTokens() {
	now := time.Now()
	elapsed := now.Sub(tb.lastUpdate).Seconds()
	tb.lastUpdate = now

	tb.tokens += elapsed * tb.rate
	if tb.tokens > float64(tb.burst) {
		tb.tokens = float64(tb.burst)
	}
}

// Unlimited returns a rate limiter that never limits
func Unlimited() RateLimiter {
	return &unlimitedLimiter{}
}

type unlimitedLimiter struct{}

func (u *unlimitedLimiter) Wait(ctx context.Context) error { return nil }
func (u *unlimitedLimiter) Allow() bool                    { return true }
