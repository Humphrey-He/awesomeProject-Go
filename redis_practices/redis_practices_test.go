// redis_practices_test.go - Redis 操作测试

package redis_practices

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// 注意：这些测试需要本地运行 Redis 服务
// 可以使用 Docker 启动: docker run -d --name redis -p 6379:6379 redis:latest

// getTestClient 获取测试客户端
func getTestClient() *redis.Client {
	return NewClient(&RedisConfig{
		Addr: "localhost:6379",
		DB:   15, // 使用 DB 15 进行测试，避免污染其他数据
	})
}

// TestStringOperations 测试字符串操作
func TestStringOperations(t *testing.T) {
	t.Skip("跳过：需要本地 Redis 环境")

	client := getTestClient()
	defer client.Close()

	ctx := context.Background()
	strOps := NewStringOperations(client)

	// 测试 Set 和 Get
	err := strOps.Set(ctx, "test:key", "hello", 10*time.Second)
	if err != nil {
		t.Fatalf("Set 失败: %v", err)
	}

	val, err := strOps.Get(ctx, "test:key")
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if val != "hello" {
		t.Errorf("期望值 'hello'，实际 '%s'", val)
	}

	// 测试 Incr
	client.Del(ctx, "test:counter")
	count, err := strOps.Incr(ctx, "test:counter")
	if err != nil {
		t.Fatalf("Incr 失败: %v", err)
	}
	if count != 1 {
		t.Errorf("期望值 1，实际 %d", count)
	}

	// 测试 SetNX
	client.Del(ctx, "test:lock")
	ok, err := strOps.SetNX(ctx, "test:lock", "locked", 10*time.Second)
	if err != nil {
		t.Fatalf("SetNX 失败: %v", err)
	}
	if !ok {
		t.Error("SetNX 应该返回 true")
	}

	// 清理
	client.Del(ctx, "test:key", "test:counter", "test:lock")
}

// TestHashOperations 测试 Hash 操作
func TestHashOperations(t *testing.T) {
	t.Skip("跳过：需要本地 Redis 环境")

	client := getTestClient()
	defer client.Close()

	ctx := context.Background()
	hashOps := NewHashOperations(client)

	// 清理
	client.Del(ctx, "test:user")

	// 测试 HSet
	_, err := hashOps.HSet(ctx, "test:user", "name", "张三", "age", 25)
	if err != nil {
		t.Fatalf("HSet 失败: %v", err)
	}

	// 测试 HGet
	name, err := hashOps.HGet(ctx, "test:user", "name")
	if err != nil {
		t.Fatalf("HGet 失败: %v", err)
	}
	if name != "张三" {
		t.Errorf("期望值 '张三'，实际 '%s'", name)
	}

	// 测试 HGetAll
	fields, err := hashOps.HGetAll(ctx, "test:user")
	if err != nil {
		t.Fatalf("HGetAll 失败: %v", err)
	}
	if len(fields) != 2 {
		t.Errorf("期望 2 个字段，实际 %d 个", len(fields))
	}

	// 清理
	client.Del(ctx, "test:user")
}

// TestListOperations 测试 List 操作
func TestListOperations(t *testing.T) {
	t.Skip("跳过：需要本地 Redis 环境")

	client := getTestClient()
	defer client.Close()

	ctx := context.Background()
	listOps := NewListOperations(client)

	// 清理
	client.Del(ctx, "test:queue")

	// 测试 RPush 和 LPop
	_, err := listOps.RPush(ctx, "test:queue", "item1", "item2", "item3")
	if err != nil {
		t.Fatalf("RPush 失败: %v", err)
	}

	item, err := listOps.LPop(ctx, "test:queue")
	if err != nil {
		t.Fatalf("LPop 失败: %v", err)
	}
	if item != "item1" {
		t.Errorf("期望值 'item1'，实际 '%s'", item)
	}

	// 测试 LLen
	len, err := listOps.LLen(ctx, "test:queue")
	if err != nil {
		t.Fatalf("LLen 失败: %v", err)
	}
	if len != 2 {
		t.Errorf("期望值 2，实际 %d", len)
	}

	// 清理
	client.Del(ctx, "test:queue")
}

// TestSetOperations 测试 Set 操作
func TestSetOperations(t *testing.T) {
	t.Skip("跳过：需要本地 Redis 环境")

	client := getTestClient()
	defer client.Close()

	ctx := context.Background()
	setOps := NewSetOperations(client)

	// 清理
	client.Del(ctx, "test:set1", "test:set2")

	// 测试 SAdd
	_, err := setOps.SAdd(ctx, "test:set1", "a", "b", "c")
	if err != nil {
		t.Fatalf("SAdd 失败: %v", err)
	}

	// 测试 SMembers
	members, err := setOps.SMembers(ctx, "test:set1")
	if err != nil {
		t.Fatalf("SMembers 失败: %v", err)
	}
	if len(members) != 3 {
		t.Errorf("期望 3 个成员，实际 %d 个", len(members))
	}

	// 测试 SIsMember
	isMember, err := setOps.SIsMember(ctx, "test:set1", "a")
	if err != nil {
		t.Fatalf("SIsMember 失败: %v", err)
	}
	if !isMember {
		t.Error("'a' 应该是集合成员")
	}

	// 清理
	client.Del(ctx, "test:set1", "test:set2")
}

// TestZSetOperations 测试 ZSet 操作
func TestZSetOperations(t *testing.T) {
	t.Skip("跳过：需要本地 Redis 环境")

	client := getTestClient()
	defer client.Close()

	ctx := context.Background()
	zsetOps := NewZSetOperations(client)

	// 清理
	client.Del(ctx, "test:ranking")

	// 测试 ZAdd
	_, err := zsetOps.ZAdd(ctx, "test:ranking",
		redis.Z{Score: 100, Member: "user1"},
		redis.Z{Score: 200, Member: "user2"},
		redis.Z{Score: 150, Member: "user3"},
	)
	if err != nil {
		t.Fatalf("ZAdd 失败: %v", err)
	}

	// 测试 ZRevRange（从大到小）
	top2, err := zsetOps.ZRevRange(ctx, "test:ranking", 0, 1)
	if err != nil {
		t.Fatalf("ZRevRange 失败: %v", err)
	}
	if len(top2) != 2 {
		t.Errorf("期望 2 个成员，实际 %d 个", len(top2))
	}
	if top2[0] != "user2" {
		t.Errorf("第一名应该是 user2，实际是 %s", top2[0])
	}

	// 测试 ZRank
	rank, err := zsetOps.ZRank(ctx, "test:ranking", "user2")
	if err != nil {
		t.Fatalf("ZRank 失败: %v", err)
	}
	if rank != 2 { // 从小到大排第3，索引为2
		t.Errorf("user2 排名应该是 2，实际是 %d", rank)
	}

	// 清理
	client.Del(ctx, "test:ranking")
}

// TestDistributedLock 测试分布式锁
func TestDistributedLock(t *testing.T) {
	t.Skip("跳过：需要本地 Redis 环境")

	client := getTestClient()
	defer client.Close()

	ctx := context.Background()

	// 清理
	client.Del(ctx, "test:lock")

	lock1 := NewDistributedLock(client, "test:lock", 10*time.Second)
	lock2 := NewDistributedLock(client, "test:lock", 10*time.Second)

	// 第一个锁应该成功
	ok1, err := lock1.Lock(ctx)
	if err != nil {
		t.Fatalf("Lock 失败: %v", err)
	}
	if !ok1 {
		t.Error("第一个锁应该成功")
	}

	// 第二个锁应该失败
	ok2, err := lock2.Lock(ctx)
	if err != nil {
		t.Fatalf("Lock 失败: %v", err)
	}
	if ok2 {
		t.Error("第二个锁应该失败")
	}

	// 释放第一个锁
	err = lock1.Unlock(ctx)
	if err != nil {
		t.Fatalf("Unlock 失败: %v", err)
	}

	// 现在第二个锁应该成功
	ok2, err = lock2.Lock(ctx)
	if err != nil {
		t.Fatalf("Lock 失败: %v", err)
	}
	if !ok2 {
		t.Error("释放后第二个锁应该成功")
	}

	// 清理
	lock2.Unlock(ctx)
	client.Del(ctx, "test:lock")
}

// TestCache 测试缓存封装
func TestCache(t *testing.T) {
	t.Skip("跳过：需要本地 Redis 环境")

	client := getTestClient()
	defer client.Close()

	ctx := context.Background()
	cache := NewCache(client, "test")

	// 测试 Set 和 Get
	err := cache.Set(ctx, "name", "test-value", 10*time.Second)
	if err != nil {
		t.Fatalf("Set 失败: %v", err)
	}

	val, err := cache.Get(ctx, "name")
	if err != nil {
		t.Fatalf("Get 失败: %v", err)
	}
	if val != "test-value" {
		t.Errorf("期望值 'test-value'，实际 '%s'", val)
	}

	// 测试 Delete
	err = cache.Delete(ctx, "name")
	if err != nil {
		t.Fatalf("Delete 失败: %v", err)
	}

	_, err = cache.Get(ctx, "name")
	if err != redis.Nil {
		t.Error("删除后应该返回 redis.Nil")
	}

	// 清理
	client.Del(ctx, "test:name")
}

// TestRateLimiter 测试限流器
func TestRateLimiter(t *testing.T) {
	t.Skip("跳过：需要本地 Redis 环境")

	client := getTestClient()
	defer client.Close()

	ctx := context.Background()
	limiter := NewRateLimiter(client)

	// 清理
	client.Del(ctx, "test:rate:limit")

	// 设置限流：5秒内最多3次请求
	for i := 0; i < 3; i++ {
		allowed, err := limiter.Allow(ctx, "test:rate:limit", 3, 5*time.Second)
		if err != nil {
			t.Fatalf("Allow 失败: %v", err)
		}
		if !allowed {
			t.Errorf("第 %d 次请求应该被允许", i+1)
		}
	}

	// 第4次应该被拒绝
	allowed, err := limiter.Allow(ctx, "test:rate:limit", 3, 5*time.Second)
	if err != nil {
		t.Fatalf("Allow 失败: %v", err)
	}
	if allowed {
		t.Error("第4次请求应该被拒绝")
	}

	// 清理
	client.Del(ctx, "test:rate:limit")
}

// TestPipeline 测试管道操作
func TestPipeline(t *testing.T) {
	t.Skip("跳过：需要本地 Redis 环境")

	client := getTestClient()
	defer client.Close()

	ctx := context.Background()
	pipeline := NewPipeline(client)

	// 清理
	client.Del(ctx, "test:pipe1", "test:pipe2")

	// 使用管道批量执行命令
	results, err := pipeline.ExecPipeline(ctx, func(pipe redis.Pipeliner) {
		pipe.Set(ctx, "test:pipe1", "value1", 10*time.Second)
		pipe.Set(ctx, "test:pipe2", "value2", 10*time.Second)
		pipe.Get(ctx, "test:pipe1")
		pipe.Get(ctx, "test:pipe2")
	})
	if err != nil {
		t.Fatalf("ExecPipeline 失败: %v", err)
	}

	// 检查结果
	if len(results) != 4 {
		t.Errorf("期望 4 个结果，实际 %d 个", len(results))
	}

	// 清理
	client.Del(ctx, "test:pipe1", "test:pipe2")
}

// TestPubSub 测试发布订阅
func TestPubSub(t *testing.T) {
	t.Skip("跳过：需要本地 Redis 环境")

	client := getTestClient()
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pubsub := NewPubSub(client)

	// 订阅频道
	sub := pubsub.Subscribe(ctx, "test:channel")
	defer sub.Close()

	// 等待订阅就绪
	_, err := sub.Receive(ctx)
	if err != nil {
		t.Fatalf("订阅失败: %v", err)
	}

	// 在另一个 goroutine 中发布消息
	go func() {
		time.Sleep(100 * time.Millisecond)
		pubsub.Publish(ctx, "test:channel", "hello")
	}()

	// 接收消息
	msg, err := sub.ReceiveMessage(ctx)
	if err != nil {
		t.Fatalf("接收消息失败: %v", err)
	}

	if msg.Payload != "hello" {
		t.Errorf("期望 'hello'，实际 '%s'", msg.Payload)
	}
}

// BenchmarkSet 基准测试 Set 操作
func BenchmarkSet(b *testing.B) {
	b.Skip("跳过：需要本地 Redis 环境")

	client := getTestClient()
	defer client.Close()

	ctx := context.Background()
	strOps := NewStringOperations(client)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench:key:%d", i)
		strOps.Set(ctx, key, "value", 10*time.Second)
	}
}

// BenchmarkGet 基准测试 Get 操作
func BenchmarkGet(b *testing.B) {
	b.Skip("跳过：需要本地 Redis 环境")

	client := getTestClient()
	defer client.Close()

	ctx := context.Background()
	strOps := NewStringOperations(client)

	// 先设置值
	client.Set(ctx, "bench:get:key", "value", 10*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strOps.Get(ctx, "bench:get:key")
	}
}

// BenchmarkPipeline 基准测试管道操作
func BenchmarkPipeline(b *testing.B) {
	b.Skip("跳过：需要本地 Redis 环境")

	client := getTestClient()
	defer client.Close()

	ctx := context.Background()
	pipeline := NewPipeline(client)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline.ExecPipeline(ctx, func(pipe redis.Pipeliner) {
			for j := 0; j < 10; j++ {
				pipe.Set(ctx, fmt.Sprintf("bench:pipe:%d:%d", i, j), "value", 10*time.Second)
			}
		})
	}
}

// ExampleStringOperations 字符串操作示例
func ExampleStringOperations() {
	client := NewClient(nil)
	defer client.Close()

	ctx := context.Background()
	strOps := NewStringOperations(client)

	// 设置值
	strOps.Set(ctx, "user:1:name", "张三", 10*time.Minute)

	// 获取值
	name, _ := strOps.Get(ctx, "user:1:name")
	fmt.Println(name)

	// 自增计数器
	strOps.Incr(ctx, "counter")
}

// ExampleDistributedLock 分布式锁示例
func ExampleDistributedLock() {
	client := NewClient(nil)
	defer client.Close()

	ctx := context.Background()
	lock := NewDistributedLock(client, "resource:lock", 30*time.Second)

	// 尝试获取锁
	ok, _ := lock.Lock(ctx)
	if ok {
		defer lock.Unlock(ctx)
		// 执行业务逻辑
		fmt.Println("获取锁成功，执行业务逻辑")
	}
}

// ExampleCache 缓存示例
func ExampleCache() {
	client := NewClient(nil)
	defer client.Close()

	ctx := context.Background()
	cache := NewCache(client, "app")

	// 设置缓存
	cache.Set(ctx, "config", "{\"key\":\"value\"}", 5*time.Minute)

	// 获取缓存
	val, err := cache.Get(ctx, "config")
	if err == nil {
		fmt.Println(val)
	}
}
