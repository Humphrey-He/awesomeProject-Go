package token_bucket

import (
	"testing"
	"time"
)

func TestDemoOutput(t *testing.T) {
	tb := NewTokenBucket(5, 10)
	t.Logf("init: capacity=%d refillRate=%d tokens=%d", tb.Capacity(), tb.RefillRate(), tb.AvailableTokens())

	for i := 0; i < 6; i++ {
		allowed := tb.Allow()
		t.Logf("request #%d allow=%v tokens=%d", i+1, allowed, tb.AvailableTokens())
	}

	time.Sleep(200 * time.Millisecond)
	t.Logf("after 200ms refill tokens=%d", tb.AvailableTokens())
}
