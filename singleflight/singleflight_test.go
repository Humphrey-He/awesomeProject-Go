package singleflight

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDo_DeduplicateConcurrentCalls(t *testing.T) {
	sf := New()
	var runCount int32

	const n = 20
	var wg sync.WaitGroup
	wg.Add(n)

	results := make([]interface{}, n)
	errs := make([]error, n)
	sharedFlags := make([]bool, n)

	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			v, err, shared := sf.Do("same-key", func() (interface{}, error) {
				atomic.AddInt32(&runCount, 1)
				time.Sleep(30 * time.Millisecond)
				return "ok", nil
			})
			results[idx] = v
			errs[idx] = err
			sharedFlags[idx] = shared
		}(i)
	}

	wg.Wait()

	if atomic.LoadInt32(&runCount) != 1 {
		t.Fatalf("expected fn run once, got %d", runCount)
	}
	for i := 0; i < n; i++ {
		if errs[i] != nil {
			t.Fatalf("unexpected err at idx=%d: %v", i, errs[i])
		}
		if results[i] != "ok" {
			t.Fatalf("unexpected value at idx=%d: %v", i, results[i])
		}
	}

	sharedCnt := 0
	for _, s := range sharedFlags {
		if s {
			sharedCnt++
		}
	}
	if sharedCnt == 0 {
		t.Fatalf("expected at least one shared caller")
	}
}

func TestDo_ErrorPropagationShared(t *testing.T) {
	sf := New()
	targetErr := errors.New("downstream failed")
	var runCount int32

	const n = 10
	var wg sync.WaitGroup
	wg.Add(n)
	errs := make([]error, n)

	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			_, err, _ := sf.Do("err-key", func() (interface{}, error) {
				atomic.AddInt32(&runCount, 1)
				time.Sleep(20 * time.Millisecond)
				return nil, targetErr
			})
			errs[idx] = err
		}(i)
	}
	wg.Wait()

	if runCount != 1 {
		t.Fatalf("expected fn run once, got %d", runCount)
	}
	for i, err := range errs {
		if !errors.Is(err, targetErr) {
			t.Fatalf("idx=%d expected targetErr, got %v", i, err)
		}
	}
}

func TestDoChan(t *testing.T) {
	sf := New()
	ch := sf.DoChan("ch-key", func() (interface{}, error) {
		time.Sleep(10 * time.Millisecond)
		return 123, nil
	})

	select {
	case res := <-ch:
		if res.Err != nil || res.Val != 123 {
			t.Fatalf("unexpected result: %+v", res)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout waiting DoChan result")
	}
}

func TestDoContext_Timeout(t *testing.T) {
	sf := New()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err, shared := sf.DoContext(ctx, "ctx-key", func(context.Context) (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		return "late", nil
	})

	if err == nil {
		t.Fatal("expected context timeout error")
	}
	if shared {
		t.Fatal("first caller should not be shared")
	}
}

func TestDoWithTimeout(t *testing.T) {
	sf := New()
	_, err, _ := sf.DoWithTimeout("timeout-key", 20*time.Millisecond, func() (interface{}, error) {
		time.Sleep(80 * time.Millisecond)
		return "slow", nil
	})
	if err == nil {
		t.Fatal("expected timeout-related error")
	}
	if !errors.Is(err, ErrTimeout) && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected ErrTimeout or DeadlineExceeded, got %v", err)
	}
}

func TestForgetAndCacheLoader(t *testing.T) {
	sf := New()
	v1, err, _ := sf.Do("k1", func() (interface{}, error) { return "v1", nil })
	if err != nil || v1 != "v1" {
		t.Fatalf("unexpected v1=%v err=%v", v1, err)
	}
	sf.Forget("k1")
	// Forget on non-inflight key should be safe and no panic.

	cl := NewCacheLoader(80 * time.Millisecond)
	var loadCount int32
	loader := func() (interface{}, error) {
		atomic.AddInt32(&loadCount, 1)
		return "cache-value", nil
	}

	v, err := cl.Load("user:1", loader)
	if err != nil || v != "cache-value" {
		t.Fatalf("first load failed: v=%v err=%v", v, err)
	}
	v, err = cl.Load("user:1", loader)
	if err != nil || v != "cache-value" {
		t.Fatalf("second load failed: v=%v err=%v", v, err)
	}
	if atomic.LoadInt32(&loadCount) != 1 {
		t.Fatalf("loader should run once before ttl expiry, got %d", loadCount)
	}

	time.Sleep(100 * time.Millisecond)
	v, err = cl.Load("user:1", loader)
	if err != nil || v != "cache-value" {
		t.Fatalf("third load failed: v=%v err=%v", v, err)
	}
	if atomic.LoadInt32(&loadCount) != 2 {
		t.Fatalf("loader should run again after ttl expiry, got %d", loadCount)
	}
}
