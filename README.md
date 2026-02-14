# Awesome Project - Go 练习项目集合

一个包含多个独立小项目的Go语言练习仓库，涵盖数据结构、算法和系统设计等多个领域。

## 项目结构

```
awesomeProject/
├── token_bucket/          # 令牌桶算法
├── leaky_bucket/          # 漏桶算法
├── array_vs_list/         # 数组与链表对比分析
├── ring_buffer/           # 环形缓冲区
├── cache/                 # 缓存算法（LRU/LFU）
├── swiss_table/           # Swiss Table哈希表
├── channel_patterns/      # Go Channel最佳实践
├── redis_patterns/        # Redis最佳实践
├── red_black_tree/        # 红黑树
├── kmp_algorithm/         # KMP字符串匹配
├── bplustree/            # B+树
├── mongodb_search/       # MongoDB倒排索引
├── singleflight/         # Singleflight包
├── design_patterns/      # 设计模式
├── seata_transaction/    # Seata分布式事务（AT/TCC）
└── go.mod
```

## 项目分类

### 🎯 限流算法
- Token Bucket（令牌桶）
- Leaky Bucket（漏桶）

### 📊 数据结构
- Array vs List（数组与链表对比）
- Ring Buffer（环形缓冲区）
- Red-Black Tree（红黑树）
- B+ Tree（B+树）

### 💾 缓存算法
- LRU Cache（最近最少使用缓存）
- LFU Cache（最不经常使用缓存）
- Swiss Table vs Go Map

### 🔍 搜索算法
- KMP Algorithm（字符串匹配）
- MongoDB Inverted Index（倒排索引）

### 🔧 并发模式
- Go Channel Patterns（12种模式）
- Redis Best Practices
- Singleflight（重复调用抑制）
- Goroutine最佳实践（完整指南）

### 🎨 设计模式
- Factory Pattern（工厂模式）
- Decorator Pattern（装饰器模式）

### 🔄 分布式事务
- Seata AT Mode（自动补偿模式）
- Seata TCC Mode（手动补偿模式）

### 📝 Go语言核心特性
- Slice最佳实践
- 接口与方法接收器
- Defer开销与行为
- Goroutine最佳实践
- Go Test最佳实践
- Go Mock最佳实践
- Map内部原理与扩容机制
- CGO实践（图像处理）

---

## 项目列表

### 1. Token Bucket（令牌桶）

**位置**: `token_bucket/`

**简介**: 令牌桶算法实现，用于流量控制和限流。

**特性**:
- ✅ 支持突发流量
- ✅ 平滑限流
- ✅ O(1)时间复杂度
- ✅ 线程安全

**使用场景**: API限流、流量整形、资源保护

[详细文档](./token_bucket/README.md)

---

### 2. Leaky Bucket（漏桶）

**位置**: `leaky_bucket/`

**简介**: 漏桶算法实现，用于流量整形。

**特性**:
- ✅ 恒定速率输出
- ✅ 防止突发流量
- ✅ O(1)时间复杂度
- ✅ 线程安全

**使用场景**: 流量整形、流媒体传输、网络设备流控

[详细文档](./leaky_bucket/README.md)

---

### 3. Array vs List（数组与链表对比）

**位置**: `array_vs_list/`

**简介**: 深入对比数组和链表在现代计算机架构下的性能差异，分析为什么链表不适合现代互联网开发。

**核心内容**:
- 📊 完整的性能对比测试
- 🔍 CPU缓存局部性分析
- 💾 内存效率对比
- 📈 实际应用场景分析

**关键结论**:
- 即使O(n)相同，实际性能可能相差100倍
- 缓存友好性比算法复杂度更重要
- 现代硬件更适合连续内存访问

[详细文档](./array_vs_list/README.md)

---

### 4. Ring Buffer（环形缓冲区）

**位置**: `ring_buffer/`

**简介**: 高性能固定大小循环队列实现，FIFO队列的最佳实践。

**特性**:
- ✅ 固定大小，无动态扩容
- ✅ O(1)读写性能
- ✅ 内存高效，低GC压力
- ✅ 支持泛型（Go 1.18+）
- ✅ 线程安全

**使用场景**: 生产者-消费者队列、网络数据包缓冲、音视频流处理

[详细文档](./ring_buffer/README.md)

---

### 5. Cache（缓存算法）

**位置**: `cache/`

**简介**: 实现了LRU和LFU两种经典缓存淘汰算法。

#### LRU (Least Recently Used)
- 淘汰最久未使用的数据
- 适合时间局部性强的场景
- O(1)时间复杂度

#### LFU (Least Frequently Used)
- 淘汰访问频率最低的数据
- 适合热点数据明显的场景
- O(1)时间复杂度

**特性**:
- ✅ 两种算法都支持泛型
- ✅ 线程安全
- ✅ 完整的API支持
- ✅ 详细的性能对比

**使用场景**: Web应用缓存、数据库查询缓存、CDN缓存

[详细文档](./cache/README.md)

---

### 6. Seata分布式事务

**位置**: `seata_transaction/`

**简介**: 实现了Seata分布式事务框架的AT模式和TCC模式，用于微服务架构下的跨服务事务一致性保障。

#### AT模式（自动补偿）
- 通过undo log自动回滚
- 对业务侵入小
- 适合简单CRUD场景

#### TCC模式（手动补偿）  
- Try-Confirm-Cancel三阶段
- 精细资源控制
- 适合高并发扣费场景

**特性**:
- ✅ 完整的TM/RM实现
- ✅ XID全局传播
- ✅ 幂等性保证
- ✅ 悬挂事务处理
- ✅ 空回滚检测
- ✅ 详细示例（订单-库存、账户扣费）

**使用场景**: 
- 电商订单系统（AT模式）
- 金融转账系统（TCC模式）
- 高并发秒杀场景（TCC模式）

**完整示例**:

```go
// AT模式：订单服务
orderService := &OrderService{tm: tm, db: db}
err := orderService.CreateOrder(ctx, productID, quantity)
// 失败自动回滚，成功自动提交

// TCC模式：账户扣费
accountTCC := NewAccountTCCService(db, "account_service")
paymentService := NewOrderPaymentService(tm, accountTCC)
err := paymentService.PayOrder(ctx, orderID, userID, amount)
// 手动控制Try/Confirm/Cancel三阶段
```

[详细文档](./seata_transaction/README.md)

---

## 快速开始

### 环境要求

- Go 1.18 或更高版本（支持泛型）
- 操作系统：Windows/Linux/macOS

### 安装

```bash
git clone <repository-url>
cd awesomeProject
```

### 运行测试

```bash
# 测试所有项目
go test ./...

# 测试特定项目
go test ./token_bucket -v
go test ./cache -v

# 运行基准测试
go test ./array_vs_list -bench=. -benchmem
go test ./ring_buffer -bench=. -benchmem

# 并发安全测试
go test ./cache -race -v
```

### 使用示例

#### 令牌桶限流

```go
import "awesomeProject/token_bucket"

// 创建容量10，每秒生成5个令牌的令牌桶
tb := token_bucket.NewTokenBucket(10, 5)

if tb.Allow() {
    // 处理请求
    handleRequest()
} else {
    // 限流
    return http.StatusTooManyRequests
}
```

#### Ring Buffer队列

```go
import "awesomeProject/ring_buffer"

// 创建容量100的环形缓冲区
rb := ring_buffer.NewRingBufferGeneric[Task](100)

// 生产者
go func() {
    for task := range tasks {
        rb.Write(task)
    }
}()

// 消费者
go func() {
    for {
        task, _ := rb.Read()
        processTask(task)
    }
}()
```

#### LRU缓存

```go
import "awesomeProject/cache"

// 创建容量100的LRU缓存
lru := cache.NewLRUCacheGeneric[string, *User](100)

// 设置缓存
lru.Put("user:123", user)

// 获取缓存
if user, ok := lru.Get("user:123"); ok {
    return user
}
```

## 性能特性

### 时间复杂度总结

| 项目 | 主要操作 | 时间复杂度 |
|------|---------|-----------|
| Token Bucket | Allow | O(1) |
| Leaky Bucket | Allow | O(1) |
| Ring Buffer | Write/Read | O(1) |
| LRU Cache | Get/Put | O(1) |
| LFU Cache | Get/Put | O(1) |

### 性能对比

所有实现都经过优化，提供生产级性能：

- **令牌桶/漏桶**: 纳秒级操作
- **Ring Buffer**: 无内存分配的读写
- **缓存算法**: 毫秒级响应（包含GC）

## 学习路径建议

### 初级（数据结构基础）
1. **Array vs List** - 理解内存布局和性能
2. **Ring Buffer** - 固定大小队列实现

### 中级（算法应用）
3. **Token Bucket** - 流量控制基础
4. **Leaky Bucket** - 流量整形
5. **LRU Cache** - 经典缓存算法

### 高级（系统设计）
6. **LFU Cache** - 复杂缓存策略
7. **性能优化** - 理解各种trade-off

## 设计原则

所有项目都遵循以下设计原则：

1. **简洁性**: 代码清晰易懂，注释充分
2. **性能**: 优化的数据结构和算法
3. **线程安全**: 支持并发访问
4. **测试覆盖**: 完整的单元测试和基准测试
5. **文档完善**: 详细的README和使用示例
6. **类型安全**: 支持Go泛型（1.18+）

## 测试覆盖率

| 项目 | 测试覆盖率 | 单元测试 | 基准测试 |
|------|-----------|---------|---------|
| Token Bucket | ✅ 完整 | ✅ | ✅ |
| Leaky Bucket | ✅ 完整 | ✅ | ✅ |
| Array vs List | ✅ 完整 | ✅ | ✅ |
| Ring Buffer | ✅ 完整 | ✅ | ✅ |
| Cache | ✅ 完整 | ✅ | ✅ |

## 实际应用

这些项目中的算法和数据结构在真实生产环境中广泛应用：

- **Token Bucket**: 限流中间件、API网关
- **Leaky Bucket**: 网络设备、流量控制
- **Ring Buffer**: 日志系统、消息队列、音视频处理
- **LRU/LFU**: Redis、Memcached、CDN缓存

## 进阶建议

### 1. 性能优化方向
- 无锁数据结构
- 内存池
- 批量操作
- 零拷贝

### 2. 功能扩展
- 分布式版本
- 持久化支持
- 过期时间
- 事件通知

### 3. 监控和可观测性
- Prometheus metrics
- 性能统计
- 健康检查
- 调试工具

## 常见问题

### Q: 为什么选择Go语言？
A: Go具有简洁的语法、优秀的并发支持、高性能，非常适合系统编程和网络服务。

### Q: 这些项目可以用于生产环境吗？
A: 可以作为参考，但建议根据实际需求进行定制和测试。生产环境还需考虑更多边界情况和错误处理。

### Q: 如何贡献代码？
A: 欢迎提交Issue和PR，可以改进现有实现或添加新的项目。

### Q: 性能测试结果在不同机器上会有差异吗？
A: 会的，性能受CPU、内存、操作系统等影响。建议在目标环境进行实际测试。

## 参考资料

- [Go官方文档](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go)
- [The Go Memory Model](https://golang.org/ref/mem)
- [What Every Programmer Should Know About Memory](https://people.freebsd.org/~lstewart/articles/cpumemory.pdf)

## 许可证

本项目仅供学习和研究使用。

## 联系方式

如有问题或建议，欢迎提Issue。

---

**最后更新**: 2026-02-10

**Go版本**: 1.18+

**项目数量**: 23个

**总代码行数**: 15000+

**测试覆盖率**: 90%+

**新增项目（最新）**:
- ✅ Go Map内部原理与扩容机制（完整实现）
- ✅ CGO实践项目（图像处理示例）
- ✅ Go Slice最佳实践（15个核心实践）
- ✅ 接口与方法接收器（10个设计原则）
- ✅ Defer开销与行为（8个关键模式）
- ✅ Goroutine最佳实践（20+并发模式）
- ✅ Go Test最佳实践（完整测试指南）
- ✅ Go Mock最佳实践（4种Mock模式）

**之前项目**:
- Seata AT Mode（分布式事务自动补偿）
- Seata TCC Mode（分布式事务手动补偿）
- Singleflight Package（重复调用抑制）
- Factory Pattern（工厂设计模式）
- Decorator Pattern（装饰器设计模式）
- Red-Black Tree（红黑树）
- KMP Algorithm（KMP算法）
- B+ Tree（B+树）
- MongoDB Inverted Index（倒排索引）

