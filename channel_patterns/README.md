

# Go Channel 最佳实践和模式

## 概述

Channel 是 Go 语言的核心特性，用于 goroutine 之间的通信。本项目总结了 Channel 的最佳实践，涵盖创建、传递、复用和关闭的各种模式。

**核心理念**："不要通过共享内存来通信，而要通过通信来共享内存"

## 目录

1. [Channel 创建模式](#1-channel-创建模式)
2. [Channel 传递模式](#2-channel-传递模式)
3. [Worker Pool 复用模式](#3-worker-pool-复用模式)
4. [Channel 关闭模式](#4-channel-关闭模式)
5. [高级模式](#5-高级模式)
6. [最佳实践](#6-最佳实践)
7. [常见陷阱](#7-常见陷阱)

## 1. Channel 创建模式

### 无缓冲 Channel（Unbuffered Channel）

**特性**：同步通信，发送和接收必须同时就绪

```go
ch := make(chan int)  // 无缓冲

// 发送会阻塞，直到有接收者
go func() {
    ch <- 42  // 阻塞直到被接收
}()

val := <-ch  // 接收阻塞直到有发送者
```

**使用场景**：
- ✅ 需要同步的场景
- ✅ 握手确认
- ✅ 保证顺序执行

### 有缓冲 Channel（Buffered Channel）

**特性**：异步通信，缓冲区未满时发送不阻塞

```go
ch := make(chan int, 3)  // 缓冲大小为3

ch <- 1  // 不阻塞
ch <- 2  // 不阻塞
ch <- 3  // 不阻塞
ch <- 4  // 阻塞！缓冲区满

<-ch     // 接收一个，腾出空间
ch <- 4  // 现在可以发送
```

**使用场景**：
- ✅ 生产者-消费者模式
- ✅ 减少goroutine阻塞
- ✅ 批处理

**缓冲大小选择**：
```go
// 太小：频繁阻塞
ch := make(chan Task, 1)

// 太大：浪费内存
ch := make(chan Task, 1000000)

// 合适：根据吞吐量
producerRate := 1000  // 每秒1000个
consumerRate := 800   // 每秒800个
buffer := (producerRate - consumerRate) * 5  // 5秒缓冲
ch := make(chan Task, buffer)
```

### 只读/只写 Channel

**限制接口，增强类型安全**：

```go
// 生产者只能发送
func producer(out chan<- int) {
    out <- 42
    // val := <-out  // 编译错误！
}

// 消费者只能接收
func consumer(in <-chan int) {
    val := <-in
    // in <- 42  // 编译错误！
}

// 使用
ch := make(chan int, 10)
go producer(ch)  // 自动转换为只写
go consumer(ch)  // 自动转换为只读
```

## 2. Channel 传递模式

### Pipeline 模式

**流式数据处理**：

```go
// 第一阶段：生成数字
func generator(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            out <- n
        }
    }()
    return out
}

// 第二阶段：平方
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

// 组合pipeline
nums := generator(1, 2, 3, 4)
squares := square(nums)

for s := range squares {
    fmt.Println(s)  // 1, 4, 9, 16
}
```

**优势**：
- ✅ 模块化
- ✅ 可组合
- ✅ 易于测试

### Fan-Out 模式

**一个输入分发到多个worker**：

```go
func fanOut(input <-chan int, workers int) []<-chan int {
    outputs := make([]<-chan int, workers)
    
    for i := 0; i < workers; i++ {
        out := make(chan int)
        outputs[i] = out
        
        go func(ch chan int) {
            defer close(ch)
            for v := range input {
                ch <- v * 2  // 处理数据
            }
        }(out)
    }
    
    return outputs
}

// 使用
input := generator(1, 2, 3, 4, 5, 6)
outputs := fanOut(input, 3)  // 3个worker
```

### Fan-In 模式

**多个输入合并到一个输出**：

```go
func fanIn(inputs ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup
    
    // 为每个输入启动一个goroutine
    for _, input := range inputs {
        wg.Add(1)
        go func(ch <-chan int) {
            defer wg.Done()
            for v := range ch {
                out <- v
            }
        }(input)
    }
    
    // 等待所有输入完成后关闭输出
    go func() {
        wg.Wait()
        close(out)
    }()
    
    return out
}

// 使用
ch1 := generator(1, 2, 3)
ch2 := generator(4, 5, 6)
merged := fanIn(ch1, ch2)
```

## 3. Worker Pool 复用模式

### 基本 Worker Pool

**固定数量的worker处理任务**：

```go
type WorkerPool struct {
    workers    int
    jobs       chan Job
    results    chan Result
    wg         sync.WaitGroup
}

func NewWorkerPool(workers int) *WorkerPool {
    return &WorkerPool{
        workers: workers,
        jobs:    make(chan Job, 100),
        results: make(chan Result, 100),
    }
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        go wp.worker(i)
    }
}

func (wp *WorkerPool) worker(id int) {
    defer wp.wg.Done()
    
    for job := range wp.jobs {
        result := job.Process()  // 处理任务
        wp.results <- result
    }
}

func (wp *WorkerPool) Stop() {
    close(wp.jobs)
    wp.wg.Wait()
    close(wp.results)
}

// 使用
pool := NewWorkerPool(5)
pool.Start()

// 提交任务
for i := 0; i < 100; i++ {
    pool.jobs <- Job{ID: i}
}

// 收集结果
go func() {
    pool.Stop()
}()

for result := range pool.results {
    fmt.Println(result)
}
```

### 带 Context 的 Worker Pool

**支持取消和超时**：

```go
func (wp *WorkerPool) worker(ctx context.Context, id int) {
    defer wp.wg.Done()
    
    for {
        select {
        case <-ctx.Done():
            return  // 被取消
        case job, ok := <-wp.jobs:
            if !ok {
                return  // jobs channel 关闭
            }
            
            // 处理任务
            result := job.Process()
            
            select {
            case wp.results <- result:
            case <-ctx.Done():
                return
            }
        }
    }
}

// 使用
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

pool.StartWithContext(ctx)
```

### 优雅关闭

**确保所有任务完成再退出**：

```go
func (wp *WorkerPool) Shutdown(timeout time.Duration) error {
    done := make(chan struct{})
    
    go func() {
        close(wp.jobs)      // 不再接受新任务
        wp.wg.Wait()        // 等待所有worker完成
        close(wp.results)   // 关闭结果channel
        close(done)
    }()
    
    select {
    case <-done:
        return nil
    case <-time.After(timeout):
        return fmt.Errorf("shutdown timeout")
    }
}
```

## 4. Channel 关闭模式

### 基本规则

**重要原则**：
1. ✅ **发送者关闭channel**，不是接收者
2. ✅ 不要向已关闭的channel发送数据（会panic）
3. ✅ 可以从已关闭的channel接收数据（返回零值）
4. ✅ 关闭已关闭的channel会panic

### 单发送者单接收者

**最简单的情况**：

```go
func singleSenderReceiver() {
    ch := make(chan int, 10)
    
    // 发送者
    go func() {
        defer close(ch)  // 发送者负责关闭
        
        for i := 0; i < 10; i++ {
            ch <- i
        }
    }()
    
    // 接收者
    for v := range ch {  // range自动处理关闭
        fmt.Println(v)
    }
}
```

### 多发送者单接收者

**使用 WaitGroup 协调关闭**：

```go
func multipleSenders() {
    ch := make(chan int, 100)
    var wg sync.WaitGroup
    
    // 启动多个发送者
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < 10; j++ {
                ch <- id*10 + j
            }
        }(i)
    }
    
    // 等待所有发送者完成后关闭
    go func() {
        wg.Wait()
        close(ch)
    }()
    
    // 接收
    for v := range ch {
        fmt.Println(v)
    }
}
```

### 单发送者多接收者

**使用信号channel通知停止**：

```go
func multipleReceivers() {
    ch := make(chan int, 10)
    stop := make(chan struct{})
    
    // 发送者
    go func() {
        defer close(ch)
        
        for i := 0; ; i++ {
            select {
            case ch <- i:
            case <-stop:
                return
            }
        }
    }()
    
    // 多个接收者
    var wg sync.WaitGroup
    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for v := range ch {
                fmt.Printf("Worker %d: %d\n", id, v)
            }
        }(i)
    }
    
    // 5秒后停止
    time.Sleep(5 * time.Second)
    close(stop)
    
    wg.Wait()
}
```

### Safe Channel（防止重复关闭）

**包装channel提供安全操作**：

```go
type SafeChannel struct {
    ch     chan int
    once   sync.Once
    closed bool
    mu     sync.RWMutex
}

func (sc *SafeChannel) Send(v int) bool {
    sc.mu.RLock()
    defer sc.mu.RUnlock()
    
    if sc.closed {
        return false  // 已关闭，发送失败
    }
    
    sc.ch <- v
    return true
}

func (sc *SafeChannel) Close() {
    sc.mu.Lock()
    defer sc.mu.Unlock()
    
    sc.once.Do(func() {
        close(sc.ch)
        sc.closed = true
    })
}
```

## 5. 高级模式

### 超时模式

**避免无限等待**：

```go
select {
case result := <-ch:
    fmt.Println(result)
case <-time.After(1 * time.Second):
    fmt.Println("timeout")
}
```

**注意**：`time.After` 会创建新的timer，频繁调用会有GC压力，建议：

```go
timer := time.NewTimer(1 * time.Second)
defer timer.Stop()

select {
case result := <-ch:
    fmt.Println(result)
case <-timer.C:
    fmt.Println("timeout")
}
```

### Context 模式

**标准的取消和超时机制**：

```go
func worker(ctx context.Context, ch <-chan int) {
    for {
        select {
        case <-ctx.Done():
            return  // 被取消或超时
        case v, ok := <-ch:
            if !ok {
                return  // channel关闭
            }
            process(v)
        }
    }
}

// 使用
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

go worker(ctx, ch)
```

### Or-Done 模式

**组合done信号和数据channel**：

```go
func orDone(done <-chan struct{}, c <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for {
            select {
            case <-done:
                return
            case v, ok := <-c:
                if !ok {
                    return
                }
                select {
                case out <- v:
                case <-done:
                    return
                }
            }
        }
    }()
    return out
}

// 使用
done := make(chan struct{})
ch := orDone(done, dataChannel)

for v := range ch {
    // 处理数据
}
```

### Tee 模式

**复制channel数据到多个输出**：

```go
func tee(in <-chan int) (<-chan int, <-chan int) {
    out1 := make(chan int)
    out2 := make(chan int)
    
    go func() {
        defer close(out1)
        defer close(out2)
        
        for v := range in {
            var wg sync.WaitGroup
            wg.Add(2)
            
            go func() {
                defer wg.Done()
                out1 <- v
            }()
            
            go func() {
                defer wg.Done()
                out2 <- v
            }()
            
            wg.Wait()
        }
    }()
    
    return out1, out2
}
```

### Bridge 模式

**展平channel of channels**：

```go
func bridge(chanStream <-chan <-chan int) <-chan int {
    out := make(chan int)
    
    go func() {
        defer close(out)
        
        for ch := range chanStream {
            for v := range ch {
                out <- v
            }
        }
    }()
    
    return out
}
```

### 限流模式

**基于channel的速率限制**：

```go
type RateLimiter struct {
    tokens chan struct{}
}

func NewRateLimiter(qps int) *RateLimiter {
    rl := &RateLimiter{
        tokens: make(chan struct{}, qps),
    }
    
    // 填充令牌
    for i := 0; i < qps; i++ {
        rl.tokens <- struct{}{}
    }
    
    // 定期补充
    go func() {
        ticker := time.NewTicker(time.Second / time.Duration(qps))
        defer ticker.Stop()
        
        for range ticker.C {
            select {
            case rl.tokens <- struct{}{}:
            default:
            }
        }
    }()
    
    return rl
}

func (rl *RateLimiter) Wait() {
    <-rl.tokens
}
```

## 6. 最佳实践

### 1. 明确所有权

```go
// ✅ 好：清晰的所有权
func producer() <-chan int {
    ch := make(chan int)
    go func() {
        defer close(ch)  // producer负责关闭
        // 生产数据
    }()
    return ch
}

// ❌ 差：不清楚谁该关闭
var GlobalChannel chan int
```

### 2. 优先使用有缓冲channel

```go
// ✅ 好：减少阻塞
ch := make(chan Task, 100)

// ⚠️  慎用：可能频繁阻塞
ch := make(chan Task)
```

### 3. 总是处理channel关闭

```go
// ✅ 好：检查channel是否关闭
v, ok := <-ch
if !ok {
    // channel已关闭
}

// ✅ 好：使用range
for v := range ch {
    // 自动处理关闭
}

// ❌ 差：忽略关闭
v := <-ch  // channel关闭后会收到零值
```

### 4. 避免goroutine泄漏

```go
// ❌ 差：goroutine可能永远阻塞
func leak() {
    ch := make(chan int)
    go func() {
        ch <- 42  // 如果没有接收者，永远阻塞
    }()
}

// ✅ 好：使用context或done channel
func noLeak(ctx context.Context) {
    ch := make(chan int)
    go func() {
        select {
        case ch <- 42:
        case <-ctx.Done():
            return
        }
    }()
}
```

### 5. 正确使用select

```go
// ✅ 好：有default防止阻塞
select {
case ch <- value:
    // 发送成功
default:
    // channel满，做其他处理
}

// ✅ 好：设置超时
select {
case result := <-ch:
    return result
case <-time.After(timeout):
    return ErrTimeout
}
```

## 7. 常见陷阱

### 陷阱1：向nil channel发送/接收

```go
var ch chan int  // nil
ch <- 42         // 永远阻塞
v := <-ch        // 永远阻塞

// 用途：在select中动态启用/禁用case
var ch chan int
if needReceive {
    ch = make(chan int)
}

select {
case v := <-ch:  // ch为nil时这个case永远不会被选中
    // ...
}
```

### 陷阱2：for-select死循环

```go
// ❌ 差：忙等待，浪费CPU
for {
    select {
    case v := <-ch:
        process(v)
    default:
        // 立即返回，循环继续
    }
}

// ✅ 好：阻塞等待
for {
    select {
    case v := <-ch:
        process(v)
    case <-done:
        return
    }
}
```

### 陷阱3：time.After泄漏

```go
// ❌ 差：每次循环创建新timer
for {
    select {
    case <-ch:
        // ...
    case <-time.After(1 * time.Second):
        // timer不会被GC，造成内存泄漏
    }
}

// ✅ 好：复用timer
timer := time.NewTimer(1 * time.Second)
defer timer.Stop()

for {
    timer.Reset(1 * time.Second)
    select {
    case <-ch:
        // ...
    case <-timer.C:
        // ...
    }
}
```

### 陷阱4：忘记关闭channel

```go
// ❌ 差：接收者不知道何时停止
func leak() <-chan int {
    ch := make(chan int)
    go func() {
        for i := 0; i < 10; i++ {
            ch <- i
        }
        // 忘记close(ch)
    }()
    return ch
}

// 接收者会永远等待
for v := range leak() {  // 永远不会退出
    process(v)
}

// ✅ 好：明确关闭
func noLeak() <-chan int {
    ch := make(chan int)
    go func() {
        defer close(ch)
        for i := 0; i < 10; i++ {
            ch <- i
        }
    }()
    return ch
}
```

### 陷阱5：在错误的地方关闭channel

```go
// ❌ 差：接收者关闭channel
func badPattern() {
    ch := make(chan int, 10)
    
    go func() {
        for i := 0; i < 10; i++ {
            ch <- i  // 可能panic！
        }
    }()
    
    for v := range ch {
        process(v)
        if shouldStop {
            close(ch)  // ❌ 接收者不应该关闭
            break
        }
    }
}

// ✅ 好：发送者关闭channel
func goodPattern() {
    ch := make(chan int, 10)
    stop := make(chan struct{})
    
    go func() {
        defer close(ch)
        for i := 0; i < 10; i++ {
            select {
            case ch <- i:
            case <-stop:
                return
            }
        }
    }()
    
    for v := range ch {
        process(v)
        if shouldStop {
            close(stop)  // ✅ 通知发送者停止
            break
        }
    }
}
```

## 8. 性能考虑

### 缓冲大小的影响

```bash
无缓冲:     最慢（每次都需要握手）
小缓冲(10): 适中（减少部分阻塞）
大缓冲(100): 较快（很少阻塞）
巨大缓冲:   浪费内存，GC压力大
```

### Worker数量选择

```go
// CPU密集型
workers := runtime.NumCPU()

// IO密集型
workers := runtime.NumCPU() * 2  // 或更多

// 混合型
workers := runtime.NumCPU() + 1
```

## 9. 调试技巧

### 检测goroutine泄漏

```go
func TestNoGoroutineLeak(t *testing.T) {
    before := runtime.NumGoroutine()
    
    // 运行测试代码
    runTest()
    
    // 等待goroutine完成
    time.Sleep(100 * time.Millisecond)
    
    after := runtime.NumGoroutine()
    if after > before {
        t.Errorf("Goroutine leak: before=%d, after=%d", before, after)
    }
}
```

### 使用race检测器

```bash
go test -race
go build -race
go run -race
```

## 10. 总结

### Channel 使用决策树

```
需要goroutine间通信？
├─ 是 → 使用channel
│  ├─ 需要同步？
│  │  ├─ 是 → 无缓冲channel
│  │  └─ 否 → 有缓冲channel
│  ├─ 需要广播？
│  │  └─ 使用close(ch)
│  └─ 需要取消？
│     └─ 使用context
└─ 否 → 考虑sync包（Mutex, WaitGroup等）
```

### 核心原则

1. **单一职责**：一个goroutine只负责发送或接收
2. **明确所有权**：谁创建谁关闭
3. **避免泄漏**：总是提供退出机制
4. **优雅关闭**：等待任务完成
5. **类型安全**：使用只读/只写channel限制操作

### 何时不用Channel

```
❌ 简单的互斥访问 → 用 sync.Mutex
❌ 等待多个goroutine → 用 sync.WaitGroup
❌ 一次性事件 → 用 sync.Once
❌ 配置数据 → 用 sync.RWMutex 或 atomic
✅ goroutine间通信 → 用 Channel
```

## 参考资料

- [Go Concurrency Patterns](https://go.dev/talks/2012/concurrency.slide)
- [Advanced Go Concurrency Patterns](https://go.dev/talks/2013/advconc.slide)
- [Share Memory By Communicating](https://go.dev/blog/codelab-share)
- [Concurrency in Go (Book)](https://www.oreilly.com/library/view/concurrency-in-go/9781491941294/)

---

**记住：Channel 是 Go 的核心特性，但不是唯一的并发工具。选择合适的工具完成任务。**

