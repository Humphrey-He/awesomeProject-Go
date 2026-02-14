package copy_semantics

import "testing"

func TestDemoOutput(t *testing.T) {
	v1 := CopyValueSliceDemo()
	t.Logf("value copy: original=%v copied=%v note=%s", v1.Original, v1.Copied, v1.Note)

	v2 := CopyNestedSliceDemo()
	t.Logf("nested copy: original=%v copied=%v note=%s", v2.Original, v2.Copied, v2.Note)

	src := [][]int{{1, 2}, {3, 4}}
	dst := DeepCopy2DInt(src)
	src[0][0] = 99
	t.Logf("deep copy: src=%v dst=%v", src, dst)
}


