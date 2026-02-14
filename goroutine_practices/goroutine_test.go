package goroutine_practices

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// ========== 功能测试 ==========

func TestWorkerPool(t *testing.T) {
	pool := NewWorkerPool(3, 10)
	pool.Start()
	defer pool.Stop()
	
	var counter int32
	
	for i := 0; i < 10; i++ {
		if !pool.Submit(func() {
			atomic.AddInt32(&counter, 1)
		}) {
			t.Error("Failed to submit task")
		}
	}
	
	time.Sleep(100 * time.Millisecond)
	
	if atomic.LoadInt32(&counter) != 10 {
		t.Errorf("Expected 10 tasks completed, got %d", counter)
	}
}

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10) // 10 requests per second
	defer rl.Stop()
	
	allowed := 0
	for i := 0; i < 20; i++ {
		if rl.Allow() {
			allowed++
		}
	}
	
	if allowed > 10 {
		t.Errorf("Too many requests allowed: %d", allowed)
	}
}

func TestErrorCollector(t *testing.T) {
	tasks := []func() error{
		func() error { return nil },
		func() error { return nil },
		func() error { return nil },
	}
	
	errs := ProcessWithErrorCollection(tasks)
	if len(errs) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(errs))
	}
}

func TestContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		close(done)
	}()
	
	cancel()
	
	select {
	case <-done:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Error("Context cancellation not propagated")
	}
}

func TestSupervisor(t *testing.T) {
	ctx := context.Background()
	sup := NewSupervisor(ctx)
	
	var counter int32
	
	for i := 0; i < 5; i++ {
		sup.Go(func(ctx context.Context) error {
			atomic.AddInt32(&counter, 1)
			return nil
		})
	}
	
	sup.Wait()
	
	if atomic.LoadInt32(&counter) != 5 {
		t.Errorf("Expected 5 tasks, got %d", counter)
	}
}

func TestRetryWithBackoff(t *testing.T) {
	attempts := 0
	f := func() error {
		attempts++
		if attempts < 3 {
			return ErrRetry
		}
		return nil
	}
	
	err := RetryWithBackoff(5, f)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

// ========== 性能基准测试 ==========

func BenchmarkGoroutineCreation(b *testing.B) {
	b.Run("WithoutGoroutine", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			func() {
				_ = 1 + 1
			}()
		}
	})
	
	b.Run("WithGoroutine", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			done := make(chan struct{})
			go func() {
				_ = 1 + 1
				close(done)
			}()
			<-done
		}
	})
}

func BenchmarkChannelCommunication(b *testing.B) {
	b.Run("UnbufferedChannel", func(b *testing.B) {
		ch := make(chan int)
		go func() {
			for i := 0; i < b.N; i++ {
				<-ch
			}
		}()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ch <- i
		}
	})
	
	b.Run("BufferedChannel", func(b *testing.B) {
		ch := make(chan int, 100)
		go func() {
			for i := 0; i < b.N; i++ {
				<-ch
			}
		}()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ch <- i
		}
	})
}

func BenchmarkMutexVsChannel(b *testing.B) {
	b.Run("Mutex", func(b *testing.B) {
		m := NewSafeMap()
		
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				m.Set("key", i)
				_, _ = m.Get("key")
				i++
			}
		})
	})
	
	b.Run("Atomic", func(b *testing.B) {
		c := &AtomicCounter{}
		
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				c.Increment()
				_ = c.Get()
			}
		})
	})
}

func BenchmarkWorkerPool(b *testing.B) {
	pool := NewWorkerPool(10, 1000)
	pool.Start()
	defer pool.Stop()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Submit(func() {
			_ = 1 + 1
		})
	}
}

func BenchmarkPipeline(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nums := make([]int, 100)
		for j := range nums {
			nums[j] = j
		}
		
		in := gen(nums...)
		out := square(in)
		
		for range out {
		}
	}
}

// ========== 竞态检测测试 ==========

func TestSafeMapRace(t *testing.T) {
	m := NewSafeMap()
	
	done := make(chan struct{})
	
	// Writer
	go func() {
		for i := 0; i < 1000; i++ {
			m.Set("key", i)
		}
		close(done)
	}()
	
	// Reader
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				_, _ = m.Get("key")
			}
		}
	}()
	
	<-done
}

func TestAtomicCounterRace(t *testing.T) {
	c := &AtomicCounter{}
	
	done := make(chan struct{})
	
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				c.Increment()
			}
			done <- struct{}{}
		}()
	}
	
	for i := 0; i < 10; i++ {
		<-done
	}
	
	expected := int64(10000)
	if c.Get() != expected {
		t.Errorf("Expected %d, got %d", expected, c.Get())
	}
}

// ========== 错误定义 ==========

var ErrRetry = func() error {
	return &retryError{}
}()

type retryError struct{}

func (e *retryError) Error() string {
	return "retry"
}

