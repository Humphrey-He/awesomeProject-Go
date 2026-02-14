package leaky_bucket

import (
	"sync"
	"time"
)

// LeakyBucket 漏桶结构
type LeakyBucket struct {
	capacity   int64         // 桶的容量（最大请求数）
	water      int64         // 当前水量（待处理的请求数）
	leakRate   int64         // 漏水速率（请求数/秒）
	lastLeakTime time.Time   // 上次漏水时间
	mu         sync.Mutex    // 互斥锁
}

// NewLeakyBucket 创建一个新的漏桶
// capacity: 桶的容量
// leakRate: 漏水速率（请求数/秒）
func NewLeakyBucket(capacity, leakRate int64) *LeakyBucket {
	return &LeakyBucket{
		capacity:     capacity,
		water:        0, // 初始时桶是空的
		leakRate:     leakRate,
		lastLeakTime: time.Now(),
	}
}

// Allow 尝试添加一个请求到桶中，成功返回 true，失败返回 false
func (lb *LeakyBucket) Allow() bool {
	return lb.AllowN(1)
}

// AllowN 尝试添加 n 个请求到桶中，成功返回 true，失败返回 false
func (lb *LeakyBucket) AllowN(n int64) bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// 先漏水（处理之前的请求）
	lb.leak()

	// 检查是否有足够的空间
	if lb.water+n <= lb.capacity {
		lb.water += n
		return true
	}
	return false
}

// leak 漏水（内部方法，调用前需要持有锁）
func (lb *LeakyBucket) leak() {
	now := time.Now()
	elapsed := now.Sub(lb.lastLeakTime)

	// 计算应该漏掉的水量
	leaked := int64(elapsed.Seconds() * float64(lb.leakRate))
	
	if leaked > 0 {
		lb.water -= leaked
		// 水量不能为负数
		if lb.water < 0 {
			lb.water = 0
		}
		lb.lastLeakTime = now
	}
}

// CurrentWater 返回当前桶中的水量
func (lb *LeakyBucket) CurrentWater() int64 {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	lb.leak()
	return lb.water
}

// AvailableSpace 返回桶中的可用空间
func (lb *LeakyBucket) AvailableSpace() int64 {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	lb.leak()
	return lb.capacity - lb.water
}

// Capacity 返回桶的容量
func (lb *LeakyBucket) Capacity() int64 {
	return lb.capacity
}

// LeakRate 返回漏水速率
func (lb *LeakyBucket) LeakRate() int64 {
	return lb.leakRate
}

// WaitTime 返回当前请求需要等待的时间（如果桶满了）
func (lb *LeakyBucket) WaitTime() time.Duration {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	lb.leak()
	
	if lb.water >= lb.capacity {
		// 计算需要等待的时间
		overflow := lb.water - lb.capacity + 1
		waitSeconds := float64(overflow) / float64(lb.leakRate)
		return time.Duration(waitSeconds * float64(time.Second))
	}
	return 0
}

