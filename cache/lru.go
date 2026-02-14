package cache

import (
	"container/list"
	"sync"
)

// LRUCache Least Recently Used 缓存
type LRUCache struct {
	capacity int
	cache    map[interface{}]*list.Element
	list     *list.List
	mu       sync.Mutex
}

// entry 缓存条目
type lruEntry struct {
	key   interface{}
	value interface{}
}

// NewLRUCache 创建一个新的LRU缓存
func NewLRUCache(capacity int) *LRUCache {
	if capacity <= 0 {
		capacity = 10
	}
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[interface{}]*list.Element),
		list:     list.New(),
	}
}

// Get 获取缓存值
func (lru *LRUCache) Get(key interface{}) (interface{}, bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if elem, ok := lru.cache[key]; ok {
		// 移动到链表头部（最近使用）
		lru.list.MoveToFront(elem)
		return elem.Value.(*lruEntry).value, true
	}

	return nil, false
}

// Put 设置缓存值
func (lru *LRUCache) Put(key, value interface{}) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	// 如果key已存在，更新值并移到前面
	if elem, ok := lru.cache[key]; ok {
		lru.list.MoveToFront(elem)
		elem.Value.(*lruEntry).value = value
		return
	}

	// 添加新元素
	entry := &lruEntry{key: key, value: value}
	elem := lru.list.PushFront(entry)
	lru.cache[key] = elem

	// 如果超过容量，移除最少使用的元素（链表尾部）
	if lru.list.Len() > lru.capacity {
		lru.removeOldest()
	}
}

// Delete 删除缓存项
func (lru *LRUCache) Delete(key interface{}) bool {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if elem, ok := lru.cache[key]; ok {
		lru.removeElement(elem)
		return true
	}

	return false
}

// removeOldest 移除最旧的元素
func (lru *LRUCache) removeOldest() {
	elem := lru.list.Back()
	if elem != nil {
		lru.removeElement(elem)
	}
}

// removeElement 移除指定元素
func (lru *LRUCache) removeElement(elem *list.Element) {
	lru.list.Remove(elem)
	entry := elem.Value.(*lruEntry)
	delete(lru.cache, entry.key)
}

// Len 返回缓存中的元素数量
func (lru *LRUCache) Len() int {
	lru.mu.Lock()
	defer lru.mu.Unlock()
	return lru.list.Len()
}

// Cap 返回缓存容量
func (lru *LRUCache) Cap() int {
	return lru.capacity
}

// Clear 清空缓存
func (lru *LRUCache) Clear() {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.list.Init()
	lru.cache = make(map[interface{}]*list.Element)
}

// Keys 返回所有key（从最近使用到最少使用）
func (lru *LRUCache) Keys() []interface{} {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	keys := make([]interface{}, 0, lru.list.Len())
	for elem := lru.list.Front(); elem != nil; elem = elem.Next() {
		keys = append(keys, elem.Value.(*lruEntry).key)
	}

	return keys
}

// Contains 检查key是否存在
func (lru *LRUCache) Contains(key interface{}) bool {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	_, ok := lru.cache[key]
	return ok
}

// Peek 查看值但不更新访问时间
func (lru *LRUCache) Peek(key interface{}) (interface{}, bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if elem, ok := lru.cache[key]; ok {
		return elem.Value.(*lruEntry).value, true
	}

	return nil, false
}

// ========== 泛型版本（Go 1.18+）==========

// LRUCacheGeneric 泛型LRU缓存
type LRUCacheGeneric[K comparable, V any] struct {
	capacity int
	cache    map[K]*list.Element
	list     *list.List
	mu       sync.Mutex
}

// lruEntryGeneric 泛型缓存条目
type lruEntryGeneric[K comparable, V any] struct {
	key   K
	value V
}

// NewLRUCacheGeneric 创建一个新的泛型LRU缓存
func NewLRUCacheGeneric[K comparable, V any](capacity int) *LRUCacheGeneric[K, V] {
	if capacity <= 0 {
		capacity = 10
	}
	return &LRUCacheGeneric[K, V]{
		capacity: capacity,
		cache:    make(map[K]*list.Element),
		list:     list.New(),
	}
}

// Get 获取缓存值
func (lru *LRUCacheGeneric[K, V]) Get(key K) (V, bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if elem, ok := lru.cache[key]; ok {
		lru.list.MoveToFront(elem)
		return elem.Value.(*lruEntryGeneric[K, V]).value, true
	}

	var zero V
	return zero, false
}

// Put 设置缓存值
func (lru *LRUCacheGeneric[K, V]) Put(key K, value V) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if elem, ok := lru.cache[key]; ok {
		lru.list.MoveToFront(elem)
		elem.Value.(*lruEntryGeneric[K, V]).value = value
		return
	}

	entry := &lruEntryGeneric[K, V]{key: key, value: value}
	elem := lru.list.PushFront(entry)
	lru.cache[key] = elem

	if lru.list.Len() > lru.capacity {
		lru.removeOldest()
	}
}

// Delete 删除缓存项
func (lru *LRUCacheGeneric[K, V]) Delete(key K) bool {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if elem, ok := lru.cache[key]; ok {
		lru.removeElement(elem)
		return true
	}

	return false
}

// removeOldest 移除最旧的元素
func (lru *LRUCacheGeneric[K, V]) removeOldest() {
	elem := lru.list.Back()
	if elem != nil {
		lru.removeElement(elem)
	}
}

// removeElement 移除指定元素
func (lru *LRUCacheGeneric[K, V]) removeElement(elem *list.Element) {
	lru.list.Remove(elem)
	entry := elem.Value.(*lruEntryGeneric[K, V])
	delete(lru.cache, entry.key)
}

// Len 返回缓存中的元素数量
func (lru *LRUCacheGeneric[K, V]) Len() int {
	lru.mu.Lock()
	defer lru.mu.Unlock()
	return lru.list.Len()
}

// Cap 返回缓存容量
func (lru *LRUCacheGeneric[K, V]) Cap() int {
	return lru.capacity
}

// Clear 清空缓存
func (lru *LRUCacheGeneric[K, V]) Clear() {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.list.Init()
	lru.cache = make(map[K]*list.Element)
}

// Keys 返回所有key（从最近使用到最少使用）
func (lru *LRUCacheGeneric[K, V]) Keys() []K {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	keys := make([]K, 0, lru.list.Len())
	for elem := lru.list.Front(); elem != nil; elem = elem.Next() {
		keys = append(keys, elem.Value.(*lruEntryGeneric[K, V]).key)
	}

	return keys
}

// Contains 检查key是否存在
func (lru *LRUCacheGeneric[K, V]) Contains(key K) bool {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	_, ok := lru.cache[key]
	return ok
}

// Peek 查看值但不更新访问时间
func (lru *LRUCacheGeneric[K, V]) Peek(key K) (V, bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if elem, ok := lru.cache[key]; ok {
		return elem.Value.(*lruEntryGeneric[K, V]).value, true
	}

	var zero V
	return zero, false
}
