package swiss_table

import (
	"fmt"
	"hash/maphash"
)

// Swiss Table 是 Google 提出的高性能哈希表实现
// 核心特性：
// 1. 使用控制字节数组（metadata）加速查找
// 2. 分组查找（每组 GROUP_SIZE 个槽位）
// 3. 开放寻址法（Open Addressing）
// 4. SIMD 友好的设计

const (
	GROUP_SIZE      = 16               // 每组的槽位数（匹配SIMD寄存器大小）
	EMPTY           = byte(0b11111111) // 空槽位标记
	DELETED         = byte(0b11111110) // 已删除槽位标记
	MAX_LOAD_FACTOR = 0.875            // 最大负载因子（14/16）
	INITIAL_GROUPS  = 4                // 初始组数
)

// SwissTable 简化版的Swiss Table实现
type SwissTable[K comparable, V any] struct {
	ctrl   []byte       // 控制字节数组（每个字节存储一个槽位的元数据）
	keys   []K          // 键数组
	values []V          // 值数组
	count  int          // 当前元素数量
	groups int          // 组数
	hash   maphash.Hash // 哈希函数
}

// NewSwissTable 创建一个新的Swiss Table
func NewSwissTable[K comparable, V any]() *SwissTable[K, V] {
	groups := INITIAL_GROUPS
	size := groups * GROUP_SIZE

	ctrl := make([]byte, size)
	for i := range ctrl {
		ctrl[i] = EMPTY
	}

	return &SwissTable[K, V]{
		ctrl:   ctrl,
		keys:   make([]K, size),
		values: make([]V, size),
		count:  0,
		groups: groups,
	}
}

// Put 插入或更新键值对
func (st *SwissTable[K, V]) Put(key K, value V) {
	// 检查是否需要扩容
	if st.shouldGrow() {
		st.grow()
	}

	hash := st.hashKey(key)
	h1 := st.h1(hash) // 组索引
	h2 := st.h2(hash) // 控制字节（高7位）

	size := len(st.ctrl)

	// 线性探测查找空槽位或已存在的key
	for i := 0; i < size; i++ {
		pos := (h1 + i) % size
		groupStart := (pos / GROUP_SIZE) * GROUP_SIZE

		// 在当前组内查找
		for j := 0; j < GROUP_SIZE; j++ {
			idx := (groupStart + j) % size
			ctrl := st.ctrl[idx]

			// 找到空槽位，插入新元素
			if ctrl == EMPTY || ctrl == DELETED {
				st.ctrl[idx] = h2
				st.keys[idx] = key
				st.values[idx] = value
				st.count++
				return
			}

			// 找到相同的key，更新值
			if ctrl == h2 && st.keys[idx] == key {
				st.values[idx] = value
				return
			}
		}
	}
}

// Get 获取键对应的值
func (st *SwissTable[K, V]) Get(key K) (V, bool) {
	hash := st.hashKey(key)
	h1 := st.h1(hash)
	h2 := st.h2(hash)

	size := len(st.ctrl)

	// 线性探测查找
	for i := 0; i < size; i++ {
		pos := (h1 + i) % size
		groupStart := (pos / GROUP_SIZE) * GROUP_SIZE

		// 在当前组内查找
		for j := 0; j < GROUP_SIZE; j++ {
			idx := (groupStart + j) % size
			ctrl := st.ctrl[idx]

			// 遇到空槽位，说明key不存在
			if ctrl == EMPTY {
				var zero V
				return zero, false
			}

			// 匹配控制字节和key
			if ctrl == h2 && st.keys[idx] == key {
				return st.values[idx], true
			}
		}
	}

	var zero V
	return zero, false
}

// Delete 删除键值对
func (st *SwissTable[K, V]) Delete(key K) bool {
	hash := st.hashKey(key)
	h1 := st.h1(hash)
	h2 := st.h2(hash)

	size := len(st.ctrl)

	for i := 0; i < size; i++ {
		pos := (h1 + i) % size
		groupStart := (pos / GROUP_SIZE) * GROUP_SIZE

		for j := 0; j < GROUP_SIZE; j++ {
			idx := (groupStart + j) % size
			ctrl := st.ctrl[idx]

			if ctrl == EMPTY {
				return false
			}

			if ctrl == h2 && st.keys[idx] == key {
				st.ctrl[idx] = DELETED
				var zeroK K
				var zeroV V
				st.keys[idx] = zeroK
				st.values[idx] = zeroV
				st.count--
				return true
			}
		}
	}

	return false
}

// Len 返回元素数量
func (st *SwissTable[K, V]) Len() int {
	return st.count
}

// Contains 检查key是否存在
func (st *SwissTable[K, V]) Contains(key K) bool {
	_, ok := st.Get(key)
	return ok
}

// Clear 清空表
func (st *SwissTable[K, V]) Clear() {
	for i := range st.ctrl {
		st.ctrl[i] = EMPTY
	}

	var zeroK K
	var zeroV V
	for i := range st.keys {
		st.keys[i] = zeroK
		st.values[i] = zeroV
	}

	st.count = 0
}

// Keys 返回所有键
func (st *SwissTable[K, V]) Keys() []K {
	keys := make([]K, 0, st.count)
	for i, ctrl := range st.ctrl {
		if ctrl != EMPTY && ctrl != DELETED {
			keys = append(keys, st.keys[i])
		}
	}
	return keys
}

// shouldGrow 判断是否需要扩容
func (st *SwissTable[K, V]) shouldGrow() bool {
	capacity := len(st.ctrl)
	return float64(st.count) >= float64(capacity)*MAX_LOAD_FACTOR
}

// grow 扩容
func (st *SwissTable[K, V]) grow() {
	oldCtrl := st.ctrl
	oldKeys := st.keys
	oldValues := st.values

	// 容量翻倍
	st.groups *= 2
	newSize := st.groups * GROUP_SIZE

	st.ctrl = make([]byte, newSize)
	for i := range st.ctrl {
		st.ctrl[i] = EMPTY
	}

	st.keys = make([]K, newSize)
	st.values = make([]V, newSize)
	st.count = 0

	// 重新插入所有元素
	for i, ctrl := range oldCtrl {
		if ctrl != EMPTY && ctrl != DELETED {
			st.Put(oldKeys[i], oldValues[i])
		}
	}
}

// hashKey 计算key的哈希值
func (st *SwissTable[K, V]) hashKey(key K) uint64 {
	st.hash.Reset()

	// 简化处理：使用fmt包将key转换为字节序列
	// 实际生产环境应该针对不同类型实现专门的哈希函数
	hashBytes(&st.hash, key)

	return st.hash.Sum64()
}

// h1 计算组索引（使用哈希值的低位）
func (st *SwissTable[K, V]) h1(hash uint64) int {
	return int(hash % uint64(len(st.ctrl)))
}

// h2 计算控制字节（使用哈希值的高7位）
func (st *SwissTable[K, V]) h2(hash uint64) byte {
	// 取高7位，确保不是EMPTY或DELETED
	h2 := byte((hash >> 57) & 0x7F)
	if h2 == EMPTY || h2 == DELETED {
		h2 = 0
	}
	return h2
}

// hashBytes 将key写入hasher（简化实现）
func hashBytes[K any](h *maphash.Hash, key K) {
	// 使用类型断言处理常见类型
	switch v := any(key).(type) {
	case string:
		h.WriteString(v)
	case int:
		var buf [8]byte
		buf[0] = byte(v)
		buf[1] = byte(v >> 8)
		buf[2] = byte(v >> 16)
		buf[3] = byte(v >> 24)
		buf[4] = byte(v >> 32)
		buf[5] = byte(v >> 40)
		buf[6] = byte(v >> 48)
		buf[7] = byte(v >> 56)
		h.Write(buf[:])
	case int64:
		var buf [8]byte
		buf[0] = byte(v)
		buf[1] = byte(v >> 8)
		buf[2] = byte(v >> 16)
		buf[3] = byte(v >> 24)
		buf[4] = byte(v >> 32)
		buf[5] = byte(v >> 40)
		buf[6] = byte(v >> 48)
		buf[7] = byte(v >> 56)
		h.Write(buf[:])
	case uint64:
		var buf [8]byte
		buf[0] = byte(v)
		buf[1] = byte(v >> 8)
		buf[2] = byte(v >> 16)
		buf[3] = byte(v >> 24)
		buf[4] = byte(v >> 32)
		buf[5] = byte(v >> 40)
		buf[6] = byte(v >> 48)
		buf[7] = byte(v >> 56)
		h.Write(buf[:])
	default:
		// 其他类型使用fmt.Sprint（性能较低，但通用）
		h.WriteString(fmt.Sprintf("%v", v))
	}
}

// ========== 与Go原生map的性能对比工具 ==========

// GoMap Go原生map的包装，用于性能对比
type GoMap[K comparable, V any] struct {
	m map[K]V
}

// NewGoMap 创建原生map
func NewGoMap[K comparable, V any]() *GoMap[K, V] {
	return &GoMap[K, V]{
		m: make(map[K]V),
	}
}

func (gm *GoMap[K, V]) Put(key K, value V) {
	gm.m[key] = value
}

func (gm *GoMap[K, V]) Get(key K) (V, bool) {
	v, ok := gm.m[key]
	return v, ok
}

func (gm *GoMap[K, V]) Delete(key K) bool {
	if _, ok := gm.m[key]; ok {
		delete(gm.m, key)
		return true
	}
	return false
}

func (gm *GoMap[K, V]) Len() int {
	return len(gm.m)
}

func (gm *GoMap[K, V]) Contains(key K) bool {
	_, ok := gm.m[key]
	return ok
}

func (gm *GoMap[K, V]) Clear() {
	gm.m = make(map[K]V)
}

func (gm *GoMap[K, V]) Keys() []K {
	keys := make([]K, 0, len(gm.m))
	for k := range gm.m {
		keys = append(keys, k)
	}
	return keys
}
