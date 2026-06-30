package memory_model_practices

import "testing"

func TestPublisher(t *testing.T) {
	var p Publisher
	if _, ok := p.Load(); ok {
		t.Fatalf("expected empty")
	}

	p.Store(Config{Version: 1, Threshold: 10, Enabled: true})
	cfg, ok := p.Load()
	if !ok || cfg.Version != 1 || !cfg.Enabled {
		t.Fatalf("load mismatch")
	}
}

func TestOnceFlag(t *testing.T) {
	var f OnceFlag
	count := 0
	if !f.Do(func() { count++ }) {
		t.Fatalf("expected first call")
	}
	if f.Do(func() { count++ }) {
		t.Fatalf("expected only once")
	}
	if count != 1 {
		t.Fatalf("count mismatch")
	}
}
