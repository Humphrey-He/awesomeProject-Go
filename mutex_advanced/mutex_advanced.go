package mutex_advanced

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ========== Go Mutex/RWMutex 深入理解与最佳实践 ==========

/*
本项目讲解 Go 同步原语的底层实现原理，包括：

一、Mutex 互斥锁
1. 锁状态与等待队列
2. 公平锁 vs 非公平锁
3. 自旋机制
4. 锁升级与降级

二、RWMutex 读写锁
1. 读锁与写锁的实现
2. 读写公平策略
3. 锁等待队列

三、WaitGroup 原理
1. 计数器实现
2. goroutine等待

四、Once 原理
1. 一次性执行

五、性能优化
1. 锁粒度控制
2. 避免死锁
*/

// ========== 1. Mutex 深入理解 ==========

/*
Go Mutex 状态：

type Mutex struct {
    state int32  // 锁状态
    sema uint32  // 信号量
}

状态位：
- bit 0: locked (1=已锁定)
- bit 1: woken (1=唤醒状态)
- 高位: waiters count (等待者数量)

状态转换：
- 0 -> 1: 获取锁
- 1 -> 0: 释放锁
- 自旋等待: CAS 循环
*/

// MutexState 模拟Mutex状态
type MutexState struct {
	state int32 // 原子操作
	sema  uint32
}

// TryLock 尝试获取锁（非阻塞）
func (m *MutexState) TryLock() bool {
	return atomic.CompareAndSwapInt32(&m.state, 0, 1)
}

// Lock 获取锁（模拟实现）
func (m *MutexState) Lock() {
	// 1. 尝试快速获取（自旋）
	for !atomic.CompareAndSwapInt32(&m.state, 0, 1) {
		// 等待信号量
		// 实际会调用 runtime_Semacquire
	}
}

// Unlock 释放锁
func (m *MutexState) Unlock() {
	atomic.StoreInt32(&m.state, 0)
}

// MutexDemo Mutex使用示例
func MutexDemo() {
	var mu sync.Mutex
	counter := 0

	// 并发增加计数器
	for i := 0; i < 1000; i++ {
		go func() {
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}

	time.Sleep(time.Millisecond)
	fmt.Println("Counter:", counter)
}

// ========== 2. RWMutex 深入理解 ==========

/*
RWMutex 结构：

type RWMutex struct {
    w           Mutex    // 写锁
    writerSem   uint32   // 写等待信号量
    readerSem   uint32   // 读等待信号量
    readerCount int32    // 当前读者数量
    readerWait  int32    // 等待的写锁读者数量
}

读锁流程：
1. readerCount++
2. 如果有写锁等待，阻塞
3. 获取读锁

写锁流程：
1. 获取互斥锁
2. 等待所有读锁释放
3. 获取写锁
*/

// RWMutexState 模拟RWMutex状态
type RWMutexState struct {
	w           int32
	writerSem   uint32
	readerSem   uint32
	readerCount int32
	readerWait  int32
}

// RLock 获取读锁
func (rw *RWMutexState) RLock() {
	// 增加读者计数
	atomic.AddInt32(&rw.readerCount, 1)
}

// RUnlock 释放读锁
func (rw *RWMutexState) RUnlock() {
	atomic.AddInt32(&rw.readerCount, -1)
}

// Lock 获取写锁
func (rw *RWMutexState) Lock() {
	// 获取互斥锁
	atomic.AddInt32(&rw.w, 1)
	// 等待所有读锁释放
	for atomic.LoadInt32(&rw.readerCount) > 0 {
		// 自旋等待
	}
}

// Unlock 释放写锁
func (rw *RWMutexState) Unlock() {
	atomic.StoreInt32(&rw.w, 0)
}

// RWMutexDemo 读写锁示例
func RWMutexDemo() {
	var rw sync.RWMutex
	counter := 0

	// 多个读 goroutine
	for i := 0; i < 10; i++ {
		go func() {
			for {
				rw.RLock()
				_ = counter
				rw.RUnlock()
			}
		}()
	}

	// 写 goroutine
	for i := 0; i < 2; i++ {
		go func() {
			for {
				rw.Lock()
				counter++
				rw.Unlock()
			}
		}()
	}

	time.Sleep(time.Millisecond)
	fmt.Println("RWMutex demo done")
}

// ========== 3. 公平锁 vs 非公平锁 ==========

/*
Go Mutex 设计：
- 默认为非公平锁
- 新请求可能"插队"等待队列
- 优点：减少唤醒延迟，提高吞吐量

公平锁：
- 严格按等待顺序获取锁
- 可能导致等待时间长

自旋：
- 多核CPU上短时间自旋等待
- 避免goroutine切换开销
*/

// FairMutex 公平锁实现
type FairMutex struct {
	mu      sync.Mutex
	queue   chan struct{}
	locked  bool
}

func NewFairMutex() *FairMutex {
	return &FairMutex{
		queue: make(chan struct{}, 1),
	}
}

func (m *FairMutex) Lock() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.locked {
		m.locked = true
		return
	}

	// 加入等待队列
	m.queue <- struct{}{}
}

func (m *FairMutex) Unlock() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.queue) > 0 {
		// 唤醒一个等待者
		<-m.queue
	} else {
		m.locked = false
	}
}

// ========== 4. 锁粒度控制 ==========

/*
锁粒度控制原则：

❌ 错误：
mu.Lock()
data := expensiveOperation()
mu.Unlock()
process(data)

✅ 正确：
data := expensiveOperation()
mu.Lock()
sharedData = data
mu.Unlock()
process(data)
*/

// BadLockExample 错误的锁使用
type BadLockExample struct {
	mu   sync.Mutex
	data map[string]int
}

func (b *BadLockExample) Process() int {
	b.mu.Lock()
	// 在锁内做耗时操作
	result := 0
	for _, v := range b.data {
		result += v * 2 // 简单计算
		time.Sleep(time.Millisecond)
	}
	b.mu.Unlock()
	return result
}

// GoodLockExample 正确的锁使用
func (b *BadLockExample) ProcessGood() int {
	// 在锁外准备数据
	items := make([]int, 0)
	b.mu.Lock()
	for _, v := range b.data {
		items = append(items, v)
	}
	b.mu.Unlock()

	// 在锁外处理
	result := 0
	for _, v := range items {
		result += v * 2
	}
	return result
}

// ========== 5. 死锁避免 ==========

/*
死锁原因：
1. 循环等待
2. 持有并等待
3. 不可抢占
4. 互斥

避免策略：
1. 固定顺序获取锁
2. 使用超时
3. 避免嵌套锁
*/

// DeadlockExample 死锁示例
func DeadlockExample() {
	var mu1, mu2 sync.Mutex

	// Goroutine 1
	go func() {
		mu1.Lock()
		time.Sleep(time.Millisecond)
		mu2.Lock()
		mu2.Unlock()
		mu1.Unlock()
	}()

	// Goroutine 2
	go func() {
		mu2.Lock()
		time.Sleep(time.Millisecond)
		mu1.Lock()
		mu1.Unlock()
		mu2.Unlock()
	}()

	time.Sleep(time.Second)
	fmt.Println("Deadlock example")
}

// SafeLockExample 安全的锁顺序
func SafeLockExample() {
	var mu1, mu2 sync.Mutex

	// 统一顺序：先锁mu1，再锁mu2
	go func() {
		mu1.Lock()
		time.Sleep(time.Millisecond)
		mu2.Lock()
		mu2.Unlock()
		mu1.Unlock()
	}()

	go func() {
		mu1.Lock()
		time.Sleep(time.Millisecond)
		mu2.Lock()
		mu2.Unlock()
		mu1.Unlock()
	}()

	time.Sleep(time.Second)
	fmt.Println("Safe lock example")
}

// ========== 6. WaitGroup 原理 ==========

/*
WaitGroup 结构：

type WaitGroup struct {
    noCopy noCopy
    state atomic.Uint64 // 高32位:counter, 低32位:waiters
}

操作：
- Add(delta): 增加计数器
- Done(): 减少计数器 (Add(-1))
- Wait(): 等待计数器为0
*/

// SimulatedWaitGroup 模拟WaitGroup
type SimulatedWaitGroup struct {
	counter  int32
	waiters int32
	ch      chan struct{}
}

func NewSimulatedWaitGroup() *SimulatedWaitGroup {
	return &SimulatedWaitGroup{ch: make(chan struct{})}
}

func (wg *SimulatedWaitGroup) Add(delta int) {
	atomic.AddInt32(&wg.counter, int32(delta))
}

func (wg *SimulatedWaitGroup) Done() {
	atomic.AddInt32(&wg.counter, -1)
	if atomic.LoadInt32(&wg.counter) == 0 {
		close(wg.ch)
	}
}

func (wg *SimulatedWaitGroup) Wait() {
	<-wg.ch
}

// WaitGroupDemo WaitGroup示例
func WaitGroupDemo() {
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			fmt.Println("Goroutine:", n)
		}(i)
	}

	wg.Wait()
	fmt.Println("All done")
}

// ========== 7. Once 原理 ==========

/*
Once 结构：

type Once struct {
    m Mutex
    done uint32
}

Do 函数：
func (o *Once) Do(f func()) {
    if atomic.LoadUint32(&o.done) == 0 {
        o.doSlow(f)
    }
}

func (o *Once) doSlow(f func()) {
    o.m.Lock()
    defer o.m.Unlock()
    if o.done == 0 {
        f()
        atomic.StoreUint32(&o.done, 1)
    }
}
*/

// SimulatedOnce 模拟Once
type SimulatedOnce struct {
	mu    sync.Mutex
	done  uint32
}

func (o *SimulatedOnce) Do(f func()) {
	if atomic.LoadUint32(&o.done) == 0 {
		o.doSlow(f)
	}
}

func (o *SimulatedOnce) doSlow(f func()) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.done == 0 {
		f()
		atomic.StoreUint32(&o.done, 1)
	}
}

// OnceDemo Once示例
func OnceDemo() {
	var once sync.Once
	once.Do(func() {
		fmt.Println("This runs only once")
	})
	once.Do(func() {
		fmt.Println("This will not run")
	})
	once.Do(func() {
		fmt.Println("This will also not run")
	})
}

// ========== 8. 条件变量 ==========

/*
Cond 用于等待某个条件发生

type Cond struct {
    L Locker
    notify  list.List
    checker list.List
}

方法：
- Wait(): 等待通知
- Signal(): 唤醒一个等待者
- Broadcast(): 唤醒所有等待者
*/

// CondDemo Cond示例
func CondDemo() {
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	ready := false

	// 等待者
	for i := 0; i < 3; i++ {
		go func(n int) {
			mu.Lock()
			for !ready {
				cond.Wait()
			}
			mu.Unlock()
			fmt.Println("Goroutine", n, "ready")
		}(i)
	}

	// 通知者
	go func() {
		time.Sleep(time.Millisecond)
		mu.Lock()
		ready = true
		cond.Broadcast()
		mu.Unlock()
	}()

	time.Sleep(time.Second)
}

// ========== 9. 性能优化技巧 ==========

/*
性能优化建议：

1. 减少锁持有时间
   - 锁内操作尽量简单
   - 复制数据到锁外处理

2. 读写分离
   - 读多写少用 RWMutex
   - 避免写锁阻塞读

3. 避免热点
   - 分散锁竞争
   - 使用分片锁

4. 无锁编程
   - 使用 atomic
   - 使用 sync.Map
*/

// OptimizedCounter 优化计数器
type OptimizedCounter struct {
	counters [10]int64
}

func NewOptimizedCounter() *OptimizedCounter {
	return &OptimizedCounter{}
}

func (c *OptimizedCounter) Add(n int) {
	idx := n % len(c.counters)
	atomic.AddInt64(&c.counters[idx], 1)
}

func (c *OptimizedCounter) Sum() int64 {
	var total int64
	for _, c := range c.counters {
		total += atomic.LoadInt64(&c)
	}
	return total
}

// ========== 10. 面试要点 ==========

/*
Mutex 面试题：

Q: Go Mutex 是公平锁吗？
A: 默认非公平，新请求可能插队，但有等待队列

Q: Mutex 和 RWMutex 的区别？
A: Mutex 独占，RWMutex 读写分离，读多写少场景性能好

Q: 自旋锁的优点？
A: 避免 goroutine 切换，短时间等待时效率高

Q: 如何避免死锁？
A: 固定锁顺序、使用超时、避免嵌套锁

Q: WaitGroup 可以复用吗？
A: 可以，Wait 后可以重新 Add
*/

// ========== 11. 完整示例 ==========

// CompleteExample 完整示例
func CompleteExample() {
	// Mutex
	MutexDemo()

	// RWMutex
	RWMutexDemo()

	// WaitGroup
	WaitGroupDemo()

	// Once
	OnceDemo()

	// Cond
	CondDemo()

	// Optimized counter
	counter := NewOptimizedCounter()
	for i := 0; i < 100; i++ {
		counter.Add(i)
	}
	fmt.Println("Sum:", counter.Sum())
}
