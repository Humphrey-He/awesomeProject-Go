package fan_patterns

import (
	"context"
	"testing"
)

func TestDemoOutput(t *testing.T) {
	ctx := context.Background()
	jobs := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		jobs <- i
	}
	close(jobs)

	results := FanOut(ctx, 3, jobs, func(v int) (int, error) { return v * v, nil })
	for r := range results {
		t.Logf("fanout input=%d output=%d err=%v", r.Input, r.Output, r.Err)
	}
}
