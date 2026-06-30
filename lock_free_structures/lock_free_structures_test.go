package lock_free_structures

import (
	"sync"
	"testing"
)

func TestStackLIFO(t *testing.T) {
	var s Stack[int]
	s.Push(1)
	s.Push(2)
	v, ok := s.Pop()
	if !ok || v != 2 {
		t.Fatalf("expected 2")
	}
	v, ok = s.Pop()
	if !ok || v != 1 {
		t.Fatalf("expected 1")
	}
}

func TestStackConcurrent(t *testing.T) {
	var s Stack[int]
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(v int) {
			defer wg.Done()
			s.Push(v)
		}(i)
	}
	wg.Wait()
	count := 0
	for {
		_, ok := s.Pop()
		if !ok {
			break
		}
		count++
	}
	if count != 100 {
		t.Fatalf("expected 100, got %d", count)
	}
}
