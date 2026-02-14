package retries_backoff

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

var ErrMaxRetriesExceeded = errors.New("max retries exceeded")

// Config controls retry and backoff behavior.
type Config struct {
	MaxRetries int
	Initial    time.Duration
	Max        time.Duration
	Multiplier float64
	Jitter     float64 // 0..1, percentage around base duration

	// RandFloat returns [0,1). Optional, defaults to math/rand.
	RandFloat func() float64
}

func (c Config) normalize() Config {
	if c.MaxRetries < 0 {
		c.MaxRetries = 0
	}
	if c.Initial <= 0 {
		c.Initial = 50 * time.Millisecond
	}
	if c.Max <= 0 {
		c.Max = 2 * time.Second
	}
	if c.Multiplier < 1 {
		c.Multiplier = 2
	}
	if c.Jitter < 0 {
		c.Jitter = 0
	}
	if c.Jitter > 1 {
		c.Jitter = 1
	}
	if c.RandFloat == nil {
		c.RandFloat = rand.Float64
	}
	return c
}

// Retry executes op with exponential backoff and jitter.
func Retry(ctx context.Context, cfg Config, op func() error) error {
	cfg = cfg.normalize()
	var lastErr error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := op(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if attempt == cfg.MaxRetries {
			break
		}

		wait := NextBackoff(cfg, attempt)
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}

	if lastErr == nil {
		return ErrMaxRetriesExceeded
	}
	return lastErr
}

// RetryValue retries operation returning value.
func RetryValue[T any](ctx context.Context, cfg Config, op func() (T, error)) (T, error) {
	var zero T
	var val T
	err := Retry(ctx, cfg, func() error {
		v, err := op()
		if err != nil {
			return err
		}
		val = v
		return nil
	})
	if err != nil {
		return zero, err
	}
	return val, nil
}

// NextBackoff computes delay for given attempt (0-based).
func NextBackoff(cfg Config, attempt int) time.Duration {
	cfg = cfg.normalize()

	base := float64(cfg.Initial)
	for i := 0; i < attempt; i++ {
		base *= cfg.Multiplier
		if base >= float64(cfg.Max) {
			base = float64(cfg.Max)
			break
		}
	}

	// jitter in [base*(1-j), base*(1+j)]
	if cfg.Jitter > 0 {
		j := cfg.Jitter
		f := cfg.RandFloat()*2 - 1 // [-1, 1)
		base = base * (1 + j*f)
	}

	if base < 0 {
		base = 0
	}
	if base > float64(cfg.Max) {
		base = float64(cfg.Max)
	}
	return time.Duration(base)
}
