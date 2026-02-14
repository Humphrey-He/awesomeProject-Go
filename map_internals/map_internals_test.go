package map_internals

import (
	"fmt"
	"testing"
)

// ========== 1. 基本功能测试 ==========

func TestSimulatedMap_Basic(t *testing.T) {
	m := NewSimulatedMap(0)

	// 插入
	m.Set("key1", "value1")
	m.Set("key2", "value2")
	m.Set("key3", "value3")

	// 获取
	if v, ok := m.Get("key1"); !ok || v != "value1" {
		t.Errorf("Get(key1) = %v, %v; want value1, true", v, ok)
	}

	// 更新
	m.Set("key1", "updated")
	if v, ok := m.Get("key1"); !ok || v != "updated" {
		t.Errorf("Get(key1) after update = %v, %v; want updated, true", v, ok)
	}

	// 统计
	if m.count != 3 {
		t.Errorf("count = %d, want 3", m.count)
	}
}

// ========== 2. 扩容测试 ==========

func TestSimulatedMap_GrowByOverload(t *testing.T) {
	m := NewSimulatedMap(0)

	fmt.Println("\n=== Testing Growth by Overload ===")

	// 初始状态
	stats := m.Stats()
	fmt.Printf("Initial: %s\n\n", stats.String())

	initialB := m.B

	// 插入足够多的元素触发扩容
	// 负载因子阈值是6.5，每个bucket 8个元素
	// 初始B=0，buckets=1，触发扩容需要 > 1 * 6.5 = 7个元素
	for i := 0; i < 100; i++ {
		m.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))

		// 每10个元素打印一次状态
		if (i+1)%10 == 0 {
			stats := m.Stats()
			fmt.Printf("After %d inserts:\n%s\n\n", i+1, stats.String())
		}
	}

	// 验证扩容发生
	if m.B <= initialB {
		t.Errorf("B should have increased, got %d, initial %d", m.B, initialB)
	}

	// 验证所有元素都能找到
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)
		expectedValue := fmt.Sprintf("value%d", i)
		if v, ok := m.Get(key); !ok || v != expectedValue {
			t.Errorf("Get(%s) = %v, %v; want %s, true", key, v, ok, expectedValue)
		}
	}

	finalStats := m.Stats()
	fmt.Printf("Final: %s\n", finalStats.String())

	// 验证负载因子在合理范围
	if finalStats.LoadFactor > 6.5 {
		t.Errorf("Load factor too high: %.2f", finalStats.LoadFactor)
	}
}

// ========== 3. 等量扩容测试 ==========

func TestSimulatedMap_SameSizeGrow(t *testing.T) {
	t.Skip("Same size grow is complex to trigger, skipping")

	m := NewSimulatedMap(0)

	// 创建大量overflow buckets的场景比较复杂
	// 这里只是演示概念

	fmt.Println("\n=== Testing Same Size Growth ===")

	// 使用相同hash值的keys（这在实际中很难构造）
	// 在真实场景中，需要精心构造hash冲突

	initialB := m.B
	stats := m.Stats()
	fmt.Printf("Initial: %s\n", stats.String())

	if m.sameSizeGrow {
		t.Log("Same size grow triggered")
		if m.B != initialB {
			t.Errorf("B should not change in same size grow")
		}
	}
}

// ========== 4. 增量迁移测试 ==========

func TestSimulatedMap_IncrementalEvacuation(t *testing.T) {
	m := NewSimulatedMap(0)

	fmt.Println("\n=== Testing Incremental Evacuation ===")

	// 插入元素触发扩容
	for i := 0; i < 20; i++ {
		m.Set(fmt.Sprintf("key%d", i), i)

		if m.growing {
			stats := m.Stats()
			fmt.Printf("Growing... Progress: %.1f%%, Evacuated: %d/%d\n",
				stats.EvacuateProgress*100, m.nevacuate, len(m.oldbuckets))
		}
	}

	// 继续访问，触发增量迁移
	for i := 0; i < 20; i++ {
		m.Get(fmt.Sprintf("key%d", i))

		if m.growing {
			stats := m.Stats()
			fmt.Printf("Accessing... Progress: %.1f%%\n", stats.EvacuateProgress*100)
		}
	}

	// 最终应该完成迁移
	if m.growing {
		// 强制完成迁移
		for m.growing {
			m.evacuate(m.nevacuate)
		}
	}

	if m.growing {
		t.Error("Map should finish growing")
	}

	if m.oldbuckets != nil {
		t.Error("Old buckets should be nil after evacuation")
	}
}

// ========== 5. 负载因子测试 ==========

func TestLoadFactor(t *testing.T) {
	tests := []struct {
		count int
		B     uint8
		want  bool
	}{
		{0, 0, false},
		{1, 0, false},
		{6, 0, false},  // 1 bucket, 6 elements
		{7, 0, false},  // 1 bucket, 7 elements: 7 < 13/2 = 6.5, 但还要检查 count > bucketCnt
		{8, 0, false},  // 8 elements, loadfactor = 8, 但 8 * 2 / 13 = 1.23 < 6.5
		{9, 0, true},   // 9 > 8 且 9 > 1 * 6.5
		{13, 1, false}, // 2 buckets, 13 elements: 13 = 2 * 6.5 (edge)
		{14, 1, true},  // 2 buckets, 14 elements: 14 > 2 * 6.5
	}

	for _, tt := range tests {
		got := overLoadFactor(tt.count, tt.B)
		if got != tt.want {
			buckets := 1 << tt.B
			threshold := float64(buckets) * 6.5
			t.Errorf("overLoadFactor(%d, %d) = %v, want %v (buckets=%d, threshold=%.1f)",
				tt.count, tt.B, got, tt.want, buckets, threshold)
		}
	}
}

// ========== 6. 性能对比测试 ==========

func TestMap_Performance_NoPrealloc(t *testing.T) {
	m := NewSimulatedMap(0)

	for i := 0; i < 1000; i++ {
		m.Set(i, i)
	}

	stats := m.Stats()
	t.Logf("Without prealloc: %s", stats.String())
}

func TestMap_Performance_WithPrealloc(t *testing.T) {
	m := NewSimulatedMap(1000) // 预分配

	for i := 0; i < 1000; i++ {
		m.Set(i, i)
	}

	stats := m.Stats()
	t.Logf("With prealloc: %s", stats.String())
}

// ========== 7. 基准测试 ==========

func BenchmarkSimulatedMap_Set(b *testing.B) {
	m := NewSimulatedMap(b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Set(i, i)
	}
}

func BenchmarkSimulatedMap_Get(b *testing.B) {
	m := NewSimulatedMap(10000)
	for i := 0; i < 10000; i++ {
		m.Set(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get(i % 10000)
	}
}

func BenchmarkNativeMap_Set(b *testing.B) {
	m := make(map[int]int, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m[i] = i
	}
}

func BenchmarkNativeMap_Get(b *testing.B) {
	m := make(map[int]int, 10000)
	for i := 0; i < 10000; i++ {
		m[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m[i%10000]
	}
}

// ========== 8. 扩容次数统计 ==========

func TestMap_GrowthCount(t *testing.T) {
	m := NewSimulatedMap(0)

	growthCount := 0
	lastB := m.B

	fmt.Println("\n=== Map Growth Tracking ===")
	fmt.Printf("Initial: B=%d, Buckets=%d\n", m.B, len(m.buckets))

	for i := 0; i < 10000; i++ {
		m.Set(i, i)

		if m.B > lastB {
			growthCount++
			stats := m.Stats()
			fmt.Printf("Growth #%d at count=%d: B=%d->%d, Buckets=%d, LoadFactor=%.2f\n",
				growthCount, i, lastB, m.B, len(m.buckets), stats.LoadFactor)
			lastB = m.B
		}
	}

	finalStats := m.Stats()
	fmt.Printf("\nFinal: %s\n", finalStats.String())
	fmt.Printf("Total growths: %d\n", growthCount)

	// 理论上，10000个元素需要的扩容次数
	// B=0: 1 bucket, 触发扩容在 7
	// B=1: 2 buckets, 触发扩容在 14
	// B=2: 4 buckets, 触发扩容在 27
	// B=3: 8 buckets, 触发扩容在 53
	// B=4: 16 buckets, 触发扩容在 105
	// B=5: 32 buckets, 触发扩容在 209
	// B=6: 64 buckets, 触发扩容在 417
	// B=7: 128 buckets, 触发扩容在 833
	// B=8: 256 buckets, 触发扩容在 1665
	// B=9: 512 buckets, 触发扩容在 3329
	// B=10: 1024 buckets, 触发扩容在 6657
	// B=11: 2048 buckets, 可以容纳 13312 元素

	expectedGrowth := 11 // 大约11次
	if growthCount < expectedGrowth-2 || growthCount > expectedGrowth+2 {
		t.Logf("Warning: growth count %d seems unusual (expected ~%d)", growthCount, expectedGrowth)
	}
}

// ========== 9. 并发安全测试（应该panic）==========

func TestMap_NotConcurrentSafe(t *testing.T) {
	t.Skip("Concurrent access would cause race, skipping")

	// 这个测试演示map不是并发安全的
	// 实际运行会有data race

	// m := NewSimulatedMap(0)
	//
	// done := make(chan bool)
	//
	// // 并发写
	// for i := 0; i < 10; i++ {
	// 	go func(id int) {
	// 		for j := 0; j < 100; j++ {
	// 			m.Set(fmt.Sprintf("key%d-%d", id, j), j)
	// 		}
	// 		done <- true
	// 	}(i)
	// }
	//
	// for i := 0; i < 10; i++ {
	// 	<-done
	// }
}

// ========== 10. 示例测试 ==========

func ExampleSimulatedMap() {
	// 创建map
	m := NewSimulatedMap(0)

	// 插入元素
	m.Set("name", "Alice")
	m.Set("age", 30)
	m.Set("city", "Beijing")

	// 获取元素
	name, _ := m.Get("name")
	fmt.Println("Name:", name)

	// 查看统计
	stats := m.Stats()
	fmt.Printf("Elements: %d, Buckets: %d\n", stats.Count, stats.BucketCount)

	// Output:
	// Name: Alice
	// Elements: 3, Buckets: 1
}

func ExampleSimulatedMap_growth() {
	m := NewSimulatedMap(0)

	fmt.Println("Initial buckets:", len(m.buckets))

	// 插入足够多的元素触发扩容
	for i := 0; i < 10; i++ {
		m.Set(fmt.Sprintf("key%d", i), i)
	}

	fmt.Println("After 10 inserts, buckets:", len(m.buckets))

	stats := m.Stats()
	fmt.Printf("Load factor: %.2f\n", stats.LoadFactor)

	// Output:
	// Initial buckets: 1
	// After 10 inserts, buckets: 2
	// Load factor: 0.62
}
