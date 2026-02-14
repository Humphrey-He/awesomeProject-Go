package resilience

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(100, 2)
	defer rl.Stop()

	if !rl.Allow() || !rl.Allow() {
		t.Fatal("initial burst tokens should be available")
	}
	if rl.Allow() {
		t.Fatal("third allow should be blocked by burst=2")
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	rl := NewRateLimiter(5, 1)
	defer rl.Stop()

	if !rl.Allow() {
		t.Fatal("first token should be available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	if err := rl.Wait(ctx); err != nil {
		t.Fatalf("wait should succeed after refill: %v", err)
	}
}

func TestCircuitBreaker_OpenAndRecover(t *testing.T) {
	cb := NewCircuitBreaker(2, 80*time.Millisecond, 1)

	_ = cb.Execute(func() error { return errors.New("x") })
	_ = cb.Execute(func() error { return errors.New("x") })

	if cb.State() != StateOpen {
		t.Fatalf("want open state, got %v", cb.State())
	}

	err := cb.Execute(func() error { return nil })
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("want ErrCircuitOpen, got %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	if cb.State() != StateHalfOpen {
		t.Fatalf("want half-open after timeout, got %v", cb.State())
	}

	err = cb.Execute(func() error { return nil })
	if err != nil {
		t.Fatalf("half-open request should pass: %v", err)
	}
	if cb.State() != StateClosed {
		t.Fatalf("want closed after successful half-open probe, got %v", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenFailureBackToOpen(t *testing.T) {
	cb := NewCircuitBreaker(1, 30*time.Millisecond, 1)
	_ = cb.Execute(func() error { return errors.New("fail") })
	time.Sleep(35 * time.Millisecond)
	if cb.State() != StateHalfOpen {
		t.Fatalf("want half-open, got %v", cb.State())
	}

	_ = cb.Execute(func() error { return errors.New("fail-again") })
	if cb.State() != StateOpen {
		t.Fatalf("want open after failure in half-open, got %v", cb.State())
	}
}


