package race_detector

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ========== Go Race 检测深入理解 ==========

/*
本项目讲解 Go race 检测器的原理和使用，包括：

一、Race 基础
1. 什么是数据竞争
2. 竞争条件类型
3. 常见场景

二、Race 检测器
1. 工作原理
2. 编译插桩
3. 阴影变量

三、检测方法
1. -race 标志
2. race detector API

四、解决方案
1. Mutex
2. atomic
3. channel
*/

// ========== 1. 数据竞争基础 ==========

/*
数据竞争（Data Race）：

定义：两个或多个 goroutine 同时访问同一内存位置，
至少有一个是写操作

竞争类型：

1. 写-写竞争
   goroutine A: x = 1
   goroutine B: x = 2
   结果不确定

2. 读-写竞争
   goroutine A: y = x + 1
   goroutine B: x = 100
   y 可能为 1 或 101
*/

// RaceDemoWriteWrite 写-写竞争（有竞态问题）
func RaceDemoWriteWrite() {
	fmt.Println("=== 写-写竞争演示（有竞态）===")
	x := 0
	
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			x = 1 // 竞态写
		}()
		go func() {
			defer wg.Done()
			x = 2 // 竞态写
		}()
	}
	wg.Wait()
	fmt.Printf("x = %d (结果不确定)\n", x)
}

// RaceDemoReadWrite 读-写竞争（有竞态问题）
func RaceDemoReadWrite() {
	fmt.Println("\n=== 读-写竞争演示（有竞态）===")
	x := 0
	var wg sync.WaitGroup
	
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = x // 竞态读
		}()
		go func() {
			defer wg.Done()
			x = i // 竞态写
		}()
	}
	wg.Wait()
	fmt.Println("读写竞争完成（结果不可预测）")
}

// ========== 2. Race 检测原理 ==========

/*
Race 检测器工作原理：

1. 编译时插桩
   - 分析内存访问
   - 添加检测代码

2. 运行时检测
   - 记录内存访问
   - 检测冲突

3. 阴影变量（Shadow Memory）
   - 为每个变量创建副本
   - 记录访问时间和goroutine

原理图：
┌─────────────────────────────────────────┐
│           原始代码                       │
│  goroutine A: x = 1                    │
│  goroutine B: y = x + 1                │
├─────────────────────────────────────────┤
│           插桩后                         │
│  goroutine A: racefunclog(&x, Write)   │
│  goroutine B: racefunclog(&x, Read)    │
└─────────────────────────────────────────┘
*/

// ========== 3. 竞态问题解决方案 ==========

/*
解决方案：

1. Mutex 互斥锁
2. RWMutex 读写锁
3. atomic 原子操作
4. channel 通道
5. sync.WaitGroup
*/

// ========== 3.1 使用 Mutex ==========

// SafeCounterMutex 使用 Mutex 的线程安全计数器
type SafeCounterMutex struct {
	mu    sync.Mutex
	count int
}

func (c *SafeCounterMutex) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
}

func (c *SafeCounterMutex) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

// MutexDemo Mutex 解决方案示例
func MutexDemo() {
	fmt.Println("\n=== Mutex 解决方案 ===")
	counter := &SafeCounterMutex{}
	
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Inc()
		}()
	}
	
	wg.Wait()
	fmt.Printf("Counter: %d (正确结果)\n", counter.Value())
}

// ========== 3.2 使用 RWMutex ==========

// SafeCounterRWMutex 使用 RWMutex 的线程安全计数器
type SafeCounterRWMutex struct {
	mu    sync.RWMutex
	count int
}

func (c *SafeCounterRWMutex) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
}

func (c *SafeCounterRWMutex) Value() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.count
}

// RWMutexDemo RWMutex 解决方案示例
func RWMutexDemo() {
	fmt.Println("\n=== RWMutex 解决方案 ===")
	counter := &SafeCounterRWMutex{}
	
	var wg sync.WaitGroup
	
	// 多个读者
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = counter.Value()
		}()
	}
	
	// 多个写者
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Inc()
		}()
	}
	
	wg.Wait()
	fmt.Printf("Counter: %d\n", counter.Value())
}

// ========== 3.3 使用 atomic ==========

// SafeCounterAtomic 使用 atomic 的线程安全计数器
type SafeCounterAtomic struct {
	count int64
}

func (c *SafeCounterAtomic) Inc() {
	atomic.AddInt64(&c.count, 1)
}

func (c *SafeCounterAtomic) Value() int64 {
	return atomic.LoadInt64(&c.count)
}

// AtomicDemo atomic 解决方案示例
func AtomicDemo() {
	fmt.Println("\n=== atomic 解决方案 ===")
	counter := &SafeCounterAtomic{}
	
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Inc()
		}()
	}
	
	wg.Wait()
	fmt.Printf("Counter: %d (正确结果)\n", counter.Value())
}

// ========== 3.4 使用 channel ==========

// SafeCounterChannel 使用 channel 的线程安全计数器
type SafeCounterChannel struct {
	ch    chan func()
	count int
}

func NewSafeCounterChannel() *SafeCounterChannel {
	c := &SafeCounterChannel{
		ch: make(chan func(), 1),
	}
	// 启动事件循环
	go func() {
		for fn := range c.ch {
			fn()
		}
	}()
	return c
}

func (c *SafeCounterChannel) Inc() {
	done := make(chan struct{})
	c.ch <- func() {
		c.count++
		close(done)
	}
	<-done
}

func (c *SafeCounterChannel) Value() int {
	done := make(chan int)
	c.ch <- func() {
		done <- c.count
	}
	return <-done
}

func (c *SafeCounterChannel) Close() {
	close(c.ch)
}

// ChannelDemo channel 解决方案示例
func ChannelDemo() {
	fmt.Println("\n=== channel 解决方案 ===")
	counter := NewSafeCounterChannel()
	defer counter.Close()
	
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Inc()
		}()
	}
	
	wg.Wait()
	fmt.Printf("Counter: %d (正确结果)\n", counter.Value())
}

// ========== 4. 常见竞态场景 ==========

// LazyInitRace 延迟初始化竞态（有竞态问题）
func LazyInitRace() {
	fmt.Println("\n=== 延迟初始化竞态演示 ===")
	var once sync.Once
	initialized := false
	initCount := 0
	
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !initialized { // 竞态检查
				once.Do(func() {
					initialized = true
					initCount++
					fmt.Println("初始化...")
				})
			}
		}()
	}
	
	wg.Wait()
	fmt.Printf("初始化次数: %d (应该为1)\n", initCount)
}

// LazyInitFixed 使用 sync.Once 修复
func LazyInitFixed() {
	fmt.Println("\n=== 正确的延迟初始化 ===")
	var once sync.Once
	initCount := 0
	
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			once.Do(func() {
				initCount++
				fmt.Println("初始化...")
			})
		}()
	}
	
	wg.Wait()
	fmt.Printf("初始化次数: %d (正确)\n", initCount)
}

// ========== 5. 检测方法 ==========

/*
Race 检测方法：

1. -race 标志（推荐）
   go run -race main.go
   go test -race ./...

输出示例：
WARNING: DATA RACE
Read at 0x00c0000a6010
Previous write at 0x00c0000a6008
Goroutine 5 (running) started at...
*/

// RaceDetectionDemo 竞态检测演示
func RaceDetectionDemo() {
	fmt.Println("\n=== 竞态检测方法 ===")
	fmt.Println("使用命令检测竞态:")
	fmt.Println("  go run -race main.go")
	fmt.Println("  go test -race ./...")
	fmt.Println("  go build -race main.go")
}

// ========== 6. 性能对比 ==========

// PerformanceComparison 性能对比
func PerformanceComparison() {
	fmt.Println("\n=== 性能对比 ===")
	
	iterations := 100000
	
	// Mutex 性能
	mutexCounter := &SafeCounterMutex{}
	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations/10; j++ {
				mutexCounter.Inc()
			}
		}()
	}
	wg.Wait()
	mutexTime := time.Since(start)
	
	// atomic 性能
	atomicCounter := &SafeCounterAtomic{}
	start = time.Now()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations/10; j++ {
				atomicCounter.Inc()
			}
		}()
	}
	wg.Wait()
	atomicTime := time.Since(start)
	
	fmt.Printf("Mutex 计数: %d, 耗时: %v\n", mutexCounter.Value(), mutexTime)
	fmt.Printf("Atomic 计数: %d, 耗时: %v\n", atomicCounter.Value(), atomicTime)
}

// ========== 7. 面试要点 ==========

/*
Race 检测面试题：

Q: 什么是数据竞争？
A: 多个 goroutine 同时访问同一内存位置，至少有一个写操作

Q: race 检测原理？
A: 编译时插桩，运行时检测，记录内存访问

Q: 如何检测 race？
A: go run -race main.go

Q: 解决 race 的方法？
A: Mutex、atomic、channel、sync.Once

Q: race 检测的性能影响？
A: 内存 2-10 倍，CPU 5-10 倍

Q: 生产环境是否开启 race？
A: 不建议，性能开销大
*/

// CompleteExample 完整示例
func CompleteExample() {
	// 竞态演示
	RaceDemoWriteWrite()
	RaceDemoReadWrite()
	
	// 解决方案
	MutexDemo()
	AtomicDemo()
	ChannelDemo()
	
	// 常见问题
	LazyInitRace()
	LazyInitFixed()
	
	// 性能对比
	PerformanceComparison()
	
	// 检测方法
	RaceDetectionDemo()
}
