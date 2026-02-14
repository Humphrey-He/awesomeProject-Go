package bplustree

import (
	"fmt"
	"sort"
)

// BPlusTree B+树实现（MySQL索引结构简化版）
type BPlusTree[K comparable, V any] struct {
	root    *BPlusNode[K, V]
	order   int                // 阶数（每个节点最多的子节点数）
	compare func(K, K) int     // 比较函数
	size    int                // 键值对数量
}

// BPlusNode B+树节点
type BPlusNode[K comparable, V any] struct {
	isLeaf   bool
	keys     []K
	values   []V                 // 仅叶子节点使用
	children []*BPlusNode[K, V]  // 仅非叶子节点使用
	next     *BPlusNode[K, V]    // 叶子节点链表指针
	parent   *BPlusNode[K, V]
}

// NewBPlusTree 创建B+树
func NewBPlusTree[K comparable, V any](order int, compare func(K, K) int) *BPlusTree[K, V] {
	if order < 3 {
		order = 3 // 最小阶数为3
	}
	
	return &BPlusTree[K, V]{
		root: &BPlusNode[K, V]{
			isLeaf: true,
			keys:   make([]K, 0),
			values: make([]V, 0),
		},
		order:   order,
		compare: compare,
	}
}

// Insert 插入键值对
func (t *BPlusTree[K, V]) Insert(key K, value V) {
	// 查找叶子节点
	leaf := t.findLeaf(key)
	
	// 在叶子节点中插入
	pos := t.findInsertPosition(leaf.keys, key)
	
	// 检查是否已存在
	if pos < len(leaf.keys) && t.compare(leaf.keys[pos], key) == 0 {
		leaf.values[pos] = value // 更新值
		return
	}
	
	// 插入新键值对
	leaf.keys = insert(leaf.keys, pos, key)
	leaf.values = insert(leaf.values, pos, value)
	t.size++
	
	// 如果节点溢出，分裂
	if len(leaf.keys) >= t.order {
		t.splitLeaf(leaf)
	}
}

// findLeaf 查找应该插入的叶子节点
func (t *BPlusTree[K, V]) findLeaf(key K) *BPlusNode[K, V] {
	node := t.root
	
	for !node.isLeaf {
		pos := t.findInsertPosition(node.keys, key)
		if pos >= len(node.children) {
			pos = len(node.children) - 1
		}
		node = node.children[pos]
	}
	
	return node
}

// findInsertPosition 二分查找插入位置
func (t *BPlusTree[K, V]) findInsertPosition(keys []K, key K) int {
	return sort.Search(len(keys), func(i int) bool {
		return t.compare(keys[i], key) >= 0
	})
}

// splitLeaf 分裂叶子节点
func (t *BPlusTree[K, V]) splitLeaf(leaf *BPlusNode[K, V]) {
	mid := len(leaf.keys) / 2
	
	// 创建新节点
	newLeaf := &BPlusNode[K, V]{
		isLeaf: true,
		keys:   make([]K, len(leaf.keys)-mid),
		values: make([]V, len(leaf.values)-mid),
		next:   leaf.next,
		parent: leaf.parent,
	}
	
	copy(newLeaf.keys, leaf.keys[mid:])
	copy(newLeaf.values, leaf.values[mid:])
	
	leaf.keys = leaf.keys[:mid]
	leaf.values = leaf.values[:mid]
	leaf.next = newLeaf
	
	// 向父节点插入分隔键
	t.insertInParent(leaf, newLeaf.keys[0], newLeaf)
}

// insertInParent 向父节点插入
func (t *BPlusTree[K, V]) insertInParent(left *BPlusNode[K, V], key K, right *BPlusNode[K, V]) {
	if left.parent == nil {
		// 创建新根节点
		t.root = &BPlusNode[K, V]{
			isLeaf:   false,
			keys:     []K{key},
			children: []*BPlusNode[K, V]{left, right},
		}
		left.parent = t.root
		right.parent = t.root
		return
	}
	
	parent := left.parent
	pos := t.findInsertPosition(parent.keys, key)
	
	parent.keys = insert(parent.keys, pos, key)
	parent.children = insert(parent.children, pos+1, right)
	right.parent = parent
	
	// 如果父节点溢出，继续分裂
	if len(parent.keys) >= t.order {
		t.splitInternal(parent)
	}
}

// splitInternal 分裂内部节点
func (t *BPlusTree[K, V]) splitInternal(node *BPlusNode[K, V]) {
	mid := len(node.keys) / 2
	splitKey := node.keys[mid]
	
	newNode := &BPlusNode[K, V]{
		isLeaf:   false,
		keys:     make([]K, len(node.keys)-mid-1),
		children: make([]*BPlusNode[K, V], len(node.children)-mid-1),
		parent:   node.parent,
	}
	
	copy(newNode.keys, node.keys[mid+1:])
	copy(newNode.children, node.children[mid+1:])
	
	// 更新子节点的父指针
	for _, child := range newNode.children {
		child.parent = newNode
	}
	
	node.keys = node.keys[:mid]
	node.children = node.children[:mid+1]
	
	t.insertInParent(node, splitKey, newNode)
}

// Search 查找
func (t *BPlusTree[K, V]) Search(key K) (V, bool) {
	leaf := t.findLeaf(key)
	pos := t.findInsertPosition(leaf.keys, key)
	
	if pos < len(leaf.keys) && t.compare(leaf.keys[pos], key) == 0 {
		return leaf.values[pos], true
	}
	
	var zero V
	return zero, false
}

// RangeSearch 范围查询 [startKey, endKey]
func (t *BPlusTree[K, V]) RangeSearch(startKey, endKey K) []V {
	result := make([]V, 0)
	
	// 找到起始叶子节点
	leaf := t.findLeaf(startKey)
	pos := t.findInsertPosition(leaf.keys, startKey)
	
	// 遍历叶子节点链表
	for leaf != nil {
		for i := pos; i < len(leaf.keys); i++ {
			if t.compare(leaf.keys[i], endKey) > 0 {
				return result
			}
			result = append(result, leaf.values[i])
		}
		
		leaf = leaf.next
		pos = 0
	}
	
	return result
}

// Delete 删除（简化实现，不处理合并）
func (t *BPlusTree[K, V]) Delete(key K) bool {
	leaf := t.findLeaf(key)
	pos := t.findInsertPosition(leaf.keys, key)
	
	if pos >= len(leaf.keys) || t.compare(leaf.keys[pos], key) != 0 {
		return false
	}
	
	leaf.keys = removeAt(leaf.keys, pos)
	leaf.values = removeAt(leaf.values, pos)
	t.size--
	
	return true
}

// Size 返回树的大小
func (t *BPlusTree[K, V]) Size() int {
	return t.size
}

// Height 计算树的高度
func (t *BPlusTree[K, V]) Height() int {
	height := 1
	node := t.root
	
	for !node.isLeaf {
		height++
		node = node.children[0]
	}
	
	return height
}

// Print 打印树结构
func (t *BPlusTree[K, V]) Print() {
	t.print(t.root, "", true)
}

func (t *BPlusTree[K, V]) print(node *BPlusNode[K, V], prefix string, isTail bool) {
	if node == nil {
		return
	}
	
	nodeType := "Internal"
	if node.isLeaf {
		nodeType = "Leaf"
	}
	
	fmt.Printf("%s%s%s: %v\n",
		prefix,
		map[bool]string{true: "└── ", false: "├── "}[isTail],
		nodeType,
		node.keys)
	
	if !node.isLeaf {
		for i, child := range node.children {
			newPrefix := prefix + map[bool]string{true: "    ", false: "│   "}[isTail]
			t.print(child, newPrefix, i == len(node.children)-1)
		}
	}
}

// ========== 辅助函数 ==========

// insert 在切片指定位置插入元素
func insert[T any](slice []T, pos int, value T) []T {
	slice = append(slice, value)
	copy(slice[pos+1:], slice[pos:])
	slice[pos] = value
	return slice
}

// removeAt 删除切片指定位置的元素
func removeAt[T any](slice []T, pos int) []T {
	return append(slice[:pos], slice[pos+1:]...)
}

// IntCompare int比较函数
func IntCompare(a, b int) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}
	return 0
}

// StringCompare string比较函数
func StringCompare(a, b string) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}
	return 0
}

