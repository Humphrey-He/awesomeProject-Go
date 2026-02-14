package leaky_bucket

import (
	"testing"
	"time"
)

func TestDemoOutput(t *testing.T) {
	lb := NewLeakyBucket(5, 10)
	t.Logf("init: capacity=%d leakRate=%d water=%d", lb.Capacity(), lb.LeakRate(), lb.CurrentWater())

	for i := 0; i < 6; i++ {
		allowed := lb.Allow()
		t.Logf("request #%d allow=%v water=%d wait=%v", i+1, allowed, lb.CurrentWater(), lb.WaitTime())
	}

	time.Sleep(200 * time.Millisecond)
	t.Logf("after 200ms leak water=%d available=%d", lb.CurrentWater(), lb.AvailableSpace())
}
