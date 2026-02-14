package resilience

import (
	"errors"
	"testing"
	"time"
)

func TestDemoOutput(t *testing.T) {
	rl := NewRateLimiter(5, 2)
	defer rl.Stop()
	t.Logf("ratelimiter allow1=%v allow2=%v allow3=%v", rl.Allow(), rl.Allow(), rl.Allow())

	cb := NewCircuitBreaker(2, 50*time.Millisecond, 1)
	_ = cb.Execute(func() error { return errors.New("fail1") })
	_ = cb.Execute(func() error { return errors.New("fail2") })
	t.Logf("circuit state after failures=%v", cb.State())
	time.Sleep(60 * time.Millisecond)
	t.Logf("circuit state after timeout=%v", cb.State())
}
