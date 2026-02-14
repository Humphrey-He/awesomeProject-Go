package copy_semantics

import "testing"

func TestCopyValueSliceDemo(t *testing.T) {
	obs := CopyValueSliceDemo()
	src := obs.Original.([]int)
	dst := obs.Copied.([]int)

	if src[0] == dst[0] {
		t.Fatalf("value slice should be independent after copy, src=%v dst=%v", src, dst)
	}
	if dst[0] != 1 {
		t.Fatalf("dst[0]=%d want=1", dst[0])
	}
}

func TestCopyNestedSliceDemo(t *testing.T) {
	obs := CopyNestedSliceDemo()
	src := obs.Original.([][]int)
	dst := obs.Copied.([][]int)

	if src[0][0] != dst[0][0] {
		t.Fatalf("nested slice should share inner data in shallow copy, src=%v dst=%v", src, dst)
	}
	if dst[0][0] != 999 {
		t.Fatalf("dst inner should reflect change, got %d", dst[0][0])
	}
}

func TestDeepCopy2DInt(t *testing.T) {
	src := [][]int{{1, 2}, {3, 4}}
	dst := DeepCopy2DInt(src)

	src[0][0] = 77
	if dst[0][0] != 1 {
		t.Fatalf("deep copy should isolate inner arrays, got=%d", dst[0][0])
	}
}

func TestCopyStructWithPointerDemo(t *testing.T) {
	obs := CopyStructWithPointerDemo()
	dst := obs.Copied.([]Wrapper)
	if dst[0].Ptr.Value != 42 {
		t.Fatalf("pointer field target should be shared in shallow copy")
	}
}
