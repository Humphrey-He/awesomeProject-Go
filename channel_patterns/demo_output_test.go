package channel_patterns

import (
	"testing"
	"time"
)

func TestDemoOutput(t *testing.T) {
	ch := NewBufferedChannel[int](2)
	ch <- 1
	ch <- 2
	t.Logf("buffered receive1=%d receive2=%d", <-ch, <-ch)

	p := NewPipeline[int]()
	p.AddStage(func(n int) int { return n * 2 })
	p.AddStage(func(n int) int { return n + 1 })
	input := make(chan int, 3)
	input <- 1
	input <- 2
	input <- 3
	close(input)
	for v := range p.Execute(input) {
		t.Logf("pipeline out=%d", v)
	}

	pool := NewWorkerPool[int, int](2, 5, func(n int) int { return n * 10 })
	pool.Start()
	_ = pool.Submit(1)
	_ = pool.Submit(2)
	time.Sleep(30 * time.Millisecond)
	pool.Stop()
	for v := range pool.Results() {
		t.Logf("worker pool out=%d", v)
	}
}


