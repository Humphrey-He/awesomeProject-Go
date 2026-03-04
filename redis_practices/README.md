# Redis 实践项目

基于 [go-redis/v9](https://github.com/redis/go-redis) 库实现的 Redis 操作示例与最佳实践。

## 项目概述

本项目演示了 Redis 在日常开发中的常用操作，涵盖所有基础数据结构、分布式锁、限流器、缓存模式等核心功能，并提供了完整的最佳实践指南。

## 核心功能

### 1. 基础数据结构

#### String 字符串操作
```go
client := NewClient(nil)
strOps := NewStringOperations(client)

// 设置值
strOps.Set(ctx, "user:name", "张三", 10*time.Minute)

// 获取值
name, _ := strOps.Get(ctx, "user:name")

// 自增计数器
strOps.Incr(ctx, "counter")
strOps.IncrBy(ctx, "counter", 10)

// 批量操作
strOps.MSet(ctx, "k1", "v1", "k2", "v2")
vals, _ := strOps.MGet(ctx, "k1", "k2")
```

#### Hash 哈希表操作
```go
hashOps := NewHashOperations(client)

// 设置字段
hashOps.HSet(ctx, "user:1", "name", "张三", "age", 25)

// 获取字段
name, _ := hashOps.HGet(ctx, "user:1", "name")

// 获取所有字段
fields, _ := hashOps.HGetAll(ctx, "user:1")

// 字段自增
hashOps.HIncrBy(ctx, "user:1", "score", 10)
```

#### List 列表操作
```go
listOps := NewListOperations(client)

// 队列操作
listOps.RPush(ctx, "queue", "task1", "task2")
task, _ := listOps.LPop(ctx, "queue")

// 阻塞弹出
result, _ := listOps.BLPop(ctx, 5*time.Second, "queue")

// 获取范围
items, _ := listOps.LRange(ctx, "queue", 0, -1)
```

#### Set 集合操作
```go
setOps := NewSetOperations(client)

// 添加元素
setOps.SAdd(ctx, "tags", "go", "redis", "kafka")

// 判断成员
isMember, _ := setOps.SIsMember(ctx, "tags", "go")

// 集合运算
intersection, _ := setOps.SInter(ctx, "set1", "set2")
union, _ := setOps.SUnion(ctx, "set1", "set2")
diff, _ := setOps.SDiff(ctx, "set1", "set2")
```

#### Sorted Set 有序集合操作
```go
zsetOps := NewZSetOperations(client)

// 添加元素
zsetOps.ZAdd(ctx, "ranking",
    redis.Z{Score: 100, Member: "user1"},
    redis.Z{Score: 200, Member: "user2"},
)

// 获取排行榜
top10, _ := zsetOps.ZRevRange(ctx, "ranking", 0, 9)

// 获取排名
rank, _ := zsetOps.ZRank(ctx, "ranking", "user1")
```

### 2. 分布式锁

```go
lock := NewDistributedLock(client, "resource:lock", 30*time.Second)

// 获取锁
ok, err := lock.Lock(ctx)
if ok {
    defer lock.Unlock(ctx)
    // 执行业务逻辑
}

// 延长锁过期时间
lock.Extend(ctx, 10*time.Second)
```

**特点**:
- 基于 SETNX 实现
- Lua 脚本保证原子性释放
- 支持锁续期

### 3. 缓存模式

```go
cache := NewCache(client, "app")

// 简单缓存
cache.Set(ctx, "config", value, 5*time.Minute)
val, _ := cache.Get(ctx, "config")

// 对象缓存
user := &User{ID: 1, Name: "张三"}
cache.SetObject(ctx, "user:1", user, 10*time.Minute)
cache.GetObject(ctx, "user:1", &user)

// 缓存穿透保护
val, _ := cache.GetOrSet(ctx, "data", 5*time.Minute, func() (interface{}, error) {
    return fetchDataFromDB()
})
```

### 4. 限流器

#### 滑动窗口限流
```go
limiter := NewRateLimiter(client)

// 5秒内最多3次请求
allowed, _ := limiter.Allow(ctx, "api:limit", 3, 5*time.Second)
if !allowed {
    return errors.New("请求过于频繁")
}
```

#### 令牌桶限流
```go
// 每秒生成10个令牌，桶容量100
allowed, _ := limiter.AllowTokenBucket(ctx, "api:limit", 10, 100)
```

### 5. 发布订阅

```go
pubsub := NewPubSub(client)

// 订阅频道
sub := pubsub.Subscribe(ctx, "notifications")
for msg := range sub.Channel() {
    fmt.Println("收到消息:", msg.Payload)
}

// 发布消息
pubsub.Publish(ctx, "notifications", "hello")
```

### 6. 管道与事务

#### 管道操作
```go
pipeline := NewPipeline(client)

// 批量执行命令（减少网络往返）
results, _ := pipeline.ExecPipeline(ctx, func(pipe redis.Pipeliner) {
    pipe.Set(ctx, "k1", "v1", 0)
    pipe.Set(ctx, "k2", "v2", 0)
    pipe.Get(ctx, "k1")
})
```

#### 事务操作
```go
tx := NewTransaction(client)

// 乐观锁事务
tx.ExecTx(ctx, []string{"counter"}, func(tx *redis.Tx) error {
    n, _ := tx.Get(ctx, "counter").Int64()
    _, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
        pipe.Set(ctx, "counter", n+1, 0)
        return nil
    })
    return err
})
```

## 配置说明

### RedisConfig

| 参数 | 默认值 | 说明 |
|------|--------|------|
| Addr | localhost:6379 | Redis 地址 |
| Password | - | 密码 |
| DB | 0 | 数据库索引 |
| PoolSize | 10 | 连接池大小 |
| MinIdleConns | 5 | 最小空闲连接 |
| MaxRetries | 3 | 最大重试次数 |
| DialTimeout | 5s | 连接超时 |
| ReadTimeout | 3s | 读取超时 |
| WriteTimeout | 3s | 写入超时 |

## 最佳实践

### 1. 连接管理

#### 连接池配置
```go
config := &RedisConfig{
    Addr:         "localhost:6379",
    PoolSize:     100,   // 高并发场景
    MinIdleConns: 20,    // 预热连接
    MaxRetries:   3,
}
client := NewClient(config)
```

#### 集群模式
```go
client := NewClusterClient([]string{
    "node1:6379",
    "node2:6379",
    "node3:6379",
}, "password")
```

### 2. 键命名规范

```go
// 推荐格式: 业务:模块:ID
const (
    KeyUser       = "user:info:%d"      // user:info:123
    KeyUserToken  = "user:token:%d"     // user:token:123
    KeyOrder      = "order:detail:%d"   // order:detail:456
    KeyCache      = "cache:data:%s"     // cache:data:abc
)

func buildKey(pattern string, args ...interface{}) string {
    return fmt.Sprintf(pattern, args...)
}
```

### 3. 过期时间策略

```go
// 热点数据 - 较长过期
cache.Set(ctx, "hot:data", value, 30*time.Minute)

// 会话数据 - 中等过期
cache.Set(ctx, "session:token", token, 2*time.Hour)

// 验证码 - 短过期
cache.Set(ctx, "sms:code", code, 5*time.Minute)

// 分布式锁 - 业务超时时间
lock := NewDistributedLock(client, "lock", 30*time.Second)
```

### 4. 防止缓存穿透

```go
func GetWithCache(ctx context.Context, key string) (*Data, error) {
    // 1. 查缓存
    val, err := cache.Get(ctx, key)
    if err == nil {
        return parseData(val), nil
    }
    
    // 2. 缓存不存在，查数据库
    data, err := queryDB(ctx, key)
    if err != nil {
        // 数据库也没有，缓存空值防止穿透
        cache.Set(ctx, key, "", 5*time.Minute)
        return nil, err
    }
    
    // 3. 写入缓存
    cache.Set(ctx, key, data, 10*time.Minute)
    return data, nil
}
```

### 5. 防止缓存雪崩

```go
func SetWithRandomExpire(ctx context.Context, key string, value interface{}) {
    // 基础过期时间 + 随机偏移
    baseExpire := 10 * time.Minute
    randomOffset := time.Duration(rand.Intn(120)) * time.Second
    cache.Set(ctx, key, value, baseExpire+randomOffset)
}
```

### 6. 优雅关闭

```go
func main() {
    client := NewClient(nil)
    defer client.Close()
    
    // 使用 context 控制超时
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    
    // 执行操作
    client.Get(ctx, "key")
}
```

## 使用场景

### 场景1: 会话存储
```go
type SessionManager struct {
    cache *Cache
}

func (s *SessionManager) Create(userID int64) (string, error) {
    token := generateToken()
    session := &Session{UserID: userID, CreatedAt: time.Now()}
    s.cache.SetObject(ctx, "session:"+token, session, 24*time.Hour)
    return token, nil
}
```

### 场景2: 排行榜
```go
type Leaderboard struct {
    zsetOps *ZSetOperations
}

func (l *Leaderboard) UpdateScore(userID string, score float64) {
    l.zsetOps.ZAdd(ctx, "game:ranking", redis.Z{
        Score:  score,
        Member: userID,
    })
}

func (l *Leaderboard) GetTopN(n int64) []string {
    top, _ := l.zsetOps.ZRevRange(ctx, "game:ranking", 0, n-1)
    return top
}
```

### 场景3: 消息队列
```go
type MessageQueue struct {
    listOps *ListOperations
}

func (q *MessageQueue) Push(message string) {
    q.listOps.RPush(ctx, "message:queue", message)
}

func (q *MessageQueue) Pop(timeout time.Duration) (string, error) {
    result, err := q.listOps.BRPop(ctx, timeout, "message:queue")
    if err != nil {
        return "", err
    }
    return result[1], nil
}
```

### 场景4: 限流器
```go
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := "rate:" + c.ClientIP()
        allowed, _ := limiter.Allow(ctx, key, 100, time.Minute)
        if !allowed {
            c.AbortWithStatusJSON(429, gin.H{"error": "请求过于频繁"})
            return
        }
        c.Next()
    }
}
```

## Docker 快速启动

```bash
# 启动 Redis
docker run -d --name redis \
    -p 6379:6379 \
    redis:latest

# 启动 Redis 带密码
docker run -d --name redis \
    -p 6379:6379 \
    redis:latest --requirepass yourpassword
```

## 测试

```bash
# 运行测试（需要本地 Redis）
go test -v

# 基准测试
go test -bench=.

# 测试特定功能
go test -v -run TestDistributedLock
```

## 性能优化建议

### 1. 使用管道减少网络往返
```go
// 不推荐：多次网络请求
client.Set(ctx, "k1", "v1", 0)
client.Set(ctx, "k2", "v2", 0)

// 推荐：管道批量操作
pipe := client.Pipeline()
pipe.Set(ctx, "k1", "v1", 0)
pipe.Set(ctx, "k2", "v2", 0)
pipe.Exec(ctx)
```

### 2. 合理设置连接池
```go
// 高并发场景
config.PoolSize = 100
config.MinIdleConns = 20

// 低并发场景
config.PoolSize = 10
config.MinIdleConns = 5
```

### 3. 使用 SCAN 替代 KEYS
```go
// 不推荐：阻塞式
keys, _ := client.Keys(ctx, "user:*").Result()

// 推荐：非阻塞式
iter := client.Scan(ctx, 0, "user:*", 100).Iterator()
for iter.Next(ctx) {
    key := iter.Val()
    // 处理 key
}
```

## 依赖

- `github.com/redis/go-redis/v9` - Redis Go 客户端

## 参考资料

- [go-redis 官方文档](https://redis.uptrace.dev/)
- [Redis 官方文档](https://redis.io/documentation)
- [Redis 命令参考](https://redis.io/commands)
