package pipeline

import (
	"context"
	"testing"
)

func TestDemoOutput(t *testing.T) {
	ctx := context.Background()
	src := Source(ctx, []int{1, 2, 3, 4, 5}, 2)
	mapped := MapStage(ctx, src, 2, 2, func(v int) int { return v * 3 })
	filtered := FilterStage(ctx, mapped, 2, func(v int) bool { return v%2 == 0 })
	out, err := Sink(ctx, filtered)
	t.Logf("pipeline output=%v err=%v", out, err)
}
