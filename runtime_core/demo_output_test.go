package runtime_core

import "testing"

func TestDemoOutput(t *testing.T) {
	s := TakeSnapshot()
	t.Logf("snapshot: cpu=%d gomaxprocs=%d goroutines=%d", s.NumCPU, s.GOMAXPROCS, s.NumGoroutine)

	before, after := ForceGCAndReadStats()
	t.Logf("gc stats: before_num_gc=%d after_num_gc=%d", before.NumGC, after.NumGC)

	m := NewGoroutineMonitor()
	t.Logf("goroutine delta=%d suspectedLeak=%v", m.Delta(), m.SuspectedLeak(100))

	cost := RunCPUTasks(4, 50000)
	t.Logf("cpu task cost=%v", cost)
}
