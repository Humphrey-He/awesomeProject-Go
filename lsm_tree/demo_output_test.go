package lsm_tree

import "testing"

func TestDemoOutput(t *testing.T) {
	lsm := NewLSMTree(2)
	lsm.Put("user:1", []byte("alice"))
	lsm.Put("user:2", []byte("bob"))
	lsm.Flush()
	v, _ := lsm.Get("user:1")
	t.Logf("user:1=%s sstables=%d", string(v), len(lsm.sstables))
}
