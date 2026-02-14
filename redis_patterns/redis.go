package redis_patterns

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ========== 1. Redis 客户端创建和连接管理 ==========

// RedisConfig Redis配置
type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DefaultRedisConfig 默认配置
func DefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// NewRedisClient 创建Redis客户端（单机模式）
func NewRedisClient(config *RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})
}

// NewRedisClusterClient 创建Redis集群客户端
func NewRedisClusterClient(addrs []string, password string) *redis.ClusterClient {
	return redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        addrs,
		Password:     password,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
	})
}

// HealthCheck 健康检查
func HealthCheck(ctx context.Context, client redis.Cmdable) error {
	return client.Ping(ctx).Err()
}

// ========== 2. 基本操作封装 ==========

// RedisCache Redis缓存封装
type RedisCache struct {
	client redis.Cmdable
}

// NewRedisCache 创建缓存实例
func NewRedisCache(client redis.Cmdable) *RedisCache {
	return &RedisCache{client: client}
}

// Set 设置缓存（带过期时间）
func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	return rc.client.Set(ctx, key, data, expiration).Err()
}

// Get 获取缓存
func (rc *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := rc.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// GetWithDefault 获取缓存，不存在时返回默认值
func (rc *RedisCache) GetWithDefault(ctx context.Context, key string, dest interface{}, defaultValue interface{}) error {
	err := rc.Get(ctx, key, dest)
	if err == redis.Nil {
		// 缓存不存在，使用默认值
		*dest.(*interface{}) = defaultValue
		return nil
	}
	return err
}

// Delete 删除缓存
func (rc *RedisCache) Delete(ctx context.Context, keys ...string) error {
	return rc.client.Del(ctx, keys...).Err()
}

// Exists 检查key是否存在
func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := rc.client.Exists(ctx, key).Result()
	return n > 0, err
}

// Expire 设置过期时间
func (rc *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return rc.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取剩余过期时间
func (rc *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return rc.client.TTL(ctx, key).Result()
}

// ========== 3. 高级操作模式 ==========

// GetOrSet 获取缓存，不存在时通过loader加载并设置
func (rc *RedisCache) GetOrSet(ctx context.Context, key string, dest interface{}, expiration time.Duration, loader func() (interface{}, error)) error {
	// 先尝试从缓存获取
	err := rc.Get(ctx, key, dest)
	if err == nil {
		return nil
	}

	if err != redis.Nil {
		return err
	}

	// 缓存不存在，通过loader加载
	value, err := loader()
	if err != nil {
		return fmt.Errorf("loader error: %w", err)
	}

	// 设置到缓存
	if err := rc.Set(ctx, key, value, expiration); err != nil {
		return fmt.Errorf("set cache error: %w", err)
	}

	// 将加载的值赋给dest
	data, _ := json.Marshal(value)
	return json.Unmarshal(data, dest)
}

// MGet 批量获取
func (rc *RedisCache) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	return rc.client.MGet(ctx, keys...).Result()
}

// MSet 批量设置
func (rc *RedisCache) MSet(ctx context.Context, pairs ...interface{}) error {
	return rc.client.MSet(ctx, pairs...).Err()
}

// ========== 4. 分布式锁 ==========

// DistributedLock 分布式锁
type DistributedLock struct {
	client redis.Cmdable
	key    string
	value  string
	ttl    time.Duration
}

// NewDistributedLock 创建分布式锁
func NewDistributedLock(client redis.Cmdable, key string, ttl time.Duration) *DistributedLock {
	return &DistributedLock{
		client: client,
		key:    key,
		value:  fmt.Sprintf("%d", time.Now().UnixNano()),
		ttl:    ttl,
	}
}

// Lock 获取锁
func (dl *DistributedLock) Lock(ctx context.Context) (bool, error) {
	return dl.client.SetNX(ctx, dl.key, dl.value, dl.ttl).Result()
}

// Unlock 释放锁（使用Lua脚本保证原子性）
func (dl *DistributedLock) Unlock(ctx context.Context) error {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	return dl.client.Eval(ctx, script, []string{dl.key}, dl.value).Err()
}

// TryLockWithRetry 尝试获取锁（带重试）
func (dl *DistributedLock) TryLockWithRetry(ctx context.Context, maxRetries int, retryInterval time.Duration) error {
	for i := 0; i < maxRetries; i++ {
		locked, err := dl.Lock(ctx)
		if err != nil {
			return err
		}

		if locked {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryInterval):
			// 继续重试
		}
	}

	return fmt.Errorf("failed to acquire lock after %d retries", maxRetries)
}

// ========== 5. Pipeline 批量操作 ==========

// PipelineExample Pipeline批量操作示例
func PipelineExample(ctx context.Context, client redis.Cmdable) error {
	pipe := client.Pipeline()

	// 添加多个命令
	pipe.Set(ctx, "key1", "value1", 0)
	pipe.Set(ctx, "key2", "value2", 0)
	pipe.Get(ctx, "key1")

	// 执行所有命令
	_, err := pipe.Exec(ctx)
	return err
}

// ========== 6. Pub/Sub 发布订阅 ==========

// Publisher 发布者
type Publisher struct {
	client redis.Cmdable
}

// NewPublisher 创建发布者
func NewPublisher(client redis.Cmdable) *Publisher {
	return &Publisher{client: client}
}

// Publish 发布消息
func (p *Publisher) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return p.client.Publish(ctx, channel, data).Err()
}

// Subscriber 订阅者
type Subscriber struct {
	client redis.Cmdable
	pubsub *redis.PubSub
}

// NewSubscriber 创建订阅者
// 注意：实际使用时需要redis.Client类型
func NewSubscriber(client redis.Cmdable, channels ...string) *Subscriber {
	// 由于Cmdable接口不包含Subscribe方法，这里需要类型断言
	// 实际使用时应传入*redis.Client
	rdb, ok := client.(*redis.Client)
	if !ok {
		// 这里简化处理，实际应该返回错误
		return &Subscriber{client: client}
	}
	pubsub := rdb.Subscribe(context.Background(), channels...)
	return &Subscriber{
		client: client,
		pubsub: pubsub,
	}
}

// Receive 接收消息
func (s *Subscriber) Receive(ctx context.Context, handler func(channel string, message []byte) error) error {
	ch := s.pubsub.Channel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-ch:
			if err := handler(msg.Channel, []byte(msg.Payload)); err != nil {
				return err
			}
		}
	}
}

// Close 关闭订阅
func (s *Subscriber) Close() error {
	return s.pubsub.Close()
}

// ========== 7. 计数器和限流 ==========

// Counter 计数器
type Counter struct {
	client redis.Cmdable
	key    string
}

// NewCounter 创建计数器
func NewCounter(client redis.Cmdable, key string) *Counter {
	return &Counter{
		client: client,
		key:    key,
	}
}

// Incr 增加计数
func (c *Counter) Incr(ctx context.Context) (int64, error) {
	return c.client.Incr(ctx, c.key).Result()
}

// IncrBy 增加指定值
func (c *Counter) IncrBy(ctx context.Context, value int64) (int64, error) {
	return c.client.IncrBy(ctx, c.key, value).Result()
}

// Decr 减少计数
func (c *Counter) Decr(ctx context.Context) (int64, error) {
	return c.client.Decr(ctx, c.key).Result()
}

// Get 获取当前值
func (c *Counter) Get(ctx context.Context) (int64, error) {
	return c.client.Get(ctx, c.key).Int64()
}

// Reset 重置计数器
func (c *Counter) Reset(ctx context.Context) error {
	return c.client.Del(ctx, c.key).Err()
}

// SlidingWindowLimiter 滑动窗口限流器
type SlidingWindowLimiter struct {
	client redis.Cmdable
	key    string
	limit  int64
	window time.Duration
}

// NewSlidingWindowLimiter 创建滑动窗口限流器
func NewSlidingWindowLimiter(client redis.Cmdable, key string, limit int64, window time.Duration) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		client: client,
		key:    key,
		limit:  limit,
		window: window,
	}
}

// Allow 检查是否允许请求
func (swl *SlidingWindowLimiter) Allow(ctx context.Context) (bool, error) {
	now := time.Now().UnixNano()
	windowStart := now - int64(swl.window)

	script := `
		redis.call('zremrangebyscore', KEYS[1], 0, ARGV[1])
		local count = redis.call('zcard', KEYS[1])
		if count < tonumber(ARGV[3]) then
			redis.call('zadd', KEYS[1], ARGV[2], ARGV[2])
			redis.call('expire', KEYS[1], ARGV[4])
			return 1
		else
			return 0
		end
	`

	result, err := swl.client.Eval(ctx, script, []string{swl.key}, windowStart, now, swl.limit, int(swl.window.Seconds())).Int()
	return result == 1, err
}

// ========== 8. 缓存预热和更新策略 ==========

// CacheWarmer 缓存预热器
type CacheWarmer struct {
	cache *RedisCache
}

// NewCacheWarmer 创建缓存预热器
func NewCacheWarmer(cache *RedisCache) *CacheWarmer {
	return &CacheWarmer{cache: cache}
}

// WarmUp 预热缓存
func (cw *CacheWarmer) WarmUp(ctx context.Context, keys []string, loader func(string) (interface{}, error), expiration time.Duration) error {
	for _, key := range keys {
		value, err := loader(key)
		if err != nil {
			return fmt.Errorf("load key %s error: %w", key, err)
		}

		if err := cw.cache.Set(ctx, key, value, expiration); err != nil {
			return fmt.Errorf("set key %s error: %w", key, err)
		}
	}

	return nil
}

// ========== 9. 错误处理和重试 ==========

// RetryableOperation 可重试的Redis操作
func RetryableOperation(ctx context.Context, maxRetries int, operation func() error) error {
	var err error

	for i := 0; i < maxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}

		// 判断是否需要重试
		if !shouldRetry(err) {
			return err
		}

		// 指数退避
		backoff := time.Duration(1<<uint(i)) * 100 * time.Millisecond
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
}

// shouldRetry 判断错误是否应该重试
func shouldRetry(err error) bool {
	// 网络错误、超时等应该重试
	// 这里简化处理
	return err != redis.Nil
}

// ========== 10. 连接池管理 ==========

// PoolStats 获取连接池统计信息
func PoolStats(client *redis.Client) *redis.PoolStats {
	return client.PoolStats()
}

// MonitorPool 监控连接池
func MonitorPool(ctx context.Context, client *redis.Client, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := client.PoolStats()
			fmt.Printf("Redis Pool Stats: Hits=%d Misses=%d Timeouts=%d TotalConns=%d IdleConns=%d StaleConns=%d\n",
				stats.Hits, stats.Misses, stats.Timeouts, stats.TotalConns, stats.IdleConns, stats.StaleConns)
		}
	}
}
