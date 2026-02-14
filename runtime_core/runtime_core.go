package runtime_core

import (
	"runtime"
	"sync"
	"time"
)

// Snapshot captures key runtime status.
type Snapshot struct {
	NumCPU       int
	GOMAXPROCS   int
	NumGoroutine int
}

func TakeSnapshot() Snapshot {
	return Snapshot{
		NumCPU:       runtime.NumCPU(),
		GOMAXPROCS:   runtime.GOMAXPROCS(0),
		NumGoroutine: runtime.NumGoroutine(),
	}
}

// TuneForCPUBound sets GOMAXPROCS to NumCPU and returns previous value.
func TuneForCPUBound() (prev int) {
	return runtime.GOMAXPROCS(runtime.NumCPU())
}

// ForceGCAndReadStats triggers GC and returns memory stats before/after.
func ForceGCAndReadStats() (before, after runtime.MemStats) {
	runtime.ReadMemStats(&before)
	runtime.GC()
	runtime.ReadMemStats(&after)
	return before, after
}

// Yield allows other goroutines to run (cooperative scheduling hint).
func Yield() {
	runtime.Gosched()
}

// GoroutineMonitor is a lightweight leak early-warning helper.
type GoroutineMonitor struct {
	baseline int
}

func NewGoroutineMonitor() *GoroutineMonitor {
	return &GoroutineMonitor{baseline: runtime.NumGoroutine()}
}

func (m *GoroutineMonitor) Delta() int {
	return runtime.NumGoroutine() - m.baseline
}

func (m *GoroutineMonitor) SuspectedLeak(threshold int) bool {
	return m.Delta() > threshold
}

// RunCPUTasks runs fixed number of CPU tasks in parallel and returns elapsed time.
// Practical use: compare different GOMAXPROCS settings in benchmarks.
func RunCPUTasks(taskCount int, iterations int) time.Duration {
	if taskCount <= 0 {
		taskCount = 1
	}
	if iterations <= 0 {
		iterations = 100000
	}

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(taskCount)
	for i := 0; i < taskCount; i++ {
		go func(seed int) {
			defer wg.Done()
			x := seed + 1
			for j := 0; j < iterations; j++ {
				x = (x*1664525 + 1013904223) & 0x7fffffff
				if j%4096 == 0 {
					Yield()
				}
			}
			_ = x
		}(i)
	}
	wg.Wait()
	return time.Since(start)
}
