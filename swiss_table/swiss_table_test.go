package swiss_table

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// ========== Swiss Table 基础测试 ==========

func TestSwissTable_Basic(t *testing.T) {
	st := NewSwissTable[string, int]()

	st.Put("a", 1)
	st.Put("b", 2)
	st.Put("c", 3)

	if st.Len() != 3 {
		t.Errorf("Expected length 3, got %d", st.Len())
	}

	val, ok := st.Get("a")
	if !ok || val != 1 {
		t.Errorf("Expected to get 1, got %v, %v", val, ok)
	}
}

func TestSwissTable_Update(t *testing.T) {
	st := NewSwissTable[string, int]()

	st.Put("key", 1)
	st.Put("key", 2)

	if st.Len() != 1 {
		t.Errorf("Expected length 1, got %d", st.Len())
	}

	val, _ := st.Get("key")
	if val != 2 {
		t.Errorf("Expected 2, got %d", val)
	}
}

func TestSwissTable_Delete(t *testing.T) {
	st := NewSwissTable[string, int]()

	st.Put("a", 1)
	st.Put("b", 2)

	deleted := st.Delete("a")
	if !deleted {
		t.Error("Expected to delete 'a'")
	}

	if st.Len() != 1 {
		t.Errorf("Expected length 1, got %d", st.Len())
	}

	_, ok := st.Get("a")
	if ok {
		t.Error("Expected 'a' to be deleted")
	}
}

func TestSwissTable_Contains(t *testing.T) {
	st := NewSwissTable[string, int]()

	st.Put("exists", 1)

	if !st.Contains("exists") {
		t.Error("Expected 'exists' to be in table")
	}

	if st.Contains("missing") {
		t.Error("Expected 'missing' not to be in table")
	}
}

func TestSwissTable_Clear(t *testing.T) {
	st := NewSwissTable[string, int]()

	for i := 0; i < 100; i++ {
		st.Put(fmt.Sprintf("key%d", i), i)
	}

	st.Clear()

	if st.Len() != 0 {
		t.Errorf("Expected length 0 after clear, got %d", st.Len())
	}
}

func TestSwissTable_Growth(t *testing.T) {
	st := NewSwissTable[int, int]()

	// 插入大量元素触发扩容
	n := 1000
	for i := 0; i < n; i++ {
		st.Put(i, i*10)
	}

	if st.Len() != n {
		t.Errorf("Expected length %d, got %d", n, st.Len())
	}

	// 验证所有元素都存在
	for i := 0; i < n; i++ {
		val, ok := st.Get(i)
		if !ok {
			t.Errorf("Key %d not found after growth", i)
		}
		if val != i*10 {
			t.Errorf("Expected %d, got %d", i*10, val)
		}
	}
}

func TestSwissTable_Keys(t *testing.T) {
	st := NewSwissTable[string, int]()

	expected := map[string]bool{
		"a": true,
		"b": true,
		"c": true,
	}

	st.Put("a", 1)
	st.Put("b", 2)
	st.Put("c", 3)

	keys := st.Keys()

	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	for _, key := range keys {
		if !expected[key] {
			t.Errorf("Unexpected key: %s", key)
		}
	}
}

// ========== Go Map 对比测试 ==========

func TestGoMap_Basic(t *testing.T) {
	gm := NewGoMap[string, int]()

	gm.Put("a", 1)
	gm.Put("b", 2)
	gm.Put("c", 3)

	if gm.Len() != 3 {
		t.Errorf("Expected length 3, got %d", gm.Len())
	}

	val, ok := gm.Get("a")
	if !ok || val != 1 {
		t.Errorf("Expected to get 1, got %v, %v", val, ok)
	}
}

// ========== 性能基准测试 ==========

const benchSize = 10000

func BenchmarkSwissTable_Put(b *testing.B) {
	st := NewSwissTable[int, int]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		st.Put(i%benchSize, i)
	}
}

func BenchmarkGoMap_Put(b *testing.B) {
	gm := NewGoMap[int, int]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gm.Put(i%benchSize, i)
	}
}

func BenchmarkSwissTable_Get(b *testing.B) {
	st := NewSwissTable[int, int]()
	for i := 0; i < benchSize; i++ {
		st.Put(i, i*10)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		st.Get(i % benchSize)
	}
}

func BenchmarkGoMap_Get(b *testing.B) {
	gm := NewGoMap[int, int]()
	for i := 0; i < benchSize; i++ {
		gm.Put(i, i*10)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gm.Get(i % benchSize)
	}
}

func BenchmarkSwissTable_Delete(b *testing.B) {
	b.StopTimer()
	st := NewSwissTable[int, int]()
	for i := 0; i < benchSize; i++ {
		st.Put(i, i)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		st.Delete(i % benchSize)
		if i%benchSize == benchSize-1 {
			b.StopTimer()
			// 重新填充
			for j := 0; j < benchSize; j++ {
				st.Put(j, j)
			}
			b.StartTimer()
		}
	}
}

func BenchmarkGoMap_Delete(b *testing.B) {
	b.StopTimer()
	gm := NewGoMap[int, int]()
	for i := 0; i < benchSize; i++ {
		gm.Put(i, i)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		gm.Delete(i % benchSize)
		if i%benchSize == benchSize-1 {
			b.StopTimer()
			// 重新填充
			for j := 0; j < benchSize; j++ {
				gm.Put(j, j)
			}
			b.StartTimer()
		}
	}
}

// ========== 字符串键性能测试 ==========

func BenchmarkSwissTable_StringKey_Put(b *testing.B) {
	st := NewSwissTable[string, int]()
	keys := generateStringKeys(benchSize)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		st.Put(keys[i%benchSize], i)
	}
}

func BenchmarkGoMap_StringKey_Put(b *testing.B) {
	gm := NewGoMap[string, int]()
	keys := generateStringKeys(benchSize)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gm.Put(keys[i%benchSize], i)
	}
}

func BenchmarkSwissTable_StringKey_Get(b *testing.B) {
	st := NewSwissTable[string, int]()
	keys := generateStringKeys(benchSize)
	for i := 0; i < benchSize; i++ {
		st.Put(keys[i], i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		st.Get(keys[i%benchSize])
	}
}

func BenchmarkGoMap_StringKey_Get(b *testing.B) {
	gm := NewGoMap[string, int]()
	keys := generateStringKeys(benchSize)
	for i := 0; i < benchSize; i++ {
		gm.Put(keys[i], i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gm.Get(keys[i%benchSize])
	}
}

// ========== 随机访问模式测试 ==========

func BenchmarkSwissTable_RandomAccess(b *testing.B) {
	st := NewSwissTable[int, int]()
	for i := 0; i < benchSize; i++ {
		st.Put(i, i)
	}

	rand.Seed(time.Now().UnixNano())
	indices := make([]int, b.N)
	for i := range indices {
		indices[i] = rand.Intn(benchSize)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		st.Get(indices[i])
	}
}

func BenchmarkGoMap_RandomAccess(b *testing.B) {
	gm := NewGoMap[int, int]()
	for i := 0; i < benchSize; i++ {
		gm.Put(i, i)
	}

	rand.Seed(time.Now().UnixNano())
	indices := make([]int, b.N)
	for i := range indices {
		indices[i] = rand.Intn(benchSize)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gm.Get(indices[i])
	}
}

// ========== 混合操作测试 ==========

func BenchmarkSwissTable_Mixed(b *testing.B) {
	st := NewSwissTable[int, int]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		op := i % 10
		key := i % benchSize

		if op < 5 { // 50% Get
			st.Get(key)
		} else if op < 9 { // 40% Put2
			st.Put(key, i)
		} else { // 10% Delete
			st.Delete(key)
		}
	}
}

func BenchmarkGoMap_Mixed(b *testing.B) {
	gm := NewGoMap[int, int]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		op := i % 10
		key := i % benchSize

		if op < 5 { // 50% Get
			gm.Get(key)
		} else if op < 9 { // 40% Put2
			gm.Put(key, i)
		} else { // 10% Delete
			gm.Delete(key)
		}
	}
}

// ========== 内存占用对比 ==========

func BenchmarkSwissTable_Memory(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		st := NewSwissTable[int, int]()
		for j := 0; j < 1000; j++ {
			st.Put(j, j)
		}
	}
}

func BenchmarkGoMap_Memory(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		gm := NewGoMap[int, int]()
		for j := 0; j < 1000; j++ {
			gm.Put(j, j)
		}
	}
}

// ========== 辅助函数 ==========

func generateStringKeys(n int) []string {
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = fmt.Sprintf("key_%d", i)
	}
	return keys
}
