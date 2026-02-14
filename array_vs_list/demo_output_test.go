package array_vs_list

import "testing"

func TestDemoOutput(t *testing.T) {
	arr := NewArrayDS(10)
	arr.Append(1)
	arr.Append(3)
	_ = arr.Insert(1, 2)
	t.Logf("array len=%d indexOf2=%d", arr.Len(), arr.Search(2))

	list := NewLinkedList()
	list.Append(1)
	list.Append(3)
	_ = list.Insert(1, 2)
	t.Logf("list len=%d indexOf2=%d", list.Len(), list.Search(2))

	t.Logf("why array better:")
	for _, reason := range WhyArrayIsBetter() {
		t.Logf("- %s", reason)
	}
	t.Logf("memory layout: %s", MemoryLayoutComparison())
}
