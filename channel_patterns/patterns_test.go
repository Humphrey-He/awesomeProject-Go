package channel_patterns

import (
	"context"
	"testing"
	"time"
)

// ========== Channel 创建测试 ==========

func TestNewUnbufferedChannel(t *testing.T) {
	ch := NewUnbufferedChannel[int]()
	
	done := make(chan bool)
	go func() {
		ch <- 42
		done <- true
	}()
	
	val := <-ch
	<-done
	
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

func TestNewBufferedChannel(t *testing.T) {
	ch := NewBufferedChannel[int](2)
	
	// 不阻塞地发送两个值
	ch <- 1
	ch <- 2
	
	if val := <-ch; val != 1 {
		t.Errorf("Expected 1, got %d", val)
	}
	if val := <-ch; val != 2 {
		t.Errorf("Expected 2, got %d", val)
	}
}

// ========== Pipeline 测试 ==========

func TestPipeline(t *testing.T) {
	input := make(chan int, 3)
	input <- 1
	input <- 2
	input <- 3
	close(input)
	
	p := NewPipeline[int]()
	p.AddStage(func(n int) int { return n * 2 })    // 乘以2
	p.AddStage(func(n int) int { return n + 1 })    // 加1
	
	output := p.Execute(input)
	
	expected := []int{3, 5, 7}
	i := 0
	for val := range output {
		if val != expected[i] {
			t.Errorf("Expected %d, got %d", expected[i], val)
		}
		i++
	}
}

// ========== Fan-Out / Fan-In 测试 ==========

func TestFanIn(t *testing.T) {
	ch1 := make(chan int, 2)
	ch2 := make(chan int, 2)
	
	ch1 <- 1
	ch1 <- 2
	close(ch1)
	
	ch2 <- 3
	ch2 <- 4
	close(ch2)
	
	output := FanIn(ch1, ch2)
	
	results := make(map[int]bool)
	for val := range output {
		results[val] = true
	}
	
	expected := []int{1, 2, 3, 4}
	for _, v := range expected {
		if !results[v] {
			t.Errorf("Expected to receive %d", v)
		}
	}
}

// ========== Worker Pool 测试 ==========

func TestWorkerPool(t *testing.T) {
	process := func(n int) int {
		return n * 2
	}
	
	pool := NewWorkerPool[int, int](3, 10, process)
	pool.Start()
	
	// 提交任务
	for i := 1; i <= 5; i++ {
		if err := pool.Submit(i); err != nil {
			t.Errorf("Submit failed: %v", err)
		}
	}
	
	// 收集结果
	go func() {
		time.Sleep(100 * time.Millisecond)
		pool.Stop()
	}()
	
	results := make(map[int]bool)
	for result := range pool.Results() {
		results[result] = true
	}
	
	// 验证结果
	expected := []int{2, 4, 6, 8, 10}
	for _, v := range expected {
		if !results[v] {
			t.Errorf("Expected result %d", v)
		}
	}
}

func TestWorkerPool_Shutdown(t *testing.T) {
	process := func(n int) int {
		time.Sleep(10 * time.Millisecond)
		return n * 2
	}
	
	pool := NewWorkerPool[int, int](2, 5, process)
	pool.Start()
	
	for i := 1; i <= 3; i++ {
		pool.Submit(i)
	}
	
	err := pool.Shutdown(500 * time.Millisecond)
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

// ========== Safe Channel 测试 ==========

func TestSafeChannel(t *testing.T) {
	sc := NewSafeChannel[int](2)
	
	// 发送数据
	if !sc.Send(1) {
		t.Error("Send should succeed")
	}
	
	// 接收数据
	val, ok := sc.Receive()
	if !ok || val != 1 {
		t.Errorf("Expected 1, got %d", val)
	}
	
	// 关闭channel
	sc.Close()
	
	// 再次关闭不应该panic
	sc.Close()
	
	// 关闭后发送应该失败
	if sc.Send(2) {
		t.Error("Send should fail after close")
	}
	
	if !sc.IsClosed() {
		t.Error("Channel should be closed")
	}
}

// ========== 超时测试 ==========

func TestTimeoutSend(t *testing.T) {
	ch := make(chan int)
	
	// 超时发送
	err := TimeoutSend(ch, 42, 10*time.Millisecond)
	if err == nil {
		t.Error("Expected timeout error")
	}
	
	// 成功发送
	go func() {
		<-ch
	}()
	
	err = TimeoutSend(ch, 42, 100*time.Millisecond)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestTimeoutReceive(t *testing.T) {
	ch := make(chan int)
	
	// 超时接收
	_, err := TimeoutReceive(ch, 10*time.Millisecond)
	if err == nil {
		t.Error("Expected timeout error")
	}
	
	// 成功接收
	go func() {
		ch <- 42
	}()
	
	val, err := TimeoutReceive(ch, 100*time.Millisecond)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}
}

// ========== Context 测试 ==========

func TestContextSend(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	
	ch := make(chan int)
	
	err := ContextSend(ctx, ch, 42)
	if err == nil {
		t.Error("Expected context error")
	}
}

func TestContextReceive(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan int)
	
	// 取消context
	cancel()
	
	_, err := ContextReceive(ctx, ch)
	if err == nil {
		t.Error("Expected context error")
	}
}

// ========== 生产者-消费者测试 ==========

func TestProducerConsumer(t *testing.T) {
	producer, ch := NewProducer[int](10)
	
	// 生产数据
	for i := 1; i <= 5; i++ {
		if err := producer.Produce(i); err != nil {
			t.Errorf("Produce failed: %v", err)
		}
	}
	producer.Close()
	
	// 消费数据
	results := make([]int, 0)
	consumer := NewConsumer(ch, func(v int) error {
		results = append(results, v)
		return nil
	})
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	
	consumer.Consume(ctx)
	
	if len(results) != 5 {
		t.Errorf("Expected 5 results, got %d", len(results))
	}
}

// ========== Done Channel 测试 ==========

func TestDoneChannel(t *testing.T) {
	dc := NewDoneChannel()
	
	if dc.IsClosed() {
		t.Error("Channel should not be closed initially")
	}
	
	dc.Close()
	
	if !dc.IsClosed() {
		t.Error("Channel should be closed")
	}
	
	// 多次关闭不应该panic
	dc.Close()
}

// ========== Rate Limiter 测试 ==========

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10) // 10 QPS
	defer rl.Stop()
	
	// 快速请求10次应该都成功
	success := 0
	for i := 0; i < 10; i++ {
		if rl.Allow() {
			success++
		}
	}
	
	if success != 10 {
		t.Errorf("Expected 10 successful requests, got %d", success)
	}
	
	// 第11次应该失败（令牌用完）
	if rl.Allow() {
		t.Error("Expected rate limit")
	}
	
	// 等待令牌补充
	time.Sleep(150 * time.Millisecond)
	
	// 应该可以再次请求
	if !rl.Allow() {
		t.Error("Expected to allow after refill")
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	rl := NewRateLimiter(5)
	defer rl.Stop()
	
	// 消耗所有令牌
	for i := 0; i < 5; i++ {
		rl.Allow()
	}
	
	// Wait应该阻塞直到有令牌
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	
	err := rl.Wait(ctx)
	if err != nil {
		t.Errorf("Wait failed: %v", err)
	}
}

// ========== Or-Done 测试 ==========

func TestOrDone(t *testing.T) {
	done := make(chan struct{})
	ch := make(chan int, 3)
	
	ch <- 1
	ch <- 2
	ch <- 3
	
	output := OrDone(done, ch)
	
	// 读取两个值
	<-output
	<-output
	
	// 关闭done
	close(done)
	
	// output应该也关闭
	_, ok := <-output
	if ok {
		t.Error("Output should be closed")
	}
}

// ========== Tee 测试 ==========

func TestTee(t *testing.T) {
	input := make(chan int, 3)
	input <- 1
	input <- 2
	input <- 3
	close(input)
	
	out1, out2 := Tee(input)
	
	results1 := make([]int, 0)
	results2 := make([]int, 0)
	
	done := make(chan bool, 2)
	
	go func() {
		for v := range out1 {
			results1 = append(results1, v)
		}
		done <- true
	}()
	
	go func() {
		for v := range out2 {
			results2 = append(results2, v)
		}
		done <- true
	}()
	
	<-done
	<-done
	
	if len(results1) != 3 || len(results2) != 3 {
		t.Errorf("Expected 3 values in each output")
	}
}

// ========== Bridge 测试 ==========

func TestBridge(t *testing.T) {
	chanStream := make(chan (<-chan int), 2)
	
	ch1 := make(chan int, 2)
	ch1 <- 1
	ch1 <- 2
	close(ch1)
	
	ch2 := make(chan int, 2)
	ch2 <- 3
	ch2 <- 4
	close(ch2)
	
	chanStream <- ch1
	chanStream <- ch2
	close(chanStream)
	
	output := Bridge(chanStream)
	
	results := make([]int, 0)
	for v := range output {
		results = append(results, v)
	}
	
	if len(results) != 4 {
		t.Errorf("Expected 4 values, got %d", len(results))
	}
}

// ========== 并发测试 ==========

func TestConcurrentSafeChannel(t *testing.T) {
	sc := NewSafeChannel[int](100)
	
	// 多个goroutine并发发送
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				sc.Send(id*10 + j)
			}
			done <- true
		}(i)
	}
	
	// 等待所有发送完成
	for i := 0; i < 10; i++ {
		<-done
	}
	
	sc.Close()
	
	// 验证没有panic
}

// ========== 基准测试 ==========

func BenchmarkUnbufferedChannel(b *testing.B) {
	ch := make(chan int)
	
	go func() {
		for i := 0; i < b.N; i++ {
			<-ch
		}
	}()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch <- i
	}
}

func BenchmarkBufferedChannel(b *testing.B) {
	ch := make(chan int, 100)
	
	go func() {
		for i := 0; i < b.N; i++ {
			<-ch
		}
	}()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch <- i
	}
}

func BenchmarkWorkerPool(b *testing.B) {
	process := func(n int) int {
		return n * 2
	}
	
	pool := NewWorkerPool[int, int](4, 100, process)
	pool.Start()
	defer pool.Stop()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Submit(i)
	}
}

func BenchmarkSafeChannel_Send(b *testing.B) {
	sc := NewSafeChannel[int](1000)
	
	go func() {
		for i := 0; i < b.N; i++ {
			sc.Receive()
		}
	}()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sc.Send(i)
	}
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	rl := NewRateLimiter(1000000)
	defer rl.Stop()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow()
	}
}

