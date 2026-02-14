package cache

import "testing"

func TestDemoOutput(t *testing.T) {
	lru := NewLRUCache(2)
	lru.Put("a", 1)
	lru.Put("b", 2)
	lru.Get("a")
	lru.Put("c", 3)
	t.Logf("lru keys(recent->old)=%v", lru.Keys())

	lfu := NewLFUCache(2)
	lfu.Put("x", 10)
	lfu.Put("y", 20)
	lfu.Get("x")
	lfu.Get("x")
	lfu.Put("z", 30)
	t.Logf("lfu freq x=%d y=%d z=%d", lfu.GetFreq("x"), lfu.GetFreq("y"), lfu.GetFreq("z"))
}
