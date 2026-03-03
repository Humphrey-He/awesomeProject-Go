package cache

import (
	"container/list"
	"sync"
)

// LFUCache2 Least Frequently Used 缓存
type LFUCache2 struct {
	capacity int
	minFreq  int                       // 最小频率
	cache    map[interface{}]*lfuNode2 // key -> node
	freqMap  map[int]*list.List        // frequency -> 该频率的元素列表
	mu       sync.Mutex
}

// lfuNode2 LFU缓存节点
type lfuNode2 struct {
	key   interface{}
	value interface{}
	freq  int           // 访问频率
	elem  *list.Element // 在频率列表中的位置
}

// NewLFUCache2 创建一个新的LFU缓存
func NewLFUCache2(capacity int) *LFUCache2 {
	if capacity <= 0 {
		capacity = 10
	}
	return &LFUCache2{
		capacity: capacity,
		minFreq:  0,
		cache:    make(map[interface{}]*lfuNode2),
		freqMap:  make(map[int]*list.List),
	}
}

// Get 获取缓存值
func (lfu *LFUCache2) Get(key interface{}) (interface{}, bool) {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	node, ok := lfu.cache[key]
	if !ok {
		return nil, false
	}

	// 增加访问频率
	lfu.increaseFreq(node)

	return node.value, true
}

// Put2 设置缓存值
func (lfu *LFUCache2) Put(key, value interface{}) {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	if lfu.capacity == 0 {
		return
	}

	// 如果key已存在，更新值和频率
	if node, ok := lfu.cache[key]; ok {
		node.value = value
		lfu.increaseFreq(node)
		return
	}

	// 如果缓存已满，移除最少使用的元素
	if len(lfu.cache) >= lfu.capacity {
		lfu.removeLFU()
	}

	// 添加新节点
	node := &lfuNode2{
		key:   key,
		value: value,
		freq:  1,
	}

	lfu.cache[key] = node
	lfu.addToFreqList(node)
	lfu.minFreq = 1
}

// Delete 删除缓存项
func (lfu *LFUCache2) Delete(key interface{}) bool {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	node, ok := lfu.cache[key]
	if !ok {
		return false
	}

	lfu.removeNode(node)
	return true
}

// increaseFreq 增加节点的访问频率
func (lfu *LFUCache2) increaseFreq(node *lfuNode2) {
	// 从旧频率列表中移除
	oldFreq := node.freq
	if freqList, ok := lfu.freqMap[oldFreq]; ok {
		freqList.Remove(node.elem)

		// 如果该频率列表为空，删除它
		if freqList.Len() == 0 {
			delete(lfu.freqMap, oldFreq)

			// 如果删除的是最小频率，更新minFreq
			if lfu.minFreq == oldFreq {
				lfu.minFreq++
			}
		}
	}

	// 增加频率并添加到新的频率列表
	node.freq++
	lfu.addToFreqList(node)
}

// addToFreqList 将节点添加到对应频率的列表
func (lfu *LFUCache2) addToFreqList(node *lfuNode2) {
	freq := node.freq

	if _, ok := lfu.freqMap[freq]; !ok {
		lfu.freqMap[freq] = list.New()
	}

	node.elem = lfu.freqMap[freq].PushBack(node)
}

// removeLFU 移除最少使用的元素
func (lfu *LFUCache2) removeLFU() {
	// 获取最小频率的列表
	freqList, ok := lfu.freqMap[lfu.minFreq]
	if !ok || freqList.Len() == 0 {
		return
	}

	// 移除列表头部的元素（最旧的）
	elem := freqList.Front()
	if elem != nil {
		node := elem.Value.(*lfuNode2)
		lfu.removeNode(node)
	}
}

// removeNode 移除指定节点
func (lfu *LFUCache2) removeNode(node *lfuNode2) {
	// 从频率列表中移除
	if freqList, ok := lfu.freqMap[node.freq]; ok {
		freqList.Remove(node.elem)

		if freqList.Len() == 0 {
			delete(lfu.freqMap, node.freq)
		}
	}

	// 从缓存中删除
	delete(lfu.cache, node.key)
}

// Len 返回缓存中的元素数量
func (lfu *LFUCache2) Len() int {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()
	return len(lfu.cache)
}

// Cap 返回缓存容量
func (lfu *LFUCache2) Cap() int {
	return lfu.capacity
}

// Clear 清空缓存
func (lfu *LFUCache2) Clear() {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	lfu.cache = make(map[interface{}]*lfuNode2)
	lfu.freqMap = make(map[int]*list.List)
	lfu.minFreq = 0
}

// Contains 检查key是否存在
func (lfu *LFUCache2) Contains(key interface{}) bool {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	_, ok := lfu.cache[key]
	return ok
}

// Peek 查看值但不更新访问频率
func (lfu *LFUCache2) Peek(key interface{}) (interface{}, bool) {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	if node, ok := lfu.cache[key]; ok {
		return node.value, true
	}

	return nil, false
}

// GetFreq 获取key的访问频率
func (lfu *LFUCache2) GetFreq(key interface{}) int {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	if node, ok := lfu.cache[key]; ok {
		return node.freq
	}

	return 0
}

// ========== 泛型版本（Go 1.18+）==========

// LFUCacheGeneric 泛型LFU缓存
type LFUCacheGeneric2[K comparable, V any] struct {
	capacity int
	minFreq  int
	cache    map[K]*lfuNodeGeneric[K, V]
	freqMap  map[int]*list.List
	mu       sync.Mutex
}

// lfuNodeGeneric 泛型LFU缓存节点
type lfuNodeGeneric2[K comparable, V any] struct {
	key   K
	value V
	freq  int
	elem  *list.Element
}

// NewLFUCacheGeneric 创建一个新的泛型LFU缓存
func NewLFUCacheGeneric2[K comparable, V any](capacity int) *LFUCacheGeneric[K, V] {
	if capacity <= 0 {
		capacity = 10
	}
	return &LFUCacheGeneric[K, V]{
		capacity: capacity,
		minFreq:  0,
		cache:    make(map[K]*lfuNodeGeneric[K, V]),
		freqMap:  make(map[int]*list.List),
	}
}

// Get 获取缓存值
func (lfu *LFUCacheGeneric[K, V]) Get2(key K) (V, bool) {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	node, ok := lfu.cache[key]
	if !ok {
		var zero V
		return zero, false
	}

	lfu.increaseFreq(node)
	return node.value, true
}

// Put 设置缓存值
func (lfu *LFUCacheGeneric[K, V]) Put2(key K, value V) {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	if lfu.capacity == 0 {
		return
	}

	if node, ok := lfu.cache[key]; ok {
		node.value = value
		lfu.increaseFreq(node)
		return
	}

	if len(lfu.cache) >= lfu.capacity {
		lfu.removeLFU()
	}

	node := &lfuNodeGeneric[K, V]{
		key:   key,
		value: value,
		freq:  1,
	}

	lfu.cache[key] = node
	lfu.addToFreqList(node)
	lfu.minFreq = 1
}

// Delete 删除缓存项
func (lfu *LFUCacheGeneric[K, V]) Delete2(key K) bool {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	node, ok := lfu.cache[key]
	if !ok {
		return false
	}

	lfu.removeNode(node)
	return true
}

// increaseFreq 增加节点的访问频率
func (lfu *LFUCacheGeneric[K, V]) increaseFreq2(node *lfuNodeGeneric[K, V]) {
	oldFreq := node.freq
	if freqList, ok := lfu.freqMap[oldFreq]; ok {
		freqList.Remove(node.elem)

		if freqList.Len() == 0 {
			delete(lfu.freqMap, oldFreq)

			if lfu.minFreq == oldFreq {
				lfu.minFreq++
			}
		}
	}

	node.freq++
	lfu.addToFreqList(node)
}

// addToFreqList 将节点添加到对应频率的列表
func (lfu *LFUCacheGeneric[K, V]) addToFreqList2(node *lfuNodeGeneric[K, V]) {
	freq := node.freq

	if _, ok := lfu.freqMap[freq]; !ok {
		lfu.freqMap[freq] = list.New()
	}

	node.elem = lfu.freqMap[freq].PushBack(node)
}

// removeLFU 移除最少使用的元素
func (lfu *LFUCacheGeneric[K, V]) removeLFU2() {
	freqList, ok := lfu.freqMap[lfu.minFreq]
	if !ok || freqList.Len() == 0 {
		return
	}

	elem := freqList.Front()
	if elem != nil {
		node := elem.Value.(*lfuNodeGeneric[K, V])
		lfu.removeNode(node)
	}
}

// removeNode 移除指定节点
func (lfu *LFUCacheGeneric[K, V]) removeNode2(node *lfuNodeGeneric[K, V]) {
	if freqList, ok := lfu.freqMap[node.freq]; ok {
		freqList.Remove(node.elem)

		if freqList.Len() == 0 {
			delete(lfu.freqMap, node.freq)
		}
	}

	delete(lfu.cache, node.key)
}

// Len 返回缓存中的元素数量
func (lfu *LFUCacheGeneric[K, V]) Len2() int {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()
	return len(lfu.cache)
}

// Cap 返回缓存容量
func (lfu *LFUCacheGeneric[K, V]) Cap2() int {
	return lfu.capacity
}

// Clear 清空缓存
func (lfu *LFUCacheGeneric[K, V]) Clear2() {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	lfu.cache = make(map[K]*lfuNodeGeneric[K, V])
	lfu.freqMap = make(map[int]*list.List)
	lfu.minFreq = 0
}

// Contains 检查key是否存在
func (lfu *LFUCacheGeneric[K, V]) Contains2(key K) bool {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	_, ok := lfu.cache[key]
	return ok
}

// Peek 查看值但不更新访问频率
func (lfu *LFUCacheGeneric[K, V]) Peek2(key K) (V, bool) {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	if node, ok := lfu.cache[key]; ok {
		return node.value, true
	}

	var zero V
	return zero, false
}

// GetFreq 获取key的访问频率
func (lfu *LFUCacheGeneric[K, V]) GetFreq2(key K) int {
	lfu.mu.Lock()
	defer lfu.mu.Unlock()

	if node, ok := lfu.cache[key]; ok {
		return node.freq
	}

	return 0
}
