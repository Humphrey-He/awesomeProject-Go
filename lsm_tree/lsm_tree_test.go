package lsm_tree

import "testing"

func TestLSMPutGet(t *testing.T) {
	lsm := NewLSMTree(2)
	lsm.Put("a", []byte("1"))
	lsm.Put("b", []byte("2"))
	if v, ok := lsm.Get("a"); !ok || string(v) != "1" {
		t.Fatalf("expected value")
	}

	lsm.Delete("a")
	if _, ok := lsm.Get("a"); ok {
		t.Fatalf("expected tombstone")
	}
}

func TestCompact(t *testing.T) {
	lsm := NewLSMTree(1)
	lsm.Put("k", []byte("v1"))
	lsm.Flush()
	lsm.Put("k", []byte("v2"))
	lsm.Flush()
	lsm.Compact()
	if v, ok := lsm.Get("k"); !ok || string(v) != "v2" {
		t.Fatalf("compact mismatch")
	}
}
