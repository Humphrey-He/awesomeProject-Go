package cond_practices

import (
	"fmt"
	"sync"
	"time"
)

// ========== Go Cond 条件变量深入理解与最佳实践 ==========

/*
本项目讲解 Go sync.Cond 条件变量的使用，包括：

一、Cond 基础
1. 什么是条件变量
2. Cond 结构与方法
3. 使用场景

二、Cond 进阶
1. 广播与单发
2. 等待超时
3. 虚假唤醒

三、Cond 实战
1. 生产者-消费者
2. 线程池任务队列
3. 连接池

四、原理剖析
1. 等待队列
2. 锁释放与获取
*/

// ========== 1. Cond 基础 ==========

/*
Cond 条件变量：

type Cond struct {
    L Locker       // 关联的锁
    notify  list.List  // 等待队列
    checker list.List  // 检查队列
}

方法：
- Wait(): 等待条件满足
- Signal(): 唤醒一个等待者
- Broadcast(): 唤醒所有等待者

使用流程：
1. Lock
2. 检查条件
3. Wait (自动释放锁)
4. 条件满足后 Unlock
*/

// CondDemo 基础示例
func CondDemo() {
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	ready := false

	// 等待者
	go func() {
		mu.Lock()
		for !ready {
			cond.Wait()
		}
		mu.Unlock()
		fmt.Println("Waiter: ready!")
	}()

	// 通知者
	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	ready = true
	cond.Signal()
	mu.Unlock()

	time.Sleep(time.Second)
}

// ========== 2. 广播与单发 ==========

/*
Signal vs Broadcast：

Signal:
- 唤醒一个等待者
- 按加入顺序唤醒（FIFO）
- 建议在有多个等待者时使用

Broadcast:
- 唤醒所有等待者
- 用于条件变化影响所有等待者
- 例如：配置更新、关闭通知
*/

// BroadcastDemo 广播示例
func BroadcastDemo() {
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	workers := 5

	// 多个worker等待任务
	for i := 0; i < workers; i++ {
		go func(id int) {
			mu.Lock()
			cond.Wait()
			mu.Unlock()
			fmt.Printf("Worker %d: received signal\n", id)
		}(i)
	}

	// 广播任务
	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	fmt.Println("Broadcasting...")
	cond.Broadcast()
	mu.Unlock()

	time.Sleep(time.Second)
}

// ========== 3. 等待超时 ==========

/*
等待超时实现：

使用 time.After 或 context 实现超时：

done := make(chan struct{})
go func() {
    mu.Lock()
    for !condition {
        cond.Wait()
    }
    mu.Unlock()
    close(done)
}()

select {
case <-done:
    // 条件满足
case <-time.After(timeout):
    // 超时
}
*/

// WaitWithTimeout 超时等待
func WaitWithTimeout(cond *sync.Cond, pred func() bool, timeout time.Duration) bool {
	done := make(chan struct{})
	
	go func() {
		cond.L.Lock()
		for !pred() {
			cond.Wait()
		}
		cond.L.Unlock()
		close(done)
	}()
	
	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}

// TimeoutDemo 超时示例
func TimeoutDemo() {
	var mu sync.Mutex
	cond := sync.NewCond(&mu)
	ready := false

	// 等待者带超时
	go func() {
		ok := WaitWithTimeout(cond, func() bool { return ready }, 2*time.Second)
		if !ok {
			fmt.Println("Wait: timeout!")
			return
		}
		fmt.Println("Wait: ready!")
	}()

	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	ready = true
	cond.Signal()
	mu.Unlock()

	time.Sleep(time.Second)
}

// ========== 4. 生产者-消费者 ==========

/*
生产者-消费者模型：

生产者：
- 生产数据
- 放入队列
- 通知消费者

消费者：
- 等待数据
- 从队列取出
- 处理数据
*/

// ProducerConsumer 生产者-消费者
type ProducerConsumer struct {
	mu       sync.Mutex
	cond     *sync.Cond
	queue    []int
	capacity int
}

func NewProducerConsumer(capacity int) *ProducerConsumer {
	pc := &ProducerConsumer{
		capacity: capacity,
		queue:    make([]int, 0),
	}
	pc.cond = sync.NewCond(&pc.mu)
	return pc
}

func (pc *ProducerConsumer) Produce(item int) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// 队列满，等待消费者
	for len(pc.queue) >= pc.capacity {
		pc.cond.Wait()
	}

	pc.queue = append(pc.queue, item)
	fmt.Printf("Produced: %d, queue len: %d\n", item, len(pc.queue))
	
	// 通知消费者
	pc.cond.Signal()
}

func (pc *ProducerConsumer) Consume() int {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	// 队列空，等待生产者
	for len(pc.queue) == 0 {
		pc.cond.Wait()
	}

	item := pc.queue[0]
	pc.queue = pc.queue[1:]
	fmt.Printf("Consumed: %d, queue len: %d\n", item, len(pc.queue))
	
	// 通知生产者
	pc.cond.Signal()
	
	return item
}

// ProducerConsumerDemo 生产者-消费者示例
func ProducerConsumerDemo() {
	pc := NewProducerConsumer(3)

	// 生产者
	for i := 0; i < 5; i++ {
		go func(n int) {
			pc.Produce(n)
		}(i)
	}

	// 消费者
	for i := 0; i < 5; i++ {
		go func() {
			pc.Consume()
		}()
	}

	time.Sleep(2 * time.Second)
}

// ========== 5. 任务队列 ==========

/*
任务队列应用：

- 线程池任务分发
- 异步任务处理
- 连接池管理
*/

// TaskQueue 任务队列
type TaskQueue struct {
	mu       sync.Mutex
	cond     *sync.Cond
	tasks    []func()
	closed   bool
}

func NewTaskQueue() *TaskQueue {
	tq := &TaskQueue{
		tasks: make([]func(), 0),
	}
	tq.cond = sync.NewCond(&tq.mu)
	return tq
}

func (tq *TaskQueue) Add(task func()) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	if tq.closed {
		return
	}

	tq.tasks = append(tq.tasks, task)
	tq.cond.Signal()
}

func (tq *TaskQueue) Get() func() {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	for len(tq.tasks) == 0 && !tq.closed {
		tq.cond.Wait()
	}

	if tq.closed && len(tq.tasks) == 0 {
		return nil
	}

	task := tq.tasks[0]
	tq.tasks = tq.tasks[1:]
	return task
}

func (tq *TaskQueue) Close() {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	tq.closed = true
	tq.cond.Broadcast()
}

// TaskQueueDemo 任务队列示例
func TaskQueueDemo() {
	queue := NewTaskQueue()

	// 工作池
	for i := 0; i < 3; i++ {
		go func(id int) {
			for {
				task := queue.Get()
				if task == nil {
					fmt.Printf("Worker %d: exiting\n", id)
					return
				}
				fmt.Printf("Worker %d: executing task\n", id)
				task()
			}
		}(i)
	}

	// 添加任务
	for i := 0; i < 10; i++ {
		n := i
		queue.Add(func() {
			fmt.Printf("Task %d done\n", n)
		})
	}

	time.Sleep(500 * time.Millisecond)
	queue.Close()
	time.Sleep(time.Second)
}

// ========== 6. 虚假唤醒 ==========

/*
虚假唤醒（Spurious Wakeup）：

原因：
- 操作系统层面的唤醒
- 不确定何时发生

解决：
- 始终使用 for 循环等待
- 不要使用 if

正确：
for !condition {
    cond.Wait()
}
*/

// CorrectWait 正确的等待方式
func CorrectWait(cond *sync.Cond, condition func() bool) {
	cond.L.Lock()
	defer cond.L.Unlock()
	
	for !condition() {
		cond.Wait()
	}
}

// ========== 7. 条件变量原理 ==========

/*
Cond 原理：

Wait 实现：
1. 释放锁
2. 阻塞等待信号
3. 重新获取锁

Signal 实现：
1. 从等待队列取出一个
2. 标记为可运行
3. 等待者会被调度

Broadcast 实现：
1. 唤醒所有等待者
2. 依次获取锁执行
*/

// ========== 8. 注意事项 ==========

/*
⚠️ 注意事项：

1. 始终在锁内使用 Wait
2. 使用 for 而非 if 检查条件
3. Signal vs Broadcast 选择
4. 等待超时处理
5. 避免在条件满足后继续 Wait
*/

// ========== 9. 面试要点 ==========

/*
Cond 面试题：

Q: Cond 和 Channel 的区别？
A: Cond 适合单一共享状态，Channel 适合传递数据

Q: 为什么要用 for 而不是 if？
A: 防止虚假唤醒

Q: Signal 和 Broadcast 区别？
A: Signal 唤醒一个，Broadcast 唤醒所有

Q: Cond 的典型应用场景？
A: 生产者-消费者、任务队列、连接池
*/

// ========== 10. 完整示例 ==========

// CompleteExample 完整示例
func CompleteExample() {
	fmt.Println("=== Cond Demo ===")
	CondDemo()

	fmt.Println("=== Broadcast Demo ===")
	BroadcastDemo()

	fmt.Println("=== Timeout Demo ===")
	TimeoutDemo()

	fmt.Println("=== Producer-Consumer Demo ===")
	ProducerConsumerDemo()

	fmt.Println("=== Task Queue Demo ===")
	TaskQueueDemo()
}
