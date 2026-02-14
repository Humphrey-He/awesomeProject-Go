package fan_patterns

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"
)

func TestFanOutAndAggregate(t *testing.T) {
	ctx := context.Background()
	jobs := make(chan int, 8)
	for i := 1; i <= 8; i++ {
		jobs <- i
	}
	close(jobs)

	results := FanOut(ctx, 3, jobs, func(v int) (int, error) {
		time.Sleep(5 * time.Millisecond)
		return v * v, nil
	})

	var got []int
	for r := range results {
		if r.Err != nil {
			t.Fatalf("unexpected error: %v", r.Err)
		}
		got = append(got, r.Output)
	}

	sort.Ints(got)
	want := []int{1, 4, 9, 16, 25, 36, 49, 64}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got=%v want=%v", got, want)
		}
	}
}

func TestFanOutError(t *testing.T) {
	ctx := context.Background()
	jobs := make(chan int, 3)
	jobs <- 1
	jobs <- 2
	jobs <- 3
	close(jobs)

	results := FanOut(ctx, 2, jobs, func(v int) (int, error) {
		if v == 2 {
			return 0, errors.New("boom")
		}
		return v, nil
	})

	foundErr := false
	for r := range results {
		if r.Err != nil {
			foundErr = true
		}
	}
	if !foundErr {
		t.Fatal("want at least one error result")
	}
}

func TestFanIn(t *testing.T) {
	ctx := context.Background()

	a := make(chan int, 2)
	b := make(chan int, 2)
	a <- 1
	a <- 2
	b <- 3
	b <- 4
	close(a)
	close(b)

	var got []int
	for v := range FanIn(ctx, a, b) {
		got = append(got, v)
	}

	sort.Ints(got)
	want := []int{1, 2, 3, 4}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got=%v want=%v", got, want)
		}
	}
}


