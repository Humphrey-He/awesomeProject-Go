package etcd_mvp

import (
	"errors"
	"testing"
	"time"
)

func TestElectionAndHeartbeat(t *testing.T) {
	c := NewCluster(50 * time.Millisecond)
	_ = c.AddMember("n2")
	_ = c.AddMember("n1")
	_ = c.AddMember("n3")

	leader, err := c.Leader()
	if err != nil {
		t.Fatalf("leader err: %v", err)
	}
	if leader != "n2" { // first add bootstrap leader
		t.Fatalf("unexpected initial leader=%s", leader)
	}

	// keep heartbeat, should stay leader
	time.Sleep(20 * time.Millisecond)
	if err := c.Heartbeat("n2"); err != nil {
		t.Fatalf("heartbeat err: %v", err)
	}
	c.Tick(time.Now())
	leader2, _ := c.Leader()
	if leader2 != "n2" {
		t.Fatalf("leader should remain n2, got %s", leader2)
	}
}

func TestLeaderTimeoutReelection(t *testing.T) {
	c := NewCluster(30 * time.Millisecond)
	_ = c.AddMember("n2")
	_ = c.AddMember("n1")
	_ = c.AddMember("n3")

	time.Sleep(40 * time.Millisecond)
	c.Tick(time.Now())
	leader, err := c.Leader()
	if err != nil {
		t.Fatalf("leader err: %v", err)
	}
	if leader != "n1" {
		t.Fatalf("expected deterministic re-election winner n1, got %s", leader)
	}
}

func TestLeaderOnlyWrite(t *testing.T) {
	c := NewCluster(100 * time.Millisecond)
	_ = c.AddMember("n1")
	_ = c.AddMember("n2")

	if err := c.Put("n2", "k", "v"); !errors.Is(err, ErrNotLeader) {
		t.Fatalf("expected not leader err, got %v", err)
	}
	if err := c.Put("n1", "k", "v"); err != nil {
		t.Fatalf("leader put err: %v", err)
	}
	if v, ok := c.Get("k"); !ok || v != "v" {
		t.Fatalf("get mismatch v=%s ok=%v", v, ok)
	}
}


