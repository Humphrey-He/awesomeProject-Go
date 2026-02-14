package config_registry

import (
	"testing"
	"time"
)

func TestDemoOutput(t *testing.T) {
	cc := NewConfigCenter()
	watchCh, cancel := cc.Watch("feature.payment", 4)
	defer cancel()

	cc.Set("feature.payment", "off")
	cc.Set("feature.payment", "on")
	e1 := <-watchCh
	e2 := <-watchCh
	t.Logf("config event1: key=%s val=%s ver=%d", e1.Key, e1.Value, e1.Version)
	t.Logf("config event2: key=%s val=%s ver=%d", e2.Key, e2.Value, e2.Version)

	reg := NewRegistry(20 * time.Millisecond)
	defer reg.Stop()
	reg.Register("user-service", "u1", "10.0.0.1:9000", 70*time.Millisecond, map[string]string{"zone": "az1"})
	t.Logf("discover after register: %+v", reg.Discover("user-service"))

	time.Sleep(90 * time.Millisecond)
	t.Logf("discover after expire: %+v", reg.Discover("user-service"))
}
