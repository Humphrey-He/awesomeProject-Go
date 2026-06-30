package lock_free_structures

import "testing"

func TestDemoOutput(t *testing.T) {
	var s Stack[string]
	s.Push("a")
	s.Push("b")
	v, _ := s.Pop()
	t.Logf("pop=%s", v)
}
