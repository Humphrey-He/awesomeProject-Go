package etcd_mvp

import (
	"testing"
	"time"
)

func TestDemoOutput(t *testing.T) {
	c := NewCluster(50 * time.Millisecond)
	_ = c.AddMember("node-2")
	_ = c.AddMember("node-1")
	_ = c.AddMember("node-3")
	leader, _ := c.Leader()
	t.Logf("initial leader=%s members=%+v", leader, c.SnapshotMembers())

	_ = c.Put(leader, "config/app", "v1")
	v, _ := c.Get("config/app")
	t.Logf("leader put/get key=config/app val=%s", v)

	time.Sleep(60 * time.Millisecond)
	c.Tick(time.Now())
	leader2, _ := c.Leader()
	t.Logf("after timeout reelected leader=%s members=%+v", leader2, c.SnapshotMembers())
}


