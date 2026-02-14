package generic_practices

import "testing"

func TestDemoOutput(t *testing.T) {
	nums := []int{1, 2, 3, 4, 5}
	squares := Map(nums, func(v int) int { return v * v })
	evens := Filter(squares, func(v int) bool { return v%2 == 0 })
	sum := Reduce(evens, 0, func(acc, v int) int { return acc + v })
	min, _ := Min(nums)
	max, _ := Max(nums)

	t.Logf("nums=%v", nums)
	t.Logf("squares=%v", squares)
	t.Logf("evens=%v sum=%d min=%d max=%d", evens, sum, min, max)
}
