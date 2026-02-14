package retries_backoff

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDemoOutput(t *testing.T) {
	attempt := 0
	cfg := Config{
		MaxRetries: 3,
		Initial:    5 * time.Millisecond,
		Max:        20 * time.Millisecond,
		Multiplier: 2,
		Jitter:     0,
	}
	err := Retry(context.Background(), cfg, func() error {
		attempt++
		t.Logf("attempt=%d backoff(next)=%v", attempt, NextBackoff(cfg, attempt-1))
		if attempt < 3 {
			return errors.New("temporary")
		}
		return nil
	})
	t.Logf("final err=%v attempts=%d", err, attempt)
}


