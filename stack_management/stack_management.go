package stack_management

import (
	"fmt"
	"runtime"
)

// ========== Go 栈管理深入理解 ==========

/*
本项目讲解 Go 栈的实现原理，包括：

一、栈基础
1. 栈帧结构
2. 函数调用过程
3. 栈操作

二、连续栈
1. 栈扩缩容
2. 栈分裂
3. 栈切换

三、逃逸分析
1. 什么是逃逸
2. 逃逸规则
3. 逃逸分析示例

四、递归优化
1. 尾递归优化
2. 栈深度控制
*/

// ========== 1. 栈基础 ==========

/*
栈帧结构：

┌─────────────────────┐
│   返回地址           │
├─────────────────────┤
│   保存的寄存器       │
├─────────────────────┤
│   局部变量           │
├─────────────────────┤
│   函数参数           │
├─────────────────────┤
│   上一帧 BP         │
└─────────────────────┘

函数调用过程：
1. 推送参数
2. 保存返回地址
3. 创建新栈帧
4. 执行函数体
5. 恢复栈帧
*/

// StackFrame 模拟栈帧
type StackFrame struct {
	ReturnAddr int      // 返回地址
	BP         int      // 基址指针
	Locals     []int    // 局部变量
	Args       []int    // 参数
}

// NewStackFrame 创建栈帧
func NewStackFrame(args []int) *StackFrame {
	return &StackFrame{
		Locals: make([]int, 0),
		Args:   args,
	}
}

// ========== 2. 连续栈 ==========

/*
连续栈（Continuous Stack）：

Go 1.3 之前：分段栈（Segemented Stack）
- 问题：栈增长时需要分配新段
- 性能：段切换开销

Go 1.3+：连续栈
- 初始大小：2KB
- 最大大小：1GB
- 扩缩容：整块搬迁

栈扩缩容：
┌────────────────────┐
│    旧栈 (2KB)      │
│  ┌──────────────┐  │
│  │    函数A     │  │
│  │  ┌────────┐  │  │
│  │  │  函数B │  │  │
│  │  └────────┘  │  │
│  └──────────────┘  │
└────────────────────┘
        ↓ 扩容
┌────────────────────┐
│    新栈 (4KB)      │
│  ┌──────────────┐  │
│  │    函数A     │  │
│  │  ┌────────┐  │  │
│  │  │  函数B │  │  │
│  │  └────────┘  │  │
│  └──────────────┘  │
└────────────────────┘
*/

// G 模拟 goroutine
type G struct {
	Stack       *Stack  // 栈
	StackGuard0 int     // 栈边界
	StackGuard1 int
}

// Stack 模拟栈
type Stack struct {
	Lo    int // 栈低（低地址）
	Hi    int // 栈顶（高地址）
	SP    int // 栈指针
}

// NewStack 创建新栈
func NewStack(size int) *Stack {
	return &Stack{
		Lo: 0,
		Hi: size,
		SP: size, // 从高地址开始
	}
}

// Size 栈大小
func (s *Stack) Size() int {
	return s.SP
}

// LeftoverSpace 剩余空间
func (s *Stack) LeftoverSpace() int {
	return s.SP - s.Lo
}

// ========== 3. 栈扩缩容 ==========

/*
栈扩容流程：
┌─────────────────────────────────────┐
│ 1. 分配新栈 (2倍大小)                │
├─────────────────────────────────────┤
│ 2. 复制旧栈数据到新栈                │
├─────────────────────────────────────┤
│ 3. 更新指针                         │
├─────────────────────────────────────┤
│ 4. 释放旧栈                         │
└─────────────────────────────────────┘
*/

// StackGrow 栈扩容
func (g *G) StackGrow(need int) {
	// 计算2倍大小
	newSize := g.Stack.Size() * 2
	if newSize < need {
		newSize = need
	}
	
	// 创建新栈
	newStack := NewStack(newSize)
	
	fmt.Printf("Stack grew from %d to %d bytes\n", 
		g.Stack.Size(), newSize)
	
	g.Stack = newStack
}

// EnsureSpace 确保栈空间
func (g *G) EnsureSpace(need int) {
	if g.Stack.LeftoverSpace() < need {
		g.StackGrow(need)
	}
}

// StackShrink 栈缩容
func (g *G) StackShrink() {
	// 当栈使用率低于 1/4 时缩容
	used := g.Stack.Hi - g.Stack.SP
	if used < g.Stack.Size()/4 {
		newSize := g.Stack.Size() / 2
		if newSize < 2048 {
			return // 最小 2KB
		}
		newStack := NewStack(newSize)
		fmt.Printf("Stack shrunk from %d to %d bytes\n", 
			g.Stack.Size(), newSize)
		g.Stack = newStack
	}
}

// ========== 4. 栈溢出检测 ==========

/*
栈溢出检测：

Go 使用"熔断"机制：
1. 栈边界检查
2. 小于 StackGuard0: 扩容
3. 大于 StackGuard1: 崩溃

栈布局：
┌────────────────────┐ ← StackGuard1 (高地址)
│    可用空间        │
├────────────────────┤
│                    │
│    函数执行区      │
│                    │
├────────────────────┤ ← StackGuard0 (低地址 + 溢出区)
│    溢出区          │
└────────────────────┘
*/

// CheckStackOverflow 检查栈溢出
func CheckStackOverflow(sp, guardLo, guardHi int) bool {
	return sp < guardLo || sp > guardHi
}

// ========== 5. 逃逸分析 ==========

/*
逃逸分析（Escape Analysis）：

作用：决定变量分配在栈还是堆

逃逸规则：
1. 返回指针 → 逃逸到堆
2. 发送指针到 channel → 逃逸到堆
3. 切片 append 超过容量 → 逃逸到堆
4. interface 值 → 可能逃逸
5. 局部变量不逃逸 → 栈分配

查看逃逸：
go build -gcflags=-m main.go
*/

// NoEscape 不逃逸（栈分配）
func NoEscape() int {
	x := 100 // 不逃逸，栈分配
	return x
}

// Escape 逃逸（堆分配）
func Escape() *int {
	x := 100 // 逃逸，堆分配
	return &x
}

// ChannelEscape 通过 channel 逃逸
func ChannelEscape(ch chan *int) {
	x := 100
	ch <- &x // 逃逸
}

// SliceEscape 切片逃逸
func SliceEscape() []int {
	s := make([]int, 0, 10) // 可能逃逸
	s = append(s, 100)
	return s
}

// ========== 6. 栈管理示例 ==========

// StackDemo 栈操作示例
func StackDemo() {
	fmt.Println("=== 栈管理演示 ===")
	
	// 创建 goroutine
	g := &G{
		Stack: NewStack(2048), // 初始 2KB
	}
	
	fmt.Printf("初始栈大小: %d bytes\n", g.Stack.Size())
	fmt.Printf("剩余空间: %d bytes\n", g.Stack.LeftoverSpace())
	
	// 确保栈空间
	g.EnsureSpace(1024)
	fmt.Printf("扩容后栈大小: %d bytes\n", g.Stack.Size())
	
	// 逃逸分析
	fmt.Println("\n逃逸分析:")
	noEscape := NoEscape()
	fmt.Printf("  NoEscape: %d (栈分配)\n", noEscape)
	
	escape := Escape()
	fmt.Printf("  Escape: %d (堆分配)\n", *escape)
}

// ========== 7. 递归优化 ==========

/*
递归优化策略：

1. 尾递归优化
   - 编译器优化
   - 减少栈帧

2. 迭代改递归
   - 使用循环代替递归

3. 栈深度控制
   - 限制递归深度
   - 使用工作队列
*/

// RecursiveSum 递归求和（可能栈溢出）
func RecursiveSum(n int) int {
	if n <= 0 {
		return 0
	}
	return n + RecursiveSum(n-1)
}

// IterativeSum 迭代求和（栈安全）
func IterativeSum(n int) int {
	sum := 0
	for i := 1; i <= n; i++ {
		sum += i
	}
	return sum
}

// TailRecursiveSum 尾递归求和
func TailRecursiveSum(n, acc int) int {
	if n <= 0 {
		return acc
	}
	return TailRecursiveSum(n-1, n+acc)
}

// RecursiveDemo 递归示例
func RecursiveDemo() {
	fmt.Println("\n=== 递归优化演示 ===")
	
	n := 1000
	
	// 递归
	sum1 := RecursiveSum(n)
	fmt.Printf("递归求和(%d): %d\n", n, sum1)
	
	// 迭代
	sum2 := IterativeSum(n)
	fmt.Printf("迭代求和(%d): %d\n", n, sum2)
	
	// 尾递归
	sum3 := TailRecursiveSum(n, 0)
	fmt.Printf("尾递归求和(%d): %d\n", n, sum3)
}

// ========== 8. 栈帧分析 ==========

// StackFrameDemo 栈帧分析演示
func StackFrameDemo() {
	fmt.Println("\n=== 栈帧分析演示 ===")
	
	// 模拟函数调用栈
	frame1 := NewStackFrame([]int{1, 2})
	frame1.Locals = []int{100, 200}
	fmt.Printf("栈帧1: Args=%v, Locals=%v\n", frame1.Args, frame1.Locals)
	
	frame2 := NewStackFrame([]int{3, 4, 5})
	frame2.Locals = []int{300}
	fmt.Printf("栈帧2: Args=%v, Locals=%v\n", frame2.Args, frame2.Locals)
}

// ========== 9. 运行时栈信息 ==========

// RuntimeStackDemo 运行时栈信息演示
func RuntimeStackDemo() {
	fmt.Println("\n=== 运行时栈信息 ===")
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("栈空间使用: %d KB\n", m.StackInuse/1024)
	fmt.Printf("系统栈空间: %d KB\n", m.StackSys/1024)
	
	// 当前 goroutine 数量
	fmt.Printf("当前 goroutine 数量: %d\n", runtime.NumGoroutine())
}

// ========== 10. 面试要点 ==========

/*
栈管理面试题：

Q: Go 栈是固定大小吗？
A: 不是，动态扩缩容

Q: 栈扩缩容发生在什么时候？
A: 函数 prologue 时检查

Q: 逃逸分析的作用？
A: 决定变量分配位置，优化内存

Q: 连续栈的优点？
A: 避免分段栈的性能开销

Q: 如何查看变量是否逃逸？
A: go build -gcflags=-m
*/

// CompleteExample 完整示例
func CompleteExample() {
	StackDemo()
	RecursiveDemo()
	StackFrameDemo()
	RuntimeStackDemo()
}
