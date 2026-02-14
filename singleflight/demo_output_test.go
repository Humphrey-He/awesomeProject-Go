package singleflight

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDemoOutput(t *testing.T) {
	sf := New()
	var runCount int32

	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			v, err, shared := sf.Do("demo-key", func() (interface{}, error) {
				atomic.AddInt32(&runCount, 1)
				time.Sleep(40 * time.Millisecond)
				return "payload", nil
			})
			t.Logf("[caller-%d] value=%v err=%v shared=%v", id, v, err, shared)
		}(i)
	}
	wg.Wait()
	t.Logf("underlying function executed times=%d", atomic.LoadInt32(&runCount))

	cl := NewCacheLoader(100 * time.Millisecond)
	loadCount := int32(0)
	loader := func() (interface{}, error) {
		atomic.AddInt32(&loadCount, 1)
		time.Sleep(15 * time.Millisecond)
		return map[string]interface{}{"id": 1, "name": "Alice"}, nil
	}

	v1, _ := cl.Load("user:1", loader)
	v2, _ := cl.Load("user:1", loader)
	t.Logf("cache load1=%v", v1)
	t.Logf("cache load2=%v", v2)
	t.Logf("loader executed times(before expire)=%d", atomic.LoadInt32(&loadCount))

	time.Sleep(120 * time.Millisecond)
	_, _ = cl.Load("user:1", loader)
	t.Logf("loader executed times(after expire)=%d", atomic.LoadInt32(&loadCount))
}


