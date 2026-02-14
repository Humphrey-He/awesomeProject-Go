package retries_backoff

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetry_SuccessAfterRetries(t *testing.T) {
	attempts := 0
	cfg := Config{
		MaxRetries: 3,
		Initial:    1 * time.Millisecond,
		Max:        5 * time.Millisecond,
		Multiplier: 2,
		Jitter:     0,
	}

	err := Retry(context.Background(), cfg, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("retry should eventually succeed: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("want 3 attempts, got %d", attempts)
	}
}

func TestRetry_FailAfterMaxRetries(t *testing.T) {
	attempts := 0
	cfg := Config{
		MaxRetries: 2,
		Initial:    1 * time.Millisecond,
		Max:        5 * time.Millisecond,
		Multiplier: 2,
		Jitter:     0,
	}

	err := Retry(context.Background(), cfg, func() error {
		attempts++
		return errors.New("always fail")
	})
	if err == nil {
		t.Fatal("want failure after max retries")
	}
	if attempts != 3 {
		t.Fatalf("want 3 attempts (1 + 2 retries), got %d", attempts)
	}
}

func TestRetry_ContextCancel(t *testing.T) {
	cfg := Config{
		MaxRetries: 10,
		Initial:    30 * time.Millisecond,
		Max:        30 * time.Millisecond,
		Multiplier: 2,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	err := Retry(ctx, cfg, func() error {
		return errors.New("fail")
	})
	if err == nil {
		t.Fatal("want context canceled/deadline exceeded")
	}
}

func TestNextBackoff_WithJitterRange(t *testing.T) {
	cfg := Config{
		MaxRetries: 3,
		Initial:    100 * time.Millisecond,
		Max:        500 * time.Millisecond,
		Multiplier: 2,
		Jitter:     0.2,
		RandFloat:  func() float64 { return 1 }, // max positive jitter
	}
	d := NextBackoff(cfg, 1) // base 200ms
	if d < 240*time.Millisecond || d > 240*time.Millisecond {
		t.Fatalf("want 240ms with deterministic jitter, got %v", d)
	}
}

func TestRetryValue(t *testing.T) {
	n := 0
	v, err := RetryValue(context.Background(), Config{
		MaxRetries: 3,
		Initial:    1 * time.Millisecond,
		Max:        3 * time.Millisecond,
		Multiplier: 2,
	}, func() (int, error) {
		n++
		if n < 2 {
			return 0, errors.New("temporary")
		}
		return 42, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 42 {
		t.Fatalf("want 42, got %d", v)
	}
}


