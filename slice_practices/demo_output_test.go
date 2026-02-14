package slice_practices

import "testing"

func TestDemoOutput(t *testing.T) {
	base := []int{1, 2, 3, 4, 5}
	t.Logf("base=%v", base)
	t.Logf("remove idx2 => %v", RemoveElement(append([]int(nil), base...), 2))
	t.Logf("remove fast idx2 => %v", RemoveElementFast(append([]int(nil), base...), 2))
	t.Logf("filter even => %v", FilterGood(base, func(v int) bool { return v%2 == 0 }))
	t.Logf("unique ordered => %v", UniqueOrdered([]int{1, 2, 2, 3, 1, 4}))
	ReverseInPlace(base)
	t.Logf("reversed => %v", base)
}


