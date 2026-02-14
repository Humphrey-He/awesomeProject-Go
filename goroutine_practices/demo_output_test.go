package goroutine_practices

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestDemoOutput(t *testing.T) {
	pool := NewWorkerPool(2, 10)
	pool.Start()
	defer pool.Stop()

	var counter int32
	for i := 0; i < 5; i++ {
		_ = pool.Submit(func() {
			atomic.AddInt32(&counter, 1)
		})
	}
	time.Sleep(50 * time.Millisecond)
	t.Logf("worker pool completed=%d", atomic.LoadInt32(&counter))

	rl := NewRateLimiter(5)
	defer rl.Stop()
	t.Logf("rate allow1=%v allow2=%v allow3=%v", rl.Allow(), rl.Allow(), rl.Allow())
}


