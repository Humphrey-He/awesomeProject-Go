package runtime_core

import (
	"runtime"
	"testing"
)

func TestTakeSnapshot(t *testing.T) {
	s := TakeSnapshot()
	if s.NumCPU <= 0 || s.GOMAXPROCS <= 0 || s.NumGoroutine <= 0 {
		t.Fatalf("invalid snapshot: %+v", s)
	}
}

func TestTuneForCPUBound(t *testing.T) {
	prev := TuneForCPUBound()
	defer runtime.GOMAXPROCS(prev)

	if got := runtime.GOMAXPROCS(0); got != runtime.NumCPU() {
		t.Fatalf("gomaxprocs=%d want=%d", got, runtime.NumCPU())
	}
}

func TestForceGCAndReadStats(t *testing.T) {
	before, after := ForceGCAndReadStats()
	if after.NumGC < before.NumGC {
		t.Fatalf("after.NumGC=%d should >= before.NumGC=%d", after.NumGC, before.NumGC)
	}
}

func TestGoroutineMonitor(t *testing.T) {
	m := NewGoroutineMonitor()
	if d := m.Delta(); d < 0 {
		t.Fatalf("delta should not be negative in stable test env, got %d", d)
	}
	_ = m.SuspectedLeak(1000)
}

func TestRunCPUTasks(t *testing.T) {
	d := RunCPUTasks(4, 300000)
	if d < 0 {
		t.Fatalf("duration should not be negative: %v", d)
	}
}
