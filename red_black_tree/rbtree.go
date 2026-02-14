package red_black_tree

import "fmt"

// Color 节点颜色
type Color bool

const (
	RED   Color = false
	BLACK Color = true
)

// Node 红黑树节点
type Node[K comparable, V any] struct {
	Key    K
	Value  V
	Color  Color
	Left   *Node[K, V]
	Right  *Node[K, V]
	Parent *Node[K, V]
}

// RBTree 红黑树
type RBTree[K comparable, V any] struct {
	root    *Node[K, V]
	size    int
	compare func(K, K) int // 比较函数：-1(less), 0(equal), 1(greater)
}

// NewRBTree 创建红黑树
func NewRBTree[K comparable, V any](compare func(K, K) int) *RBTree[K, V] {
	return &RBTree[K, V]{
		compare: compare,
	}
}

// Insert 插入键值对
func (t *RBTree[K, V]) Insert(key K, value V) {
	node := &Node[K, V]{
		Key:   key,
		Value: value,
		Color: RED, // 新节点默认为红色
	}

	if t.root == nil {
		node.Color = BLACK // 根节点必须是黑色
		t.root = node
		t.size++
		return
	}

	// 查找插入位置
	parent := t.findParent(key)
	node.Parent = parent

	if t.compare(key, parent.Key) < 0 {
		parent.Left = node
	} else if t.compare(key, parent.Key) > 0 {
		parent.Right = node
	} else {
		// key已存在，更新value
		parent.Value = value
		return
	}

	t.size++
	t.fixAfterInsertion(node)
}

// findParent 查找插入位置的父节点
func (t *RBTree[K, V]) findParent(key K) *Node[K, V] {
	current := t.root
	var parent *Node[K, V]

	for current != nil {
		parent = current
		cmp := t.compare(key, current.Key)
		if cmp < 0 {
			current = current.Left
		} else if cmp > 0 {
			current = current.Right
		} else {
			return current
		}
	}

	return parent
}

// fixAfterInsertion 插入后修复红黑树性质
func (t *RBTree[K, V]) fixAfterInsertion(node *Node[K, V]) {
	for node != t.root && node.Parent.Color == RED {
		if node.Parent == node.Parent.Parent.Left {
			uncle := node.Parent.Parent.Right

			if uncle != nil && uncle.Color == RED {
				// Case 1: 叔叔节点是红色
				node.Parent.Color = BLACK
				uncle.Color = BLACK
				node.Parent.Parent.Color = RED
				node = node.Parent.Parent
			} else {
				if node == node.Parent.Right {
					// Case 2: 节点是右孩子
					node = node.Parent
					t.rotateLeft(node)
				}
				// Case 3: 节点是左孩子
				node.Parent.Color = BLACK
				node.Parent.Parent.Color = RED
				t.rotateRight(node.Parent.Parent)
			}
		} else {
			// 镜像情况
			uncle := node.Parent.Parent.Left

			if uncle != nil && uncle.Color == RED {
				node.Parent.Color = BLACK
				uncle.Color = BLACK
				node.Parent.Parent.Color = RED
				node = node.Parent.Parent
			} else {
				if node == node.Parent.Left {
					node = node.Parent
					t.rotateRight(node)
				}
				node.Parent.Color = BLACK
				node.Parent.Parent.Color = RED
				t.rotateLeft(node.Parent.Parent)
			}
		}
	}

	t.root.Color = BLACK
}

// rotateLeft 左旋
func (t *RBTree[K, V]) rotateLeft(node *Node[K, V]) {
	right := node.Right
	node.Right = right.Left

	if right.Left != nil {
		right.Left.Parent = node
	}

	right.Parent = node.Parent

	if node.Parent == nil {
		t.root = right
	} else if node == node.Parent.Left {
		node.Parent.Left = right
	} else {
		node.Parent.Right = right
	}

	right.Left = node
	node.Parent = right
}

// rotateRight 右旋
func (t *RBTree[K, V]) rotateRight(node *Node[K, V]) {
	left := node.Left
	node.Left = left.Right

	if left.Right != nil {
		left.Right.Parent = node
	}

	left.Parent = node.Parent

	if node.Parent == nil {
		t.root = left
	} else if node == node.Parent.Right {
		node.Parent.Right = left
	} else {
		node.Parent.Left = left
	}

	left.Right = node
	node.Parent = left
}

// Search 查找
func (t *RBTree[K, V]) Search(key K) (V, bool) {
	node := t.searchNode(key)
	if node == nil {
		var zero V
		return zero, false
	}
	return node.Value, true
}

func (t *RBTree[K, V]) searchNode(key K) *Node[K, V] {
	current := t.root

	for current != nil {
		cmp := t.compare(key, current.Key)
		if cmp < 0 {
			current = current.Left
		} else if cmp > 0 {
			current = current.Right
		} else {
			return current
		}
	}

	return nil
}

// Delete 删除
func (t *RBTree[K, V]) Delete(key K) bool {
	node := t.searchNode(key)
	if node == nil {
		return false
	}

	t.deleteNode(node)
	t.size--
	return true
}

func (t *RBTree[K, V]) deleteNode(node *Node[K, V]) {
	// 如果有两个子节点，找到后继节点替换
	if node.Left != nil && node.Right != nil {
		successor := t.minimum(node.Right)
		node.Key = successor.Key
		node.Value = successor.Value
		node = successor
	}

	// 现在node最多有一个子节点
	var replacement *Node[K, V]
	if node.Left != nil {
		replacement = node.Left
	} else {
		replacement = node.Right
	}

	if replacement != nil {
		replacement.Parent = node.Parent

		if node.Parent == nil {
			t.root = replacement
		} else if node == node.Parent.Left {
			node.Parent.Left = replacement
		} else {
			node.Parent.Right = replacement
		}

		if node.Color == BLACK {
			t.fixAfterDeletion(replacement)
		}
	} else if node.Parent == nil {
		t.root = nil
	} else {
		if node.Color == BLACK {
			t.fixAfterDeletion(node)
		}

		if node.Parent != nil {
			if node == node.Parent.Left {
				node.Parent.Left = nil
			} else {
				node.Parent.Right = nil
			}
		}
	}
}

func (t *RBTree[K, V]) fixAfterDeletion(node *Node[K, V]) {
	for node != t.root && (node == nil || node.Color == BLACK) {
		if node == node.Parent.Left {
			sibling := node.Parent.Right

			if sibling.Color == RED {
				sibling.Color = BLACK
				node.Parent.Color = RED
				t.rotateLeft(node.Parent)
				sibling = node.Parent.Right
			}

			if (sibling.Left == nil || sibling.Left.Color == BLACK) &&
				(sibling.Right == nil || sibling.Right.Color == BLACK) {
				sibling.Color = RED
				node = node.Parent
			} else {
				if sibling.Right == nil || sibling.Right.Color == BLACK {
					if sibling.Left != nil {
						sibling.Left.Color = BLACK
					}
					sibling.Color = RED
					t.rotateRight(sibling)
					sibling = node.Parent.Right
				}

				sibling.Color = node.Parent.Color
				node.Parent.Color = BLACK
				if sibling.Right != nil {
					sibling.Right.Color = BLACK
				}
				t.rotateLeft(node.Parent)
				node = t.root
			}
		} else {
			// 镜像情况
			sibling := node.Parent.Left

			if sibling.Color == RED {
				sibling.Color = BLACK
				node.Parent.Color = RED
				t.rotateRight(node.Parent)
				sibling = node.Parent.Left
			}

			if (sibling.Right == nil || sibling.Right.Color == BLACK) &&
				(sibling.Left == nil || sibling.Left.Color == BLACK) {
				sibling.Color = RED
				node = node.Parent
			} else {
				if sibling.Left == nil || sibling.Left.Color == BLACK {
					if sibling.Right != nil {
						sibling.Right.Color = BLACK
					}
					sibling.Color = RED
					t.rotateLeft(sibling)
					sibling = node.Parent.Left
				}

				sibling.Color = node.Parent.Color
				node.Parent.Color = BLACK
				if sibling.Left != nil {
					sibling.Left.Color = BLACK
				}
				t.rotateRight(node.Parent)
				node = t.root
			}
		}
	}

	if node != nil {
		node.Color = BLACK
	}
}

// minimum 找到最小节点
func (t *RBTree[K, V]) minimum(node *Node[K, V]) *Node[K, V] {
	for node.Left != nil {
		node = node.Left
	}
	return node
}

// Size 返回树的大小
func (t *RBTree[K, V]) Size() int {
	return t.size
}

// Height 计算树的高度
func (t *RBTree[K, V]) Height() int {
	return t.height(t.root)
}

func (t *RBTree[K, V]) height(node *Node[K, V]) int {
	if node == nil {
		return 0
	}

	leftHeight := t.height(node.Left)
	rightHeight := t.height(node.Right)

	if leftHeight > rightHeight {
		return leftHeight + 1
	}
	return rightHeight + 1
}

// InOrder 中序遍历
func (t *RBTree[K, V]) InOrder() []K {
	result := make([]K, 0, t.size)
	t.inOrder(t.root, &result)
	return result
}

func (t *RBTree[K, V]) inOrder(node *Node[K, V], result *[]K) {
	if node == nil {
		return
	}

	t.inOrder(node.Left, result)
	*result = append(*result, node.Key)
	t.inOrder(node.Right, result)
}

// Print 打印树结构（用于调试）
func (t *RBTree[K, V]) Print() {
	t.print(t.root, "", true)
}

func (t *RBTree[K, V]) print(node *Node[K, V], prefix string, isTail bool) {
	if node == nil {
		return
	}

	color := "R"
	if node.Color == BLACK {
		color = "B"
	}

	fmt.Printf("%s%s%v(%s)\n", prefix, map[bool]string{true: "└── ", false: "├── "}[isTail], node.Key, color)

	if node.Left != nil || node.Right != nil {
		if node.Right != nil {
			newPrefix := prefix + map[bool]string{true: "    ", false: "│   "}[isTail]
			t.print(node.Right, newPrefix, false)
		}

		if node.Left != nil {
			newPrefix := prefix + map[bool]string{true: "    ", false: "│   "}[isTail]
			t.print(node.Left, newPrefix, true)
		}
	}
}

// IntCompare int类型的比较函数
func IntCompare(a, b int) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}
	return 0
}

// StringCompare string类型的比较函数
func StringCompare(a, b string) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}
	return 0
}
