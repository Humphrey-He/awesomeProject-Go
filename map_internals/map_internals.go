package map_internals

import (
	"fmt"
	"hash/maphash"
	"math"
	"unsafe"
)

// ========== Go Map 内部原理与扩容机制 ==========

/*
本文件深入讲解Go map的内部实现机制，包括：
1. 数据结构（hmap, bmap）
2. 哈希函数与冲突解决
3. 扩容触发条件
4. 增量扩容过程
5. 等量扩容（内存优化）
6. 负载因子计算
7. 性能分析

注意：这是教学性质的模拟实现，Go运行时的真实实现在 runtime/map.go
*/

// ========== 1. Go Map 数据结构模拟 ==========

// hmap 是map的header结构（简化版）
// 真实的runtime.hmap在 src/runtime/map.go
type hmap struct {
	count     int      // map中的元素数量
	B         uint8    // log_2(buckets数量)，即buckets = 2^B
	noverflow uint16   // overflow buckets数量的近似值
	hash0     uint32   // 哈希种子
	buckets   unsafe.Pointer // 指向bucket数组
	oldbuckets unsafe.Pointer // 扩容时的旧bucket数组
	nevacuate uintptr  // 扩容进度，小于此值的bucket已迁移
}

// bmap 是bucket结构（简化版）
// Go的真实bmap包含8个key-value对
const (
	bucketCntBits = 3
	bucketCnt     = 1 << bucketCntBits // 8个元素
	loadFactorNum = 13                  // 负载因子分子
	loadFactorDen = 2                   // 负载因子分母，即 6.5
)

// bmap bucket结构模拟
type bmap struct {
	tophash  [bucketCnt]uint8 // 每个key的hash高8位
	keys     [bucketCnt]interface{}
	values   [bucketCnt]interface{}
	overflow *bmap // 溢出桶指针
}

// ========== 2. 模拟Map实现 ==========

// SimulatedMap 模拟Go map的内部实现
type SimulatedMap struct {
	buckets      []*bmap
	B            uint8      // log_2(buckets数量)
	count        int        // 元素数量
	hashSeed     maphash.Seed // 哈希种子
	oldbuckets   []*bmap    // 扩容时的旧buckets
	nevacuate    int        // 扩容进度
	growing      bool       // 是否正在扩容
	sameSizeGrow bool       // 是否等量扩容
}

// NewSimulatedMap 创建模拟map
func NewSimulatedMap(hint int) *SimulatedMap {
	B := uint8(0)
	// 计算初始bucket数量
	for overLoadFactor(hint, B) {
		B++
	}
	
	bucketCount := 1 << B
	buckets := make([]*bmap, bucketCount)
	for i := range buckets {
		buckets[i] = &bmap{}
	}
	
	return &SimulatedMap{
		buckets:  buckets,
		B:        B,
		hashSeed: maphash.MakeSeed(),
	}
}

// overLoadFactor 检查是否超过负载因子
// Go map的负载因子是 6.5 (13/2)
func overLoadFactor(count int, B uint8) bool {
	return count > bucketCnt && uintptr(count) > loadFactorNum*(1<<B)/loadFactorDen
}

// tooManyOverflowBuckets 检查是否有太多overflow buckets
func tooManyOverflowBuckets(noverflow uint16, B uint8) bool {
	if B > 15 {
		B = 15
	}
	return noverflow >= uint16(1<<(B&15))
}

// hashKey 计算key的hash值
func (m *SimulatedMap) hashKey(key interface{}) uint64 {
	var h maphash.Hash
	h.SetSeed(m.hashSeed) // 使用固定种子
	h.WriteString(fmt.Sprintf("%v", key))
	return h.Sum64()
}

// bucketMask 返回bucket索引掩码
func (m *SimulatedMap) bucketMask() uintptr {
	return (1 << m.B) - 1
}

// tophash 计算hash的高8位
func tophash(hash uint64) uint8 {
	top := uint8(hash >> 56)
	if top < 5 {
		top += 5
	}
	return top
}

// Set 插入key-value
func (m *SimulatedMap) Set(key, value interface{}) {
	// 1. 计算hash值
	hash := m.hashKey(key)
	
	// 2. 计算bucket索引
	bucketIdx := hash & uint64(m.bucketMask())
	
	// 3. 如果正在扩容，先迁移一个bucket
	if m.growing {
		m.growWork(int(bucketIdx))
	}
	
	// 4. 找到bucket
	b := m.buckets[bucketIdx]
	top := tophash(hash)
	
	// 5. 在bucket中查找或插入
	inserted := false
	for {
		for i := 0; i < bucketCnt; i++ {
			if b.tophash[i] == 0 {
				// 找到空位，插入
				b.tophash[i] = top
				b.keys[i] = key
				b.values[i] = value
				m.count++
				inserted = true
				break
			}
			if b.tophash[i] == top && b.keys[i] == key {
				// 更新已存在的key
				b.values[i] = value
				inserted = true
				break
			}
		}
		
		if inserted {
			break
		}
		
		// 需要overflow bucket
		if b.overflow == nil {
			b.overflow = &bmap{}
		}
		b = b.overflow
	}
	
	// 6. 检查是否需要扩容
	if !m.growing && m.shouldGrow() {
		m.hashGrow()
	}
}

// Get 获取value
func (m *SimulatedMap) Get(key interface{}) (interface{}, bool) {
	if m.count == 0 {
		return nil, false
	}
	
	hash := m.hashKey(key)
	bucketIdx := hash & uint64(m.bucketMask())
	
	// 如果正在扩容，可能需要查旧buckets
	b := m.buckets[bucketIdx]
	if m.growing && int(bucketIdx) < m.nevacuate {
		// 数据可能还在旧bucket中
		oldBucketIdx := bucketIdx & uint64((1<<(m.B-1))-1)
		if int(oldBucketIdx) < len(m.oldbuckets) {
			b = m.oldbuckets[oldBucketIdx]
		}
	}
	
	top := tophash(hash)
	
	for b != nil {
		for i := 0; i < bucketCnt; i++ {
			if b.tophash[i] == top && b.keys[i] == key {
				return b.values[i], true
			}
		}
		b = b.overflow
	}
	
	return nil, false
}

// shouldGrow 判断是否需要扩容
func (m *SimulatedMap) shouldGrow() bool {
	// 条件1：负载因子 > 6.5
	if overLoadFactor(m.count, m.B) {
		return true
	}
	
	// 条件2：overflow buckets太多（等量扩容）
	noverflow := m.countOverflowBuckets()
	if tooManyOverflowBuckets(uint16(noverflow), m.B) {
		return true
	}
	
	return false
}

// countOverflowBuckets 统计overflow buckets数量
func (m *SimulatedMap) countOverflowBuckets() int {
	count := 0
	for _, b := range m.buckets {
		curr := b
		for curr.overflow != nil {
			count++
			curr = curr.overflow
		}
	}
	return count
}

// hashGrow 开始扩容
func (m *SimulatedMap) hashGrow() {
	// 判断是增量扩容还是等量扩容
	bigger := uint8(1)
	if !overLoadFactor(m.count, m.B) {
		// 等量扩容：overflow buckets太多，但负载因子正常
		// 目的是整理内存，减少overflow buckets
		bigger = 0
		m.sameSizeGrow = true
	}
	
	// 保存旧buckets
	m.oldbuckets = m.buckets
	
	// 创建新buckets（容量翻倍或不变）
	newB := m.B + bigger
	newBucketCount := 1 << newB
	newBuckets := make([]*bmap, newBucketCount)
	for i := range newBuckets {
		newBuckets[i] = &bmap{}
	}
	
	m.buckets = newBuckets
	m.B = newB
	m.growing = true
	m.nevacuate = 0
}

// growWork 执行增量扩容工作
// 每次访问map时，迁移一个bucket
func (m *SimulatedMap) growWork(bucket int) {
	// 迁移正在访问的bucket
	m.evacuate(m.nevacuate)
	
	// 如果还在扩容，多迁移一个bucket
	if m.growing {
		m.evacuate(m.nevacuate)
	}
}

// evacuate 迁移一个bucket
func (m *SimulatedMap) evacuate(oldbucket int) {
	if oldbucket >= len(m.oldbuckets) {
		m.growing = false
		m.oldbuckets = nil
		return
	}
	
	b := m.oldbuckets[oldbucket]
	if b == nil {
		m.nevacuate++
		return
	}
	
	// 计算新的bucket索引
	newbit := uintptr(1) << (m.B - 1)
	if m.sameSizeGrow {
		newbit = 0
	}
	
	// 遍历旧bucket的所有元素
	for b != nil {
		for i := 0; i < bucketCnt; i++ {
			if b.tophash[i] == 0 {
				continue
			}
			
			// 重新计算hash
			hash := m.hashKey(b.keys[i])
			var newBucketIdx uintptr
			
			if m.sameSizeGrow {
				// 等量扩容：索引不变
				newBucketIdx = uintptr(oldbucket)
			} else {
				// 增量扩容：根据hash决定新位置
				// 可能是原位置，也可能是原位置+旧容量
				if hash&uint64(newbit) == 0 {
					newBucketIdx = uintptr(oldbucket)
				} else {
					newBucketIdx = uintptr(oldbucket) + newbit
				}
			}
			
			// 插入到新bucket
			newBucket := m.buckets[newBucketIdx]
			inserted := false
			
			for !inserted {
				for j := 0; j < bucketCnt; j++ {
					if newBucket.tophash[j] == 0 {
						newBucket.tophash[j] = b.tophash[i]
						newBucket.keys[j] = b.keys[i]
						newBucket.values[j] = b.values[i]
						inserted = true
						break
					}
				}
				
				if !inserted {
					if newBucket.overflow == nil {
						newBucket.overflow = &bmap{}
					}
					newBucket = newBucket.overflow
				}
			}
		}
		b = b.overflow
	}
	
	// 标记此bucket已迁移
	m.nevacuate++
	
	// 检查是否全部迁移完成
	if m.nevacuate >= len(m.oldbuckets) {
		m.growing = false
		m.oldbuckets = nil
		m.sameSizeGrow = false
	}
}

// Stats 返回map统计信息
func (m *SimulatedMap) Stats() MapStats {
	stats := MapStats{
		Count:           m.count,
		BucketCount:     len(m.buckets),
		B:               m.B,
		LoadFactor:      float64(m.count) / float64(len(m.buckets)*bucketCnt),
		Growing:         m.growing,
		SameSizeGrow:    m.sameSizeGrow,
		EvacuateProgress: float64(m.nevacuate) / math.Max(1, float64(len(m.oldbuckets))),
	}
	
	// 统计overflow buckets
	for _, b := range m.buckets {
		curr := b
		for curr.overflow != nil {
			stats.OverflowBuckets++
			curr = curr.overflow
		}
	}
	
	// 计算平均链长
	totalChainLen := 0
	maxChainLen := 0
	for _, b := range m.buckets {
		chainLen := 0
		curr := b
		for curr != nil {
			chainLen++
			curr = curr.overflow
		}
		totalChainLen += chainLen
		if chainLen > maxChainLen {
			maxChainLen = chainLen
		}
	}
	stats.AvgChainLength = float64(totalChainLen) / float64(len(m.buckets))
	stats.MaxChainLength = maxChainLen
	
	return stats
}

// MapStats map统计信息
type MapStats struct {
	Count            int     // 元素数量
	BucketCount      int     // bucket数量
	B                uint8   // log_2(buckets)
	LoadFactor       float64 // 负载因子
	OverflowBuckets  int     // overflow bucket数量
	AvgChainLength   float64 // 平均链长
	MaxChainLength   int     // 最大链长
	Growing          bool    // 是否正在扩容
	SameSizeGrow     bool    // 是否等量扩容
	EvacuateProgress float64 // 迁移进度 (0-1)
}

func (s MapStats) String() string {
	return fmt.Sprintf(`Map Statistics:
  Elements: %d
  Buckets: %d (2^%d)
  Load Factor: %.2f (threshold: 6.5)
  Overflow Buckets: %d
  Avg Chain Length: %.2f
  Max Chain Length: %d
  Growing: %v
  Same Size Grow: %v
  Evacuate Progress: %.1f%%`,
		s.Count, s.BucketCount, s.B, s.LoadFactor,
		s.OverflowBuckets, s.AvgChainLength, s.MaxChainLength,
		s.Growing, s.SameSizeGrow, s.EvacuateProgress*100)
}

// ========== 3. Go Map 扩容机制总结 ==========

/*
Go Map 扩容机制详解：

📊 扩容触发条件：

1. 负载因子过大（增量扩容）
   - 条件：count > buckets * 6.5
   - 动作：容量翻倍（B++）
   - 目的：保持性能

2. overflow buckets过多（等量扩容）
   - 条件：noverflow >= 2^min(B, 15)
   - 动作：容量不变
   - 目的：整理内存，减少碎片

🔄 扩容过程（增量扩容）：

1. 分配新buckets（容量翻倍）
2. 保存旧buckets到oldbuckets
3. 设置growing标志
4. 渐进式迁移：
   - 每次访问map时迁移1-2个bucket
   - 不是一次性迁移（避免STW）
   - 读写操作同时进行

📦 元素重新分布：

增量扩容时，元素的新位置有两种可能：
- 位置1：原bucket索引
- 位置2：原bucket索引 + 旧容量

决定因素：hash值的第B位
- 如果第B位是0 -> 位置1
- 如果第B位是1 -> 位置2

🔍 查找过程（扩容期间）：

1. 计算bucket索引
2. 检查是否已迁移
3. 如果未迁移，在oldbuckets中查找
4. 如果已迁移，在新buckets中查找

⚡ 性能特点：

优点：
✅ 增量扩容避免STW
✅ 读写可以并行
✅ 内存利用率高

代价：
❌ 扩容期间需要额外内存
❌ 查找可能需要访问两个buckets
❌ 写入需要触发迁移工作

📈 负载因子选择（6.5）：

- 太小：空间浪费
- 太大：性能下降
- 6.5是经过测试的平衡点

🎯 最佳实践：

1. 预分配容量
   m := make(map[K]V, hint)
   避免多次扩容

2. 避免大量删除后再插入
   - 删除不会缩容
   - 大量删除后内存不释放
   - 建议重新创建map

3. 并发访问需要加锁
   - map不是并发安全的
   - 使用sync.RWMutex
   - 或使用sync.Map

4. 注意内存泄漏
   - 大value会一直占用内存
   - 考虑用指针
   - 或定期重建map

🔬 实现细节：

真实的runtime.hmap包含：
- count: 元素数量
- B: log_2(buckets数量)
- noverflow: overflow buckets数量
- hash0: 哈希种子
- buckets: bucket数组指针
- oldbuckets: 旧bucket数组
- nevacuate: 迁移进度
- extra: 额外信息（GC相关）

真实的runtime.bmap包含：
- tophash[8]: 8个元素的hash高8位
- keys[8]: 8个key
- values[8]: 8个value
- overflow: 溢出桶指针

内存布局优化：
- keys和values分开存储（不是KV对）
- 好处：内存对齐，缓存友好
- tophash用于快速过滤

🎓 面试要点：

Q: map如何扩容？
A: 两种扩容：增量扩容（负载因子>6.5，容量翻倍）和等量扩容（overflow过多，整理内存）

Q: 为什么是渐进式扩容？
A: 避免一次性迁移导致的STW，保证低延迟

Q: map为什么无序？
A: 扩容时元素位置改变，且Go故意随机化遍历起点

Q: map并发安全吗？
A: 不安全，并发读写会panic，需要加锁或使用sync.Map

Q: 如何优化map性能？
A: 预分配容量、避免频繁扩容、考虑使用指针value
*/

