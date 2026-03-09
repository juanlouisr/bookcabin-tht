package ratelimit

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestTokenBucket_Wait_Success(t *testing.T) {
	tb := NewTokenBucket(10, 1) // 10 tokens per second, burst of 1

	ctx := context.Background()
	err := tb.Wait(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestTokenBucket_Wait_ContextCancellation(t *testing.T) {
	tb := NewTokenBucket(0.1, 1) // 0.1 tokens per second (slow)

	// Consume the initial token
	tb.Wait(context.Background())

	// Next wait should take 10 seconds, but we cancel after 50ms
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := tb.Wait(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got: %v", err)
	}
}

func TestTokenBucket_Allow_Immediate(t *testing.T) {
	tb := NewTokenBucket(10, 10) // 10 tokens per second, burst of 10

	// First 10 calls should succeed immediately
	for i := 0; i < 10; i++ {
		if !tb.Allow() {
			t.Errorf("Expected Allow() to return true on call %d", i+1)
		}
	}

	// 11th call should fail (no tokens left)
	if tb.Allow() {
		t.Error("Expected Allow() to return false after consuming all tokens")
	}
}

func TestTokenBucket_Allow_Refill(t *testing.T) {
	tb := NewTokenBucket(100, 1) // 100 tokens per second, burst of 1

	// Consume the token
	if !tb.Allow() {
		t.Error("Expected Allow() to return true initially")
	}

	// Should fail immediately
	if tb.Allow() {
		t.Error("Expected Allow() to return false immediately after consumption")
	}

	// Wait for refill
	time.Sleep(20 * time.Millisecond)

	// Should succeed after refill
	if !tb.Allow() {
		t.Error("Expected Allow() to return true after refill")
	}
}

func TestTokenBucket_ConcurrentAccess(t *testing.T) {
	tb := NewTokenBucket(1000, 100) // 1000 tokens per second, burst of 100

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if tb.Allow() {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// All 50 should succeed since burst is 100
	if successCount != 50 {
		t.Errorf("Expected 50 successful allows, got %d", successCount)
	}
}

func TestTokenBucket_UpdateTokens(t *testing.T) {
	tb := NewTokenBucket(10, 5) // 10 tokens per second, burst of 5

	// Consume all tokens
	for i := 0; i < 5; i++ {
		tb.Allow()
	}

	if tb.tokens != 0 {
		t.Errorf("Expected 0 tokens, got %f", tb.tokens)
	}

	// Wait for some tokens to refill
	time.Sleep(200 * time.Millisecond)

	// Update tokens manually (normally done in Allow/Wait)
	tb.updateTokens()

	// Should have some tokens now (2 tokens = 10 * 0.2 seconds)
	if tb.tokens < 1 {
		t.Errorf("Expected at least 1 token after refill, got %f", tb.tokens)
	}

	// Should not exceed burst
	if tb.tokens > float64(tb.burst) {
		t.Errorf("Token count %f exceeds burst %d", tb.tokens, tb.burst)
	}
}

func TestUnlimited(t *testing.T) {
	ul := Unlimited()

	ctx := context.Background()
	if err := ul.Wait(ctx); err != nil {
		t.Errorf("Expected no error from unlimited limiter, got: %v", err)
	}

	if !ul.Allow() {
		t.Error("Expected Allow() to always return true for unlimited")
	}

	// Should work many times without blocking
	for i := 0; i < 1000; i++ {
		if !ul.Allow() {
			t.Error("Expected Allow() to always return true")
		}
	}
}

func TestTokenBucket_RateLimiting(t *testing.T) {
	tb := NewTokenBucket(10, 1) // 10 tokens per second, burst of 1

	start := time.Now()

	// First call should be immediate
	tb.Wait(context.Background())

	// Second call should wait ~100ms for token refill
	tb.Wait(context.Background())

	elapsed := time.Since(start)

	// Should have waited at least 80ms (allowing for some timing variance)
	if elapsed < 80*time.Millisecond {
		t.Errorf("Expected to wait at least 80ms, but only waited %v", elapsed)
	}
}

func TestTokenBucket_ZeroRate(t *testing.T) {
	tb := NewTokenBucket(0, 0) // 0 tokens per second

	if tb.Allow() {
		t.Error("Expected Allow() to return false with zero rate")
	}
}

func TestTokenBucket_BurstSize(t *testing.T) {
	// Test various burst sizes
	burstSizes := []int{1, 5, 10, 100}

	for _, burst := range burstSizes {
		tb := NewTokenBucket(1000, burst)

		// Consume all burst tokens
		count := 0
		for tb.Allow() {
			count++
		}

		if count != burst {
			t.Errorf("Expected %d tokens for burst %d, got %d", burst, burst, count)
		}
	}
}
