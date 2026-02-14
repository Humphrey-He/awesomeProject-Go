package array_vs_list

import (
	"testing"
)

// ========== 数组测试 ==========

func TestArrayDS_Append(t *testing.T) {
	arr := NewArrayDS(10)
	for i := 0; i < 5; i++ {
		arr.Append(i)
	}
	
	if arr.Len() != 5 {
		t.Errorf("Expected length 5, got %d", arr.Len())
	}
	
	for i := 0; i < 5; i++ {
		val, _ := arr.Get(i)
		if val != i {
			t.Errorf("Expected value %d at index %d, got %d", i, i, val)
		}
	}
}

func TestArrayDS_Insert(t *testing.T) {
	arr := NewArrayDS(10)
	arr.Append(1)
	arr.Append(3)
	arr.Insert(1, 2)
	
	expected := []int{1, 2, 3}
	for i, exp := range expected {
		val, _ := arr.Get(i)
		if val != exp {
			t.Errorf("Expected %d at index %d, got %d", exp, i, val)
		}
	}
}

func TestArrayDS_Delete(t *testing.T) {
	arr := NewArrayDS(10)
	for i := 1; i <= 3; i++ {
		arr.Append(i)
	}
	
	arr.Delete(1)
	
	if arr.Len() != 2 {
		t.Errorf("Expected length 2, got %d", arr.Len())
	}
	
	val, _ := arr.Get(1)
	if val != 3 {
		t.Errorf("Expected 3 at index 1, got %d", val)
	}
}

func TestArrayDS_Search(t *testing.T) {
	arr := NewArrayDS(10)
	for i := 0; i < 5; i++ {
		arr.Append(i * 10)
	}
	
	index := arr.Search(20)
	if index != 2 {
		t.Errorf("Expected index 2, got %d", index)
	}
	
	index = arr.Search(99)
	if index != -1 {
		t.Errorf("Expected -1 for not found, got %d", index)
	}
}

// ========== 链表测试 ==========

func TestLinkedList_Append(t *testing.T) {
	list := NewLinkedList()
	for i := 0; i < 5; i++ {
		list.Append(i)
	}
	
	if list.Len() != 5 {
		t.Errorf("Expected length 5, got %d", list.Len())
	}
	
	for i := 0; i < 5; i++ {
		val, _ := list.Get(i)
		if val != i {
			t.Errorf("Expected value %d at index %d, got %d", i, i, val)
		}
	}
}

func TestLinkedList_Insert(t *testing.T) {
	list := NewLinkedList()
	list.Append(1)
	list.Append(3)
	list.Insert(1, 2)
	
	expected := []int{1, 2, 3}
	for i, exp := range expected {
		val, _ := list.Get(i)
		if val != exp {
			t.Errorf("Expected %d at index %d, got %d", exp, i, val)
		}
	}
}

func TestLinkedList_Delete(t *testing.T) {
	list := NewLinkedList()
	for i := 1; i <= 3; i++ {
		list.Append(i)
	}
	
	list.Delete(1)
	
	if list.Len() != 2 {
		t.Errorf("Expected length 2, got %d", list.Len())
	}
	
	val, _ := list.Get(1)
	if val != 3 {
		t.Errorf("Expected 3 at index 1, got %d", val)
	}
}

func TestLinkedList_Search(t *testing.T) {
	list := NewLinkedList()
	for i := 0; i < 5; i++ {
		list.Append(i * 10)
	}
	
	index := list.Search(20)
	if index != 2 {
		t.Errorf("Expected index 2, got %d", index)
	}
	
	index = list.Search(99)
	if index != -1 {
		t.Errorf("Expected -1 for not found, got %d", index)
	}
}

// ========== 性能基准测试 ==========

const benchSize = 10000

// Append 操作对比
func BenchmarkArray_Append(b *testing.B) {
	for i := 0; i < b.N; i++ {
		arr := NewArrayDS(benchSize)
		for j := 0; j < benchSize; j++ {
			arr.Append(j)
		}
	}
}

func BenchmarkLinkedList_Append(b *testing.B) {
	for i := 0; i < b.N; i++ {
		list := NewLinkedList()
		for j := 0; j < benchSize; j++ {
			list.Append(j)
		}
	}
}

// 随机访问对比
func BenchmarkArray_RandomAccess(b *testing.B) {
	arr := NewArrayDS(benchSize)
	for i := 0; i < benchSize; i++ {
		arr.Append(i)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr.Get(i % benchSize)
	}
}

func BenchmarkLinkedList_RandomAccess(b *testing.B) {
	list := NewLinkedList()
	for i := 0; i < benchSize; i++ {
		list.Append(i)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list.Get(i % benchSize)
	}
}

// 顺序遍历对比
func BenchmarkArray_Traverse(b *testing.B) {
	arr := NewArrayDS(benchSize)
	for i := 0; i < benchSize; i++ {
		arr.Append(i)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < arr.Len(); j++ {
			val, _ := arr.Get(j)
			sum += val
		}
	}
}

func BenchmarkLinkedList_Traverse(b *testing.B) {
	list := NewLinkedList()
	for i := 0; i < benchSize; i++ {
		list.Append(i)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < list.Len(); j++ {
			val, _ := list.Get(j)
			sum += val
		}
	}
}

// 查找操作对比
func BenchmarkArray_Search(b *testing.B) {
	arr := NewArrayDS(benchSize)
	for i := 0; i < benchSize; i++ {
		arr.Append(i)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr.Search(benchSize / 2)
	}
}

func BenchmarkLinkedList_Search(b *testing.B) {
	list := NewLinkedList()
	for i := 0; i < benchSize; i++ {
		list.Append(i)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list.Search(benchSize / 2)
	}
}

// 头部插入对比
func BenchmarkArray_InsertHead(b *testing.B) {
	arr := NewArrayDS(1000)
	for i := 0; i < 1000; i++ {
		arr.Append(i)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr.Insert(0, i)
		arr.Delete(0) // 保持大小不变
	}
}

func BenchmarkLinkedList_InsertHead(b *testing.B) {
	list := NewLinkedList()
	for i := 0; i < 1000; i++ {
		list.Append(i)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list.Insert(0, i)
		list.Delete(0) // 保持大小不变
	}
}

// 中间插入对比
func BenchmarkArray_InsertMiddle(b *testing.B) {
	arr := NewArrayDS(1000)
	for i := 0; i < 1000; i++ {
		arr.Append(i)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		arr.Insert(500, i)
		arr.Delete(500) // 保持大小不变
	}
}

func BenchmarkLinkedList_InsertMiddle(b *testing.B) {
	list := NewLinkedList()
	for i := 0; i < 1000; i++ {
		list.Append(i)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		list.Insert(500, i)
		list.Delete(500) // 保持大小不变
	}
}

// 内存分配对比
func BenchmarkArray_MemoryAllocation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		arr := NewArrayDS(100)
		for j := 0; j < 100; j++ {
			arr.Append(j)
		}
	}
}

func BenchmarkLinkedList_MemoryAllocation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		list := NewLinkedList()
		for j := 0; j < 100; j++ {
			list.Append(j)
		}
	}
}

// Cache locality 测试
func BenchmarkArray_CacheLocality_Sequential(b *testing.B) {
	size := 1000000
	arr := NewArrayDS(size)
	for i := 0; i < size; i++ {
		arr.Append(i)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		for j := 0; j < size; j++ {
			val, _ := arr.Get(j)
			sum += val
		}
	}
}

func BenchmarkLinkedList_CacheLocality_Sequential(b *testing.B) {
	size := 1000000
	list := NewLinkedList()
	for i := 0; i < size; i++ {
		list.Append(i)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum := 0
		curr := list.head
		for curr != nil {
			sum += curr.Val
			curr = curr.Next
		}
	}
}

