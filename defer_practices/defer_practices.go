package defer_practices

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// ========== Defer 基础行为 ==========

// 1. Defer 执行顺序：LIFO（后进先出）
func DeferExecutionOrder() {
	fmt.Println("Start")
	defer fmt.Println("Defer 1")
	defer fmt.Println("Defer 2")
	defer fmt.Println("Defer 3")
	fmt.Println("End")
	// 输出顺序：Start -> End -> Defer 3 -> Defer 2 -> Defer 1
}

// 2. Defer 参数求值时机：立即求值
func DeferArgumentEvaluation() {
	x := 1
	defer fmt.Println("Deferred x:", x) // x=1 立即求值
	
	x = 2
	fmt.Println("Current x:", x)
	// 输出：Current x: 2
	//      Deferred x: 1
}

// 3. Defer 与闭包：延迟求值
func DeferWithClosure() {
	x := 1
	defer func() {
		fmt.Println("Deferred x (closure):", x) // x=2 延迟求值
	}()
	
	x = 2
	fmt.Println("Current x:", x)
	// 输出：Current x: 2
	//      Deferred x (closure): 2
}

// ========== Defer 常见用途 ==========

// 1. 资源释放
func ResourceCleanup(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close() // 确保文件关闭
	
	// 处理文件...
	data := make([]byte, 1024)
	_, err = file.Read(data)
	return err
}

// 2. 互斥锁释放
type SafeCounter struct {
	mu    sync.Mutex
	count int
}

func (c *SafeCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock() // 确保锁释放
	
	c.count++
	// 即使这里 panic，锁也会被释放
}

// 3. 错误处理和恢复
func SafeFunction() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered: %v", r)
		}
	}()
	
	// 可能 panic 的代码
	// panic("something went wrong")
	
	return nil
}

// 4. 日志记录
func TracedFunction(name string) {
	start := time.Now()
	defer func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}()
	
	// 函数逻辑...
	time.Sleep(100 * time.Millisecond)
}

// ========== Defer 陷阱 ==========

// 陷阱1：循环中的 defer
func DeferInLoopBad() {
	// Bad: 所有 defer 在函数结束时执行，可能耗尽资源
	for i := 0; i < 1000; i++ {
		file, _ := os.Open(fmt.Sprintf("file%d.txt", i))
		defer file.Close() // 1000 个文件都打开着！
		// 处理文件...
	}
	// 所有文件在这里才关闭
}

// Good: 使用函数包装
func DeferInLoopGood() {
	for i := 0; i < 1000; i++ {
		func() {
			file, _ := os.Open(fmt.Sprintf("file%d.txt", i))
			defer file.Close() // 每次迭代结束时关闭
			// 处理文件...
		}()
	}
}

// 陷阱2：Defer 和返回值
func DeferReturnValue1() int {
	x := 0
	defer func() {
		x++ // 不影响返回值
	}()
	return x // 返回 0
}

func DeferReturnValue2() (x int) {
	defer func() {
		x++ // 影响返回值！
	}()
	return 0 // 返回 1 (0 被修改为 1)
}

func DeferReturnValue3() *int {
	x := 0
	defer func() {
		x++ // 不影响返回的指针指向的值
	}()
	return &x // 返回的指针指向 x，但值已经是 1
}

// 陷阱3：Defer 与指针
func DeferWithPointer() {
	x := 1
	defer func(p *int) {
		fmt.Println("Deferred value:", *p) // 2
	}(&x)
	
	x = 2
}

// 陷阱4：Defer 捕获循环变量
func DeferLoopVariable() {
	for i := 0; i < 3; i++ {
		defer fmt.Println("Bad:", i) // 都是 2 (最后一个值)
	}
	
	for i := 0; i < 3; i++ {
		i := i // 创建新变量
		defer fmt.Println("Good:", i) // 0, 1, 2
	}
	
	for i := 0; i < 3; i++ {
		defer func(v int) {
			fmt.Println("Good2:", v) // 0, 1, 2
		}(i)
	}
}

// ========== Defer 性能开销 ==========

// 简单函数（无 defer）
func NoDefer() int {
	x := 1
	x++
	return x
}

// 使用 defer
func WithDefer() int {
	x := 1
	defer func() {
		// 空函数
	}()
	x++
	return x
}

// 多个 defer
func MultipleDefer() int {
	x := 1
	defer func() {}()
	defer func() {}()
	defer func() {}()
	x++
	return x
}

// ========== Defer 的开销来源 ==========

/*
Defer 的性能开销：

1. 函数调用开销
   - defer 需要保存函数指针和参数
   - 运行时需要维护 defer 链表

2. 栈操作
   - defer 信息压入栈
   - 函数返回时从栈弹出并执行

3. Go 1.13+ 优化
   - 简单 defer 可以优化为内联
   - Open-coded defer (栈上 defer)
   - 减少了约 30% 的开销

性能影响：
- 简单 defer：约 1-2 纳秒开销（现代 Go）
- 复杂 defer：约 35 纳秒开销（旧版本可能更高）
- 对大多数应用可以忽略
- 热路径（百万次/秒）需要考虑
*/

// ========== Defer 最佳实践 ==========

// 1. 资源管理模式
type Resource struct {
	name string
}

func (r *Resource) Open() error {
	fmt.Printf("Opening %s\n", r.name)
	return nil
}

func (r *Resource) Close() error {
	fmt.Printf("Closing %s\n", r.name)
	return nil
}

func UseResource() error {
	r := &Resource{name: "database"}
	if err := r.Open(); err != nil {
		return err
	}
	defer r.Close() // 在成功 Open 后立即 defer Close
	
	// 使用资源...
	return nil
}

// 2. 错误处理模式
func ReadFile(filename string) (content []byte, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		// 合并 Close 错误
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	
	content, err = io.ReadAll(f)
	return
}

// 3. 事务回滚模式
type Transaction struct {
	committed bool
}

func (tx *Transaction) Commit() error {
	tx.committed = true
	fmt.Println("Transaction committed")
	return nil
}

func (tx *Transaction) Rollback() error {
	if !tx.committed {
		fmt.Println("Transaction rolled back")
	}
	return nil
}

func ExecuteTransaction() error {
	tx := &Transaction{}
	defer tx.Rollback() // 如果未提交，自动回滚
	
	// 执行操作...
	// if 某个条件 {
	//     return error // 自动回滚
	// }
	
	return tx.Commit()
}

// 4. 性能测量模式
func MeasureTime(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}

func SomeOperation() {
	defer MeasureTime("SomeOperation")()
	
	// 执行操作...
	time.Sleep(100 * time.Millisecond)
}

// 5. 调试追踪模式
var depth int

func Trace(name string) func() {
	fmt.Printf("%s%s begin\n", indent(), name)
	depth++
	return func() {
		depth--
		fmt.Printf("%s%s end\n", indent(), name)
	}
}

func indent() string {
	return ">  "[0:depth*2]
}

func A() {
	defer Trace("A")()
	B()
}

func B() {
	defer Trace("B")()
	C()
}

func C() {
	defer Trace("C")()
	// do something
}

// ========== Defer vs 手动清理 ==========

// 手动清理（容易出错）
func ManualCleanup() error {
	f, err := os.Open("file.txt")
	if err != nil {
		return err
	}
	
	data := make([]byte, 1024)
	n, err := f.Read(data)
	if err != nil {
		f.Close() // 需要记得关闭
		return err
	}
	
	if n == 0 {
		f.Close() // 需要记得关闭
		return fmt.Errorf("empty file")
	}
	
	// 更多逻辑...
	
	f.Close() // 最后关闭
	return nil
}

// 使用 defer（推荐）
func DeferCleanup() error {
	f, err := os.Open("file.txt")
	if err != nil {
		return err
	}
	defer f.Close() // 只需一次，所有路径都会执行
	
	data := make([]byte, 1024)
	n, err := f.Read(data)
	if err != nil {
		return err
	}
	
	if n == 0 {
		return fmt.Errorf("empty file")
	}
	
	// 更多逻辑...
	
	return nil
}

// ========== Defer 与 Panic/Recover ==========

// 基本的 panic 恢复
func SafeExecute() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()
	
	panic("something went wrong")
}

// 完整的错误处理
func CompleteErrorHandling() (err error) {
	defer func() {
		if r := recover(); r != nil {
			// 转换 panic 为 error
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	
	// 可能 panic 的代码
	riskyOperation()
	
	return nil
}

func riskyOperation() {
	// 某些操作
}

// Recover 只能在 defer 中直接调用
func RecoverRules() {
	// Bad: recover 不在 defer 中
	recover() // 无效
	
	// Bad: recover 不是 defer 函数的直接调用
	defer func() {
		helper() // recover 在 helper 中无效
	}()
	
	// Good: recover 是 defer 函数的直接调用
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered:", r)
		}
	}()
}

func helper() {
	recover() // 无效
}

// ========== 高级 Defer 模式 ==========

// 1. 条件性清理
func ConditionalCleanup(shouldClean bool) {
	resource := acquireResource()
	
	if shouldClean {
		defer releaseResource(resource)
	}
	
	// 使用资源...
}

func acquireResource() string {
	return "resource"
}

func releaseResource(r string) {
	fmt.Println("Releasing:", r)
}

// 2. 多资源清理
func MultiResourceCleanup() error {
	r1, err := openResource1()
	if err != nil {
		return err
	}
	defer r1.Close()
	
	r2, err := openResource2()
	if err != nil {
		return err // r1 会自动关闭
	}
	defer r2.Close()
	
	r3, err := openResource3()
	if err != nil {
		return err // r1 和 r2 会自动关闭
	}
	defer r3.Close()
	
	// 使用资源...
	return nil
}

type resource struct{}

func (r *resource) Close() {}

func openResource1() (*resource, error) { return &resource{}, nil }
func openResource2() (*resource, error) { return &resource{}, nil }
func openResource3() (*resource, error) { return &resource{}, nil }

// 3. 嵌套 defer 与作用域
func NestedDefer() {
	defer fmt.Println("Outer defer 1")
	
	func() {
		defer fmt.Println("Inner defer 1")
		defer fmt.Println("Inner defer 2")
		fmt.Println("Inner function")
	}() // Inner defer 2 -> Inner defer 1
	
	defer fmt.Println("Outer defer 2")
	fmt.Println("Outer function")
	// 输出顺序：
	// Inner function
	// Inner defer 2
	// Inner defer 1
	// Outer function
	// Outer defer 2
	// Outer defer 1
}

// ========== Defer 最佳实践总结 ==========

/*
Defer 使用指南：

✅ 何时使用 defer：

1. 资源清理
   - 文件关闭：defer file.Close()
   - 锁释放：defer mu.Unlock()
   - 连接关闭：defer conn.Close()

2. 错误恢复
   - Panic 恢复
   - 错误包装

3. 追踪和日志
   - 函数执行时间
   - 调用追踪

4. 状态恢复
   - 事务回滚
   - 临时状态恢复

❌ 避免的陷阱：

1. 循环中使用 defer
   - 会累积到函数结束
   - 可能耗尽资源
   - 解决：使用函数包装

2. 忽略 defer 的参数求值时机
   - 参数立即求值
   - 闭包延迟求值

3. Defer 与返回值
   - 命名返回值可以被修改
   - 理解返回值语义

4. 循环变量捕获
   - 使用局部变量
   - 或作为参数传递

⚡ 性能考虑：

1. 现代 Go（1.13+）优化
   - 简单 defer 开销小（1-2ns）
   - Open-coded defer
   
2. 热路径优化
   - 百万次/秒的调用考虑避免 defer
   - 其他场景忽略性能影响

3. 可读性 > 微优化
   - Defer 提高代码可维护性
   - 减少资源泄漏风险

🎯 推荐模式：

1. 资源获取后立即 defer 释放
2. 使用命名返回值处理清理错误
3. Defer 函数中使用 recover
4. 性能测量使用 defer
5. 事务自动回滚

❗ 重要规则：

1. Defer 执行顺序：LIFO（后进先出）
2. Defer 在函数返回前执行（包括 panic）
3. Recover 只在 defer 中有效
4. Defer 可以修改命名返回值
5. 循环中的 defer 要小心
*/

