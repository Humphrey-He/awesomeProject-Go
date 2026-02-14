package generic_practices

import "sync"

// Ordered is a local constraint for sortable primitives.
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~string
}

// Map transforms []T to []R while keeping order.
func Map[T any, R any](in []T, fn func(T) R) []R {
	out := make([]R, len(in))
	for i := range in {
		out[i] = fn(in[i])
	}
	return out
}

// Filter keeps elements satisfying predicate.
func Filter[T any](in []T, pred func(T) bool) []T {
	out := make([]T, 0, len(in))
	for i := range in {
		if pred(in[i]) {
			out = append(out, in[i])
		}
	}
	return out
}

// Reduce folds []T into R.
func Reduce[T any, R any](in []T, init R, fn func(R, T) R) R {
	acc := init
	for i := range in {
		acc = fn(acc, in[i])
	}
	return acc
}

func Min[T Ordered](in []T) (T, bool) {
	var zero T
	if len(in) == 0 {
		return zero, false
	}
	m := in[0]
	for i := 1; i < len(in); i++ {
		if in[i] < m {
			m = in[i]
		}
	}
	return m, true
}

func Max[T Ordered](in []T) (T, bool) {
	var zero T
	if len(in) == 0 {
		return zero, false
	}
	m := in[0]
	for i := 1; i < len(in); i++ {
		if in[i] > m {
			m = in[i]
		}
	}
	return m, true
}

// Set is a generic set requiring comparable key.
type Set[T comparable] struct {
	m map[T]struct{}
}

func NewSet[T comparable](capHint int) *Set[T] {
	if capHint < 0 {
		capHint = 0
	}
	return &Set[T]{m: make(map[T]struct{}, capHint)}
}

func (s *Set[T]) Add(v T)      { s.m[v] = struct{}{} }
func (s *Set[T]) Has(v T) bool { _, ok := s.m[v]; return ok }
func (s *Set[T]) Len() int     { return len(s.m) }

// SafeCache demonstrates generics + concurrency.
type SafeCache[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

func NewSafeCache[K comparable, V any]() *SafeCache[K, V] {
	return &SafeCache[K, V]{m: make(map[K]V)}
}

func (c *SafeCache[K, V]) Set(k K, v V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[k] = v
}

func (c *SafeCache[K, V]) Get(k K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.m[k]
	return v, ok
}


