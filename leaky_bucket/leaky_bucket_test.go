package leaky_bucket

import (
	"testing"
	"time"
)

func TestNewLeakyBucket(t *testing.T) {
	capacity := int64(10)
	leakRate := int64(5)
	
	lb := NewLeakyBucket(capacity, leakRate)
	
	if lb.Capacity() != capacity {
		t.Errorf("Expected capacity %d, got %d", capacity, lb.Capacity())
	}
	
	if lb.LeakRate() != leakRate {
		t.Errorf("Expected leakRate %d, got %d", leakRate, lb.LeakRate())
	}
	
	if lb.CurrentWater() != 0 {
		t.Errorf("Expected initial water 0, got %d", lb.CurrentWater())
	}
}

func TestLeakyBucket_Allow(t *testing.T) {
	lb := NewLeakyBucket(10, 5)
	
	// 初始状态应该能接受 10 个请求
	for i := 0; i < 10; i++ {
		if !lb.Allow() {
			t.Errorf("Expected to allow request %d", i+1)
		}
	}
	
	// 第 11 个请求应该被拒绝
	if lb.Allow() {
		t.Error("Expected to deny request when bucket is full")
	}
}

func TestLeakyBucket_AllowN(t *testing.T) {
	lb := NewLeakyBucket(10, 5)
	
	// 一次性添加 5 个请求
	if !lb.AllowN(5) {
		t.Error("Expected to allow 5 requests")
	}
	
	// 当前水量应该是 5
	if lb.CurrentWater() != 5 {
		t.Errorf("Expected 5 water, got %d", lb.CurrentWater())
	}
	
	// 再添加 5 个请求
	if !lb.AllowN(5) {
		t.Error("Expected to allow 5 requests")
	}
	
	// 尝试再添加 1 个请求应该失败
	if lb.AllowN(1) {
		t.Error("Expected to deny when bucket is full")
	}
}

func TestLeakyBucket_Leak(t *testing.T) {
	lb := NewLeakyBucket(10, 10) // 每秒漏出 10 个请求
	
	// 填满桶
	lb.AllowN(10)
	
	if lb.CurrentWater() != 10 {
		t.Errorf("Expected 10 water, got %d", lb.CurrentWater())
	}
	
	// 等待 0.5 秒，应该漏出 5 个请求
	time.Sleep(500 * time.Millisecond)
	
	water := lb.CurrentWater()
	if water < 4 || water > 6 {
		t.Errorf("Expected around 5 water after 0.5s, got %d", water)
	}
	
	// 等待 1 秒，桶应该变空
	time.Sleep(1 * time.Second)
	
	if lb.CurrentWater() != 0 {
		t.Errorf("Expected 0 water, got %d", lb.CurrentWater())
	}
}

func TestLeakyBucket_LeakDoesNotGoBelowZero(t *testing.T) {
	lb := NewLeakyBucket(5, 10)
	
	// 不添加任何请求，等待一段时间
	time.Sleep(500 * time.Millisecond)
	
	// 水量不应该是负数
	if lb.CurrentWater() < 0 {
		t.Errorf("Water should not be negative, got %d", lb.CurrentWater())
	}
}

func TestLeakyBucket_AvailableSpace(t *testing.T) {
	lb := NewLeakyBucket(10, 5)
	
	// 初始可用空间应该是容量
	if lb.AvailableSpace() != 10 {
		t.Errorf("Expected available space 10, got %d", lb.AvailableSpace())
	}
	
	// 添加 3 个请求
	lb.AllowN(3)
	
	// 可用空间应该是 7
	if lb.AvailableSpace() != 7 {
		t.Errorf("Expected available space 7, got %d", lb.AvailableSpace())
	}
}

func TestLeakyBucket_WaitTime(t *testing.T) {
	lb := NewLeakyBucket(10, 10) // 每秒漏出 10 个请求
	
	// 桶未满时，等待时间应该是 0
	if lb.WaitTime() != 0 {
		t.Errorf("Expected wait time 0, got %v", lb.WaitTime())
	}
	
	// 填满桶
	lb.AllowN(10)
	
	// 桶满时，应该有等待时间
	waitTime := lb.WaitTime()
	if waitTime == 0 {
		t.Error("Expected non-zero wait time when bucket is full")
	}
	
	// 等待时间应该在合理范围内（大约 0.1 秒）
	if waitTime < 50*time.Millisecond || waitTime > 200*time.Millisecond {
		t.Errorf("Expected wait time around 100ms, got %v", waitTime)
	}
}

func TestLeakyBucket_ConcurrentAccess(t *testing.T) {
	lb := NewLeakyBucket(100, 50)
	
	done := make(chan bool)
	numGoroutines := 10
	requestsPerGoroutine := 10
	
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < requestsPerGoroutine; j++ {
				lb.Allow()
			}
			done <- true
		}()
	}
	
	// 等待所有 goroutine 完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	// 验证水量一致性（应该没有数据竞争）
	water := lb.CurrentWater()
	if water < 0 || water > lb.Capacity() {
		t.Errorf("Water count inconsistent: %d", water)
	}
}

func TestLeakyBucket_SmoothRateLimiting(t *testing.T) {
	lb := NewLeakyBucket(5, 10) // 容量 5，每秒漏出 10 个
	
	allowed := 0
	denied := 0
	
	// 快速发送 20 个请求
	for i := 0; i < 20; i++ {
		if lb.Allow() {
			allowed++
		} else {
			denied++
		}
		time.Sleep(10 * time.Millisecond)
	}
	
	// 前 5 个应该通过，后面的大部分应该被拒绝
	if allowed < 5 || allowed > 8 {
		t.Errorf("Expected around 5-8 requests allowed, got %d", allowed)
	}
}

func BenchmarkLeakyBucket_Allow(b *testing.B) {
	lb := NewLeakyBucket(1000000, 1000000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb.Allow()
	}
}

func BenchmarkLeakyBucket_AllowN(b *testing.B) {
	lb := NewLeakyBucket(1000000, 1000000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb.AllowN(10)
	}
}

func BenchmarkLeakyBucket_CurrentWater(b *testing.B) {
	lb := NewLeakyBucket(1000000, 1000000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb.CurrentWater()
	}
}

