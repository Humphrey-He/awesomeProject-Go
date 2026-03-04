// kafka_practices_test.go - Kafka 操作测试

package kafka_practices

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/IBM/sarama"
)

// 注意：这些测试需要本地运行 Kafka 服务
// 可以使用 Docker 启动: docker run -d --name kafka -p 9092:9092 apache/kafka:latest

// TestSyncProducer 测试同步生产者
func TestSyncProducer(t *testing.T) {
	// 跳过测试如果没有 Kafka 环境
	t.Skip("跳过：需要本地 Kafka 环境")

	config := DefaultProducerConfig()
	producer, err := NewSyncProducer(config)
	if err != nil {
		t.Fatalf("创建同步生产者失败: %v", err)
	}
	defer producer.Close()

	// 发送消息
	key := []byte("test-key")
	value := []byte("test-value")
	partition, offset, err := producer.SendMessage("test-topic", key, value)
	if err != nil {
		t.Fatalf("发送消息失败: %v", err)
	}

	log.Printf("消息发送成功: partition=%d, offset=%d", partition, offset)
}

// TestAsyncProducer 测试异步生产者
func TestAsyncProducer(t *testing.T) {
	t.Skip("跳过：需要本地 Kafka 环境")

	config := DefaultProducerConfig()
	producer, err := NewAsyncProducer(config)
	if err != nil {
		t.Fatalf("创建异步生产者失败: %v", err)
	}
	defer producer.Close()

	// 发送多条消息
	for i := 0; i < 10; i++ {
		key := []byte(fmt.Sprintf("key-%d", i))
		value := []byte(fmt.Sprintf("value-%d", i))
		producer.SendMessageAsync("test-topic", key, value)
	}

	// 等待一段时间让消息发送
	time.Sleep(2 * time.Second)
}

// TestConsumerGroup 测试消费者组
func TestConsumerGroup(t *testing.T) {
	t.Skip("跳过：需要本地 Kafka 环境")

	handler := &DefaultConsumerHandler{
		ProcessFunc: func(msg *sarama.ConsumerMessage) error {
			log.Printf("收到消息: topic=%s, partition=%d, offset=%d, key=%s, value=%s",
				msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
			return nil
		},
	}

	config := &ConsumerConfig{
		Brokers:       []string{"localhost:9092"},
		GroupID:       "test-group",
		Topics:        []string{"test-topic"},
		InitialOffset: sarama.OffsetNewest,
	}

	consumer, err := NewConsumerGroup(config, handler)
	if err != nil {
		t.Fatalf("创建消费者组失败: %v", err)
	}

	// 启动消费者
	consumer.Start()

	// 消费10秒后停止
	time.Sleep(10 * time.Second)
	consumer.Stop()
}

// TestRetryWithBackoff 测试重试机制
func TestRetryWithBackoff(t *testing.T) {
	callCount := 0
	err := RetryWithBackoff(3, 10*time.Millisecond, func() error {
		callCount++
		if callCount < 3 {
			return fmt.Errorf("模拟错误")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("重试后应该成功: %v", err)
	}

	if callCount != 3 {
		t.Fatalf("应该重试3次，实际 %d 次", callCount)
	}
}

// TestValidateMessage 测试消息验证
func TestValidateMessage(t *testing.T) {
	tests := []struct {
		name    string
		key     []byte
		value   []byte
		maxSize int
		wantErr bool
	}{
		{
			name:    "正常消息",
			key:     []byte("key"),
			value:   []byte("value"),
			maxSize: 1024,
			wantErr: false,
		},
		{
			name:    "key 过大",
			key:     make([]byte, 100),
			value:   []byte("value"),
			maxSize: 50,
			wantErr: true,
		},
		{
			name:    "value 过大",
			key:     []byte("key"),
			value:   make([]byte, 100),
			maxSize: 50,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMessage(tt.key, tt.value, tt.maxSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestProducerMetrics 测试生产者指标
func TestProducerMetrics(t *testing.T) {
	metrics := &ProducerMetrics{}

	// 记录成功
	metrics.RecordSuccess(100)
	metrics.RecordSuccess(200)

	// 记录失败
	metrics.RecordFailure()

	sent, failed, bytes := metrics.GetStats()
	if sent != 2 {
		t.Errorf("MessagesSent = %d, want 2", sent)
	}
	if failed != 1 {
		t.Errorf("MessagesFailed = %d, want 1", failed)
	}
	if bytes != 300 {
		t.Errorf("BytesSent = %d, want 300", bytes)
	}
}

// TestPartitioners 测试分区器
func TestPartitioners(t *testing.T) {
	topic := "test-topic"

	// 测试哈希分区器
	hashPartitioner := NewHashPartitioner(topic)
	if hashPartitioner == nil {
		t.Error("哈希分区器创建失败")
	}

	// 测试随机分区器
	randomPartitioner := NewRandomPartitioner(topic)
	if randomPartitioner == nil {
		t.Error("随机分区器创建失败")
	}

	// 测试轮询分区器
	roundRobinPartitioner := NewRoundRobinPartitioner(topic)
	if roundRobinPartitioner == nil {
		t.Error("轮询分区器创建失败")
	}
}

// BenchmarkSyncProducer 基准测试同步生产者
func BenchmarkSyncProducer(b *testing.B) {
	b.Skip("跳过：需要本地 Kafka 环境")

	config := DefaultProducerConfig()
	producer, err := NewSyncProducer(config)
	if err != nil {
		b.Fatalf("创建同步生产者失败: %v", err)
	}
	defer producer.Close()

	key := []byte("benchmark-key")
	value := []byte("benchmark-value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := producer.SendMessage("benchmark-topic", key, value)
		if err != nil {
			b.Fatalf("发送消息失败: %v", err)
		}
	}
}

// BenchmarkAsyncProducer 基准测试异步生产者
func BenchmarkAsyncProducer(b *testing.B) {
	b.Skip("跳过：需要本地 Kafka 环境")

	config := DefaultProducerConfig()
	producer, err := NewAsyncProducer(config)
	if err != nil {
		b.Fatalf("创建异步生产者失败: %v", err)
	}
	defer producer.Close()

	key := []byte("benchmark-key")
	value := []byte("benchmark-value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		producer.SendMessageAsync("benchmark-topic", key, value)
	}
}

// ExampleSyncProducer 同步生产者示例
func ExampleSyncProducer() {
	// 创建生产者配置
	config := DefaultProducerConfig()
	config.Brokers = []string{"localhost:9092"}

	// 创建同步生产者
	producer, err := NewSyncProducer(config)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	// 发送消息
	partition, offset, err := producer.SendMessage("my-topic",
		[]byte("my-key"),
		[]byte("my-value"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("消息已发送: partition=%d, offset=%d\n", partition, offset)
}

// ExampleConsumerGroup 消费者组示例
func ExampleConsumerGroup() {
	// 创建消息处理器
	handler := &DefaultConsumerHandler{
		ProcessFunc: func(msg *sarama.ConsumerMessage) error {
			fmt.Printf("处理消息: %s\n", string(msg.Value))
			return nil
		},
	}

	// 创建消费者配置
	config := &ConsumerConfig{
		Brokers:       []string{"localhost:9092"},
		GroupID:       "my-group",
		Topics:        []string{"my-topic"},
		InitialOffset: sarama.OffsetNewest,
	}

	// 创建消费者组
	consumer, err := NewConsumerGroup(config, handler)
	if err != nil {
		log.Fatal(err)
	}

	// 启动消费者
	consumer.Start()

	// 运行一段时间后停止
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	<-ctx.Done()
	consumer.Stop()
}
