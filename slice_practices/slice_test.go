package slice_practices

import (
	"testing"
)

// ========== 功能测试 ==========

func TestAppendWithPrealloc(t *testing.T) {
	n := 1000
	result := AppendWithPrealloc(n)

	if len(result) != n {
		t.Errorf("Expected length %d, got %d", n, len(result))
	}

	for i, v := range result {
		if v != i {
			t.Errorf("Expected %d at index %d, got %d", i, i, v)
		}
	}
}

func TestRemoveElement(t *testing.T) {
	tests := []struct {
		name  string
		slice []int
		index int
		want  []int
	}{
		{"remove first", []int{1, 2, 3, 4}, 0, []int{2, 3, 4}},
		{"remove middle", []int{1, 2, 3, 4}, 2, []int{1, 2, 4}},
		{"remove last", []int{1, 2, 3, 4}, 3, []int{1, 2, 3}},
		{"out of bounds", []int{1, 2, 3}, 10, []int{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveElement(append([]int(nil), tt.slice...), tt.index)
			if !sliceEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveElementFast(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	result := RemoveElementFast(s, 2) // 移除 3

	if len(result) != 4 {
		t.Errorf("Expected length 4, got %d", len(result))
	}

	// 快速删除不保持顺序，但长度正确
	if !contains(result, 1) || !contains(result, 2) || !contains(result, 4) || !contains(result, 5) {
		t.Errorf("Missing expected elements in %v", result)
	}
}

func TestFilterGood(t *testing.T) {
	s := []int{1, 2, 3, 4, 5, 6}
	result := FilterGood(s, func(v int) bool {
		return v%2 == 0
	})

	want := []int{2, 4, 6}
	if !sliceEqual(result, want) {
		t.Errorf("got %v, want %v", result, want)
	}
}

func TestUniqueOrdered(t *testing.T) {
	s := []int{1, 2, 2, 3, 1, 4, 3, 5}
	result := UniqueOrdered(s)

	want := []int{1, 2, 3, 4, 5}
	if !sliceEqual(result, want) {
		t.Errorf("got %v, want %v", result, want)
	}
}

func TestReverseInPlace(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	ReverseInPlace(s)

	want := []int{5, 4, 3, 2, 1}
	if !sliceEqual(s, want) {
		t.Errorf("got %v, want %v", s, want)
	}
}

func TestGenericFunctions(t *testing.T) {
	// Test Map
	ints := []int{1, 2, 3}
	strings := Map(ints, func(v int) string {
		return string(rune('A' + v - 1))
	})
	wantStrings := []string{"A", "B", "C"}
	if !sliceEqualGeneric(strings, wantStrings) {
		t.Errorf("Map: got %v, want %v", strings, wantStrings)
	}

	// Test Filter
	nums := []int{1, 2, 3, 4, 5, 6}
	evens := Filter(nums, func(v int) bool { return v%2 == 0 })
	wantEvens := []int{2, 4, 6}
	if !sliceEqual(evens, wantEvens) {
		t.Errorf("Filter: got %v, want %v", evens, wantEvens)
	}

	// Test Reduce
	sum := Reduce(nums, 0, func(acc, v int) int { return acc + v })
	wantSum := 21
	if sum != wantSum {
		t.Errorf("Reduce: got %d, want %d", sum, wantSum)
	}

	// Test Contains
	if !Contains(nums, 3) {
		t.Error("Contains: should contain 3")
	}
	if Contains(nums, 10) {
		t.Error("Contains: should not contain 10")
	}
}

// ========== 性能测试 ==========

func BenchmarkAppendWithoutPrealloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		AppendWithoutPrealloc(1000)
	}
}

func BenchmarkAppendWithPrealloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		AppendWithPrealloc(1000)
	}
}

func BenchmarkRemoveElement(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmp := append([]int(nil), s...)
		_ = RemoveElement(tmp, 500)
	}
}

func BenchmarkRemoveElementFast(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmp := append([]int(nil), s...)
		_ = RemoveElementFast(tmp, 500)
	}
}

func BenchmarkFilterBad(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}
	keep := func(v int) bool { return v%2 == 0 }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmp := append([]int(nil), s...)
		_ = FilterBad(tmp, keep)
	}
}

func BenchmarkFilterGood(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}
	keep := func(v int) bool { return v%2 == 0 }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = FilterGood(s, keep)
	}
}

func BenchmarkCopy(b *testing.B) {
	src := make([]int, 1000)
	dst := make([]int, 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(dst, src)
	}
}

func BenchmarkSumWithRange(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SumWithRange(s)
	}
}

func BenchmarkSumWithIndex(b *testing.B) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SumWithIndex(s)
	}
}

func BenchmarkAppendInLoopBad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = AppendInLoopBad(100)
	}
}

func BenchmarkAppendInLoopGood(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = AppendInLoopGood(100)
	}
}

// ========== 辅助函数 ==========

func sliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sliceEqualGeneric[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func contains(s []int, v int) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
