package token_bucket

import (
	"testing"
	"time"
)

func TestNewTokenBucket(t *testing.T) {
	capacity := int64(10)
	refillRate := int64(5)
	
	tb := NewTokenBucket(capacity, refillRate)
	
	if tb.Capacity() != capacity {
		t.Errorf("Expected capacity %d, got %d", capacity, tb.Capacity())
	}
	
	if tb.RefillRate() != refillRate {
		t.Errorf("Expected refillRate %d, got %d", refillRate, tb.RefillRate())
	}
	
	if tb.AvailableTokens() != capacity {
		t.Errorf("Expected initial tokens %d, got %d", capacity, tb.AvailableTokens())
	}
}

func TestTokenBucket_Allow(t *testing.T) {
	tb := NewTokenBucket(10, 5)
	
	// 初始状态应该有 10 个令牌
	for i := 0; i < 10; i++ {
		if !tb.Allow() {
			t.Errorf("Expected to allow request %d", i+1)
		}
	}
	
	// 第 11 个请求应该被拒绝
	if tb.Allow() {
		t.Error("Expected to deny request when tokens are exhausted")
	}
}

func TestTokenBucket_AllowN(t *testing.T) {
	tb := NewTokenBucket(10, 5)
	
	// 一次性获取 5 个令牌
	if !tb.AllowN(5) {
		t.Error("Expected to allow 5 tokens")
	}
	
	// 还剩 5 个令牌
	if tb.AvailableTokens() != 5 {
		t.Errorf("Expected 5 tokens remaining, got %d", tb.AvailableTokens())
	}
	
	// 再获取 5 个令牌
	if !tb.AllowN(5) {
		t.Error("Expected to allow 5 tokens")
	}
	
	// 尝试获取 1 个令牌应该失败
	if tb.AllowN(1) {
		t.Error("Expected to deny when tokens are exhausted")
	}
}

func TestTokenBucket_Refill(t *testing.T) {
	tb := NewTokenBucket(10, 10) // 每秒生成 10 个令牌
	
	// 消耗所有令牌
	tb.AllowN(10)
	
	if tb.AvailableTokens() != 0 {
		t.Errorf("Expected 0 tokens, got %d", tb.AvailableTokens())
	}
	
	// 等待 0.5 秒，应该生成 5 个令牌
	time.Sleep(500 * time.Millisecond)
	
	available := tb.AvailableTokens()
	if available < 4 || available > 6 {
		t.Errorf("Expected around 5 tokens after 0.5s, got %d", available)
	}
	
	// 等待 1 秒，应该达到容量上限
	time.Sleep(1 * time.Second)
	
	if tb.AvailableTokens() != 10 {
		t.Errorf("Expected 10 tokens (capacity), got %d", tb.AvailableTokens())
	}
}

func TestTokenBucket_RefillDoesNotExceedCapacity(t *testing.T) {
	tb := NewTokenBucket(5, 10)
	
	// 等待足够长的时间
	time.Sleep(1 * time.Second)
	
	// 令牌数不应该超过容量
	if tb.AvailableTokens() > 5 {
		t.Errorf("Expected tokens not to exceed capacity 5, got %d", tb.AvailableTokens())
	}
}

func TestTokenBucket_ConcurrentAccess(t *testing.T) {
	tb := NewTokenBucket(100, 50)
	
	done := make(chan bool)
	numGoroutines := 10
	requestsPerGoroutine := 10
	
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < requestsPerGoroutine; j++ {
				tb.Allow()
			}
			done <- true
		}()
	}
	
	// 等待所有 goroutine 完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	// 验证令牌数一致性（应该没有数据竞争）
	available := tb.AvailableTokens()
	if available < 0 || available > tb.Capacity() {
		t.Errorf("Token count inconsistent: %d", available)
	}
}

func BenchmarkTokenBucket_Allow(b *testing.B) {
	tb := NewTokenBucket(1000000, 1000000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.Allow()
	}
}

func BenchmarkTokenBucket_AllowN(b *testing.B) {
	tb := NewTokenBucket(1000000, 1000000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.AllowN(10)
	}
}

