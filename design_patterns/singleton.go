package design_patterns

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// ========== 基础单例模式 ==========

// Singleton 基础单例（非线程安全，仅演示）
type Singleton struct {
	value string
}

var instance *Singleton

// GetInstance 获取单例实例（非线程安全）
func GetInstance() *Singleton {
	if instance == nil {
		instance = &Singleton{value: "default"}
	}
	return instance
}

// ========== 线程安全的单例模式 ==========

// ThreadSafeSingleton 线程安全单例
type ThreadSafeSingleton struct {
	value string
}

var (
	threadSafeInstance *ThreadSafeSingleton
	once               sync.Once
)

// GetThreadSafeInstance 获取线程安全单例
func GetThreadSafeInstance() *ThreadSafeSingleton {
	once.Do(func() {
		threadSafeInstance = &ThreadSafeSingleton{value: "thread-safe"}
	})
	return threadSafeInstance
}

// ========== 双重检查锁定单例 ==========

// DoubleCheckSingleton 双重检查锁定单例
type DoubleCheckSingleton struct {
	value string
}

var (
	doubleCheckInstance *DoubleCheckSingleton
	doubleCheckMutex    sync.Mutex
	initialized         int32
)

// GetDoubleCheckInstance 双重检查锁定获取实例
func GetDoubleCheckInstance() *DoubleCheckSingleton {
	if atomic.LoadInt32(&initialized) == 1 {
		return doubleCheckInstance
	}

	doubleCheckMutex.Lock()
	defer doubleCheckMutex.Unlock()

	if doubleCheckInstance == nil {
		doubleCheckInstance = &DoubleCheckSingleton{value: "double-check"}
		atomic.StoreInt32(&initialized, 1)
	}
	return doubleCheckInstance
}

// ========== 实际应用：数据库连接池单例 ==========

// DBConnectionPool 数据库连接池单例
type DBConnectionPool struct {
	connections []interface{}
	maxSize     int
	mu          sync.Mutex
}

var (
	dbPoolInstance *DBConnectionPool
	dbPoolOnce     sync.Once
)

// GetDBConnectionPool 获取数据库连接池单例
func GetDBConnectionPool(maxSize int) *DBConnectionPool {
	dbPoolOnce.Do(func() {
		dbPoolInstance = &DBConnectionPool{
			connections: make([]interface{}, 0, maxSize),
			maxSize:     maxSize,
		}
		fmt.Printf("[DB_POOL] Created connection pool with max size: %d\n", maxSize)
	})
	return dbPoolInstance
}

// GetConnection 获取连接
func (p *DBConnectionPool) GetConnection() interface{} {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.connections) == 0 {
		fmt.Println("[DB_POOL] Creating new connection")
		return fmt.Sprintf("connection-%d", len(p.connections)+1)
	}

	conn := p.connections[len(p.connections)-1]
	p.connections = p.connections[:len(p.connections)-1]
	fmt.Printf("[DB_POOL] Reusing connection: %v\n", conn)
	return conn
}

// ReleaseConnection 释放连接
func (p *DBConnectionPool) ReleaseConnection(conn interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.connections) < p.maxSize {
		p.connections = append(p.connections, conn)
		fmt.Printf("[DB_POOL] Released connection: %v\n", conn)
	} else {
		fmt.Println("[DB_POOL] Connection pool full, discarding connection")
	}
}

// ========== 实际应用：配置管理单例 ==========

// ConfigManager 配置管理单例
type ConfigManager struct {
	config map[string]string
	mu     sync.RWMutex
}

var (
	configInstance *ConfigManager
	configOnce     sync.Once
)

// GetConfigManager 获取配置管理单例
func GetConfigManager() *ConfigManager {
	configOnce.Do(func() {
		configInstance = &ConfigManager{
			config: make(map[string]string),
		}
		// 加载默认配置
		configInstance.config["app_name"] = "MyApp"
		configInstance.config["version"] = "1.0.0"
		configInstance.config["env"] = "development"
	})
	return configInstance
}

// Get 获取配置
func (c *ConfigManager) Get(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config[key]
}

// Set 设置配置
func (c *ConfigManager) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config[key] = value
	fmt.Printf("[CONFIG] Set %s = %s\n", key, value)
}

// GetAll 获取所有配置
func (c *ConfigManager) GetAll() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]string)
	for k, v := range c.config {
		result[k] = v
	}
	return result
}

// ========== 实际应用：日志管理单例 ==========

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

func (l LogLevel) String() string {
	return [...]string{"DEBUG", "INFO", "WARN", "ERROR"}[l]
}

// Logger 日志单例
type Logger struct {
	level LogLevel
	mu    sync.Mutex
}

var (
	loggerInstance *Logger
	loggerOnce     sync.Once
)

// GetLogger 获取日志单例
func GetLogger() *Logger {
	loggerOnce.Do(func() {
		loggerInstance = &Logger{
			level: INFO,
		}
	})
	return loggerInstance
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Log 记录日志
func (l *Logger) Log(level LogLevel, message string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level >= l.level {
		fmt.Printf("[%s] %s\n", level.String(), message)
	}
}

// Debug 记录调试日志
func (l *Logger) Debug(message string) {
	l.Log(DEBUG, message)
}

// Info 记录信息日志
func (l *Logger) Info(message string) {
	l.Log(INFO, message)
}

// Warn 记录警告日志
func (l *Logger) Warn(message string) {
	l.Log(WARN, message)
}

// Error 记录错误日志
func (l *Logger) Error(message string) {
	l.Log(ERROR, message)
}

// ========== 实际应用：缓存管理单例 ==========

// CacheItem 缓存项
type CacheItem struct {
	Value interface{}
}

// CacheManager 缓存管理单例
type CacheManager struct {
	items map[string]*CacheItem
	mu    sync.RWMutex
}

var (
	cacheInstance *CacheManager
	cacheOnce     sync.Once
)

// GetCacheManager 获取缓存管理单例
func GetCacheManager() *CacheManager {
	cacheOnce.Do(func() {
		cacheInstance = &CacheManager{
			items: make(map[string]*CacheItem),
		}
	})
	return cacheInstance
}

// Set 设置缓存
func (c *CacheManager) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = &CacheItem{Value: value}
	fmt.Printf("[CACHE] Set: %s\n", key)
}

// Get 获取缓存
func (c *CacheManager) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if item, ok := c.items[key]; ok {
		fmt.Printf("[CACHE] Hit: %s\n", key)
		return item.Value, true
	}

	fmt.Printf("[CACHE] Miss: %s\n", key)
	return nil, false
}

// Delete 删除缓存
func (c *CacheManager) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
	fmt.Printf("[CACHE] Deleted: %s\n", key)
}

// Clear 清空缓存
func (c *CacheManager) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*CacheItem)
	fmt.Println("[CACHE] Cleared all")
}

// ========== 泛型单例（Go 1.18+）==========

// SingletonHolder 泛型单例持有者
type SingletonHolder[T any] struct {
	instance *T
	once     sync.Once
	create   func() *T
}

// NewSingletonHolder 创建泛型单例持有者
func NewSingletonHolder[T any](create func() *T) *SingletonHolder[T] {
	return &SingletonHolder[T]{
		create: create,
	}
}

// Get 获取实例
func (h *SingletonHolder[T]) Get() *T {
	h.once.Do(func() {
		h.instance = h.create()
	})
	return h.instance
}

// ========== 实际应用：ID生成器单例 ==========

// IDGenerator ID生成器单例
type IDGenerator struct {
	counter int64
	mu      sync.Mutex
}

var (
	idGenInstance *IDGenerator
	idGenOnce     sync.Once
)

// GetIDGenerator 获取ID生成器单例
func GetIDGenerator() *IDGenerator {
	idGenOnce.Do(func() {
		idGenInstance = &IDGenerator{
			counter: 0,
		}
	})
	return idGenInstance
}

// NextID 生成下一个ID
func (g *IDGenerator) NextID() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.counter++
	return g.counter
}

// NextIDWithPrefix 生成带前缀的ID
func (g *IDGenerator) NextIDWithPrefix(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, g.NextID())
}

// ========== 实际应用：事件总线单例 ==========

// EventHandler 事件处理函数
type EventHandler func(data interface{})

// EventBus 事件总线单例
type EventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

var (
	eventBusInstance *EventBus
	eventBusOnce     sync.Once
)

// GetEventBus 获取事件总线单例
func GetEventBus() *EventBus {
	eventBusOnce.Do(func() {
		eventBusInstance = &EventBus{
			handlers: make(map[string][]EventHandler),
		}
	})
	return eventBusInstance
}

// Subscribe 订阅事件
func (b *EventBus) Subscribe(eventName string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
	fmt.Printf("[EVENT_BUS] Subscribed to: %s\n", eventName)
}

// Publish 发布事件
func (b *EventBus) Publish(eventName string, data interface{}) {
	b.mu.RLock()
	handlers := b.handlers[eventName]
	b.mu.RUnlock()

	fmt.Printf("[EVENT_BUS] Publishing: %s\n", eventName)
	for _, handler := range handlers {
		handler(data)
	}
}

// Unsubscribe 取消订阅
func (b *EventBus) Unsubscribe(eventName string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.handlers, eventName)
	fmt.Printf("[EVENT_BUS] Unsubscribed from: %s\n", eventName)
}

