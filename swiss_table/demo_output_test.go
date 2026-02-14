package swiss_table

import "testing"

func TestDemoOutput(t *testing.T) {
	st := NewSwissTable[string, int]()
	st.Put("a", 1)
	st.Put("b", 2)
	st.Put("c", 3)
	v, ok := st.Get("b")
	t.Logf("swiss get b => %d ok=%v len=%d keys=%v", v, ok, st.Len(), st.Keys())

	gm := NewGoMap[string, int]()
	gm.Put("a", 1)
	gm.Put("b", 2)
	t.Logf("gomap len=%d contains(b)=%v", gm.Len(), gm.Contains("b"))
}
