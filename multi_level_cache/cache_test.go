package multi_level_cache

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type countingBackend struct {
	base *InMemoryBackend
	cnt  int32
}

func (b *countingBackend) Get(key string) (string, error) {
	atomic.AddInt32(&b.cnt, 1)
	time.Sleep(20 * time.Millisecond)
	return b.base.Get(key)
}

func (b *countingBackend) Set(key, value string, ttl time.Duration) error {
	return b.base.Set(key, value, ttl)
}

func (b *countingBackend) Delete(key string) error {
	return b.base.Delete(key)
}

func TestSetGetDelete(t *testing.T) {
	l2 := NewInMemoryBackend()
	c := New(2, l2)

	if err := c.Set("k1", "v1", time.Second); err != nil {
		t.Fatalf("set err: %v", err)
	}
	v, err := c.Get("k1")
	if err != nil || v != "v1" {
		t.Fatalf("get failed v=%s err=%v", v, err)
	}
	if err := c.Delete("k1"); err != nil {
		t.Fatalf("delete err: %v", err)
	}
	if _, err = c.Get("k1"); err == nil {
		t.Fatal("expected not found after delete")
	}
}

func TestL1EvictionAndTTL(t *testing.T) {
	l2 := NewInMemoryBackend()
	c := New(2, l2)

	_ = c.Set("a", "1", 30*time.Millisecond)
	_ = c.Set("b", "2", time.Second)
	_ = c.Set("c", "3", time.Second) // evict one

	l1Size, _ := c.Stats()
	if l1Size != 2 {
		t.Fatalf("expected l1 size 2, got %d", l1Size)
	}

	time.Sleep(50 * time.Millisecond)
	if _, err := c.Get("a"); err == nil {
		t.Fatalf("a should expire")
	}
}

func TestAntiStampedeSingleflight(t *testing.T) {
	base := NewInMemoryBackend()
	b := &countingBackend{base: base}
	_ = b.Set("hot", "value", time.Second)

	c := New(64, b)
	var wg sync.WaitGroup
	const n = 20
	wg.Add(n)

	// force miss in L1 by clearing.
	_ = c.Delete("hot")
	_ = b.Set("hot", "value", time.Second)

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			v, err := c.Get("hot")
			if err != nil || v != "value" {
				t.Errorf("unexpected v=%s err=%v", v, err)
			}
		}()
	}
	wg.Wait()
	if atomic.LoadInt32(&b.cnt) != 1 {
		t.Fatalf("backend should be hit once, got %d", atomic.LoadInt32(&b.cnt))
	}
}
