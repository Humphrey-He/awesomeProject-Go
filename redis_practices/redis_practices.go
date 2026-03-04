// redis_practices.go - Redis 操作示例与最佳实践
// 使用 go-redis/v9 库实现 Redis 常用操作

package redis_practices

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// ==================== 基础客户端 ====================

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr         string        // Redis 地址，如 "localhost:6379"
	Password     string        // 密码
	DB           int           // 数据库索引
	PoolSize     int           // 连接池大小
	MinIdleConns int           // 最小空闲连接数
	MaxRetries   int           // 最大重试次数
	DialTimeout  time.Duration // 连接超时
	ReadTimeout  time.Duration // 读取超时
	WriteTimeout time.Duration // 写入超时
}

// DefaultRedisConfig 返回默认配置
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

// NewClient 创建 Redis 客户端
func NewClient(cfg *RedisConfig) *redis.Client {
	if cfg == nil {
		cfg = DefaultRedisConfig()
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	return client
}

// NewClusterClient 创建 Redis 集群客户端
func NewClusterClient(addrs []string, password string) *redis.ClusterClient {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        addrs,
		Password:     password,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	return client
}

// ==================== String 操作 ====================

// StringOperations 字符串操作
type StringOperations struct {
	client *redis.Client
}

// NewStringOperations 创建字符串操作实例
func NewStringOperations(client *redis.Client) *StringOperations {
	return &StringOperations{client: client}
}

// Set 设置字符串值
func (s *StringOperations) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return s.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取字符串值
func (s *StringOperations) Get(ctx context.Context, key string) (string, error) {
	return s.client.Get(ctx, key).Result()
}

// SetNX 仅在 key 不存在时设置（分布式锁基础）
func (s *StringOperations) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return s.client.SetNX(ctx, key, value, expiration).Result()
}

// SetEX 设置值并指定过期时间
func (s *StringOperations) SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return s.client.SetEx(ctx, key, value, expiration).Err()
}

// Incr 自增
func (s *StringOperations) Incr(ctx context.Context, key string) (int64, error) {
	return s.client.Incr(ctx, key).Result()
}

// IncrBy 按指定值自增
func (s *StringOperations) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return s.client.IncrBy(ctx, key, value).Result()
}

// Decr 自减
func (s *StringOperations) Decr(ctx context.Context, key string) (int64, error) {
	return s.client.Decr(ctx, key).Result()
}

// MSet 批量设置
func (s *StringOperations) MSet(ctx context.Context, values ...interface{}) error {
	return s.client.MSet(ctx, values...).Err()
}

// MGet 批量获取
func (s *StringOperations) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	return s.client.MGet(ctx, keys...).Result()
}

// ==================== Hash 操作 ====================

// HashOperations Hash 操作
type HashOperations struct {
	client *redis.Client
}

// NewHashOperations 创建 Hash 操作实例
func NewHashOperations(client *redis.Client) *HashOperations {
	return &HashOperations{client: client}
}

// HSet 设置 Hash 字段
func (h *HashOperations) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return h.client.HSet(ctx, key, values...).Result()
}

// HGet 获取 Hash 字段
func (h *HashOperations) HGet(ctx context.Context, key, field string) (string, error) {
	return h.client.HGet(ctx, key, field).Result()
}

// HGetAll 获取所有 Hash 字段
func (h *HashOperations) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return h.client.HGetAll(ctx, key).Result()
}

// HMSet 批量设置 Hash 字段
func (h *HashOperations) HMSet(ctx context.Context, key string, values map[string]interface{}) error {
	return h.client.HMSet(ctx, key, values).Err()
}

// HMGet 批量获取 Hash 字段
func (h *HashOperations) HMGet(ctx context.Context, key string, fields ...string) ([]interface{}, error) {
	return h.client.HMGet(ctx, key, fields...).Result()
}

// HIncrBy Hash 字段自增
func (h *HashOperations) HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) {
	return h.client.HIncrBy(ctx, key, field, incr).Result()
}

// HDel 删除 Hash 字段
func (h *HashOperations) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	return h.client.HDel(ctx, key, fields...).Result()
}

// HExists 判断字段是否存在
func (h *HashOperations) HExists(ctx context.Context, key, field string) (bool, error) {
	return h.client.HExists(ctx, key, field).Result()
}

// ==================== List 操作 ====================

// ListOperations List 操作
type ListOperations struct {
	client *redis.Client
}

// NewListOperations 创建 List 操作实例
func NewListOperations(client *redis.Client) *ListOperations {
	return &ListOperations{client: client}
}

// LPush 从左侧推入元素
func (l *ListOperations) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return l.client.LPush(ctx, key, values...).Result()
}

// RPush 从右侧推入元素
func (l *ListOperations) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return l.client.RPush(ctx, key, values...).Result()
}

// LPop 从左侧弹出元素
func (l *ListOperations) LPop(ctx context.Context, key string) (string, error) {
	return l.client.LPop(ctx, key).Result()
}

// RPop 从右侧弹出元素
func (l *ListOperations) RPop(ctx context.Context, key string) (string, error) {
	return l.client.RPop(ctx, key).Result()
}

// BLPop 阻塞式左侧弹出
func (l *ListOperations) BLPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	return l.client.BLPop(ctx, timeout, keys...).Result()
}

// BRPop 阻塞式右侧弹出
func (l *ListOperations) BRPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	return l.client.BRPop(ctx, timeout, keys...).Result()
}

// LLen 获取列表长度
func (l *ListOperations) LLen(ctx context.Context, key string) (int64, error) {
	return l.client.LLen(ctx, key).Result()
}

// LRange 获取列表范围
func (l *ListOperations) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return l.client.LRange(ctx, key, start, stop).Result()
}

// LTrim 修剪列表
func (l *ListOperations) LTrim(ctx context.Context, key string, start, stop int64) error {
	return l.client.LTrim(ctx, key, start, stop).Err()
}

// ==================== Set 操作 ====================

// SetOperations Set 操作
type SetOperations struct {
	client *redis.Client
}

// NewSetOperations 创建 Set 操作实例
func NewSetOperations(client *redis.Client) *SetOperations {
	return &SetOperations{client: client}
}

// SAdd 添加元素到集合
func (s *SetOperations) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return s.client.SAdd(ctx, key, members...).Result()
}

// SRem 从集合移除元素
func (s *SetOperations) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return s.client.SRem(ctx, key, members...).Result()
}

// SMembers 获取集合所有元素
func (s *SetOperations) SMembers(ctx context.Context, key string) ([]string, error) {
	return s.client.SMembers(ctx, key).Result()
}

// SIsMember 判断元素是否在集合中
func (s *SetOperations) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return s.client.SIsMember(ctx, key, member).Result()
}

// SCard 获取集合元素数量
func (s *SetOperations) SCard(ctx context.Context, key string) (int64, error) {
	return s.client.SCard(ctx, key).Result()
}

// SPop 随机弹出元素
func (s *SetOperations) SPop(ctx context.Context, key string) (string, error) {
	return s.client.SPop(ctx, key).Result()
}

// SInter 获取集合交集
func (s *SetOperations) SInter(ctx context.Context, keys ...string) ([]string, error) {
	return s.client.SInter(ctx, keys...).Result()
}

// SUnion 获取集合并集
func (s *SetOperations) SUnion(ctx context.Context, keys ...string) ([]string, error) {
	return s.client.SUnion(ctx, keys...).Result()
}

// SDiff 获取集合差集
func (s *SetOperations) SDiff(ctx context.Context, keys ...string) ([]string, error) {
	return s.client.SDiff(ctx, keys...).Result()
}

// ==================== Sorted Set 操作 ====================

// ZSetOperations Sorted Set 操作
type ZSetOperations struct {
	client *redis.Client
}

// NewZSetOperations 创建 ZSet 操作实例
func NewZSetOperations(client *redis.Client) *ZSetOperations {
	return &ZSetOperations{client: client}
}

// ZAdd 添加元素到有序集合
func (z *ZSetOperations) ZAdd(ctx context.Context, key string, members ...redis.Z) (int64, error) {
	return z.client.ZAdd(ctx, key, members...).Result()
}

// ZRem 从有序集合移除元素
func (z *ZSetOperations) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return z.client.ZRem(ctx, key, members...).Result()
}

// ZRange 按分数范围获取元素（从小到大）
func (z *ZSetOperations) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return z.client.ZRange(ctx, key, start, stop).Result()
}

// ZRevRange 按分数范围获取元素（从大到小）
func (z *ZSetOperations) ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return z.client.ZRevRange(ctx, key, start, stop).Result()
}

// ZRangeByScore 按分数范围获取元素
func (z *ZSetOperations) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return z.client.ZRangeByScore(ctx, key, opt).Result()
}

// ZRank 获取元素排名（从小到大）
func (z *ZSetOperations) ZRank(ctx context.Context, key, member string) (int64, error) {
	return z.client.ZRank(ctx, key, member).Result()
}

// ZRevRank 获取元素排名（从大到小）
func (z *ZSetOperations) ZRevRank(ctx context.Context, key, member string) (int64, error) {
	return z.client.ZRevRank(ctx, key, member).Result()
}

// ZScore 获取元素分数
func (z *ZSetOperations) ZScore(ctx context.Context, key, member string) (float64, error) {
	return z.client.ZScore(ctx, key, member).Result()
}

// ZCard 获取有序集合元素数量
func (z *ZSetOperations) ZCard(ctx context.Context, key string) (int64, error) {
	return z.client.ZCard(ctx, key).Result()
}

// ZIncrBy 增加元素分数
func (z *ZSetOperations) ZIncrBy(ctx context.Context, key string, increment float64, member string) (float64, error) {
	return z.client.ZIncrBy(ctx, key, increment, member).Result()
}

// ==================== 分布式锁 ====================

// DistributedLock 分布式锁
type DistributedLock struct {
	client     *redis.Client
	key        string
	value      string
	expiration time.Duration
	mu         sync.Mutex
}

// NewDistributedLock 创建分布式锁
func NewDistributedLock(client *redis.Client, key string, expiration time.Duration) *DistributedLock {
	return &DistributedLock{
		client:     client,
		key:        key,
		value:      generateLockValue(),
		expiration: expiration,
	}
}

// generateLockValue 生成锁值（唯一标识）
func generateLockValue() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}

// Lock 获取锁
func (l *DistributedLock) Lock(ctx context.Context) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	ok, err := l.client.SetNX(ctx, l.key, l.value, l.expiration).Result()
	if err != nil {
		return false, err
	}
	return ok, nil
}

// Unlock 释放锁（使用 Lua 脚本保证原子性）
func (l *DistributedLock) Unlock(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	_, err := l.client.Eval(ctx, script, []string{l.key}, l.value).Result()
	return err
}

// Extend 延长锁过期时间
func (l *DistributedLock) Extend(ctx context.Context, expiration time.Duration) (bool, error) {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`
	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value, expiration.Seconds()).Result()
	if err != nil {
		return false, err
	}
	return result.(int64) == 1, nil
}

// ==================== 缓存模式 ====================

// Cache 缓存封装
type Cache struct {
	client *redis.Client
	prefix string
}

// NewCache 创建缓存实例
func NewCache(client *redis.Client, prefix string) *Cache {
	return &Cache{
		client: client,
		prefix: prefix,
	}
}

// keyWithPrefix 添加前缀
func (c *Cache) keyWithPrefix(key string) string {
	if c.prefix == "" {
		return key
	}
	return c.prefix + ":" + key
}

// Get 获取缓存
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, c.keyWithPrefix(key)).Result()
}

// Set 设置缓存
func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, c.keyWithPrefix(key), value, expiration).Err()
}

// Delete 删除缓存
func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = c.keyWithPrefix(key)
	}
	return c.client.Del(ctx, prefixedKeys...).Err()
}

// SetObject 设置对象（JSON序列化）
func (c *Cache) SetObject(ctx context.Context, key string, obj interface{}, expiration time.Duration) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("序列化对象失败: %w", err)
	}
	return c.Set(ctx, key, data, expiration)
}

// GetObject 获取对象（JSON反序列化）
func (c *Cache) GetObject(ctx context.Context, key string, obj interface{}) error {
	data, err := c.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), obj)
}

// GetOrSet 缓存模式：不存在则设置
func (c *Cache) GetOrSet(ctx context.Context, key string, expiration time.Duration, fn func() (interface{}, error)) (string, error) {
	// 先尝试获取
	val, err := c.Get(ctx, key)
	if err == nil {
		return val, nil
	}

	// 不存在，执行函数获取值
	newVal, err := fn()
	if err != nil {
		return "", err
	}

	// 设置缓存
	if err := c.Set(ctx, key, newVal, expiration); err != nil {
		log.Printf("设置缓存失败: %v", err)
	}

	return fmt.Sprintf("%v", newVal), nil
}

// ==================== 限流器 ====================

// RateLimiter 基于 Redis 的限流器
type RateLimiter struct {
	client *redis.Client
}

// NewRateLimiter 创建限流器
func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

// Allow 滑动窗口限流
func (r *RateLimiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	now := time.Now().UnixMilli()
	windowStart := now - window.Milliseconds()

	pipe := r.client.Pipeline()

	// 移除窗口外的请求记录
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// 获取当前窗口内的请求数
	pipe.ZCard(ctx, key)

	// 添加当前请求
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})

	// 设置 key 过期时间
	pipe.Expire(ctx, key, window)

	results, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	// 获取当前请求数（ZCard 的结果）
	currentCount := results[1].(*redis.IntCmd).Val()

	return currentCount <= limit, nil
}

// AllowTokenBucket 令牌桶限流
func (r *RateLimiter) AllowTokenBucket(ctx context.Context, key string, rate float64, capacity int64) (bool, error) {
	script := `
		local key = KEYS[1]
		local rate = tonumber(ARGV[1])
		local capacity = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])

		local tokens = redis.call("get", key)
		if tokens == false then
			tokens = capacity
		else
			tokens = tonumber(tokens)
		end

		local last_updated = redis.call("get", key .. ":last_updated")
		if last_updated == false then
			last_updated = 0
		else
			last_updated = tonumber(last_updated)
		end

		local delta = math.max(0, now - last_updated)
		tokens = math.min(capacity, tokens + delta * rate)

		local allowed = 0
		if tokens >= 1 then
			tokens = tokens - 1
			allowed = 1
		end

		redis.call("set", key, tokens)
		redis.call("set", key .. ":last_updated", now)
		redis.call("expire", key, 60)
		redis.call("expire", key .. ":last_updated", 60)

		return allowed
	`

	now := float64(time.Now().UnixMilli()) / 1000
	result, err := r.client.Eval(ctx, script, []string{key}, rate, capacity, now).Result()
	if err != nil {
		return false, err
	}

	return result.(int64) == 1, nil
}

// ==================== 发布订阅 ====================

// PubSub 发布订阅
type PubSub struct {
	client *redis.Client
}

// NewPubSub 创建发布订阅实例
func NewPubSub(client *redis.Client) *PubSub {
	return &PubSub{client: client}
}

// Publish 发布消息
func (p *PubSub) Publish(ctx context.Context, channel string, message interface{}) error {
	return p.client.Publish(ctx, channel, message).Err()
}

// Subscribe 订阅频道
func (p *PubSub) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return p.client.Subscribe(ctx, channels...)
}

// PSubscribe 按模式订阅
func (p *PubSub) PSubscribe(ctx context.Context, patterns ...string) *redis.PubSub {
	return p.client.PSubscribe(ctx, patterns...)
}

// ==================== 管道与事务 ====================

// Pipeline 管道操作
type Pipeline struct {
	client *redis.Client
}

// NewPipeline 创建管道
func NewPipeline(client *redis.Client) *Pipeline {
	return &Pipeline{client: client}
}

// ExecPipeline 执行管道命令
func (p *Pipeline) ExecPipeline(ctx context.Context, fn func(pipe redis.Pipeliner)) ([]redis.Cmder, error) {
	pipe := p.client.Pipeline()
	fn(pipe)
	return pipe.Exec(ctx)
}

// Transaction 事务操作
type Transaction struct {
	client *redis.Client
}

// NewTransaction 创建事务
func NewTransaction(client *redis.Client) *Transaction {
	return &Transaction{client: client}
}

// ExecTx 执行事务（使用 Watch 实现乐观锁）
func (t *Transaction) ExecTx(ctx context.Context, keys []string, fn func(tx *redis.Tx) error) error {
	return t.client.Watch(ctx, func(tx *redis.Tx) error {
		return fn(tx)
	}, keys...)
}

// ==================== 工具函数 ====================

// Ping 测试连接
func Ping(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}

// FlushDB 清空当前数据库
func FlushDB(ctx context.Context, client *redis.Client) error {
	return client.FlushDB(ctx).Err()
}

// Info 获取服务器信息
func Info(ctx context.Context, client *redis.Client, section string) (string, error) {
	return client.Info(ctx, section).Result()
}

// Scan 扫描键
func Scan(ctx context.Context, client *redis.Client, cursor uint64, match string, count int64) ([]string, uint64, error) {
	return client.Scan(ctx, cursor, match, count).Result()
}

// Expire 设置过期时间
func Expire(ctx context.Context, client *redis.Client, key string, expiration time.Duration) (bool, error) {
	return client.Expire(ctx, key, expiration).Result()
}

// TTL 获取剩余过期时间
func TTL(ctx context.Context, client *redis.Client, key string) (time.Duration, error) {
	return client.TTL(ctx, key).Result()
}

// Exists 检查键是否存在
func Exists(ctx context.Context, client *redis.Client, keys ...string) (int64, error) {
	return client.Exists(ctx, keys...).Result()
}

// Del 删除键
func Del(ctx context.Context, client *redis.Client, keys ...string) (int64, error) {
	return client.Del(ctx, keys...).Result()
}
