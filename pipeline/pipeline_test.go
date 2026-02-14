package pipeline

import (
	"context"
	"sort"
	"testing"
	"time"
)

func TestPipeline_EndToEnd(t *testing.T) {
	ctx := context.Background()

	src := Source(ctx, []int{1, 2, 3, 4, 5, 6}, 2)
	mapped := MapStage(ctx, src, 3, 2, func(v int) int { return v * 2 })
	filtered := FilterStage(ctx, mapped, 1, func(v int) bool { return v%4 == 0 })
	got, err := Sink(ctx, filtered)
	if err != nil {
		t.Fatalf("sink error: %v", err)
	}

	sort.Ints(got)
	want := []int{4, 8, 12}
	if len(got) != len(want) {
		t.Fatalf("len mismatch: got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestPipeline_Cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	src := Source(ctx, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 1)
	mapped := MapStage(ctx, src, 2, 1, func(v int) int {
		time.Sleep(10 * time.Millisecond)
		return v
	})

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	_, err := Sink(ctx, mapped)
	if err == nil {
		t.Fatal("want canceled error")
	}
}

func TestPipeline_BackpressureWithSmallBuffer(t *testing.T) {
	ctx := context.Background()
	src := Source(ctx, []int{1, 2, 3, 4}, 0)
	mapped := MapStage(ctx, src, 1, 0, func(v int) int { return v + 1 })
	got, err := Sink(ctx, mapped)
	if err != nil {
		t.Fatalf("sink error: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("want 4 items, got %d", len(got))
	}
}

