package generic_advanced

import (
	"fmt"
	"sort"
)

// ========== Go 泛型进阶：深入理解与最佳实践 ==========

/*
本项目讲解 Go 泛型的进阶特性，包括：

一、泛型基础回顾
1. 类型参数
2. 类型约束
3. 泛型函数

二、泛型进阶特性
1. 泛型方法
2. 泛型结构体
3. 泛型接口
4. 类型列表与枚举

三、泛型高级用法
1. 泛型 Map 和 Set
2. 泛型算法实现
3. 泛型与接口的组合

四、性能优化
1. 避免动态类型分发
2. 内存优化

五、泛型编译原理
1. 类型参数实例化
2. 代码生成
*/

// ========== 1. 泛型基础 ==========

// ========== 1.1 类型约束 ==========

// Number 数值类型约束（使用接口）
type Number interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		float32 | float64
}

// SignedNumber 有符号数约束
type SignedNumber interface {
	int | int8 | int16 | int32 | int64 | float32 | float64
}

// Comparable 可比较约束（内置约束）
type Comparable interface {
	comparable
}

// Stringable 可转换为字符串
type Stringable interface {
	String() string
}

// ========== 1.2 泛型函数 ==========

// Sum 求和函数 - 泛型实现
func Sum[T Number](vals []T) T {
	var sum T
	for _, v := range vals {
		sum += v
	}
	return sum
}

// Max 获取最大值
func Max[T Number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Min 获取最小值
func Min[T Number](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Clamp 限制值在范围内
func Clamp[T Number](val, min, max T) T {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// Contains 检查切片是否包含元素
func Contains[T comparable](vals []T, target T) bool {
	for _, v := range vals {
		if v == target {
			return true
		}
	}
	return false
}

// IndexOf 查找元素索引
func IndexOf[T comparable](vals []T, target T) int {
	for i, v := range vals {
		if v == target {
			return i
		}
	}
	return -1
}

// RemoveDuplicates 去重
func RemoveDuplicates[T comparable](vals []T) []T {
	seen := make(map[T]bool)
	result := make([]T, 0)
	for _, v := range vals {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// Reverse 反转切片
func Reverse[T any](vals []T) {
	for i, j := 0, len(vals)-1; i < j; i, j = i+1, j-1 {
		vals[i], vals[j] = vals[j], vals[i]
	}
}

// ========== 2. 泛型结构体 ==========

// ========== 2.1 泛型容器 ==========

// Stack 泛型栈
type Stack[T any] struct {
	items []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{items: make([]T, 0)}
}

func (s *Stack[T]) Push(val T) {
	s.items = append(s.items, val)
}

func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	val := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return val, true
}

func (s *Stack[T]) Peek() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

func (s *Stack[T]) Len() int {
	return len(s.items)
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}

// Queue 泛型队列
type Queue[T any] struct {
	items []T
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{items: make([]T, 0)}
}

func (q *Queue[T]) Enqueue(val T) {
	q.items = append(q.items, val)
}

func (q *Queue[T]) Dequeue() (T, bool) {
	if len(q.items) == 0 {
		var zero T
		return zero, false
	}
	val := q.items[0]
	q.items = q.items[1:]
	return val, true
}

func (q *Queue[T]) Len() int {
	return len(q.items)
}

// ========== 2.2 泛型Map和Set ==========

// Map 泛型Map（包装标准map）
type Map[K comparable, V any] struct {
	data map[K]V
}

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{data: make(map[K]V)}
}

func (m *Map[K, V]) Set(key K, value V) {
	m.data[key] = value
}

func (m *Map[K, V]) Get(key K) (V, bool) {
	val, ok := m.data[key]
	return val, ok
}

func (m *Map[K, V]) Delete(key K) {
	delete(m.data, key)
}

func (m *Map[K, V]) Has(key K) bool {
	_, ok := m.data[key]
	return ok
}

func (m *Map[K, V]) Keys() []K {
	keys := make([]K, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

func (m *Map[K, V]) Values() []V {
	values := make([]V, 0, len(m.data))
	for _, v := range m.data {
		values = append(values, v)
	}
	return values
}

// Set 泛型Set
type Set[T comparable] struct {
	data map[T]struct{}
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{data: make(map[T]struct{})}
}

func (s *Set[T]) Add(val T) {
	s.data[val] = struct{}{}
}

func (s *Set[T]) Remove(val T) {
	delete(s.data, val)
}

func (s *Set[T]) Contains(val T) bool {
	_, ok := s.data[val]
	return ok
}

func (s *Set[T]) Len() int {
	return len(s.data)
}

func (s *Set[T]) Union(other *Set[T]) *Set[T] {
	result := NewSet[T]()
	for v := range s.data {
		result.Add(v)
	}
	for v := range other.data {
		result.Add(v)
	}
	return result
}

func (s *Set[T]) Intersection(other *Set[T]) *Set[T] {
	result := NewSet[T]()
	if s.Len() > other.Len() {
		s, other = other, s
	}
	for v := range s.data {
		if other.Contains(v) {
			result.Add(v)
		}
	}
	return result
}

// ========== 3. 泛型方法 ==========

// ========== 3.1 带接收者的泛型方法 ==========

// Container 容器类型
type Container[T any] struct {
	items []T
}

func NewContainer[T any]() *Container[T] {
	return &Container[T]{items: make([]T, 0)}
}

// Filter 过滤元素（泛型方法）
func (c *Container[T]) Filter(pred func(T) bool) *Container[T] {
	result := NewContainer[T]()
	for _, item := range c.items {
		if pred(item) {
			result.items = append(result.items, item)
		}
	}
	return result
}

// Map 转换元素（泛型方法）
func (c *Container[T]) Map(func(func(T) T)) *Container[T] {
	result := NewContainer[T]()
	for _, item := range c.items {
		result.items = append(result.items, func(item))
	}
	return result
}

// Transform 转换类型（泛型方法）
func (c *Container[T]) Transform[U any](f func(T) U) *Container[U] {
	result := NewContainer[U]()
	for _, item := range c.items {
		result.items = append(result.items, f(item))
	}
	return result
}

// ========== 4. 泛型接口 ==========

// ========== 4.1 泛型接口定义 ==========

// Adder 可相加接口
type Adder[T any] interface {
	Add(T)
}

// AdderImpl 加法器实现
type AdderImpl[T Number] struct {
	Value T
}

func (a *AdderImpl[T]) Add(val T) {
	a.Value += val
}

func (a *AdderImpl[T]) Get() T {
	return a.Value
}

// ========== 4.2 泛型接口实现 ==========

// Iterable 可迭代接口
type Iterable[T any] interface {
	Iter() []T
}

// StringIterable 可迭代字符串
type StringIterable struct {
	values []string
}

func NewStringIterable(values ...string) *StringIterable {
	return &StringIterable{values: values}
}

func (s *StringIterable) Iter() []string {
	return s.values
}

// ========== 5. 泛型算法实现 ==========

// ========== 5.1 排序算法 ==========

// Sortable 可排序接口
type Sortable[T any] interface {
	Less(i, j int) bool
	Swap(i, j int)
	Len() int
}

// IntSliceSortable int切片排序
type IntSliceSortable []int

func (s IntSliceSortable) Len() int           { return len(s) }
func (s IntSliceSortable) Less(i, j int) bool { return s[i] < s[j] }
func (s IntSliceSortable) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// GenericSort 泛型排序
func GenericSort[T Sortable](data T) {
	sort.Sort(data)
}

// QuickSort 快速排序（泛型）
func QuickSort[T any](data []T, less func(i, j int) bool) {
	if len(data) <= 1 {
		return
	}
	
	pivot := len(data) / 2
	data[0], data[pivot] = data[pivot], data[0]
	
	i := 1
	for j := 1; j < len(data); j++ {
		if less(j, 0) {
			data[i], data[j] = data[j], data[i]
			i++
		}
	}
	data[0], data[i-1] = data[i-1], data[0]
	
	QuickSort(data[:i-1], less)
	QuickSort(data[i:], less)
}

// ========== 5.2 二分查找 ==========

// BinarySearch 二分查找
func BinarySearch[T comparable](data []T, target T, less func(i, j int) bool) int {
	left, right := 0, len(data)-1
	for left <= right {
		mid := left + (right-left)/2
		if data[mid] == target {
			return mid
		}
		if less(mid, 0) { // data[mid] < target
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return -1
}

// BinarySearchInt 二分查找（int版本）
func BinarySearchInt(data []int, target int) int {
	left, right := 0, len(data)-1
	for left <= right {
		mid := left + (right-left)/2
		if data[mid] == target {
			return mid
		}
		if data[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return -1
}

// ========== 6. 泛型进阶技巧 ==========

// ========== 6.1 类型列表 ==========

// TypeList 类型列表（编译期类型操作）
type TypeList[T any] []T

// ========== 6.2 泛型与接口组合 ==========

// Processor 处理接口
type Processor[T any] interface {
	Process(T) T
}

// Pipeline 管道
type Pipeline[T any] struct {
	processors []Processor[T]
}

func NewPipeline[T any]() *Pipeline[T] {
	return &Pipeline[T]{processors: make([]Processor[T], 0)}
}

func (p *Pipeline[T]) Add(proc Processor[T]) *Pipeline[T] {
	p.processors = append(p.processors, proc)
	return p
}

func (p *Pipeline[T]) Execute(val T) T {
	result := val
	for _, proc := range p.processors {
		result = proc.Process(result)
	}
	return result
}

// ========== 7. 泛型性能优化 ==========

/*
泛型性能特点：

✅ 优势：
- 编译时类型展开，无运行时类型开销
- 比 interface{} 更高效
- 内存布局更紧凑

⚠️ 注意事项：
- 泛型函数每次调用会生成一份代码
- 过多泛型实例化可能导致二进制膨胀
- 某些边界情况性能可能不如预期
*/

// ========== 7.1 性能对比示例 ==========

// GenericAdd 泛型版本（编译时展开）
func GenericAdd[T Number](a, b T) T {
	return a + b
}

// InterfaceAdd interface{}版本（运行时类型分发）
func InterfaceAdd(a, b interface{}) interface{} {
	// 需要类型断言，性能较低
	switch a.(type) {
	case int:
		return a.(int) + b.(int)
	case float64:
		return a.(float64) + b.(float64)
	}
	return nil
}

// ========== 8. 泛型最佳实践 ==========

/*
泛型使用建议：

✅ 推荐：
1. 泛型用于数据结构：栈、队列、map、set
2. 泛型用于通用算法：排序、搜索
3. 泛型用于减少重复代码

❌ 避免：
1. 过度泛化
2. 泛型嵌套过深
3. 与接口混用导致类型复杂

🔧 约束选择：
- 使用内置约束：comparable
- 常用约束组合：Number, String
- 必要时自定义约束
*/

// ========== 9. 面试要点 ==========

/*
泛型面试题：

Q: Go 泛型和 Java 泛型的区别？
A: Go 使用类型参数，编译时单态化；Java 使用类型擦除

Q: 泛型的性能优势？
A: 无运行时类型分发，接近具体类型性能

Q: 泛型约束的作用？
A: 限制类型参数的范围，提供编译期检查

Q: comparable 约束的使用？
A: 用于需要比较操作（==, <, >）的场景
*/

// ========== 10. 完整示例 ==========

// CompleteExample 完整示例
func CompleteExample() {
	// 1. 泛型函数
	nums := []int{1, 2, 3, 4, 5}
	fmt.Println("Sum:", Sum(nums))
	fmt.Println("Max:", Max(3, 7))
	fmt.Println("Contains:", Contains(nums, 3))
	
	// 2. 泛型容器
	stack := NewStack[int]()
	stack.Push(1)
	stack.Push(2)
	val, _ := stack.Pop()
	fmt.Println("Stack pop:", val)
	
	// 3. 泛型Map/Set
	m := NewMap[string, int]()
	m.Set("a", 1)
	m.Set("b", 2)
	if v, ok := m.Get("a"); ok {
		fmt.Println("Map get:", v)
	}
	
	set := NewSet[int]()
	set.Add(1)
	set.Add(2)
	fmt.Println("Set contains:", set.Contains(1))
	
	// 4. 泛型方法
	container := NewContainer[int]()
	container.items = []int{1, 2, 3, 4, 5}
	filtered := container.Filter(func(i int) bool {
		return i%2 == 0
	})
	fmt.Println("Filtered:", filtered.items)
	
	// 5. 泛型接口
	adder := &AdderImpl[int]{}
	adder.Add(10)
	fmt.Println("Adder value:", adder.Get())
}
