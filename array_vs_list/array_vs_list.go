package array_vs_list

import (
	"fmt"
)

// ========== 数组实现 ==========

// ArrayDS 基于数组的数据结构
type ArrayDS struct {
	data []int
}

// NewArrayDS 创建一个新的数组数据结构
func NewArrayDS(capacity int) *ArrayDS {
	return &ArrayDS{
		data: make([]int, 0, capacity),
	}
}

// Append 在末尾添加元素 O(1) amortized
func (a *ArrayDS) Append(val int) {
	a.data = append(a.data, val)
}

// Insert 在指定位置插入元素 O(n)
func (a *ArrayDS) Insert(index, val int) error {
	if index < 0 || index > len(a.data) {
		return fmt.Errorf("index out of range")
	}
	a.data = append(a.data[:index], append([]int{val}, a.data[index:]...)...)
	return nil
}

// Delete 删除指定位置的元素 O(n)
func (a *ArrayDS) Delete(index int) error {
	if index < 0 || index >= len(a.data) {
		return fmt.Errorf("index out of range")
	}
	a.data = append(a.data[:index], a.data[index+1:]...)
	return nil
}

// Get 获取指定位置的元素 O(1)
func (a *ArrayDS) Get(index int) (int, error) {
	if index < 0 || index >= len(a.data) {
		return 0, fmt.Errorf("index out of range")
	}
	return a.data[index], nil
}

// Search 查找元素 O(n)
func (a *ArrayDS) Search(val int) int {
	for i, v := range a.data {
		if v == val {
			return i
		}
	}
	return -1
}

// Len 返回数组长度
func (a *ArrayDS) Len() int {
	return len(a.data)
}

// ========== 链表实现 ==========

// Node 链表节点
type Node struct {
	Val  int
	Next *Node
}

// LinkedList 单向链表
type LinkedList struct {
	head *Node
	tail *Node
	size int
}

// NewLinkedList 创建一个新的链表
func NewLinkedList() *LinkedList {
	return &LinkedList{}
}

// Append 在末尾添加元素 O(1)
func (l *LinkedList) Append(val int) {
	node := &Node{Val: val}
	if l.head == nil {
		l.head = node
		l.tail = node
	} else {
		l.tail.Next = node
		l.tail = node
	}
	l.size++
}

// Insert 在指定位置插入元素 O(n)
func (l *LinkedList) Insert(index, val int) error {
	if index < 0 || index > l.size {
		return fmt.Errorf("index out of range")
	}

	node := &Node{Val: val}

	if index == 0 {
		node.Next = l.head
		l.head = node
		if l.tail == nil {
			l.tail = node
		}
		l.size++
		return nil
	}

	curr := l.head
	for i := 0; i < index-1; i++ {
		curr = curr.Next
	}

	node.Next = curr.Next
	curr.Next = node

	if node.Next == nil {
		l.tail = node
	}

	l.size++
	return nil
}

// Delete 删除指定位置的元素 O(n)
func (l *LinkedList) Delete(index int) error {
	if index < 0 || index >= l.size {
		return fmt.Errorf("index out of range")
	}

	if index == 0 {
		l.head = l.head.Next
		if l.head == nil {
			l.tail = nil
		}
		l.size--
		return nil
	}

	curr := l.head
	for i := 0; i < index-1; i++ {
		curr = curr.Next
	}

	curr.Next = curr.Next.Next
	if curr.Next == nil {
		l.tail = curr
	}

	l.size--
	return nil
}

// Get 获取指定位置的元素 O(n)
func (l *LinkedList) Get(index int) (int, error) {
	if index < 0 || index >= l.size {
		return 0, fmt.Errorf("index out of range")
	}

	curr := l.head
	for i := 0; i < index; i++ {
		curr = curr.Next
	}

	return curr.Val, nil
}

// Search 查找元素 O(n)
func (l *LinkedList) Search(val int) int {
	curr := l.head
	index := 0
	for curr != nil {
		if curr.Val == val {
			return index
		}
		curr = curr.Next
		index++
	}
	return -1
}

// Len 返回链表长度
func (l *LinkedList) Len() int {
	return l.size
}

// ========== 性能分析 ==========

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	Operation   string
	ArrayTime   int64  // 纳秒
	ListTime    int64  // 纳秒
	ArrayMemory uint64 // 字节
	ListMemory  uint64 // 字节
}

// WhyArrayIsBetter 解释为什么数组更适合现代互联网开发
func WhyArrayIsBetter() []string {
	return []string{
		"1. CPU缓存友好性（Cache Locality）",
		"   - 数组元素在内存中连续存储，访问时能充分利用CPU缓存行（Cache Line）",
		"   - 链表节点分散在堆内存中，每次访问都可能导致缓存未命中（Cache Miss）",
		"   - 现代CPU的L1/L2/L3缓存对连续内存访问有巨大性能优势",
		"",
		"2. 内存效率",
		"   - 数组只存储数据本身",
		"   - 链表每个节点需要额外存储指针（64位系统中每个指针8字节）",
		"   - 对于int类型，链表的内存开销是数组的3倍以上（数据+指针+内存对齐）",
		"",
		"3. 内存分配效率",
		"   - 数组一次性分配连续内存，分配和释放都很快",
		"   - 链表每个节点都需要单独分配内存，增加GC压力",
		"   - Go的内存分配器对大块连续内存分配有优化",
		"",
		"4. 遍历性能",
		"   - 数组遍历可以利用CPU预取（Prefetch）和SIMD指令",
		"   - 链表遍历需要频繁解引用指针，无法预测下一个节点位置",
		"   - 数组遍历速度通常是链表的5-10倍",
		"",
		"5. 随机访问",
		"   - 数组支持O(1)的随机访问",
		"   - 链表必须从头遍历，O(n)时间复杂度",
		"   - 现代应用中随机访问场景非常常见",
		"",
		"6. 编译器优化",
		"   - 数组操作更容易被编译器优化（循环展开、向量化等）",
		"   - 链表的指针跳转限制了编译器优化空间",
		"",
		"7. 现代硬件特性",
		"   - 现代CPU的分支预测器对连续内存访问更友好",
		"   - TLB（Translation Lookaside Buffer）对连续内存访问更高效",
		"   - 内存带宽利用率：数组接近100%，链表通常<20%",
		"",
		"8. 实际应用场景",
		"   - 互联网应用中，读操作远多于写操作",
		"   - 即使需要频繁插入/删除，Go的slice扩容机制也很高效",
		"   - 大部分场景下，数组的整体性能优于链表",
		"",
		"9. Go语言特性",
		"   - Go的slice底层就是数组，append操作经过高度优化",
		"   - Go的GC对大量小对象（链表节点）不友好",
		"   - Go标准库很少使用链表，container/list性能较差",
		"",
		"10. 什么时候才考虑链表？",
		"    - 需要频繁在中间位置插入/删除，且数据量很大",
		"    - 需要实现特殊数据结构（如LRU缓存的双向链表）",
		"    - 元素大小不确定，无法预估容量",
		"    - 但即使这些场景，也可能有更好的替代方案（如ring buffer）",
	}
}

// CacheLineExample 演示缓存行的影响
type CacheLineExample struct {
	// 现代CPU的缓存行通常是64字节
	// 一次可以加载8个int64或16个int32
	data [16]int32 // 恰好一个缓存行
}

// MemoryLayoutComparison 内存布局对比
//func MemoryLayoutComparison() string {
//	return `
//内存布局对比：
//
//数组（连续内存）：
//[0|1|2|3|4|5|6|7|8|9] <- 所有元素紧密排列，一次可加载多个到缓存
//
//链表（分散内存）：
//[0|ptr] -> [1|ptr] -> [2|ptr] -> [3|ptr] -> [4|ptr]
//  堆地址1    堆地址n    堆地址m    堆地址x    堆地址y
//
//每个节点分散在不同的内存位置，每次访问都可能触发缓存未命中
//
//内存占用对比（int32为例）：
//- 数组：10个元素 = 40字节
//- 链表：10个节点 = 10*(4+8+padding) ≈ 160字节（4倍差距）
//`
//}
