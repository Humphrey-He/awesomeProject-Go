package copy_semantics

// Go 的内建 copy 并不是简单一句“深拷贝”或“浅拷贝”可以概括。
//
// 结论：
// 1) copy 会“逐元素复制”切片元素。
// 2) 如果元素本身是值类型（如 int、struct 不含引用字段），效果接近深拷贝。
// 3) 如果元素含引用语义（指针、slice、map、chan、func、interface 内含引用对象），
//    则只复制“引用”，底层对象共享，表现为浅拷贝。

type CopyObservation struct {
	Original any
	Copied   any
	Note     string
}

// CopyValueSliceDemo 展示 []int copy 后互不影响。
func CopyValueSliceDemo() CopyObservation {
	src := []int{1, 2, 3}
	dst := make([]int, len(src))
	copy(dst, src)

	src[0] = 100
	return CopyObservation{
		Original: src,
		Copied:   dst,
		Note:     "[]int 元素是值类型，copy 后修改 src 不影响 dst 对应元素",
	}
}

// CopyNestedSliceDemo 展示 [][]int copy 后内层共享。
func CopyNestedSliceDemo() CopyObservation {
	src := [][]int{{1, 2}, {3, 4}}
	dst := make([][]int, len(src))
	copy(dst, src)

	// 修改内层切片，src/dst 都会看到变化（共享内层数组）
	src[0][0] = 999
	return CopyObservation{
		Original: src,
		Copied:   dst,
		Note:     "[][]int 仅复制到内层切片头，内层底层数组共享，属于浅层复制",
	}
}

// DeepCopy2DInt 手写深拷贝二维切片。
func DeepCopy2DInt(src [][]int) [][]int {
	dst := make([][]int, len(src))
	for i := range src {
		if src[i] == nil {
			continue
		}
		dst[i] = make([]int, len(src[i]))
		copy(dst[i], src[i])
	}
	return dst
}

type Node struct {
	Value int
}

type Wrapper struct {
	Name string
	Ptr  *Node
}

// CopyStructWithPointerDemo 展示 struct 含指针字段时 copy 依然共享指针目标。
func CopyStructWithPointerDemo() CopyObservation {
	src := []Wrapper{
		{Name: "a", Ptr: &Node{Value: 1}},
	}
	dst := make([]Wrapper, len(src))
	copy(dst, src)

	src[0].Ptr.Value = 42
	return CopyObservation{
		Original: src,
		Copied:   dst,
		Note:     "Wrapper 的 Ptr 指向同一 Node，修改指针目标会互相可见",
	}
}
