package log_structured_storage

import "testing"

func TestLogStorePutGet(t *testing.T) {
	store := NewLogStore(2)
	store.Put("a", []byte("1"))
	store.Put("b", []byte("2"))

	if v, ok := store.Get("a"); !ok || string(v) != "1" {
		t.Fatalf("get mismatch")
	}
	store.Delete("a")
	if _, ok := store.Get("a"); ok {
		t.Fatalf("expected deleted")
	}
}

func TestLogStoreCompact(t *testing.T) {
	store := NewLogStore(1)
	store.Put("k", []byte("v1"))
	store.Put("k", []byte("v2"))
	store.Compact()
	if v, ok := store.Get("k"); !ok || string(v) != "v2" {
		t.Fatalf("compact mismatch")
	}
	if len(store.segments) != 1 {
		t.Fatalf("expected single segment")
	}
}
