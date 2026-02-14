package worker_pool

import (
	"context"
	"testing"
	"time"
)

func TestDemoOutput(t *testing.T) {
	p := New(2, 4)
	defer p.Stop()

	resCh, err := p.Submit(context.Background(), func(ctx context.Context) (any, error) {
		time.Sleep(20 * time.Millisecond)
		return "job-ok", nil
	})
	if err != nil {
		t.Fatalf("submit err: %v", err)
	}
	res := <-resCh
	t.Logf("worker=%d value=%v err=%v", res.WorkerID, res.Value, res.Err)
}
