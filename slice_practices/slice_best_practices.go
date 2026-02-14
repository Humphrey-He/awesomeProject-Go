package slice_practices

import (
	"fmt"
)

// ========== Slice 最佳实践 ==========

// 1. 预分配容量
// Bad: 动态扩容会导致多次内存分配和复制
func AppendWithoutPrealloc(n int) []int {
	var result []int
	for i := 0; i < n; i++ {
		result = append(result, i)
	}
	return result
}

// Good: 预分配足够容量，避免扩容
func AppendWithPrealloc(n int) []int {
	result := make([]int, 0, n)
	for i := 0; i < n; i++ {
		result = append(result, i)
	}
	return result
}

// 2. 使用 make 初始化时明确长度和容量
// Bad: 未明确容量
func MakeBad() []int {
	return make([]int, 10) // 长度和容量都是10，可能不够
}

// Good: 明确指定容量
func MakeGood() []int {
	return make([]int, 0, 100) // 长度0，容量100，避免扩容
}

// 3. 避免切片内存泄漏
// Bad: 保留了整个底层数组的引用
func SubsliceBad(s []int) []int {
	// 只需要前10个元素，但底层数组可能很大
	return s[:10]
}

// Good: 复制需要的部分，释放原数组
func SubsliceGood(s []int) []int {
	result := make([]int, 10)
	copy(result, s[:10])
	return result
}

// 4. 正确的删除元素方式
// 删除单个元素（保持顺序）
func RemoveElement(s []int, index int) []int {
	if index < 0 || index >= len(s) {
		return s
	}
	// 使用 copy 移动元素
	return append(s[:index], s[index+1:]...)
}

// 删除单个元素（快速方式，不保持顺序）
func RemoveElementFast(s []int, index int) []int {
	if index < 0 || index >= len(s) {
		return s
	}
	// 用最后一个元素替换要删除的元素
	s[index] = s[len(s)-1]
	return s[:len(s)-1]
}

// 5. 过滤切片
// Bad: 原地修改可能导致问题
func FilterBad(s []int, keep func(int) bool) []int {
	j := 0
	for i := 0; i < len(s); i++ {
		if keep(s[i]) {
			s[j] = s[i]
			j++
		}
	}
	return s[:j]
}

// Good: 创建新切片
func FilterGood(s []int, keep func(int) bool) []int {
	result := make([]int, 0, len(s))
	for _, v := range s {
		if keep(v) {
			result = append(result, v)
		}
	}
	return result
}

// 6. 清空切片
// Bad: 重新赋值会丢失容量
func ClearBad(s []int) []int {
	return []int{} // 容量丢失
}

// Good: 保留容量
func ClearGood(s []int) []int {
	return s[:0] // 长度为0，但容量保留
}

// 7. 复制切片
// Bad: 直接赋值只是引用
func CopyBad(src []int) []int {
	return src // 共享底层数组
}

// Good: 深拷贝
func CopyGood(src []int) []int {
	dst := make([]int, len(src))
	copy(dst, src)
	return dst
}

// 8. 安全的切片操作
// 检查边界，避免panic
func SafeSlice(s []int, start, end int) []int {
	if start < 0 {
		start = 0
	}
	if end > len(s) {
		end = len(s)
	}
	if start > end {
		start = end
	}
	return s[start:end]
}

// 9. 反转切片
// 原地反转（高效）
func ReverseInPlace(s []int) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// 创建反转的新切片
func Reverse(s []int) []int {
	result := make([]int, len(s))
	for i, v := range s {
		result[len(s)-1-i] = v
	}
	return result
}

// 10. 去重
// 保持顺序的去重
func UniqueOrdered(s []int) []int {
	seen := make(map[int]bool)
	result := make([]int, 0, len(s))
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// 不保持顺序的去重（使用map）
func Unique(s []int) []int {
	seen := make(map[int]bool)
	for _, v := range s {
		seen[v] = true
	}
	result := make([]int, 0, len(seen))
	for v := range seen {
		result = append(result, v)
	}
	return result
}

// ========== 高级技巧 ==========

// 11. 扩容策略理解
func DemonstrateCapacityGrowth() {
	s := make([]int, 0)
	fmt.Println("初始容量:", cap(s))

	for i := 0; i < 100; i++ {
		oldCap := cap(s)
		s = append(s, i)
		newCap := cap(s)
		if oldCap != newCap {
			fmt.Printf("长度 %d: 容量从 %d 增长到 %d (%.2fx)\n",
				len(s), oldCap, newCap, float64(newCap)/float64(oldCap))
		}
	}
}

// 12. 切片作为参数传递
// Bad: 修改会影响原切片的内容（但不影响长度）
func ModifySliceBad(s []int) {
	s[0] = 999 // 会影响原切片
}

// Good: 明确返回新切片
func ModifySliceGood(s []int) []int {
	result := make([]int, len(s))
	copy(result, s)
	result[0] = 999
	return result
}

// 13. 避免在循环中使用 append
// Bad: 多次append可能触发扩容
func AppendInLoopBad(n int) []int {
	var result []int
	for i := 0; i < n; i++ {
		for j := 0; j < 10; j++ {
			result = append(result, j)
		}
	}
	return result
}

// Good: 预先计算总大小
func AppendInLoopGood(n int) []int {
	result := make([]int, 0, n*10)
	for i := 0; i < n; i++ {
		for j := 0; j < 10; j++ {
			result = append(result, j)
		}
	}
	return result
}

// 14. 多维切片初始化
// Bad: 共享内部切片
func Make2DSliceBad(rows, cols int) [][]int {
	matrix := make([][]int, rows)
	row := make([]int, cols)
	for i := range matrix {
		matrix[i] = row // 错误：所有行共享同一个底层数组
	}
	return matrix
}

// Good: 每行独立
func Make2DSliceGood(rows, cols int) [][]int {
	matrix := make([][]int, rows)
	for i := range matrix {
		matrix[i] = make([]int, cols)
	}
	return matrix
}

// 15. 批量操作优化
// 使用 copy 比逐个赋值快
func BatchCopyGood(dst, src []int) {
	copy(dst, src)
}

func BatchCopyBad(dst, src []int) {
	for i := 0; i < len(src) && i < len(dst); i++ {
		dst[i] = src[i]
	}
}

// ========== 切片陷阱 ==========

// 陷阱1: 切片扩容后的底层数组变化
func TrapCapacityChange() {
	s1 := []int{1, 2, 3}
	s2 := s1

	fmt.Printf("s1 初始地址: %p\n", s1)

	// 扩容
	s1 = append(s1, 4, 5, 6, 7, 8)

	fmt.Printf("s1 扩容后地址: %p\n", s1)
	fmt.Printf("s2 地址: %p\n", s2)

	// s1 和 s2 现在指向不同的底层数组
	s1[0] = 999
	fmt.Printf("s1[0]=%d, s2[0]=%d\n", s1[0], s2[0])
}

// 陷阱2: 切片作为函数参数
func TrapSliceParameter() {
	s := []int{1, 2, 3}
	modifySlice(s)
	fmt.Println("外部:", s) // [999, 2, 3]

	appendSlice(s)
	fmt.Println("外部:", s) // 仍然是 [999, 2, 3]
}

func modifySlice(s []int) {
	s[0] = 999 // 会影响原切片
}

func appendSlice(s []int) {
	s = append(s, 4) // 不影响原切片的长度
	fmt.Println("内部:", s)
}

// 陷阱3: range 循环中的指针
func TrapRangePointer() {
	s := []int{1, 2, 3}
	var ptrs []*int

	// Bad: 所有指针都指向同一个变量
	for _, v := range s {
		ptrs = append(ptrs, &v) // v 是循环变量，地址不变
	}

	for _, p := range ptrs {
		fmt.Print(*p, " ") // 输出: 3 3 3
	}
	fmt.Println()

	// Good: 使用索引或创建新变量
	ptrs = nil
	for i := range s {
		v := s[i] // 创建新变量
		ptrs = append(ptrs, &v)
	}

	for _, p := range ptrs {
		fmt.Print(*p, " ") // 输出: 1 2 3
	}
	fmt.Println()
}

// ========== 性能优化技巧 ==========

// 技巧1: 使用索引而非 range（某些场景更快）
func SumWithRange(s []int) int {
	sum := 0
	for _, v := range s {
		sum += v
	}
	return sum
}

func SumWithIndex(s []int) int {
	sum := 0
	for i := 0; i < len(s); i++ {
		sum += s[i]
	}
	return sum
}

// 技巧2: 避免不必要的切片操作
func ProcessBad(s []int) int {
	// 每次都创建新切片
	filtered := FilterGood(s, func(v int) bool { return v > 0 })
	sum := 0
	for _, v := range filtered {
		sum += v
	}
	return sum
}

func ProcessGood(s []int) int {
	// 直接处理，避免创建新切片
	sum := 0
	for _, v := range s {
		if v > 0 {
			sum += v
		}
	}
	return sum
}

// ========== 泛型切片操作（Go 1.18+）==========

// Map: 转换切片元素
func Map[T, U any](s []T, f func(T) U) []U {
	result := make([]U, len(s))
	for i, v := range s {
		result[i] = f(v)
	}
	return result
}

// Filter: 过滤切片元素
func Filter[T any](s []T, keep func(T) bool) []T {
	result := make([]T, 0, len(s))
	for _, v := range s {
		if keep(v) {
			result = append(result, v)
		}
	}
	return result
}

// Reduce: 归约操作
func Reduce[T, U any](s []T, init U, f func(U, T) U) U {
	acc := init
	for _, v := range s {
		acc = f(acc, v)
	}
	return acc
}

// Contains: 检查元素是否存在
func Contains[T comparable](s []T, target T) bool {
	for _, v := range s {
		if v == target {
			return true
		}
	}
	return false
}

// ========== 最佳实践总结 ==========

/*
Slice 最佳实践清单：

✅ 1. 预分配容量
   - 已知大小时使用 make([]T, 0, capacity)
   - 避免多次扩容带来的性能损失

✅ 2. 理解底层数组
   - 切片是引用类型，共享底层数组
   - 扩容会创建新数组

✅ 3. 避免内存泄漏
   - 大切片的小子切片会保留整个底层数组
   - 使用 copy 创建独立切片

✅ 4. 正确删除元素
   - 保持顺序: append(s[:i], s[i+1:]...)
   - 快速删除: s[i] = s[len(s)-1]; s = s[:len(s)-1]

✅ 5. 切片作为参数
   - 修改元素会影响原切片
   - append 不会影响原切片长度
   - 需要修改长度时返回新切片

✅ 6. 避免常见陷阱
   - range 循环中的指针问题
   - 切片扩容后的地址变化
   - nil 切片 vs 空切片

✅ 7. 性能优化
   - 批量操作使用 copy
   - 减少不必要的切片创建
   - 预分配足够的容量

✅ 8. 并发安全
   - 切片本身不是并发安全的
   - 需要使用互斥锁或通道保护

✅ 9. 使用泛型（Go 1.18+）
   - 编写通用的切片操作函数
   - 提高代码复用性

✅ 10. 测试和基准测试
   - 验证正确性
   - 对比不同实现的性能
*/
