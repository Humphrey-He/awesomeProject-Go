package lsm_tree

import (
	"sort"
	"sync"
)

type Value struct {
	Data      []byte
	Tombstone bool
}

type KV struct {
	Key       string
	Value     []byte
	Tombstone bool
}

type SSTable struct {
	items []KV
}

func (s SSTable) Get(key string) (Value, bool) {
	idx := sort.Search(len(s.items), func(i int) bool {
		return s.items[i].Key >= key
	})
	if idx < len(s.items) && s.items[idx].Key == key {
		item := s.items[idx]
		return Value{Data: item.Value, Tombstone: item.Tombstone}, true
	}
	return Value{}, false
}

type LSMTree struct {
	mu           sync.RWMutex
	mem          map[string]Value
	sstables     []SSTable // newest first
	memSize      int
	memThreshold int
}

func NewLSMTree(memThreshold int) *LSMTree {
	if memThreshold <= 0 {
		memThreshold = 1024
	}
	return &LSMTree{
		mem:          map[string]Value{},
		memThreshold: memThreshold,
	}
}

func (t *LSMTree) Put(key string, value []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.mem[key] = Value{Data: value}
	t.memSize++
	if t.memSize >= t.memThreshold {
		t.flushLocked()
	}
}

func (t *LSMTree) Delete(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.mem[key] = Value{Tombstone: true}
	t.memSize++
	if t.memSize >= t.memThreshold {
		t.flushLocked()
	}
}

func (t *LSMTree) Get(key string) ([]byte, bool) {
	t.mu.RLock()
	if v, ok := t.mem[key]; ok {
		t.mu.RUnlock()
		if v.Tombstone {
			return nil, false
		}
		return v.Data, true
	}
	t.mu.RUnlock()

	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, s := range t.sstables {
		if v, ok := s.Get(key); ok {
			if v.Tombstone {
				return nil, false
			}
			return v.Data, true
		}
	}
	return nil, false
}

func (t *LSMTree) Flush() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.flushLocked()
}

func (t *LSMTree) flushLocked() {
	if len(t.mem) == 0 {
		return
	}
	items := make([]KV, 0, len(t.mem))
	for k, v := range t.mem {
		items = append(items, KV{Key: k, Value: v.Data, Tombstone: v.Tombstone})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Key < items[j].Key })
	t.sstables = append([]SSTable{{items: items}}, t.sstables...)
	t.mem = map[string]Value{}
	t.memSize = 0
}

func (t *LSMTree) Compact() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.sstables) <= 1 {
		return
	}
	latest := map[string]KV{}
	for _, s := range t.sstables {
		for _, kv := range s.items {
			if _, ok := latest[kv.Key]; !ok {
				latest[kv.Key] = kv
			}
		}
	}
	items := make([]KV, 0, len(latest))
	for _, kv := range latest {
		items = append(items, kv)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Key < items[j].Key })
	t.sstables = []SSTable{{items: items}}
}
