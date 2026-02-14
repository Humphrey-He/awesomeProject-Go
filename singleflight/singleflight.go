package singleflight

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Singleflight 防止缓存击穿，确保同一时刻只有一个请求在执行
// 其他相同的请求会等待并共享结果
type Singleflight struct {
	mu    sync.Mutex
	calls map[string]*call
}

// call 表示一个正在进行的或已完成的请求
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
	
	// 记录调用次数
	dups int
}

// Result 包装返回结果
type Result struct {
	Val    interface{}
	Err    error
	Shared bool // 是否是共享的结果
}

// New 创建Singleflight实例
func New() *Singleflight {
	return &Singleflight{
		calls: make(map[string]*call),
	}
}

// Do 执行函数，相同key的并发调用会共享结果
func (g *Singleflight) Do(key string, fn func() (interface{}, error)) (interface{}, error, bool) {
	g.mu.Lock()
	
	// 如果已有相同key的请求在执行，等待结果
	if c, ok := g.calls[key]; ok {
		c.dups++
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err, true
	}
	
	// 创建新的call
	c := &call{}
	c.wg.Add(1)
	g.calls[key] = c
	g.mu.Unlock()
	
	// 执行函数
	c.val, c.err = fn()
	c.wg.Done()
	
	// 清理
	g.mu.Lock()
	delete(g.calls, key)
	g.mu.Unlock()
	
	return c.val, c.err, c.dups > 0
}

// DoChan 异步执行，返回channel
func (g *Singleflight) DoChan(key string, fn func() (interface{}, error)) <-chan Result {
	ch := make(chan Result, 1)
	
	go func() {
		val, err, shared := g.Do(key, fn)
		ch <- Result{Val: val, Err: err, Shared: shared}
	}()
	
	return ch
}

// DoContext 带context的执行
func (g *Singleflight) DoContext(ctx context.Context, key string, fn func(context.Context) (interface{}, error)) (interface{}, error, bool) {
	g.mu.Lock()
	
	if c, ok := g.calls[key]; ok {
		c.dups++
		g.mu.Unlock()
		
		// 等待结果或context取消
		done := make(chan struct{})
		go func() {
			c.wg.Wait()
			close(done)
		}()
		
		select {
		case <-ctx.Done():
			return nil, ctx.Err(), false
		case <-done:
			return c.val, c.err, true
		}
	}
	
	c := &call{}
	c.wg.Add(1)
	g.calls[key] = c
	g.mu.Unlock()
	
	// 在goroutine中执行
	done := make(chan struct{})
	go func() {
		c.val, c.err = fn(ctx)
		c.wg.Done()
		close(done)
	}()
	
	// 等待完成或取消
	select {
	case <-ctx.Done():
		// Context取消，但不中断已启动的请求
		return nil, ctx.Err(), false
	case <-done:
		g.mu.Lock()
		delete(g.calls, key)
		g.mu.Unlock()
		return c.val, c.err, c.dups > 0
	}
}

// Forget 主动忘记某个key，使下次调用重新执行
func (g *Singleflight) Forget(key string) {
	g.mu.Lock()
	delete(g.calls, key)
	g.mu.Unlock()
}

// ========== 缓存集成示例 ==========

// CacheLoader 缓存加载器（集成singleflight防止缓存击穿）
type CacheLoader struct {
	sf    *Singleflight
	cache sync.Map // 简单的内存缓存
	ttl   time.Duration
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Value      interface{}
	ExpireTime time.Time
}

// NewCacheLoader 创建缓存加载器
func NewCacheLoader(ttl time.Duration) *CacheLoader {
	return &CacheLoader{
		sf:  New(),
		ttl: ttl,
	}
}

// Load 加载数据（带singleflight保护）
func (cl *CacheLoader) Load(key string, loader func() (interface{}, error)) (interface{}, error) {
	// 先查缓存
	if entry, ok := cl.cache.Load(key); ok {
		ce := entry.(*CacheEntry)
		if time.Now().Before(ce.ExpireTime) {
			return ce.Value, nil
		}
		// 缓存过期，删除
		cl.cache.Delete(key)
	}
	
	// 使用singleflight加载数据
	val, err, _ := cl.sf.Do(key, func() (interface{}, error) {
		value, err := loader()
		if err != nil {
			return nil, err
		}
		
		// 写入缓存
		cl.cache.Store(key, &CacheEntry{
			Value:      value,
			ExpireTime: time.Now().Add(cl.ttl),
		})
		
		return value, nil
	})
	
	return val, err
}

// Invalidate 使缓存失效
func (cl *CacheLoader) Invalidate(key string) {
	cl.cache.Delete(key)
	cl.sf.Forget(key)
}

// ========== 高级用法：超时控制 ==========

var ErrTimeout = errors.New("singleflight: timeout")

// DoWithTimeout 带超时的执行
func (g *Singleflight) DoWithTimeout(key string, timeout time.Duration, fn func() (interface{}, error)) (interface{}, error, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	return g.DoContext(ctx, key, func(ctx context.Context) (interface{}, error) {
		done := make(chan struct {
			val interface{}
			err error
		}, 1)
		
		go func() {
			val, err := fn()
			done <- struct {
				val interface{}
				err error
			}{val, err}
		}()
		
		select {
		case <-ctx.Done():
			return nil, ErrTimeout
		case result := <-done:
			return result.val, result.err
		}
	})
}

// ========== 统计信息 ==========

// Stats 统计信息
type Stats struct {
	Calls      int // 总调用次数
	InFlight   int // 正在执行的请求数
	Duplicates int // 重复请求数（被合并的）
}

// GetStats 获取统计信息
func (g *Singleflight) GetStats() Stats {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	stats := Stats{
		InFlight: len(g.calls),
	}
	
	for _, c := range g.calls {
		stats.Duplicates += c.dups
	}
	
	return stats
}

// ========== 使用示例 ==========

// Example 展示基本用法
func Example() {
	sf := New()
	
	// 模拟多个并发请求
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			val, err, shared := sf.Do("key", func() (interface{}, error) {
				// 模拟耗时操作（如数据库查询）
				time.Sleep(100 * time.Millisecond)
				return "result", nil
			})
			
			if shared {
				// 这个请求等待了其他请求的结果
				_ = val
				_ = err
			}
		}(i)
	}
	
	wg.Wait()
}

// ExampleWithCache 缓存集成示例
func ExampleWithCache() {
	loader := NewCacheLoader(5 * time.Minute)
	
	// 多次加载相同的key
	for i := 0; i < 100; i++ {
		val, err := loader.Load("user:123", func() (interface{}, error) {
			// 只有第一次会真正执行
			// 模拟数据库查询
			time.Sleep(100 * time.Millisecond)
			return map[string]interface{}{
				"id":   123,
				"name": "Alice",
			}, nil
		})
		
		_ = val
		_ = err
	}
}

