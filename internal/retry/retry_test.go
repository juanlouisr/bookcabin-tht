package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDo_Success(t *testing.T) {
	callCount := 0
	fn := func() error {
		callCount++
		return nil
	}

	config := DefaultConfig()
	config.MaxRetries = 3

	ctx := context.Background()
	err := Do(ctx, fn, nil, config)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestDo_MaxRetriesExceeded(t *testing.T) {
	callCount := 0
	testErr := errors.New("test error")
	fn := func() error {
		callCount++
		return testErr
	}

	config := DefaultConfig()
	config.MaxRetries = 2
	config.InitialInterval = 10 * time.Millisecond

	ctx := context.Background()
	err := Do(ctx, fn, nil, config)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	// Should be called MaxRetries + 1 times (initial + retries)
	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
}

func TestDo_ContextCancellation(t *testing.T) {
	callCount := 0
	fn := func() error {
		callCount++
		return errors.New("test error")
	}

	config := DefaultConfig()
	config.MaxRetries = 10
	config.InitialInterval = 1 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := Do(ctx, fn, nil, config)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded, got: %v", err)
	}
}

func TestDo_NonRetryableError(t *testing.T) {
	callCount := 0
	nonRetryableErr := errors.New("non-retryable error")
	fn := func() error {
		callCount++
		return nonRetryableErr
	}

	isRetryable := func(err error) bool {
		return err != nonRetryableErr
	}

	config := DefaultConfig()
	config.MaxRetries = 3

	ctx := context.Background()
	err := Do(ctx, fn, isRetryable, config)

	if err != nonRetryableErr {
		t.Errorf("Expected nonRetryableErr, got: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call (non-retryable), got %d", callCount)
	}
}

func TestDo_RetryableThenSuccess(t *testing.T) {
	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary error")
		}
		return nil
	}

	config := DefaultConfig()
	config.MaxRetries = 5
	config.InitialInterval = 10 * time.Millisecond

	ctx := context.Background()
	err := Do(ctx, fn, nil, config)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
}

func TestCalculateBackoff(t *testing.T) {
	config := DefaultConfig()
	config.InitialInterval = 100 * time.Millisecond
	config.Multiplier = 2.0
	config.MaxInterval = 1 * time.Second
	config.RandomizationFactor = 0.0 // Disable jitter for predictable tests

	tests := []struct {
		attempt     int
		minExpected time.Duration
		maxExpected time.Duration
	}{
		{0, 100 * time.Millisecond, 100 * time.Millisecond},
		{1, 200 * time.Millisecond, 200 * time.Millisecond},
		{2, 400 * time.Millisecond, 400 * time.Millisecond},
		{3, 800 * time.Millisecond, 800 * time.Millisecond},
		{4, 1000 * time.Millisecond, 1000 * time.Millisecond},  // Capped at max
		{10, 1000 * time.Millisecond, 1000 * time.Millisecond}, // Capped at max
	}

	for _, test := range tests {
		backoff := calculateBackoff(test.attempt, config)
		if backoff < test.minExpected || backoff > test.maxExpected {
			t.Errorf("Attempt %d: backoff %v not in range [%v, %v]",
				test.attempt, backoff, test.minExpected, test.maxExpected)
		}
	}
}

func TestCalculateBackoff_NegativeAttempt(t *testing.T) {
	config := DefaultConfig()
	config.InitialInterval = 100 * time.Millisecond
	config.RandomizationFactor = 0.0

	backoff := calculateBackoff(-1, config)
	if backoff != 100*time.Millisecond {
		t.Errorf("Expected 100ms for negative attempt, got %v", backoff)
	}
}

func TestCalculateBackoff_WithJitter(t *testing.T) {
	config := DefaultConfig()
	config.InitialInterval = 100 * time.Millisecond
	config.Multiplier = 2.0
	config.RandomizationFactor = 0.1 // 10% jitter

	// Run multiple times to account for randomness
	for i := 0; i < 10; i++ {
		backoff := calculateBackoff(1, config)
		// With 10% jitter on 200ms: range is 180ms to 220ms
		minExpected := 180 * time.Millisecond
		maxExpected := 220 * time.Millisecond

		if backoff < minExpected || backoff > maxExpected {
			t.Errorf("Backoff %v not in expected range [%v, %v]",
				backoff, minExpected, maxExpected)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries 3, got %d", config.MaxRetries)
	}
	if config.InitialInterval != 100*time.Millisecond {
		t.Errorf("Expected InitialInterval 100ms, got %v", config.InitialInterval)
	}
	if config.MaxInterval != 30*time.Second {
		t.Errorf("Expected MaxInterval 30s, got %v", config.MaxInterval)
	}
	if config.Multiplier != 2.0 {
		t.Errorf("Expected Multiplier 2.0, got %f", config.Multiplier)
	}
	if config.RandomizationFactor != 0.1 {
		t.Errorf("Expected RandomizationFactor 0.1, got %f", config.RandomizationFactor)
	}
}

func TestDo_ZeroMaxRetries(t *testing.T) {
	callCount := 0
	fn := func() error {
		callCount++
		return errors.New("error")
	}

	config := DefaultConfig()
	config.MaxRetries = 0
	config.InitialInterval = 10 * time.Millisecond

	ctx := context.Background()
	err := Do(ctx, fn, nil, config)

	if err == nil {
		t.Error("Expected error")
	}
	// Should be called once (initial attempt only, no retries)
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestIsRetryableFunc(t *testing.T) {
	retryableErr := errors.New("retryable")
	nonRetryableErr := errors.New("non-retryable")

	isRetryable := func(err error) bool {
		return err == retryableErr
	}

	callCount := 0
	fn := func() error {
		callCount++
		if callCount == 1 {
			return retryableErr
		}
		return nonRetryableErr
	}

	config := DefaultConfig()
	config.MaxRetries = 5
	config.InitialInterval = 10 * time.Millisecond

	ctx := context.Background()
	err := Do(ctx, fn, isRetryable, config)

	if err != nonRetryableErr {
		t.Errorf("Expected nonRetryableErr, got: %v", err)
	}
	// Should stop after non-retryable error
	if callCount != 2 {
		t.Errorf("Expected 2 calls, got %d", callCount)
	}
}
