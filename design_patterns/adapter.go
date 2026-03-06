package design_patterns

import (
	"fmt"
	"io"
)

// ========== 基础适配器模式 ==========

// Target 目标接口
type Target interface {
	Request() string
}

// Adaptee 被适配者
type Adaptee struct{}

func (a *Adaptee) SpecificRequest() string {
	return "Specific request from Adaptee"
}

// Adapter 适配器
type Adapter struct {
	adaptee *Adaptee
}

func NewAdapter(adaptee *Adaptee) *Adapter {
	return &Adapter{adaptee: adaptee}
}

func (a *Adapter) Request() string {
	return fmt.Sprintf("Adapter: %s", a.adaptee.SpecificRequest())
}

// ========== 实际应用：日志适配器 ==========

// LoggerInterface 日志接口
type LoggerInterface interface {
	Log(message string)
}

// ZapLogger Zap日志库
type ZapLogger struct{}

func (l *ZapLogger) ZapLog(msg string, level string) {
	fmt.Printf("[ZAP][%s] %s\n", level, msg)
}

// LogrusLogger Logrus日志库
type LogrusLogger struct{}

func (l *LogrusLogger) LogrusEntry(msg string, level string) {
	fmt.Printf("[LOGRUS][%s] %s\n", level, msg)
}

// ZapAdapter Zap适配器
type ZapAdapter struct {
	zap *ZapLogger
}

func NewZapAdapter(zap *ZapLogger) *ZapAdapter {
	return &ZapAdapter{zap: zap}
}

func (a *ZapAdapter) Log(message string) {
	a.zap.ZapLog(message, "INFO")
}

// LogrusAdapter Logrus适配器
type LogrusAdapter struct {
	logrus *LogrusLogger
}

func NewLogrusAdapter(logrus *LogrusLogger) *LogrusAdapter {
	return &LogrusAdapter{logrus: logrus}
}

func (a *LogrusAdapter) Log(message string) {
	a.logrus.LogrusEntry(message, "info")
}

// ========== 实际应用：数据库适配器 ==========

// DatabaseDriver 数据库驱动接口
type DatabaseDriver interface {
	Connect(host string, port int, database string) error
	Execute(query string) (interface{}, error)
	Disconnect() error
}

// MySQLDriver MySQL驱动
type MySQLDriver struct {
	connected bool
}

func (d *MySQLDriver) MySQLConnect(host string, port int, user, password, database string) error {
	fmt.Printf("[MySQL] Connecting to %s:%d/%s\n", host, port, database)
	d.connected = true
	return nil
}

func (d *MySQLDriver) MySQLQuery(sql string) (interface{}, error) {
	if !d.connected {
		return nil, fmt.Errorf("not connected")
	}
	fmt.Printf("[MySQL] Executing: %s\n", sql)
	return fmt.Sprintf("MySQL Result: %s", sql), nil
}

func (d *MySQLDriver) MySQLClose() error {
	d.connected = false
	fmt.Println("[MySQL] Connection closed")
	return nil
}

// MySQLDriverAdapter MySQL驱动适配器
type MySQLDriverAdapter struct {
	mysql    *MySQLDriver
	user     string
	password string
}

func NewMySQLDriverAdapter(mysql *MySQLDriver, user, password string) *MySQLDriverAdapter {
	return &MySQLDriverAdapter{
		mysql:    mysql,
		user:     user,
		password: password,
	}
}

func (a *MySQLDriverAdapter) Connect(host string, port int, database string) error {
	return a.mysql.MySQLConnect(host, port, a.user, a.password, database)
}

func (a *MySQLDriverAdapter) Execute(query string) (interface{}, error) {
	return a.mysql.MySQLQuery(query)
}

func (a *MySQLDriverAdapter) Disconnect() error {
	return a.mysql.MySQLClose()
}

// PostgreSQLDriver PostgreSQL驱动
type PostgreSQLDriver struct {
	connStr string
}

func (d *PostgreSQLDriver) PGConnect(connStr string) error {
	d.connStr = connStr
	fmt.Printf("[PostgreSQL] Connecting with: %s\n", connStr)
	return nil
}

func (d *PostgreSQLDriver) PGQuery(sql string) (interface{}, error) {
	if d.connStr == "" {
		return nil, fmt.Errorf("not connected")
	}
	fmt.Printf("[PostgreSQL] Executing: %s\n", sql)
	return fmt.Sprintf("PostgreSQL Result: %s", sql), nil
}

func (d *PostgreSQLDriver) PGClose() error {
	d.connStr = ""
	fmt.Println("[PostgreSQL] Connection closed")
	return nil
}

// PostgreSQLDriverAdapter PostgreSQL驱动适配器
type PostgreSQLDriverAdapter struct {
	pg *PostgreSQLDriver
}

func NewPostgreSQLDriverAdapter(pg *PostgreSQLDriver) *PostgreSQLDriverAdapter {
	return &PostgreSQLDriverAdapter{pg: pg}
}

func (a *PostgreSQLDriverAdapter) Connect(host string, port int, database string) error {
	connStr := fmt.Sprintf("host=%s port=%d dbname=%s", host, port, database)
	return a.pg.PGConnect(connStr)
}

func (a *PostgreSQLDriverAdapter) Execute(query string) (interface{}, error) {
	return a.pg.PGQuery(query)
}

func (a *PostgreSQLDriverAdapter) Disconnect() error {
	return a.pg.PGClose()
}

// ========== 实际应用：第三方支付适配器 ==========

// PaymentProcessor 支付处理器接口
type PaymentProcessor interface {
	ProcessPayment(orderID string, amount float64) (string, error)
	Refund(orderID string, amount float64) error
	QueryStatus(orderID string) (string, error)
}

// AlipaySDK 支付宝SDK
type AlipaySDK struct{}

func (s *AlipaySDK) CreateTrade(orderID string, totalAmount float64) (string, error) {
	return fmt.Sprintf("ALIPAY_TRADE_%s", orderID), nil
}

func (s *AlipaySDK) RefundTrade(tradeNo string, refundAmount float64) error {
	fmt.Printf("[ALIPAY] Refunding %.2f for %s\n", refundAmount, tradeNo)
	return nil
}

func (s *AlipaySDK) QueryTrade(tradeNo string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"trade_no": tradeNo,
		"status":   "TRADE_SUCCESS",
	}, nil
}

// AlipayAdapter 支付宝适配器
type AlipayAdapter struct {
	sdk *AlipaySDK
}

func NewAlipayAdapter(sdk *AlipaySDK) *AlipayAdapter {
	return &AlipayAdapter{sdk: sdk}
}

func (a *AlipayAdapter) ProcessPayment(orderID string, amount float64) (string, error) {
	tradeNo, err := a.sdk.CreateTrade(orderID, amount)
	if err != nil {
		return "", err
	}
	fmt.Printf("[ALIPAY_ADAPTER] Created trade: %s\n", tradeNo)
	return tradeNo, nil
}

func (a *AlipayAdapter) Refund(orderID string, amount float64) error {
	return a.sdk.RefundTrade(fmt.Sprintf("ALIPAY_TRADE_%s", orderID), amount)
}

func (a *AlipayAdapter) QueryStatus(orderID string) (string, error) {
	result, err := a.sdk.QueryTrade(fmt.Sprintf("ALIPAY_TRADE_%s", orderID))
	if err != nil {
		return "", err
	}
	return result["status"].(string), nil
}

// WeChatPaySDK 微信支付SDK
type WeChatPaySDK struct{}

func (s *WeChatPaySDK) UnifiedOrder(outTradeNo string, totalFee int) (string, error) {
	return fmt.Sprintf("WECHAT_PREPAY_%s", outTradeNo), nil
}

func (s *WeChatPaySDK) RefundOrder(outTradeNo string, refundFee int) error {
	fmt.Printf("[WECHAT] Refunding %d cents for %s\n", refundFee, outTradeNo)
	return nil
}

func (s *WeChatPaySDK) QueryOrder(outTradeNo string) (string, error) {
	return "SUCCESS", nil
}

// WeChatPayAdapter 微信支付适配器
type WeChatPayAdapter struct {
	sdk *WeChatPaySDK
}

func NewWeChatPayAdapter(sdk *WeChatPaySDK) *WeChatPayAdapter {
	return &WeChatPayAdapter{sdk: sdk}
}

func (a *WeChatPayAdapter) ProcessPayment(orderID string, amount float64) (string, error) {
	// 微信支付以分为单位
	totalFee := int(amount * 100)
	prepayID, err := a.sdk.UnifiedOrder(orderID, totalFee)
	if err != nil {
		return "", err
	}
	fmt.Printf("[WECHAT_ADAPTER] Created prepay: %s\n", prepayID)
	return prepayID, nil
}

func (a *WeChatPayAdapter) Refund(orderID string, amount float64) error {
	return a.sdk.RefundOrder(orderID, int(amount*100))
}

func (a *WeChatPayAdapter) QueryStatus(orderID string) (string, error) {
	return a.sdk.QueryOrder(orderID)
}

// ========== 实际应用：消息队列适配器 ==========

// MessageQueue 消息队列接口
type MessageQueue interface {
	Publish(topic string, message []byte) error
	Subscribe(topic string, handler func([]byte)) error
	Close() error
}

// KafkaClient Kafka客户端
type KafkaClient struct {
	brokers []string
}

func (c *KafkaClient) KafkaProduce(topic string, key, value []byte) error {
	fmt.Printf("[KAFKA] Producing to topic %s: %s\n", topic, string(value))
	return nil
}

func (c *KafkaClient) KafkaConsume(topic string, handler func(key, value []byte)) error {
	fmt.Printf("[KAFKA] Consuming from topic %s\n", topic)
	return nil
}

func (c *KafkaClient) KafkaClose() error {
	fmt.Println("[KAFKA] Closing")
	return nil
}

// KafkaAdapter Kafka适配器
type KafkaAdapter struct {
	client *KafkaClient
}

func NewKafkaAdapter(client *KafkaClient) *KafkaAdapter {
	return &KafkaAdapter{client: client}
}

func (a *KafkaAdapter) Publish(topic string, message []byte) error {
	return a.client.KafkaProduce(topic, nil, message)
}

func (a *KafkaAdapter) Subscribe(topic string, handler func([]byte)) error {
	return a.client.KafkaConsume(topic, func(key, value []byte) {
		handler(value)
	})
}

func (a *KafkaAdapter) Close() error {
	return a.client.KafkaClose()
}

// RabbitMQClient RabbitMQ客户端
type RabbitMQClient struct {
	url string
}

func (c *RabbitMQClient) RabbitPublish(queue string, body []byte) error {
	fmt.Printf("[RABBITMQ] Publishing to queue %s: %s\n", queue, string(body))
	return nil
}

func (c *RabbitMQClient) RabbitConsume(queue string, handler func([]byte)) error {
	fmt.Printf("[RABBITMQ] Consuming from queue %s\n", queue)
	return nil
}

func (c *RabbitMQClient) RabbitClose() error {
	fmt.Println("[RABBITMQ] Closing")
	return nil
}

// RabbitMQAdapter RabbitMQ适配器
type RabbitMQAdapter struct {
	client *RabbitMQClient
}

func NewRabbitMQAdapter(client *RabbitMQClient) *RabbitMQAdapter {
	return &RabbitMQAdapter{client: client}
}

func (a *RabbitMQAdapter) Publish(topic string, message []byte) error {
	return a.client.RabbitPublish(topic, message)
}

func (a *RabbitMQAdapter) Subscribe(topic string, handler func([]byte)) error {
	return a.client.RabbitConsume(topic, handler)
}

func (a *RabbitMQAdapter) Close() error {
	return a.client.RabbitClose()
}

// ========== 实际应用：文件存储适配器 ==========

// FileStorage 文件存储接口
type FileStorage interface {
	Upload(key string, data []byte) error
	Download(key string) ([]byte, error)
	Delete(key string) error
}

// S3Client S3客户端
type S3Client struct {
	bucket string
	region string
}

func (c *S3Client) PutObject(key string, body io.Reader) error {
	fmt.Printf("[S3] Uploading %s to bucket %s in %s\n", key, c.bucket, c.region)
	return nil
}

func (c *S3Client) GetObject(key string) (io.ReadCloser, error) {
	fmt.Printf("[S3] Downloading %s\n", key)
	return nil, nil
}

func (c *S3Client) DeleteObject(key string) error {
	fmt.Printf("[S3] Deleting %s\n", key)
	return nil
}

// S3Adapter S3适配器
type S3Adapter struct {
	client *S3Client
}

func NewS3Adapter(client *S3Client) *S3Adapter {
	return &S3Adapter{client: client}
}

func (a *S3Adapter) Upload(key string, data []byte) error {
	return a.client.PutObject(key, nil)
}

func (a *S3Adapter) Download(key string) ([]byte, error) {
	_, err := a.client.GetObject(key)
	return nil, err
}

func (a *S3Adapter) Delete(key string) error {
	return a.client.DeleteObject(key)
}

// OSSClient 阿里云OSS客户端
type OSSClient struct {
	bucket string
	endpoint string
}

func (c *OSSClient) OSSPutObject(key string, data []byte) error {
	fmt.Printf("[OSS] Uploading %s to bucket %s at %s\n", key, c.bucket, c.endpoint)
	return nil
}

func (c *OSSClient) OSSGetObject(key string) ([]byte, error) {
	fmt.Printf("[OSS] Downloading %s\n", key)
	return nil, nil
}

func (c *OSSClient) OSSDeleteObject(key string) error {
	fmt.Printf("[OSS] Deleting %s\n", key)
	return nil
}

// OSSAdapter OSS适配器
type OSSAdapter struct {
	client *OSSClient
}

func NewOSSAdapter(client *OSSClient) *OSSAdapter {
	return &OSSAdapter{client: client}
}

func (a *OSSAdapter) Upload(key string, data []byte) error {
	return a.client.OSSPutObject(key, data)
}

func (a *OSSAdapter) Download(key string) ([]byte, error) {
	return a.client.OSSGetObject(key)
}

func (a *OSSAdapter) Delete(key string) error {
	return a.client.OSSDeleteObject(key)
}

// ========== 实际应用：缓存适配器 ==========

// CacheClient 缓存客户端接口
type CacheClient interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Del(key string)
}

// RedisClient Redis客户端
type RedisClient struct {
	addr string
}

func (c *RedisClient) RedisGet(key string) (string, bool) {
	fmt.Printf("[REDIS] Getting %s from %s\n", key, c.addr)
	return "value", true
}

func (c *RedisClient) RedisSet(key, value string) {
	fmt.Printf("[REDIS] Setting %s = %s\n", key, value)
}

func (c *RedisClient) RedisDel(key string) {
	fmt.Printf("[REDIS] Deleting %s\n", key)
}

// RedisAdapter Redis适配器
type RedisAdapter struct {
	client *RedisClient
}

func NewRedisAdapter(client *RedisClient) *RedisAdapter {
	return &RedisAdapter{client: client}
}

func (a *RedisAdapter) Get(key string) (interface{}, bool) {
	return a.client.RedisGet(key)
}

func (a *RedisAdapter) Set(key string, value interface{}) {
	a.client.RedisSet(key, fmt.Sprintf("%v", value))
}

func (a *RedisAdapter) Del(key string) {
	a.client.RedisDel(key)
}

// MemcachedClient Memcached客户端
type MemcachedClient struct {
	servers []string
}

func (c *MemcachedClient) MCGet(key string) ([]byte, bool) {
	fmt.Printf("[MEMCACHED] Getting %s\n", key)
	return []byte("value"), true
}

func (c *MemcachedClient) MCSet(key string, value []byte) {
	fmt.Printf("[MEMCACHED] Setting %s\n", key)
}

func (c *MemcachedClient) MCDel(key string) {
	fmt.Printf("[MEMCACHED] Deleting %s\n", key)
}

// MemcachedAdapter Memcached适配器
type MemcachedAdapter struct {
	client *MemcachedClient
}

func NewMemcachedAdapter(client *MemcachedClient) *MemcachedAdapter {
	return &MemcachedAdapter{client: client}
}

func (a *MemcachedAdapter) Get(key string) (interface{}, bool) {
	return a.client.MCGet(key)
}

func (a *MemcachedAdapter) Set(key string, value interface{}) {
	a.client.MCSet(key, []byte(fmt.Sprintf("%v", value)))
}

func (a *MemcachedAdapter) Del(key string) {
	a.client.MCDel(key)
}

// ========== 实际应用：HTTP客户端适配器 ==========

// HTTPClient HTTP客户端接口
type HTTPClient interface {
	Do(method, url string, body []byte, headers map[string]string) ([]byte, error)
}

// DefaultHTTPClient 默认HTTP客户端
type DefaultHTTPClient struct{}

func (c *DefaultHTTPClient) Do(method, url string, body []byte, headers map[string]string) ([]byte, error) {
	fmt.Printf("[HTTP] %s %s\n", method, url)
	return []byte("response"), nil
}

// RetryHTTPClientAdapter 带重试的HTTP客户端适配器
type RetryHTTPClientAdapter struct {
	client     *DefaultHTTPClient
	maxRetries int
}

func NewRetryHTTPClientAdapter(client *DefaultHTTPClient, maxRetries int) *RetryHTTPClientAdapter {
	return &RetryHTTPClientAdapter{
		client:     client,
		maxRetries: maxRetries,
	}
}

func (a *RetryHTTPClientAdapter) Do(method, url string, body []byte, headers map[string]string) ([]byte, error) {
	var result []byte
	var err error

	for i := 0; i < a.maxRetries; i++ {
		result, err = a.client.Do(method, url, body, headers)
		if err == nil {
			return result, nil
		}
		fmt.Printf("[RETRY_HTTP] Attempt %d failed: %v\n", i+1, err)
	}

	return nil, err
}
