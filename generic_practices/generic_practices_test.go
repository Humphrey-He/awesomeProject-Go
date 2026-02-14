package generic_practices

import "testing"

func TestMapFilterReduce(t *testing.T) {
	in := []int{1, 2, 3, 4}
	sq := Map(in, func(v int) int { return v * v })
	if sq[2] != 9 {
		t.Fatalf("sq[2]=%d", sq[2])
	}
	evens := Filter(sq, func(v int) bool { return v%2 == 0 })
	if len(evens) != 2 {
		t.Fatalf("evens len=%d", len(evens))
	}
	sum := Reduce(evens, 0, func(acc, v int) int { return acc + v })
	if sum != 20 {
		t.Fatalf("sum=%d want=20", sum)
	}
}

func TestMinMax(t *testing.T) {
	min, ok := Min([]int{3, 9, 1, 7})
	if !ok || min != 1 {
		t.Fatalf("min=%d ok=%v", min, ok)
	}
	max, ok := Max([]string{"a", "z", "b"})
	if !ok || max != "z" {
		t.Fatalf("max=%s ok=%v", max, ok)
	}
}

func TestSetAndSafeCache(t *testing.T) {
	s := NewSet[string](2)
	s.Add("a")
	s.Add("b")
	if !s.Has("a") || s.Len() != 2 {
		t.Fatalf("set incorrect")
	}

	c := NewSafeCache[string, int]()
	c.Set("x", 1)
	v, ok := c.Get("x")
	if !ok || v != 1 {
		t.Fatalf("cache get failed")
	}
}


