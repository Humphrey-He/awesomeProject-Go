package goroutine_practices

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ========== 1. Goroutine 启动与停止 ==========

// 1.1 基本启动
func BasicStart() {
	// 简单启动
	go func() {
		fmt.Println("Hello from goroutine")
	}()
	
	// 等待goroutine完成（不推荐用sleep）
	time.Sleep(100 * time.Millisecond)
}

// 1.2 使用 WaitGroup 等待完成
func StartWithWaitGroup() {
	var wg sync.WaitGroup
	
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("Worker %d completed\n", id)
		}(i) // 注意：传递参数避免闭包陷阱
	}
	
	wg.Wait() // 等待所有goroutine完成
	fmt.Println("All workers completed")
}

// 1.3 使用 Context 取消
func StartWithContext() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	go worker(ctx, 1)
	go worker(ctx, 2)
	go worker(ctx, 3)
	
	// 运行一段时间后取消
	time.Sleep(2 * time.Second)
	cancel() // 通知所有goroutine停止
	
	// 给goroutine时间清理
	time.Sleep(100 * time.Millisecond)
}

func worker(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Worker %d stopped\n", id)
			return
		default:
			fmt.Printf("Worker %d working...\n", id)
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// 1.4 优雅停止模式
type Server struct {
	stopCh chan struct{}
	wg     sync.WaitGroup
}

func NewServer() *Server {
	return &Server{
		stopCh: make(chan struct{}),
	}
}

func (s *Server) Start() {
	s.wg.Add(1)
	go s.run()
}

func (s *Server) run() {
	defer s.wg.Done()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.stopCh:
			fmt.Println("Server stopping...")
			return
		case <-ticker.C:
			fmt.Println("Server tick")
		}
	}
}

func (s *Server) Stop() {
	close(s.stopCh) // 通知停止
	s.wg.Wait()     // 等待goroutine完成
	fmt.Println("Server stopped")
}

// ========== 2. Goroutine 同步与通信 ==========

// 2.1 使用 Channel 通信
func ChannelCommunication() {
	ch := make(chan int, 10)
	
	// 生产者
	go func() {
		for i := 0; i < 10; i++ {
			ch <- i
		}
		close(ch) // 关闭channel
	}()
	
	// 消费者
	for v := range ch {
		fmt.Println("Received:", v)
	}
}

// 2.2 使用 Mutex 同步
type SafeMap struct {
	mu   sync.RWMutex
	data map[string]int
}

func NewSafeMap() *SafeMap {
	return &SafeMap{
		data: make(map[string]int),
	}
}

func (m *SafeMap) Set(key string, value int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

func (m *SafeMap) Get(key string) (int, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.data[key]
	return v, ok
}

// 2.3 使用 atomic 原子操作
type AtomicCounter struct {
	count int64
}

func (c *AtomicCounter) Increment() {
	atomic.AddInt64(&c.count, 1)
}

func (c *AtomicCounter) Get() int64 {
	return atomic.LoadInt64(&c.count)
}

// 2.4 使用 sync.Once 单次执行
var (
	instance *Singleton
	once     sync.Once
)

type Singleton struct {
	data string
}

func GetInstance() *Singleton {
	once.Do(func() {
		instance = &Singleton{data: "initialized"}
		fmt.Println("Singleton created")
	})
	return instance
}

// ========== 3. 限流与并发控制 ==========

// 3.1 使用带缓冲的 channel 限制并发数
func LimitConcurrency(tasks []func(), maxConcurrent int) {
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	
	for _, task := range tasks {
		wg.Add(1)
		go func(t func()) {
			defer wg.Done()
			
			semaphore <- struct{}{} // 获取信号量
			defer func() { <-semaphore }() // 释放信号量
			
			t() // 执行任务
		}(task)
	}
	
	wg.Wait()
}

// 3.2 Worker Pool 模式
type WorkerPool struct {
	workerCount int
	taskQueue   chan func()
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

func NewWorkerPool(workerCount, queueSize int) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		taskQueue:   make(chan func(), queueSize),
		stopCh:      make(chan struct{}),
	}
}

func (p *WorkerPool) Start() {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
}

func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()
	
	for {
		select {
		case <-p.stopCh:
			return
		case task, ok := <-p.taskQueue:
			if !ok {
				return
			}
			task() // 执行任务
		}
	}
}

func (p *WorkerPool) Submit(task func()) bool {
	select {
	case p.taskQueue <- task:
		return true
	case <-p.stopCh:
		return false
	default:
		return false // 队列满
	}
}

func (p *WorkerPool) Stop() {
	close(p.stopCh)
	p.wg.Wait()
}

// 3.3 令牌桶限流（复用已实现的）
type RateLimiter struct {
	ticker   *time.Ticker
	bucket   chan struct{}
	stopOnce sync.Once
}

func NewRateLimiter(rate int) *RateLimiter {
	rl := &RateLimiter{
		ticker: time.NewTicker(time.Second / time.Duration(rate)),
		bucket: make(chan struct{}, rate),
	}
	
	// 填充bucket
	for i := 0; i < rate; i++ {
		rl.bucket <- struct{}{}
	}
	
	// 定期补充令牌
	go func() {
		for range rl.ticker.C {
			select {
			case rl.bucket <- struct{}{}:
			default:
			}
		}
	}()
	
	return rl
}

func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.bucket:
		return true
	default:
		return false
	}
}

func (rl *RateLimiter) Stop() {
	rl.stopOnce.Do(func() {
		rl.ticker.Stop()
	})
}

// ========== 4. 错误与 Panic 处理 ==========

// 4.1 Goroutine 中的 panic 恢复
func SafeGoroutine(f func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Recovered from panic: %v\n", r)
				// 记录堆栈
				buf := make([]byte, 1024)
				n := runtime.Stack(buf, false)
				fmt.Printf("Stack trace:\n%s\n", buf[:n])
			}
		}()
		
		f()
	}()
}

// 4.2 错误收集模式
type ErrorCollector struct {
	mu     sync.Mutex
	errors []error
}

func (ec *ErrorCollector) Add(err error) {
	if err == nil {
		return
	}
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.errors = append(ec.errors, err)
}

func (ec *ErrorCollector) Errors() []error {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	return append([]error(nil), ec.errors...)
}

func ProcessWithErrorCollection(tasks []func() error) []error {
	var wg sync.WaitGroup
	collector := &ErrorCollector{}
	
	for _, task := range tasks {
		wg.Add(1)
		go func(t func() error) {
			defer wg.Done()
			if err := t(); err != nil {
				collector.Add(err)
			}
		}(task)
	}
	
	wg.Wait()
	return collector.Errors()
}

// 4.3 使用 errgroup
type Group struct {
	cancel  func()
	wg      sync.WaitGroup
	errOnce sync.Once
	err     error
}

func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{cancel: cancel}, ctx
}

func (g *Group) Go(f func() error) {
	g.wg.Add(1)
	
	go func() {
		defer g.wg.Done()
		
		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	}()
}

func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}

// ========== 5. 资源释放与清理 ==========

// 5.1 确保资源释放
func ResourceCleanup() {
	// 使用 defer 确保清理
	resource := acquireResource()
	defer resource.Release()
	
	// 使用资源...
}

type Resource struct {
	name string
}

func acquireResource() *Resource {
	return &Resource{name: "database connection"}
}

func (r *Resource) Release() {
	fmt.Printf("Releasing %s\n", r.name)
}

// 5.2 清理 goroutine 泄漏
func PreventGoroutineLeakBad() {
	ch := make(chan int)
	
	// Bad: goroutine 永远阻塞
	go func() {
		val := <-ch // 如果没有发送者，永远阻塞
		fmt.Println(val)
	}()
	
	// 函数返回，goroutine 泄漏
}

func PreventGoroutineLeakGood() {
	ch := make(chan int)
	done := make(chan struct{})
	
	go func() {
		select {
		case val := <-ch:
			fmt.Println(val)
		case <-done:
			fmt.Println("Goroutine stopped")
			return
		}
	}()
	
	// 清理
	close(done)
	time.Sleep(10 * time.Millisecond)
}

// 5.3 Context 管理资源
func ContextResourceManagement(ctx context.Context) error {
	// 创建子context
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel() // 确保资源释放
	
	// 启动goroutine
	done := make(chan error, 1)
	go func() {
		done <- doWork(ctx)
	}()
	
	// 等待完成或超时
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func doWork(ctx context.Context) error {
	// 模拟工作
	time.Sleep(1 * time.Second)
	return nil
}

// ========== 6. 超时与重试策略 ==========

// 6.1 简单超时
func WithTimeout(timeout time.Duration, f func() error) error {
	done := make(chan error, 1)
	
	go func() {
		done <- f()
	}()
	
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return errors.New("timeout")
	}
}

// 6.2 使用 Context 超时
func WithContextTimeout(ctx context.Context, timeout time.Duration, f func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	return f(ctx)
}

// 6.3 指数退避重试
func RetryWithBackoff(maxRetries int, f func() error) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		if err = f(); err == nil {
			return nil
		}
		
		if i < maxRetries-1 {
			// 指数退避：100ms, 200ms, 400ms, 800ms...
			backoff := time.Duration(100*(1<<uint(i))) * time.Millisecond
			fmt.Printf("Retry %d/%d after %v\n", i+1, maxRetries, backoff)
			time.Sleep(backoff)
		}
	}
	return fmt.Errorf("max retries exceeded: %w", err)
}

// 6.4 带超时的重试
func RetryWithTimeout(maxRetries int, timeout time.Duration, f func() error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	errCh := make(chan error, 1)
	
	go func() {
		errCh <- RetryWithBackoff(maxRetries, f)
	}()
	
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ========== 7. 结构化并发 ==========

// 7.1 父子关系管理
type Supervisor struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewSupervisor(ctx context.Context) *Supervisor {
	ctx, cancel := context.WithCancel(ctx)
	return &Supervisor{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Supervisor) Go(f func(context.Context) error) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := f(s.ctx); err != nil {
			fmt.Printf("Task failed: %v\n", err)
		}
	}()
}

func (s *Supervisor) Wait() error {
	s.wg.Wait()
	return nil
}

func (s *Supervisor) Cancel() {
	s.cancel()
}

// 7.2 Pipeline 模式
func Pipeline() {
	// Stage 1: 生成数字
	gen := func(nums ...int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for _, n := range nums {
				out <- n
			}
		}()
		return out
	}
	
	// Stage 2: 平方
	sq := func(in <-chan int) <-chan int {
		out := make(chan int)
		go func() {
			defer close(out)
			for n := range in {
				out <- n * n
			}
		}()
		return out
	}
	
	// Stage 3: 打印
	print := func(in <-chan int) {
		for n := range in {
			fmt.Println(n)
		}
	}
	
	// 连接 pipeline
	numbers := gen(1, 2, 3, 4, 5)
	squared := sq(numbers)
	print(squared)
}

// 7.3 Fan-out/Fan-in 模式
func FanOutFanIn() {
	// 输入
	in := gen(2, 3, 4, 5)
	
	// Fan-out: 启动多个worker
	c1 := square(in)
	c2 := square(in)
	c3 := square(in)
	
	// Fan-in: 合并结果
	for n := range merge(c1, c2, c3) {
		fmt.Println(n)
	}
}

func gen(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

func merge(cs ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	
	wg.Add(len(cs))
	for _, c := range cs {
		go func(ch <-chan int) {
			defer wg.Done()
			for n := range ch {
				out <- n
			}
		}(c)
	}
	
	go func() {
		wg.Wait()
		close(out)
	}()
	
	return out
}

// ========== 8. 常见模式与最佳实践 ==========

// 8.1 Generator 模式
func Generator(ctx context.Context, start, end int) <-chan int {
	ch := make(chan int)
	
	go func() {
		defer close(ch)
		for i := start; i < end; i++ {
			select {
			case ch <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	
	return ch
}

// 8.2 Or-done 模式
func OrDone(ctx context.Context, c <-chan int) <-chan int {
	out := make(chan int)
	
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-c:
				if !ok {
					return
				}
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	
	return out
}

// 8.3 Tee 模式（分流）
func Tee(ctx context.Context, in <-chan int) (<-chan int, <-chan int) {
	out1 := make(chan int)
	out2 := make(chan int)
	
	go func() {
		defer close(out1)
		defer close(out2)
		
		for val := range OrDone(ctx, in) {
			var out1, out2 = out1, out2
			for i := 0; i < 2; i++ {
				select {
				case out1 <- val:
					out1 = nil
				case out2 <- val:
					out2 = nil
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	
	return out1, out2
}

// 8.4 Bridge 模式（合并channel的channel）
func Bridge(ctx context.Context, chanStream <-chan <-chan int) <-chan int {
	out := make(chan int)
	
	go func() {
		defer close(out)
		
		for {
			var stream <-chan int
			select {
			case maybeStream, ok := <-chanStream:
				if !ok {
					return
				}
				stream = maybeStream
			case <-ctx.Done():
				return
			}
			
			for val := range OrDone(ctx, stream) {
				select {
				case out <- val:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	
	return out
}

// ========== Goroutine 最佳实践总结 ==========

/*
Goroutine 开发最佳实践：

✅ 1. 启动与停止
   - 使用 WaitGroup 等待完成
   - 使用 Context 取消操作
   - 提供优雅停止机制
   - 避免 goroutine 泄漏

✅ 2. 同步与通信
   - 优先使用 channel 通信
   - Mutex 保护共享状态
   - atomic 用于简单计数
   - sync.Once 单次初始化

✅ 3. 限流与并发控制
   - Worker Pool 模式
   - 带缓冲的 channel 作为信号量
   - 令牌桶限流
   - 控制并发数量

✅ 4. 错误处理
   - Goroutine 中捕获 panic
   - 收集并返回错误
   - 使用 errgroup 模式
   - 记录堆栈信息

✅ 5. 资源管理
   - 使用 defer 释放资源
   - Context 管理生命周期
   - 避免资源泄漏
   - 确保 goroutine 正常退出

✅ 6. 超时与重试
   - Context 超时控制
   - 指数退避重试
   - 最大重试次数
   - 超时与重试组合

✅ 7. 结构化并发
   - 父子关系管理
   - Pipeline 模式
   - Fan-out/Fan-in
   - 清晰的所有权

✅ 8. 常见模式
   - Generator
   - Or-done
   - Tee
   - Bridge
   - Pipeline

❌ 避免的陷阱：

1. Goroutine 泄漏
   - 忘记取消 context
   - Channel 永远阻塞
   - 没有退出机制

2. 竞态条件
   - 未保护的共享状态
   - 忘记加锁
   - 使用 -race 检测

3. 死锁
   - 循环等待
   - Channel 容量不足
   - 锁顺序不一致

4. 过多的 goroutine
   - 无限制启动
   - 没有限流
   - 消耗过多资源

5. Context 误用
   - 不传递 context
   - 忘记 cancel
   - Context 存储不当

⚡ 性能优化：

1. 控制 goroutine 数量
   - 不要启动过多 goroutine
   - 使用 Worker Pool
   - 根据 CPU 核心数调整

2. Channel 优化
   - 使用带缓冲的 channel
   - 避免频繁创建
   - 及时关闭 channel

3. 避免锁竞争
   - 减小临界区
   - 使用 RWMutex
   - 考虑无锁设计

4. 内存优化
   - 复用对象（sync.Pool）
   - 避免逃逸到堆
   - 控制 goroutine 栈大小

🎯 通用建议：

1. 明确 goroutine 的生命周期
2. 始终考虑如何停止 goroutine
3. 使用 Context 传递取消信号
4. 测试并发代码（-race, -count=100）
5. 监控 goroutine 数量（runtime.NumGoroutine()）
6. 简单优于复杂
7. 可测试性优先
8. 文档化并发行为
*/

