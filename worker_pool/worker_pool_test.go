package worker_pool

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestPool_Basic(t *testing.T) {
	p := New(3, 10)
	defer p.Stop()

	ch, err := p.Submit(context.Background(), func(ctx context.Context) (any, error) {
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	res := <-ch
	if res.Err != nil {
		t.Fatalf("task error: %v", res.Err)
	}
	if res.Value != "ok" {
		t.Fatalf("want ok, got %v", res.Value)
	}
}

func TestPool_ConcurrencyLimit(t *testing.T) {
	workers := 4
	p := New(workers, 100)
	defer p.Stop()

	var current int64
	var maxSeen int64

	total := 40
	results := make([]<-chan Result, 0, total)
	for i := 0; i < total; i++ {
		ch, err := p.Submit(context.Background(), func(ctx context.Context) (any, error) {
			v := atomic.AddInt64(&current, 1)
			for {
				old := atomic.LoadInt64(&maxSeen)
				if v <= old || atomic.CompareAndSwapInt64(&maxSeen, old, v) {
					break
				}
			}
			time.Sleep(15 * time.Millisecond)
			atomic.AddInt64(&current, -1)
			return 1, nil
		})
		if err != nil {
			t.Fatalf("submit failed: %v", err)
		}
		results = append(results, ch)
	}

	for _, ch := range results {
		res := <-ch
		if res.Err != nil {
			t.Fatalf("task error: %v", res.Err)
		}
	}

	if got := atomic.LoadInt64(&maxSeen); got > int64(workers) {
		t.Fatalf("max concurrency = %d, workers = %d", got, workers)
	}
}

func TestPool_ShutdownRejectsNewTask(t *testing.T) {
	p := New(2, 2)
	err := p.Shutdown(context.Background())
	if err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}

	_, err = p.Submit(context.Background(), func(ctx context.Context) (any, error) {
		return nil, nil
	})
	if err == nil {
		t.Fatal("want submit error after shutdown")
	}
}

func TestPool_StopCancelsRunningTask(t *testing.T) {
	p := New(1, 1)

	ch, err := p.Submit(context.Background(), func(ctx context.Context) (any, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
			return "done", nil
		}
	})
	if err != nil {
		t.Fatalf("submit failed: %v", err)
	}

	time.Sleep(30 * time.Millisecond)
	p.Stop()

	res := <-ch
	if res.Err == nil {
		t.Fatal("want canceled error after stop")
	}
}

func TestPool_SubmitWithCanceledContext(t *testing.T) {
	p := New(1, 1)
	defer p.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := p.Submit(ctx, func(ctx context.Context) (any, error) {
		return nil, nil
	})
	if err == nil {
		t.Fatal("want submit canceled error")
	}
}

