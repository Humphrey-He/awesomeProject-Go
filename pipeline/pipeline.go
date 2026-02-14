package pipeline

import (
	"context"
	"sync"
)

// Source emits values to downstream with optional buffer to control backpressure.
func Source(ctx context.Context, values []int, buffer int) <-chan int {
	out := make(chan int, max(0, buffer))
	go func() {
		defer close(out)
		for _, v := range values {
			select {
			case <-ctx.Done():
				return
			case out <- v:
			}
		}
	}()
	return out
}

// MapStage transforms input in parallel with fixed workers.
func MapStage(ctx context.Context, in <-chan int, workers, buffer int, fn func(int) int) <-chan int {
	if workers <= 0 {
		workers = 1
	}
	out := make(chan int, max(0, buffer))
	var wg sync.WaitGroup
	wg.Add(workers)

	worker := func() {
		defer wg.Done()
		for v := range in {
			mapped := fn(v)
			select {
			case <-ctx.Done():
				return
			case out <- mapped:
			}
		}
	}

	for i := 0; i < workers; i++ {
		go worker()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// FilterStage keeps values that satisfy predicate.
func FilterStage(ctx context.Context, in <-chan int, buffer int, predicate func(int) bool) <-chan int {
	out := make(chan int, max(0, buffer))
	go func() {
		defer close(out)
		for v := range in {
			if !predicate(v) {
				continue
			}
			select {
			case <-ctx.Done():
				return
			case out <- v:
			}
		}
	}()
	return out
}

// Sink collects all values from upstream.
func Sink(ctx context.Context, in <-chan int) ([]int, error) {
	var out []int
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case v, ok := <-in:
			if !ok {
				return out, nil
			}
			out = append(out, v)
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

