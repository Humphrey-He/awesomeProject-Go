# Kafka 实践项目

基于 [IBM/sarama](https://github.com/IBM/sarama) 库实现的 Kafka 操作示例与最佳实践。

## 项目概述

本项目演示了 Kafka 在日常开发中的常用操作，包括生产者、消费者、事务消息等核心功能，并提供了完整的最佳实践指南。

## 核心功能

### 1. 生产者

#### 同步生产者 (SyncProducer)
```go
// 创建同步生产者
config := DefaultProducerConfig()
producer, err := NewSyncProducer(config)
if err != nil {
    log.Fatal(err)
}
defer producer.Close()

// 发送消息
partition, offset, err := producer.SendMessage("my-topic", 
    []byte("key"), 
    []byte("value"))
```

**特点**:
- 等待所有副本确认 (WaitForAll)
- 自动重试机制
- 适合重要消息场景

#### 异步生产者 (AsyncProducer)
```go
// 创建异步生产者
producer, err := NewAsyncProducer(config)
defer producer.Close()

// 异步发送消息
producer.SendMessageAsync("my-topic", []byte("key"), []byte("value"))
```

**特点**:
- 批量发送优化
- 更高吞吐量
- 非阻塞调用

#### 事务生产者 (TransactionalProducer)
```go
// 创建事务生产者
producer, _ := NewTransactionalProducer(config, "my-transaction-id")

// 开始事务
producer.BeginTransaction()

// 发送多条消息
producer.SendMessageInTransaction("topic1", []byte("k1"), []byte("v1"))
producer.SendMessageInTransaction("topic2", []byte("k2"), []byte("v2"))

// 提交或回滚
producer.CommitTransaction()  // 或 producer.AbortTransaction()
```

### 2. 消费者

#### 消费者组 (ConsumerGroup)
```go
// 创建消息处理器
handler := &DefaultConsumerHandler{
    ProcessFunc: func(msg *sarama.ConsumerMessage) error {
        fmt.Printf("收到消息: %s\n", string(msg.Value))
        return nil
    },
}

// 创建消费者组
config := &ConsumerConfig{
    Brokers:       []string{"localhost:9092"},
    GroupID:       "my-group",
    Topics:        []string{"my-topic"},
    InitialOffset: sarama.OffsetNewest,
}
consumer, _ := NewConsumerGroup(config, handler)

// 启动消费
consumer.Start()
defer consumer.Stop()
```

### 3. 分区策略

| 策略 | 说明 | 使用场景 |
|------|------|----------|
| Hash | 根据 key 哈希分配 | 保证相同 key 进入同一分区 |
| Random | 随机分配 | 负载均衡场景 |
| RoundRobin | 轮询分配 | 均匀分布场景 |
| Manual | 手动指定分区 | 特殊业务需求 |

```go
// 使用哈希分区器
partitioner := NewHashPartitioner("my-topic")
```

### 4. 管理操作

```go
// 创建 Topic
CreateTopic([]string{"localhost:9092"}, "new-topic", 3, 2)

// 列出所有 Topic
topics, _ := ListTopics([]string{"localhost:9092"})

// 获取消费者延迟
lag, _ := GetConsumerLag([]string{"localhost:9092"}, "my-group")
```

## 配置说明

### ProducerConfig 生产者配置

| 参数 | 默认值 | 说明 |
|------|--------|------|
| Brokers | ["localhost:9092"] | Broker 地址列表 |
| RequiredAcks | WaitForAll | 确认级别 |
| RetryMax | 3 | 最大重试次数 |
| RetryBackoff | 100ms | 重试间隔 |
| FlushBytes | 1MB | 批量发送字节数 |
| FlushMessages | 1000 | 批量发送消息数 |
| FlushFrequency | 100ms | 批量发送频率 |

### ConsumerConfig 消费者配置

| 参数 | 默认值 | 说明 |
|------|--------|------|
| Brokers | ["localhost:9092"] | Broker 地址列表 |
| GroupID | - | 消费者组ID |
| Topics | - | 订阅的Topic列表 |
| InitialOffset | OffsetNewest | 初始偏移量 |
| SessionTimeout | 10s | 会话超时时间 |
| HeartbeatInterval | 3s | 心跳间隔 |

## 最佳实践

### 1. 生产者最佳实践

#### 幂等性配置
```go
config := sarama.NewConfig()
config.Producer.Idempotent = true       // 开启幂等性
config.Net.MaxOpenRequests = 1          // 必须设置为1
config.Producer.RequiredAcks = sarama.WaitForAll
```

#### 批量发送优化
```go
config.Producer.Flush.Bytes = 1024 * 1024     // 1MB
config.Producer.Flush.Messages = 1000         // 1000条
config.Producer.Flush.Frequency = 100 * time.Millisecond
```

#### 重试策略
```go
config.Producer.Retry.Max = 3
config.Producer.Retry.Backoff = 100 * time.Millisecond
```

### 2. 消费者最佳实践

#### 手动提交 Offset
```go
handler := &DefaultConsumerHandler{
    ProcessFunc: func(msg *sarama.ConsumerMessage) error {
        // 处理消息
        if err := processMessage(msg); err != nil {
            return err // 不提交 offset，稍后重试
        }
        return nil
    },
}
```

#### 优雅关闭
```go
ctx, cancel := context.WithCancel(context.Background())

// 监听信号
go func() {
    sig := <-signalChan
    log.Println("收到信号:", sig)
    cancel()
}()

// 消费循环
for {
    select {
    case <-ctx.Done():
        return
    default:
        // 消费消息
    }
}
```

### 3. 错误处理

#### 带退避的重试
```go
err := RetryWithBackoff(3, 100*time.Millisecond, func() error {
    _, _, err := producer.SendMessage(topic, key, value)
    return err
})
```

#### 消息验证
```go
if err := ValidateMessage(key, value, 1024*1024); err != nil {
    return fmt.Errorf("消息验证失败: %w", err)
}
```

### 4. 监控指标

```go
metrics := &ProducerMetrics{}

// 发送成功后记录
metrics.RecordSuccess(len(value))

// 发送失败后记录
metrics.RecordFailure()

// 获取统计
sent, failed, bytes := metrics.GetStats()
```

## 使用场景

### 场景1: 日志收集
```go
// 异步生产者 + 批量发送
producer, _ := NewAsyncProducer(config)
for _, log := range logs {
    producer.SendMessageAsync("logs-topic", 
        []byte(log.Service), 
        []byte(log.Content))
}
```

### 场景2: 订单系统
```go
// 事务生产者 + 精确一次语义
producer.BeginTransaction()
producer.SendMessageInTransaction("orders", orderKey, orderValue)
producer.SendMessageInTransaction("inventory", inventoryKey, inventoryValue)
producer.CommitTransaction()
```

### 场景3: 实时流处理
```go
// 消费者组 + 并行处理
handler := &DefaultConsumerHandler{
    ProcessFunc: func(msg *sarama.ConsumerMessage) error {
        return processStream(msg.Value)
    },
}
```

## Docker 快速启动

```bash
# 启动 Kafka
docker run -d --name kafka \
    -p 9092:9092 \
    -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092 \
    apache/kafka:latest
```

## 测试

```bash
# 运行测试（需要本地 Kafka）
go test -v

# 基准测试
go test -bench=.
```

## 依赖

- `github.com/IBM/sarama` - Kafka Go 客户端

## 参考资料

- [sarama 官方文档](https://github.com/IBM/sarama)
- [Kafka 官方文档](https://kafka.apache.org/documentation/)
- [Kafka 设计原理](https://kafka.apache.org/documentation/#design)
