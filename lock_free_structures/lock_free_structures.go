package lock_free_structures

import "sync/atomic"

type node[T any] struct {
	value T
	next  *node[T]
}

type Stack[T any] struct {
	head atomic.Pointer[node[T]]
}

func (s *Stack[T]) Push(v T) {
	for {
		old := s.head.Load()
		n := &node[T]{value: v, next: old}
		if s.head.CompareAndSwap(old, n) {
			return
		}
	}
}

func (s *Stack[T]) Pop() (T, bool) {
	for {
		old := s.head.Load()
		if old == nil {
			var zero T
			return zero, false
		}
		next := old.next
		if s.head.CompareAndSwap(old, next) {
			return old.value, true
		}
	}
}
