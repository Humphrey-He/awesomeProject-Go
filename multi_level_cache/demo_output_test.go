package multi_level_cache

import (
	"sync/atomic"
	"testing"
	"time"
)

type demoBackend struct {
	base *InMemoryBackend
	hits int32
}

func (d *demoBackend) Get(key string) (string, error) {
	atomic.AddInt32(&d.hits, 1)
	return d.base.Get(key)
}
func (d *demoBackend) Set(key, value string, ttl time.Duration) error {
	return d.base.Set(key, value, ttl)
}
func (d *demoBackend) Delete(key string) error { return d.base.Delete(key) }

func TestDemoOutput(t *testing.T) {
	b := &demoBackend{base: NewInMemoryBackend()}
	_ = b.Set("hot", "v1", time.Second)
	c := New(2, b)

	v, _ := c.Get("hot")
	t.Logf("first get value=%s l2_hits=%d", v, atomic.LoadInt32(&b.hits))
	v, _ = c.Get("hot")
	t.Logf("second get(l1 hit) value=%s l2_hits=%d", v, atomic.LoadInt32(&b.hits))

	_ = c.Set("k2", "v2", 20*time.Millisecond)
	time.Sleep(30 * time.Millisecond)
	_, err := c.Get("k2")
	t.Logf("after ttl expiry get k2 err=%v", err)
}
