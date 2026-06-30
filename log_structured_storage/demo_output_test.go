package log_structured_storage

import "testing"

func TestDemoOutput(t *testing.T) {
	store := NewLogStore(2)
	store.Put("u:1", []byte("alice"))
	store.Put("u:2", []byte("bob"))
	v, _ := store.Get("u:2")
	t.Logf("segments=%d u:2=%s", len(store.segments), string(v))
}
