package multi_level_cache

import (
	"container/list"
	"errors"
	"sync"
	"time"

	"awesomeProject/singleflight"
)

var ErrNotFound = errors.New("cache: key not found")

// Backend is the L2 storage abstraction.
type Backend interface {
	Get(key string) (string, error)
	Set(key, value string, ttl time.Duration) error
	Delete(key string) error
}

type l1Entry struct {
	key      string
	value    string
	expireAt time.Time
	elem     *list.Element
}

// MultiLevelCache: L1(memory) + L2(backend) + singleflight anti-stampede.
type MultiLevelCache struct {
	mu       sync.RWMutex
	l1       map[string]*l1Entry
	lru      *list.List
	capacity int

	sf *singleflight.Singleflight
	l2 Backend
}

func New(capacity int, l2 Backend) *MultiLevelCache {
	if capacity <= 0 {
		capacity = 128
	}
	return &MultiLevelCache{
		l1:       make(map[string]*l1Entry, capacity),
		lru:      list.New(),
		capacity: capacity,
		sf:       singleflight.New(),
		l2:       l2,
	}
}

func (c *MultiLevelCache) Get(key string) (string, error) {
	// Fast path: L1 hit
	if v, ok := c.getL1(key); ok {
		return v, nil
	}

	// Slow path: deduplicate concurrent misses.
	val, err, _ := c.sf.Do(key, func() (interface{}, error) {
		// Double-check L1 after entering singleflight.
		if v, ok := c.getL1(key); ok {
			return v, nil
		}
		if c.l2 == nil {
			return "", ErrNotFound
		}
		v, err := c.l2.Get(key)
		if err != nil {
			return "", err
		}
		// Backfill L1 with default short ttl.
		c.setL1(key, v, 30*time.Second)
		return v, nil
	})
	if err != nil {
		return "", err
	}
	return val.(string), nil
}

func (c *MultiLevelCache) Set(key, value string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	c.setL1(key, value, ttl)
	if c.l2 != nil {
		return c.l2.Set(key, value, ttl)
	}
	return nil
}

func (c *MultiLevelCache) Delete(key string) error {
	c.mu.Lock()
	if e, ok := c.l1[key]; ok {
		c.lru.Remove(e.elem)
		delete(c.l1, key)
	}
	c.mu.Unlock()

	if c.l2 != nil {
		return c.l2.Delete(key)
	}
	return nil
}

func (c *MultiLevelCache) Stats() (l1Size int, l1Cap int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.l1), c.capacity
}

func (c *MultiLevelCache) getL1(key string) (string, bool) {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.l1[key]
	if !ok {
		return "", false
	}
	if now.After(e.expireAt) {
		c.lru.Remove(e.elem)
		delete(c.l1, key)
		return "", false
	}
	c.lru.MoveToFront(e.elem)
	return e.value, true
}

func (c *MultiLevelCache) setL1(key, value string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.l1[key]; ok {
		e.value = value
		e.expireAt = time.Now().Add(ttl)
		c.lru.MoveToFront(e.elem)
		return
	}

	elem := c.lru.PushFront(key)
	c.l1[key] = &l1Entry{
		key:      key,
		value:    value,
		expireAt: time.Now().Add(ttl),
		elem:     elem,
	}
	if len(c.l1) > c.capacity {
		c.evictOldestLocked()
	}
}

func (c *MultiLevelCache) evictOldestLocked() {
	back := c.lru.Back()
	if back == nil {
		return
	}
	key := back.Value.(string)
	c.lru.Remove(back)
	delete(c.l1, key)
}

// InMemoryBackend is an in-memory L2 mock backend.
type InMemoryBackend struct {
	mu sync.RWMutex
	m  map[string]backendEntry
}

type backendEntry struct {
	v        string
	expireAt time.Time
}

func NewInMemoryBackend() *InMemoryBackend {
	return &InMemoryBackend{m: make(map[string]backendEntry)}
}

func (b *InMemoryBackend) Get(key string) (string, error) {
	b.mu.RLock()
	e, ok := b.m[key]
	b.mu.RUnlock()
	if !ok {
		return "", ErrNotFound
	}
	if !e.expireAt.IsZero() && time.Now().After(e.expireAt) {
		b.mu.Lock()
		delete(b.m, key)
		b.mu.Unlock()
		return "", ErrNotFound
	}
	return e.v, nil
}

func (b *InMemoryBackend) Set(key, value string, ttl time.Duration) error {
	exp := time.Time{}
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	b.mu.Lock()
	b.m[key] = backendEntry{v: value, expireAt: exp}
	b.mu.Unlock()
	return nil
}

func (b *InMemoryBackend) Delete(key string) error {
	b.mu.Lock()
	delete(b.m, key)
	b.mu.Unlock()
	return nil
}
