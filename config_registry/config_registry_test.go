package config_registry

import (
	"testing"
	"time"
)

func TestConfigCenter_SetGetWatch(t *testing.T) {
	cc := NewConfigCenter()
	ch, cancel := cc.Watch("db.dsn", 2)
	defer cancel()

	ev := cc.Set("db.dsn", "mysql://root@127.0.0.1")
	if ev.Version != 1 {
		t.Fatalf("version=%d want=1", ev.Version)
	}
	ev2 := cc.Set("db.dsn", "mysql://root@localhost")
	if ev2.Version != 2 {
		t.Fatalf("version=%d want=2", ev2.Version)
	}

	got, err := cc.Get("db.dsn")
	if err != nil || got.Value != "mysql://root@localhost" {
		t.Fatalf("get failed: %+v err=%v", got, err)
	}

	recv1 := <-ch
	recv2 := <-ch
	if recv1.Version != 1 || recv2.Version != 2 {
		t.Fatalf("unexpected watch events: %+v %+v", recv1, recv2)
	}
}

func TestRegistry_RegisterHeartbeatExpire(t *testing.T) {
	r := NewRegistry(20 * time.Millisecond)
	defer r.Stop()

	r.Register("order", "ins-1", "10.0.0.1:8080", 80*time.Millisecond, map[string]string{"zone": "a"})
	if list := r.Discover("order"); len(list) != 1 {
		t.Fatalf("discover len=%d want=1", len(list))
	}

	time.Sleep(50 * time.Millisecond)
	if err := r.Heartbeat("order", "ins-1", 80*time.Millisecond); err != nil {
		t.Fatalf("heartbeat err: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	if list := r.Discover("order"); len(list) != 1 {
		t.Fatalf("instance should still be alive, len=%d", len(list))
	}

	time.Sleep(120 * time.Millisecond)
	if list := r.Discover("order"); len(list) != 0 {
		t.Fatalf("instance should expire, len=%d", len(list))
	}
}

func TestRegistry_Deregister(t *testing.T) {
	r := NewRegistry(50 * time.Millisecond)
	defer r.Stop()
	r.Register("stock", "s1", "10.0.0.2:8081", time.Second, nil)
	r.Deregister("stock", "s1")
	if list := r.Discover("stock"); len(list) != 0 {
		t.Fatalf("expect 0 instance after deregister, got %d", len(list))
	}
}
