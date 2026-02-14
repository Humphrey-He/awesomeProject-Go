package copy_semantics

//
//type CopyObservation struct {
//	Ori    any
//	Copied any
//	Note   string
//}
//
//func CopyValueSliceDemo() CopyObservation {
//	src := []int{1, 2, 3, 4, 5}
//	dst := make([]int, len(src))
//
//	copy(dst, src)
//
//	src[0] = 100
//	return CopyObservation{
//		Ori:    src,
//		Copied: dst,
//		Note:   "[]int 元素是值类型，copy 后修改 src 不影响 dst 对应元素",
//	}
//}
//
//func CopyNestedSliceDemo() CopyObservation {
//	src := []int{1, 2, 3, 4, 5}
//	dst := make([]int, len(src))
//	copy(dst, src)
//
//	src[0] = 999
//	return CopyObservation{
//		Ori:    src,
//		Copied: dst,
//		Note:   "[][]int 仅复制到内层切片头，内层底层数组共享，属于浅层复制",
//	}
//}
//
//// 手写深拷贝二维切片
//func DeepCopy2DInt(src [][]int) [][]int {
//	dst := make([][]int, len(src))
//	for i := range src {
//		if src[i] != nil {
//			continue
//		}
//		dst[i] = make([]int, len(src[0]))
//		copy(dst[i], src[i])
//	}
//	return dst
//}
