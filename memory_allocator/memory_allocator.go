package memory_allocator

import (
	"fmt"
)

// ========== Go 内存分配器深入理解 ==========

/*
本项目讲解 Go 内存分配器的实现原理，包括：

一、内存管理层次
1. mcache - 每P本地缓存
2. mcentral - 中心缓存
3. mheap - 堆内存管理
4. mspan - 内存span

二、内存分配流程
1. 小对象分配（<32KB）
2. 大对象分配（>32KB）
3. 内存对齐

三、span 大小类
1. 67种大小类
2. 8KB ~ 32KB
*/

// ========== 1. 内存管理结构 ==========

/*
Go 内存管理架构：

┌─────────────────────────────────────────┐
│                  mheap                  │
│              (全局堆内存)                │
├─────────────────────────────────────────┤
│  ┌───────────┐    ┌───────────┐       │
│  │ mcentral   │    │ mcentral   │ ...  │
│  │ sizeclass │    │ sizeclass │       │
│  │  8byte   │    │  16byte   │       │
│  └───────────┘    └───────────┘       │
└─────────────────────────────────────────┘
            ▲
            │
┌─────────────────────────────────────────┐
│                  mcache                  │
│              (每P本地缓存)                │
├─────────────────────────────────────────┤
│  ┌───────────┐    ┌───────────┐       │
│  │  mspan    │    │  mspan    │ ...  │
│  │  sizeclass│    │  sizeclass│       │
│  └───────────┘    └───────────┘       │
└─────────────────────────────────────────┘
*/

// ========== 2. mspan 结构 ==========

// Span 模拟 mspan
type Span struct {
	StartAddr int      // 起始地址
	NPages    int      // 页数
	Nelems    int      // 元素个数
	SizeClass int      // 大小类
	FreeIndex int      // 下一个可用位置
	AllocBits []byte   // 分配位图
	Next      *Span    // 链表next
	Prev      *Span    // 链表prev
}

// NewSpan 创建新的span
func NewSpan(sizeClass int, nelems int) *Span {
	return &Span{
		SizeClass: sizeClass,
		Nelems:    nelems,
		AllocBits: make([]byte, (nelems+7)/8),
	}
}

// Alloc 分配一个对象
func (s *Span) Alloc() int {
	for s.FreeIndex < s.Nelems {
		bitIndex := s.FreeIndex / 8
		bitOffset := s.FreeIndex % 8
		if s.AllocBits[bitIndex]&(1<<bitOffset) == 0 {
			// 标记为已分配
			s.AllocBits[bitIndex] |= 1 << bitOffset
			idx := s.FreeIndex
			s.FreeIndex++
			return idx
		}
		s.FreeIndex++
	}
	return -1
}

// Free 释放一个对象
func (s *Span) Free(idx int) bool {
	if idx < 0 || idx >= s.Nelems {
		return false
	}
	bitIndex := idx / 8
	bitOffset := idx % 8
	s.AllocBits[bitIndex] &^= (1 << bitOffset)
	return true
}

// IsFull 检查span是否已满
func (s *Span) IsFull() bool {
	return s.FreeIndex >= s.Nelems
}

// ========== 3. mcache ==========

/*
mcache 结构（每P本地缓存）：
- 无锁分配
- 每个P一个
- 定期从 mcentral refill
*/

// MCache 模拟 mcache
type MCache struct {
	Spans [67]*Span // 67种大小类
}

// NewMCache 创建 mcache
func NewMCache() *MCache {
	return &MCache{}
}

// Alloc 从 mcache 分配
func (m *MCache) Alloc(sizeClass int) int {
	if sizeClass < 0 || sizeClass >= 67 {
		return -1
	}
	
	span := m.Spans[sizeClass]
	if span == nil {
		return -1
	}
	
	return span.Alloc()
}

// GetSpan 获取或创建span
func (m *MCache) GetSpan(sizeClass int, nelems int) *Span {
	if m.Spans[sizeClass] == nil {
		m.Spans[sizeClass] = NewSpan(sizeClass, nelems)
	}
	return m.Spans[sizeClass]
}

// ========== 4. mcentral ==========

/*
mcentral 结构（中心缓存）：
- 每种大小类一个
- span 在 nonempty 和 empty 之间移动
- 有锁保护
*/

// MCentral 模拟 mcentral
type MCentral struct {
	SizeClass int
	NonEmpty  *Span // 有空闲对象的span链表
	Empty     *Span // 已满的span链表
	NMalloc   uint64
}

// NewMCentral 创建 mcentral
func NewMCentral(sizeClass int) *MCentral {
	return &MCentral{SizeClass: sizeClass}
}

// Alloc 分配 span
func (c *MCentral) Alloc(nelems int) *Span {
	// 优先从 nonempty 链表获取
	if c.NonEmpty != nil {
		span := c.NonEmpty
		c.NonEmpty = span.Next
		return span
	}
	
	// 创建新 span
	span := NewSpan(c.SizeClass, nelems)
	c.NMalloc++
	return span
}

// Free 释放 span 到 nonempty 链表
func (c *MCentral) Free(span *Span) {
	span.Next = c.NonEmpty
	c.NonEmpty = span
}

// ========== 5. mheap ==========

/*
mheap 结构（全局堆）：
- 虚拟内存管理
- span 分配/回收
- 页面映射
*/

// MHeap 模拟 mheap
type MHeap struct {
	Central [67]*MCentral
	Spans   []*Span
}

// NewMHeap 创建 heap
func NewMHeap() *MHeap {
	h := &MHeap{
		Spans: make([]*Span, 0),
	}
	for i := 0; i < 67; i++ {
		h.Central[i] = NewMCentral(i)
	}
	return h
}

// AllocSpan 从 heap 分配 span
func (h *MHeap) AllocSpan(sizeClass int, nelems int) *Span {
	if sizeClass >= 0 && sizeClass < 67 {
		return h.Central[sizeClass].Alloc(nelems)
	}
	// 大对象
	span := NewSpan(-1, 1)
	h.Spans = append(h.Spans, span)
	return span
}

// ========== 6. 内存分配流程 ==========

/*
内存分配流程：

1. 计算大小类
   - < 32KB: 使用 mcache
   - >= 32KB: 直接从 heap

2. 小对象分配 (mcache):
   a. 计算 sizeclass
   b. 检查 mcache 对应 span
   c. 有空闲: 直接分配
   d. 无空闲: 从 mcentral 获取

3. 大对象分配 (heap):
   a. 查找合适的 span
   b. 不足则向 OS 申请
   c. 返回指针
*/

// MemoryAllocator 内存分配器
type MemoryAllocator struct {
	mcache *MCache
	mheap  *MHeap
}

// NewMemoryAllocator 创建内存分配器
func NewMemoryAllocator() *MemoryAllocator {
	return &MemoryAllocator{
		mcache: NewMCache(),
		mheap:  NewMHeap(),
	}
}

// Alloc 分配内存
func (a *MemoryAllocator) Alloc(size int) (int, int) {
	if size <= 0 {
		return -1, -1
	}
	
	// 计算大小类
	sizeClass := calcSizeClass(size)
	
	if size > 32*1024 {
		// 大对象
		span := a.mheap.AllocSpan(-1, 1)
		return -1, span.Alloc()
	}
	
	// 小对象
	nelems := 512 / ((sizeClass + 1) * 8)
	span := a.mcache.GetSpan(sizeClass, nelems)
	if span == nil {
		span = a.mheap.AllocSpan(sizeClass, nelems)
		a.mcache.Spans[sizeClass] = span
	}
	
	return sizeClass, span.Alloc()
}

// calcSizeClass 计算大小类
func calcSizeClass(size int) int {
	if size <= 8 {
		return 0
	}
	if size >= 32*1024 {
		return 66
	}
	return (size + 7) / 8
}

// ========== 7. 大小类表 ==========

/*
Go 67种大小类 (sizeclasses)：

sizeclass 0: 8字节
sizeclass 1: 16字节
sizeclass 2: 24字节
...
sizeclass 66: 32768字节
*/

// SizeClassInfo 大小类信息
type SizeClassInfo struct {
	Size     int // 对象大小
	Nelems   int // 对象数量
	SpanSize int // span大小
}

// GetSizeClassInfo 获取大小类信息
func GetSizeClassInfo(sizeClass int) SizeClassInfo {
	if sizeClass < 0 || sizeClass >= 67 {
		return SizeClassInfo{}
	}
	size := getSizeForClass(sizeClass)
	return SizeClassInfo{
		Size:     size,
		Nelems:   8192 / size,
		SpanSize: 8192,
	}
}

// getSizeForClass 根据大小类获取大小
func getSizeForClass(sizeClass int) int {
	if sizeClass == 0 {
		return 8
	}
	return (sizeClass + 1) * 8
}

// ========== 8. 内存对齐 ==========

/*
内存对齐规则：
1. 基本类型对齐：8字节类型8字节对齐
2. 结构体对齐：最大字段的倍数
3. 数组对齐：元素类型对齐
*/

// AlignSize 对齐到指定倍数
func AlignSize(size, align int) int {
	return ((size + align - 1) / align) * align
}

// ========== 9. 内存分配示例 ==========

// MemoryAllocDemo 内存分配示例
func MemoryAllocDemo() {
	fmt.Println("=== 内存分配器演示 ===")
	
	allocator := NewMemoryAllocator()
	
	// 分配小对象
	fmt.Println("\n分配小对象:")
	for i := 0; i < 5; i++ {
		sizeClass, idx := allocator.Alloc(100)
		fmt.Printf("  分配100字节: sizeClass=%d, index=%d\n", sizeClass, idx)
	}
	
	// 分配不同大小
	fmt.Println("\n分配不同大小:")
	sizes := []int{8, 16, 32, 64, 128, 256, 512, 1024}
	for _, size := range sizes {
		sizeClass, idx := allocator.Alloc(size)
		fmt.Printf("  分配%d字节: sizeClass=%d, index=%d\n", size, sizeClass, idx)
	}
	
	// 大小类信息
	fmt.Println("\n大小类信息:")
	for i := 0; i < 5; i++ {
		info := GetSizeClassInfo(i)
		fmt.Printf("  SizeClass %d: size=%d, nelems=%d\n", i, info.Size, info.Nelems)
	}
}

// ========== 10. Span 操作示例 ==========

// SpanDemo Span操作演示
func SpanDemo() {
	fmt.Println("\n=== Span 操作演示 ===")
	
	// 创建 span
	span := NewSpan(1, 32) // sizeClass=1 (16字节), 32个元素
	fmt.Printf("创建 Span: sizeClass=%d, nelems=%d\n", span.SizeClass, span.Nelems)
	
	// 分配对象
	fmt.Println("\n分配对象:")
	for i := 0; i < 5; i++ {
		idx := span.Alloc()
		fmt.Printf("  分配: index=%d\n", idx)
	}
	
	// 释放对象
	fmt.Println("\n释放对象:")
	span.Free(2)
	fmt.Println("  释放 index=2")
	span.Free(4)
	fmt.Println("  释放 index=4")
}

// CompleteExample 完整示例
func CompleteExample() {
	MemoryAllocDemo()
	SpanDemo()
}
