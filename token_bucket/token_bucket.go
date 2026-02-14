package token_bucket

import (
	"sync"
	"time"
)

// TokenBucket 令牌桶结构
type TokenBucket struct {
	capacity       int64      // 桶的容量（最大令牌数）
	tokens         int64      // 当前令牌数
	refillRate     int64      // 令牌生成速率（令牌数/秒）
	lastRefillTime time.Time  // 上次填充时间
	mu             sync.Mutex // 互斥锁
}

// NewTokenBucket 创建一个新的令牌桶
// capacity: 桶的容量
// refillRate: 令牌生成速率（令牌数/秒）
func NewTokenBucket(capacity, refillRate int64) *TokenBucket {
	return &TokenBucket{
		capacity:       capacity,
		tokens:         capacity, // 初始化时桶是满的
		refillRate:     refillRate,
		lastRefillTime: time.Now(),
	}
}

// Allow 尝试获取一个令牌，成功返回 true，失败返回 false
func (tb *TokenBucket) Allow() bool {
	return tb.AllowN(1)
}

// AllowN 尝试获取 n 个令牌，成功返回 true，失败返回 false
func (tb *TokenBucket) AllowN(n int64) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// 先填充令牌
	tb.refill()

	// 检查是否有足够的令牌
	if tb.tokens >= n {
		tb.tokens -= n
		return true
	}
	return false
}

// refill 填充令牌（内部方法，调用前需要持有锁）
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime)

	// 计算应该添加的令牌数
	tokensToAdd := int64(elapsed.Seconds() * float64(tb.refillRate))

	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		// 令牌数不能超过容量
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefillTime = now
	}
}

// AvailableTokens 返回当前可用的令牌数
func (tb *TokenBucket) AvailableTokens() int64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()
	return tb.tokens
}

// Capacity 返回桶的容量
func (tb *TokenBucket) Capacity() int64 {
	return tb.capacity
}

// RefillRate 返回令牌生成速率
func (tb *TokenBucket) RefillRate() int64 {
	return tb.refillRate
}
