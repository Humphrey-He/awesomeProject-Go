package channel_patterns

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ========== 1. Channel 创建模式 ==========

// 无缓冲 channel - 用于同步通信

func NewUnbufferedChannel[T any]() chan T {
	return make(chan T)
}

// 有缓冲 channel - 用于异步通信

func NewBufferedChannel[T any](size int) chan T {
	return make(chan T, size)
}

// 只读 channel - 限制接收方不能发送

func AsReadOnly[T any](ch chan T) <-chan T {
	return ch
}

// 只写 channel - 限制发送方不能接收

func AsWriteOnly[T any](ch chan T) chan<- T {
	return ch
}

// ========== 2. Channel 传递模式 ==========

// Pipeline 模式 - 流式处理数据
type Pipeline[T any] struct {
	stages []func(T) T
}

func NewPipeline[T any]() *Pipeline[T] {
	return &Pipeline[T]{}
}

func (p *Pipeline[T]) AddStage(fn func(T) T) *Pipeline[T] {
	p.stages = append(p.stages, fn)
	return p
}

func (p *Pipeline[T]) Execute(input <-chan T) <-chan T {
	out := input
	for _, stage := range p.stages {
		out = p.applyStage(out, stage)
	}
	return out
}

func (p *Pipeline[T]) applyStage(in <-chan T, fn func(T) T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for v := range in {
			out <- fn(v)
		}
	}()
	return out
}

// Fan-Out 模式 - 一个输入分发到多个worker
func FanOut[T any](input <-chan T, workers int) []<-chan T {
	outputs := make([]<-chan T, workers)
	for i := 0; i < workers; i++ {
		out := make(chan T)
		outputs[i] = out
		go func(ch chan T) {
			defer close(ch)
			for v := range input {
				ch <- v
			}
		}(out)
	}
	return outputs
}

// Fan-In 模式 - 多个输入合并到一个输出
func FanIn[T any](inputs ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup

	for _, input := range inputs {
		wg.Add(1)
		go func(ch <-chan T) {
			defer wg.Done()
			for v := range ch {
				out <- v
			}
		}(input)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// ========== 3. Worker Pool 模式（Channel 复用）==========

// WorkerPool 工作池
type WorkerPool[In any, Out any] struct {
	workers    int
	jobQueue   chan In
	resultChan chan Out
	process    func(In) Out
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewWorkerPool 创建工作池
func NewWorkerPool[In any, Out any](workers, queueSize int, process func(In) Out) *WorkerPool[In, Out] {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool[In, Out]{
		workers:    workers,
		jobQueue:   make(chan In, queueSize),
		resultChan: make(chan Out, queueSize),
		process:    process,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start 启动工作池
func (wp *WorkerPool[In, Out]) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// worker 工作协程
func (wp *WorkerPool[In, Out]) worker(id int) {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case job, ok := <-wp.jobQueue:
			if !ok {
				return
			}
			result := wp.process(job)
			select {
			case wp.resultChan <- result:
			case <-wp.ctx.Done():
				return
			}
		}
	}
}

// Submit 提交任务
func (wp *WorkerPool[In, Out]) Submit(job In) error {
	select {
	case wp.jobQueue <- job:
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool closed")
	}
}

// Results 获取结果channel
func (wp *WorkerPool[In, Out]) Results() <-chan Out {
	return wp.resultChan
}

// Stop 停止工作池
func (wp *WorkerPool[In, Out]) Stop() {
	close(wp.jobQueue)
	wp.wg.Wait()
	close(wp.resultChan)
}

// Shutdown 优雅关闭（带超时）
func (wp *WorkerPool[In, Out]) Shutdown(timeout time.Duration) error {
	done := make(chan struct{})

	go func() {
		wp.Stop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		wp.cancel()
		return fmt.Errorf("shutdown timeout")
	}
}

// ========== 4. Channel 关闭模式 ==========

// SafeClose 安全关闭channel（防止重复关闭）
type SafeChannel[T any] struct {
	ch     chan T
	once   sync.Once
	closed bool
	mu     sync.RWMutex
}

func NewSafeChannel[T any](size int) *SafeChannel[T] {
	return &SafeChannel[T]{
		ch: make(chan T, size),
	}
}

func (sc *SafeChannel[T]) Send(v T) bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if sc.closed {
		return false
	}

	sc.ch <- v
	return true
}

func (sc *SafeChannel[T]) Receive() (T, bool) {
	v, ok := <-sc.ch
	return v, ok
}

func (sc *SafeChannel[T]) Close() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.once.Do(func() {
		close(sc.ch)
		sc.closed = true
	})
}

func (sc *SafeChannel[T]) IsClosed() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.closed
}

func (sc *SafeChannel[T]) Chan() <-chan T {
	return sc.ch
}

// ========== 5. 超时和取消模式 ==========

// TimeoutChannel 带超时的channel操作
func TimeoutSend[T any](ch chan T, value T, timeout time.Duration) error {
	select {
	case ch <- value:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("send timeout")
	}
}

func TimeoutReceive[T any](ch <-chan T, timeout time.Duration) (T, error) {
	select {
	case v := <-ch:
		return v, nil
	case <-time.After(timeout):
		var zero T
		return zero, fmt.Errorf("receive timeout")
	}
}

// ContextChannel 带context的channel操作
func ContextSend[T any](ctx context.Context, ch chan T, value T) error {
	select {
	case ch <- value:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func ContextReceive[T any](ctx context.Context, ch <-chan T) (T, error) {
	select {
	case v := <-ch:
		return v, nil
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}

// ========== 6. 生产者-消费者模式 ==========

// Producer 生产者
type Producer[T any] struct {
	output chan<- T
	closed bool
	mu     sync.Mutex
}

func NewProducer[T any](size int) (*Producer[T], <-chan T) {
	ch := make(chan T, size)
	return &Producer[T]{output: ch}, ch
}

func (p *Producer[T]) Produce(value T) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("producer closed")
	}

	p.output <- value
	return nil
}

func (p *Producer[T]) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.closed {
		close(p.output)
		p.closed = true
	}
}

// Consumer 消费者
type Consumer[T any] struct {
	input  <-chan T
	handle func(T) error
}

func NewConsumer[T any](input <-chan T, handle func(T) error) *Consumer[T] {
	return &Consumer[T]{
		input:  input,
		handle: handle,
	}
}

func (c *Consumer[T]) Consume(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case v, ok := <-c.input:
			if !ok {
				return nil // Channel closed normally
			}
			if err := c.handle(v); err != nil {
				return err
			}
		}
	}
}

// ========== 7. Select 多路复用模式 ==========

// Multiplexer 多路复用器
type Multiplexer[T any] struct {
	inputs []<-chan T
}

func NewMultiplexer[T any](inputs ...<-chan T) *Multiplexer[T] {
	return &Multiplexer[T]{inputs: inputs}
}

// Merge 合并所有输入到一个输出
func (m *Multiplexer[T]) Merge() <-chan T {
	return FanIn(m.inputs...)
}

// ========== 8. Done Channel 模式 ==========

// DoneChannel 用于信号通知的channel
type DoneChannel struct {
	done chan struct{}
	once sync.Once
}

func NewDoneChannel() *DoneChannel {
	return &DoneChannel{
		done: make(chan struct{}),
	}
}

func (dc *DoneChannel) Done() <-chan struct{} {
	return dc.done
}

func (dc *DoneChannel) Close() {
	dc.once.Do(func() {
		close(dc.done)
	})
}

func (dc *DoneChannel) IsClosed() bool {
	select {
	case <-dc.done:
		return true
	default:
		return false
	}
}

// ========== 9. 限流模式 ==========

// RateLimiter 基于channel的限流器
type RateLimiter struct {
	tokens chan struct{}
	rate   time.Duration
	stop   chan struct{}
}

func NewRateLimiter(qps int) *RateLimiter {
	rl := &RateLimiter{
		tokens: make(chan struct{}, qps),
		rate:   time.Second / time.Duration(qps),
		stop:   make(chan struct{}),
	}

	// 填充初始令牌
	for i := 0; i < qps; i++ {
		rl.tokens <- struct{}{}
	}

	// 启动令牌生成器
	go rl.refill()

	return rl
}

func (rl *RateLimiter) refill() {
	ticker := time.NewTicker(rl.rate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			select {
			case rl.tokens <- struct{}{}:
			default:
				// 令牌桶满
			}
		case <-rl.stop:
			return
		}
	}
}

func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (rl *RateLimiter) Stop() {
	close(rl.stop)
}

// ========== 10. Or-Done 模式 ==========

// OrDone 将done channel与数据channel组合
func OrDone[T any](done <-chan struct{}, c <-chan T) <-chan T {
	out := make(chan T)
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

// ========== 11. Tee 模式 ==========

// Tee 将一个channel的数据复制到两个输出
func Tee[T any](input <-chan T) (<-chan T, <-chan T) {
	out1 := make(chan T)
	out2 := make(chan T)

	go func() {
		defer close(out1)
		defer close(out2)
		for v := range input {
			// 复制到两个输出
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

// ========== 12. Bridge 模式 ==========

// Bridge 将channel of channels展平为单个channel
func Bridge[T any](chanStream <-chan <-chan T) <-chan T {
	out := make(chan T)

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
