package cache

import (
	"fmt"
	"sync"
	"testing"
)

// ========== LRU 测试 ==========

func TestLRUCache_Basic(t *testing.T) {
	lru := NewLRUCache(2)

	lru.Put("a", 1)
	lru.Put("b", 2)

	val, ok := lru.Get("a")
	if !ok || val != 1 {
		t.Errorf("Expected to get 1, got %v, %v", val, ok)
	}

	val, ok = lru.Get("b")
	if !ok || val != 2 {
		t.Errorf("Expected to get 2, got %v, %v", val, ok)
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	lru := NewLRUCache(2)

	lru.Put("a", 1)
	lru.Put("b", 2)
	lru.Put("c", 3) // 应该淘汰 a

	// a应该被淘汰
	_, ok := lru.Get("a")
	if ok {
		t.Error("Expected 'a' to be evicted")
	}

	// b和c应该还在
	val, ok := lru.Get("b")
	if !ok || val != 2 {
		t.Error("Expected 'b' to exist")
	}

	val, ok = lru.Get("c")
	if !ok || val != 3 {
		t.Error("Expected 'c' to exist")
	}
}

func TestLRUCache_Update(t *testing.T) {
	lru := NewLRUCache(2)

	lru.Put("a", 1)
	lru.Put("b", 2)
	lru.Put("a", 10) // 更新a的值

	val, ok := lru.Get("a")
	if !ok || val != 10 {
		t.Errorf("Expected to get 10, got %v", val)
	}

	if lru.Len() != 2 {
		t.Errorf("Expected length 2, got %d", lru.Len())
	}
}

func TestLRUCache_LRUOrder(t *testing.T) {
	lru := NewLRUCache(3)

	lru.Put("a", 1)
	lru.Put("b", 2)
	lru.Put("c", 3)

	// 访问a，使其成为最近使用的
	lru.Get("a")

	// 添加d，应该淘汰b（最少使用的）
	lru.Put("d", 4)

	_, ok := lru.Get("b")
	if ok {
		t.Error("Expected 'b' to be evicted")
	}

	// a应该还在
	val, ok := lru.Get("a")
	if !ok || val != 1 {
		t.Error("Expected 'a' to exist")
	}
}

func TestLRUCache_Delete(t *testing.T) {
	lru := NewLRUCache(2)

	lru.Put("a", 1)
	lru.Put("b", 2)

	deleted := lru.Delete("a")
	if !deleted {
		t.Error("Expected to delete 'a'")
	}

	if lru.Len() != 1 {
		t.Errorf("Expected length 1, got %d", lru.Len())
	}

	_, ok := lru.Get("a")
	if ok {
		t.Error("Expected 'a' to be deleted")
	}
}

func TestLRUCache_Peek(t *testing.T) {
	lru := NewLRUCache(2)

	lru.Put("a", 1)
	lru.Put("b", 2)

	// Peek不应该改变访问顺序
	val, ok := lru.Peek("a")
	if !ok || val != 1 {
		t.Error("Peek failed")
	}

	// 添加c，应该淘汰a（因为peek不更新访问时间）
	lru.Put("c", 3)

	_, ok = lru.Get("a")
	if ok {
		t.Error("Expected 'a' to be evicted after peek")
	}
}

func TestLRUCache_Keys(t *testing.T) {
	lru := NewLRUCache(3)

	lru.Put("a", 1)
	lru.Put("b", 2)
	lru.Put("c", 3)

	keys := lru.Keys()

	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// 最近添加的应该在前面
	if keys[0] != "c" {
		t.Errorf("Expected first key to be 'c', got %v", keys[0])
	}
}

// ========== LFU 测试 ==========

func TestLFUCache_Basic(t *testing.T) {
	lfu := NewLFUCache(2)

	lfu.Put("a", 1)
	lfu.Put("b", 2)

	val, ok := lfu.Get("a")
	if !ok || val != 1 {
		t.Errorf("Expected to get 1, got %v, %v", val, ok)
	}

	val, ok = lfu.Get("b")
	if !ok || val != 2 {
		t.Errorf("Expected to get 2, got %v, %v", val, ok)
	}
}

func TestLFUCache_Eviction(t *testing.T) {
	lfu := NewLFUCache(2)

	lfu.Put("a", 1)
	lfu.Put("b", 2)

	// 访问a两次，使其频率更高
	lfu.Get("a")
	lfu.Get("a")

	// 只访问b一次
	lfu.Get("b")

	// 添加c，应该淘汰b（频率更低）
	lfu.Put("c", 3)

	_, ok := lfu.Get("b")
	if ok {
		t.Error("Expected 'b' to be evicted (lower frequency)")
	}

	// a应该还在
	val, ok := lfu.Get("a")
	if !ok || val != 1 {
		t.Error("Expected 'a' to exist")
	}
}

func TestLFUCache_SameFrequency(t *testing.T) {
	lfu := NewLFUCache(2)

	lfu.Put("a", 1)
	lfu.Put("b", 2)

	// 两者频率相同，应该淘汰最旧的（a）
	lfu.Put("c", 3)

	_, ok := lfu.Get("a")
	if ok {
		t.Error("Expected 'a' to be evicted (oldest with same frequency)")
	}

	val, ok := lfu.Get("b")
	if !ok || val != 2 {
		t.Error("Expected 'b' to exist")
	}
}

func TestLFUCache_Update(t *testing.T) {
	lfu := NewLFUCache(2)

	lfu.Put("a", 1)

	// 访问a，频率变为2
	lfu.Get("a")

	// 更新a的值，频率应该增加
	lfu.Put("a", 10)

	// 频率应该是3（初始Put 1 + Get 1 + 更新Put 1）
	freq := lfu.GetFreq("a")
	if freq != 3 {
		t.Errorf("Expected frequency 3, got %d", freq)
	}

	// 验证值已更新
	val, ok := lfu.Get("a")
	if !ok || val != 10 {
		t.Errorf("Expected to get 10, got %v", val)
	}

	// Get后频率应该是4
	if lfu.GetFreq("a") != 4 {
		t.Errorf("Expected frequency 4 after Get, got %d", lfu.GetFreq("a"))
	}
}

func TestLFUCache_GetFreq(t *testing.T) {
	lfu := NewLFUCache(2)

	lfu.Put("a", 1)

	if lfu.GetFreq("a") != 1 {
		t.Error("Expected initial frequency to be 1")
	}

	lfu.Get("a")
	lfu.Get("a")

	if lfu.GetFreq("a") != 3 {
		t.Errorf("Expected frequency 3, got %d", lfu.GetFreq("a"))
	}
}

func TestLFUCache_Peek(t *testing.T) {
	lfu := NewLFUCache(2)

	lfu.Put("a", 1)

	// Peek不应该增加频率
	val, ok := lfu.Peek("a")
	if !ok || val != 1 {
		t.Error("Peek failed")
	}

	if lfu.GetFreq("a") != 1 {
		t.Error("Peek should not increase frequency")
	}
}

// ========== 泛型版本测试 ==========

func TestLRUCacheGeneric_Int(t *testing.T) {
	lru := NewLRUCacheGeneric[string, int](2)

	lru.Put("a", 1)
	lru.Put("b", 2)

	val, ok := lru.Get("a")
	if !ok || val != 1 {
		t.Errorf("Expected 1, got %d", val)
	}
}

func TestLRUCacheGeneric_String(t *testing.T) {
	lru := NewLRUCacheGeneric[int, string](2)

	lru.Put(1, "one")
	lru.Put(2, "two")

	val, ok := lru.Get(1)
	if !ok || val != "one" {
		t.Errorf("Expected 'one', got '%s'", val)
	}
}

func TestLFUCacheGeneric_Int(t *testing.T) {
	lfu := NewLFUCacheGeneric[string, int](2)

	lfu.Put("a", 1)
	lfu.Put("b", 2)

	val, ok := lfu.Get("a")
	if !ok || val != 1 {
		t.Errorf("Expected 1, got %d", val)
	}
}

// ========== 并发测试 ==========

func TestLRUCache_Concurrent(t *testing.T) {
	lru := NewLRUCache(100)

	var wg sync.WaitGroup
	goroutines := 10
	operations := 1000

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				key := fmt.Sprintf("key%d", j%50)
				lru.Put(key, j)
				lru.Get(key)
			}
		}(i)
	}

	wg.Wait()

	// 验证缓存大小不超过容量
	if lru.Len() > lru.Cap() {
		t.Errorf("Cache size %d exceeds capacity %d", lru.Len(), lru.Cap())
	}
}

func TestLFUCache_Concurrent(t *testing.T) {
	lfu := NewLFUCache(100)

	var wg sync.WaitGroup
	goroutines := 10
	operations := 1000

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				key := fmt.Sprintf("key%d", j%50)
				lfu.Put(key, j)
				lfu.Get(key)
			}
		}(i)
	}

	wg.Wait()

	if lfu.Len() > lfu.Cap() {
		t.Errorf("Cache size %d exceeds capacity %d", lfu.Len(), lfu.Cap())
	}
}

// ========== 性能基准测试 ==========

func BenchmarkLRUCache_Put(b *testing.B) {
	lru := NewLRUCache(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lru.Put(i, i)
	}
}

func BenchmarkLRUCache_Get(b *testing.B) {
	lru := NewLRUCache(1000)
	for i := 0; i < 1000; i++ {
		lru.Put(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lru.Get(i % 1000)
	}
}

func BenchmarkLFUCache_Put(b *testing.B) {
	lfu := NewLFUCache(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lfu.Put(i, i)
	}
}

func BenchmarkLFUCache_Get(b *testing.B) {
	lfu := NewLFUCache(1000)
	for i := 0; i < 1000; i++ {
		lfu.Put(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lfu.Get(i % 1000)
	}
}

func BenchmarkLRUCacheGeneric_Put(b *testing.B) {
	lru := NewLRUCacheGeneric[int, int](1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lru.Put(i, i)
	}
}

func BenchmarkLRUCacheGeneric_Get(b *testing.B) {
	lru := NewLRUCacheGeneric[int, int](1000)
	for i := 0; i < 1000; i++ {
		lru.Put(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lru.Get(i % 1000)
	}
}

// ========== LRU vs LFU 对比测试 ==========

func TestLRU_vs_LFU_Scenario1(t *testing.T) {
	// 场景：重复访问少数几个热点数据
	t.Log("Scenario: Frequently accessed hot data")

	lru := NewLRUCache(3)
	lfu := NewLFUCache(3)

	// 初始化
	for i := 1; i <= 3; i++ {
		lru.Put(i, i*10)
		lfu.Put(i, i*10)
	}

	// 频繁访问1和2
	for i := 0; i < 5; i++ {
		lru.Get(1)
		lru.Get(2)
		lfu.Get(1)
		lfu.Get(2)
	}

	// 添加新数据
	lru.Put(4, 40)
	lfu.Put(4, 40)

	// LRU：淘汰3（最久未使用）
	// LFU：也淘汰3（频率最低）

	_, lruHas3 := lru.Get(3)
	_, lfuHas3 := lfu.Get(3)

	t.Logf("LRU has 3: %v, LFU has 3: %v", lruHas3, lfuHas3)

	if lruHas3 || lfuHas3 {
		t.Error("Both should evict 3")
	}
}

func TestLRU_vs_LFU_Scenario2(t *testing.T) {
	// 场景：顺序扫描
	t.Log("Scenario: Sequential scan")

	lru := NewLRUCache(3)
	lfu := NewLFUCache(3)

	// 顺序访问1-6
	for i := 1; i <= 6; i++ {
		lru.Put(i, i*10)
		lfu.Put(i, i*10)
	}

	// LRU：保留最后3个（4,5,6）
	// LFU：保留最后3个（4,5,6）

	t.Logf("LRU keys: %v", lru.Keys())

	for i := 1; i <= 3; i++ {
		_, lruHas := lru.Get(i)
		_, lfuHas := lfu.Get(i)

		if lruHas || lfuHas {
			t.Errorf("Both should not have %d", i)
		}
	}
}
