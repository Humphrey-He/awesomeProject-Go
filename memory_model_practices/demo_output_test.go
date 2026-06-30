package memory_model_practices

import "testing"

func TestDemoOutput(t *testing.T) {
	var p Publisher
	p.Store(Config{Version: 42, Threshold: 100, Enabled: true})
	cfg, _ := p.Load()
	t.Logf("version=%d threshold=%d enabled=%v", cfg.Version, cfg.Threshold, cfg.Enabled)
}
