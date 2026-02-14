package fan_patterns

import (
	"context"
	"sync"
)

// JobResult is an aggregated result for fan-out workers.
type JobResult struct {
	Input  int
	Output int
	Err    error
}

// FanOut starts fixed workers to process jobs in parallel.
func FanOut(ctx context.Context, workers int, jobs <-chan int, fn func(int) (int, error)) <-chan JobResult {
	if workers <= 0 {
		workers = 1
	}

	out := make(chan JobResult)
	var wg sync.WaitGroup
	wg.Add(workers)

	worker := func() {
		defer wg.Done()
		for job := range jobs {
			select {
			case <-ctx.Done():
				return
			default:
			}

			v, err := fn(job)
			res := JobResult{Input: job, Output: v, Err: err}
			select {
			case <-ctx.Done():
				return
			case out <- res:
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

// FanIn merges many channels into one output channel.
func FanIn[T any](ctx context.Context, inputs ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup
	wg.Add(len(inputs))

	for _, ch := range inputs {
		c := ch
		go func() {
			defer wg.Done()
			for v := range c {
				select {
				case <-ctx.Done():
					return
				case out <- v:
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}


