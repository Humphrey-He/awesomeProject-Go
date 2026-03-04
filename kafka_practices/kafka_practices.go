// kafka_practices.go - Kafka 操作示例与最佳实践
// 使用 IBM/sarama 库实现 Kafka 生产者、消费者、事务等操作

package kafka_practices

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

// ==================== 生产者 ====================

// ProducerConfig 生产者配置
type ProducerConfig struct {
	Brokers       []string      // Kafka broker 地址列表
	RequiredAcks  sarama.RequiredAcks // 确认级别
	RetryMax      int           // 最大重试次数
	RetryBackoff  time.Duration // 重试间隔
	FlushBytes    int           // 批量发送字节数
	FlushMessages int           // 批量发送消息数
	FlushFrequency time.Duration // 批量发送频率
}

// DefaultProducerConfig 返回默认生产者配置
func DefaultProducerConfig() *ProducerConfig {
	return &ProducerConfig{
		Brokers:        []string{"localhost:9092"},
		RequiredAcks:   sarama.WaitForAll, // 等待所有副本确认
		RetryMax:       3,
		RetryBackoff:   100 * time.Millisecond,
		FlushBytes:     1024 * 1024, // 1MB
		FlushMessages:  1000,
		FlushFrequency: 100 * time.Millisecond,
	}
}

// SyncProducer 同步生产者
type SyncProducer struct {
	producer sarama.SyncProducer
	config   *ProducerConfig
}

// NewSyncProducer 创建同步生产者
func NewSyncProducer(cfg *ProducerConfig) (*SyncProducer, error) {
	if cfg == nil {
		cfg = DefaultProducerConfig()
	}

	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.RequiredAcks = cfg.RequiredAcks
	saramaCfg.Producer.Retry.Max = cfg.RetryMax
	saramaCfg.Producer.Retry.Backoff = cfg.RetryBackoff
	saramaCfg.Producer.Return.Successes = true // 同步生产者需要设置为true

	producer, err := sarama.NewSyncProducer(cfg.Brokers, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("创建同步生产者失败: %w", err)
	}

	return &SyncProducer{
		producer: producer,
		config:   cfg,
	}, nil
}

// SendMessage 发送单条消息
func (p *SyncProducer) SendMessage(topic string, key, value []byte) (int32, int64, error) {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return 0, 0, fmt.Errorf("发送消息失败: %w", err)
	}

	return partition, offset, nil
}

// SendMessageWithHeaders 发送带消息头的消息
func (p *SyncProducer) SendMessageWithHeaders(topic string, key, value []byte, headers map[string]string) (int32, int64, error) {
	msgHeaders := make([]sarama.RecordHeader, 0, len(headers))
	for k, v := range headers {
		msgHeaders = append(msgHeaders, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		})
	}

	msg := &sarama.ProducerMessage{
		Topic:   topic,
		Key:     sarama.ByteEncoder(key),
		Value:   sarama.ByteEncoder(value),
		Headers: msgHeaders,
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return 0, 0, fmt.Errorf("发送消息失败: %w", err)
	}

	return partition, offset, nil
}

// Close 关闭生产者
func (p *SyncProducer) Close() error {
	return p.producer.Close()
}

// AsyncProducer 异步生产者
type AsyncProducer struct {
	producer sarama.AsyncProducer
	config   *ProducerConfig
	errors   chan error
}

// NewAsyncProducer 创建异步生产者
func NewAsyncProducer(cfg *ProducerConfig) (*AsyncProducer, error) {
	if cfg == nil {
		cfg = DefaultProducerConfig()
	}

	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.RequiredAcks = cfg.RequiredAcks
	saramaCfg.Producer.Retry.Max = cfg.RetryMax
	saramaCfg.Producer.Retry.Backoff = cfg.RetryBackoff
	saramaCfg.Producer.Return.Successes = false // 异步生产者通常不需要返回成功
	saramaCfg.Producer.Return.Errors = true     // 但需要返回错误
	saramaCfg.Producer.Flush.Bytes = cfg.FlushBytes
	saramaCfg.Producer.Flush.Messages = cfg.FlushMessages
	saramaCfg.Producer.Flush.Frequency = cfg.FlushFrequency

	producer, err := sarama.NewAsyncProducer(cfg.Brokers, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("创建异步生产者失败: %w", err)
	}

	ap := &AsyncProducer{
		producer: producer,
		config:   cfg,
		errors:   make(chan error, 100),
	}

	// 启动错误处理 goroutine
	go ap.handleErrors()

	return ap, nil
}

// handleErrors 处理异步发送错误
func (p *AsyncProducer) handleErrors() {
	for err := range p.producer.Errors() {
		log.Printf("消息发送失败: topic=%s, partition=%d, offset=%d, error=%v",
			err.Msg.Topic, err.Msg.Partition, err.Msg.Offset, err.Err)
		select {
		case p.errors <- err.Err:
		default:
			// 错误通道已满，丢弃错误
		}
	}
}

// SendMessageAsync 异步发送消息
func (p *AsyncProducer) SendMessageAsync(topic string, key, value []byte) {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}
	p.producer.Input() <- msg
}

// Errors 返回错误通道
func (p *AsyncProducer) Errors() <-chan error {
	return p.errors
}

// Close 关闭异步生产者
func (p *AsyncProducer) Close() error {
	return p.producer.Close()
}

// ==================== 消费者 ====================

// ConsumerConfig 消费者配置
type ConsumerConfig struct {
	Brokers         []string
	GroupID         string
	Topics          []string
	InitialOffset   int64 // sarama.OffsetNewest 或 sarama.OffsetOldest
	SessionTimeout  time.Duration
	HeartbeatInterval time.Duration
	MaxProcessingTime time.Duration
}

// DefaultConsumerConfig 返回默认消费者配置
func DefaultConsumerConfig() *ConsumerConfig {
	return &ConsumerConfig{
		Brokers:           []string{"localhost:9092"},
		InitialOffset:     sarama.OffsetNewest,
		SessionTimeout:    10 * time.Second,
		HeartbeatInterval: 3 * time.Second,
		MaxProcessingTime: 10 * time.Second,
	}
}

// ConsumerHandler 消息处理器接口
type ConsumerHandler interface {
	Setup(sarama.ConsumerGroupSession) error
	Cleanup(sarama.ConsumerGroupSession) error
	ConsumeClaim(sarama.ConsumerGroupSession, sarama.ConsumerGroupClaim) error
}

// ConsumerGroup 消费者组
type ConsumerGroup struct {
	group   sarama.ConsumerGroup
	config  *ConsumerConfig
	handler ConsumerHandler
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// NewConsumerGroup 创建消费者组
func NewConsumerGroup(cfg *ConsumerConfig, handler ConsumerHandler) (*ConsumerGroup, error) {
	if cfg == nil {
		cfg = DefaultConsumerConfig()
	}

	saramaCfg := sarama.NewConfig()
	saramaCfg.Consumer.Return.Errors = true
	saramaCfg.Consumer.Offsets.Initial = cfg.InitialOffset
	saramaCfg.Consumer.Group.Session.Timeout = cfg.SessionTimeout
	saramaCfg.Consumer.Group.Heartbeat.Interval = cfg.HeartbeatInterval
	saramaCfg.Consumer.MaxProcessingTime = cfg.MaxProcessingTime

	group, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("创建消费者组失败: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ConsumerGroup{
		group:   group,
		config:  cfg,
		handler: handler,
		ctx:     ctx,
		cancel:  cancel,
	}, nil
}

// Start 启动消费者
func (c *ConsumerGroup) Start() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			// 当 ctx 被取消时，Consume 会返回
			err := c.group.Consume(c.ctx, c.config.Topics, c.handler)
			if err != nil {
				log.Printf("消费者组消费错误: %v", err)
			}

			// 检查是否应该退出
			if c.ctx.Err() != nil {
				return
			}
		}
	}()
}

// Stop 停止消费者
func (c *ConsumerGroup) Stop() {
	c.cancel()
	c.wg.Wait()
	c.group.Close()
}

// DefaultConsumerHandler 默认消息处理器
type DefaultConsumerHandler struct {
	ProcessFunc func(msg *sarama.ConsumerMessage) error
}

// Setup 在消费开始前调用
func (h *DefaultConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	log.Println("消费者组开始消费")
	return nil
}

// Cleanup 在消费结束后调用
func (h *DefaultConsumerHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("消费者组停止消费")
	return nil
}

// ConsumeClaim 消费消息
func (h *DefaultConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if h.ProcessFunc != nil {
			if err := h.ProcessFunc(msg); err != nil {
				log.Printf("处理消息失败: %v", err)
				// 根据业务需求决定是否提交 offset
				continue
			}
		}
		// 手动提交 offset
		session.MarkMessage(msg, "")
	}
	return nil
}

// ==================== 事务生产者 ====================

// TransactionalProducer 事务生产者
type TransactionalProducer struct {
	producer sarama.AsyncProducer
	config   *ProducerConfig
	transactionID string
}

// NewTransactionalProducer 创建事务生产者
func NewTransactionalProducer(cfg *ProducerConfig, transactionID string) (*TransactionalProducer, error) {
	if cfg == nil {
		cfg = DefaultProducerConfig()
	}

	saramaCfg := sarama.NewConfig()
	saramaCfg.Producer.RequiredAcks = sarama.WaitForAll
	saramaCfg.Producer.Retry.Max = cfg.RetryMax
	saramaCfg.Producer.Idempotent = true // 幂等性
	saramaCfg.Producer.Transaction.ID = transactionID
	saramaCfg.Net.MaxOpenRequests = 1    // 事务生产者必须设置为1
	saramaCfg.Producer.Return.Successes = true
	saramaCfg.Producer.Return.Errors = true

	producer, err := sarama.NewAsyncProducer(cfg.Brokers, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("创建事务生产者失败: %w", err)
	}

	return &TransactionalProducer{
		producer:      producer,
		config:        cfg,
		transactionID: transactionID,
	}, nil
}

// BeginTransaction 开始事务
func (p *TransactionalProducer) BeginTransaction() error {
	return p.producer.BeginTxn()
}

// CommitTransaction 提交事务
func (p *TransactionalProducer) CommitTransaction() error {
	return p.producer.CommitTxn()
}

// AbortTransaction 回滚事务
func (p *TransactionalProducer) AbortTransaction() error {
	return p.producer.AbortTxn()
}

// SendMessageInTransaction 在事务中发送消息
func (p *TransactionalProducer) SendMessageInTransaction(topic string, key, value []byte) {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}
	p.producer.Input() <- msg
}

// Close 关闭事务生产者
func (p *TransactionalProducer) Close() error {
	return p.producer.Close()
}

// ==================== 分区分配策略 ====================

// PartitionAssigner 分区分配器
type PartitionAssigner struct{}

// NewHashPartitioner 基于 key 的哈希分区器
func NewHashPartitioner(topic string) sarama.Partitioner {
	return sarama.NewHashPartitioner(topic)
}

// NewRandomPartitioner 随机分区器
func NewRandomPartitioner(topic string) sarama.Partitioner {
	return sarama.NewRandomPartitioner(topic)
}

// NewRoundRobinPartitioner 轮询分区器
func NewRoundRobinPartitioner(topic string) sarama.Partitioner {
	return sarama.NewRoundRobinPartitioner(topic)
}

// NewManualPartitioner 手动指定分区
func NewManualPartitioner(topic string) sarama.Partitioner {
	return sarama.NewManualPartitioner(topic)
}

// ==================== 监控与指标 ====================

// ProducerMetrics 生产者指标
type ProducerMetrics struct {
	MessagesSent    uint64
	MessagesFailed  uint64
	BytesSent       uint64
	LatencyMs       []float64
	mu              sync.RWMutex
}

// RecordSuccess 记录成功发送
func (m *ProducerMetrics) RecordSuccess(bytes int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesSent++
	m.BytesSent += uint64(bytes)
}

// RecordFailure 记录发送失败
func (m *ProducerMetrics) RecordFailure() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.MessagesFailed++
}

// GetStats 获取统计信息
func (m *ProducerMetrics) GetStats() (sent, failed, bytes uint64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.MessagesSent, m.MessagesFailed, m.BytesSent
}

// ==================== 最佳实践工具函数 ====================

// CreateTopic 创建 Topic（需要 Kafka 管理员权限）
func CreateTopic(brokers []string, topic string, partitions int32, replicationFactor int16) error {
	config := sarama.NewConfig()
	admin, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		return fmt.Errorf("创建集群管理员失败: %w", err)
	}
	defer admin.Close()

	detail := &sarama.TopicDetail{
		NumPartitions:     partitions,
		ReplicationFactor: replicationFactor,
	}

	if err := admin.CreateTopic(topic, detail, false); err != nil {
		return fmt.Errorf("创建 topic 失败: %w", err)
	}

	return nil
}

// ListTopics 列出所有 Topic
func ListTopics(brokers []string) (map[string]sarama.TopicDetail, error) {
	config := sarama.NewConfig()
	admin, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("创建集群管理员失败: %w", err)
	}
	defer admin.Close()

	topics, err := admin.ListTopics()
	if err != nil {
		return nil, fmt.Errorf("列出 topics 失败: %w", err)
	}

	return topics, nil
}

// GetConsumerLag 获取消费者延迟
func GetConsumerLag(brokers []string, groupID string) (map[string]map[int32]int64, error) {
	config := sarama.NewConfig()
	admin, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("创建集群管理员失败: %w", err)
	}
	defer admin.Close()

	// 获取消费者组描述
	descs, err := admin.DescribeConsumerGroups([]string{groupID})
	if err != nil {
		return nil, fmt.Errorf("描述消费者组失败: %w", err)
	}

	lag := make(map[string]map[int32]int64)
	for _, desc := range descs {
		for _, member := range desc.Members {
			assignment, err := member.GetMemberAssignment()
			if err != nil {
				continue
			}
			for topic, partitions := range assignment.Topics {
				if lag[topic] == nil {
					lag[topic] = make(map[int32]int64)
				}
				for _, partition := range partitions {
					// 这里简化处理，实际需要获取 offset 信息计算 lag
					lag[topic][partition] = 0
				}
			}
		}
	}

	return lag, nil
}

// RetryWithBackoff 带退避的重试函数
func RetryWithBackoff(maxRetries int, initialBackoff time.Duration, fn func() error) error {
	backoff := initialBackoff
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
			if i < maxRetries-1 {
				time.Sleep(backoff)
				backoff *= 2 // 指数退避
			}
		}
	}

	return fmt.Errorf("重试 %d 次后仍然失败: %w", maxRetries, lastErr)
}

// ValidateMessage 验证消息
func ValidateMessage(key, value []byte, maxSize int) error {
	if len(key) > maxSize {
		return fmt.Errorf("key 大小超过限制: %d > %d", len(key), maxSize)
	}
	if len(value) > maxSize {
		return fmt.Errorf("value 大小超过限制: %d > %d", len(value), maxSize)
	}
	return nil
}

// SerializeJSON 序列化 JSON 消息（示例）
func SerializeJSON(data interface{}) ([]byte, error) {
	// 实际项目中使用 json.Marshal
	// 这里仅作为示例
	return []byte(fmt.Sprintf("%v", data)), nil
}

// DeserializeJSON 反序列化 JSON 消息（示例）
func DeserializeJSON(data []byte, v interface{}) error {
	// 实际项目中使用 json.Unmarshal
	// 这里仅作为示例
	return nil
}
