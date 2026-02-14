package design_patterns

import (
	"fmt"
	"time"
)

// ========== 基础装饰器模式 ==========

// Component 组件接口
type Component interface {
	Operation() string
}

// ConcreteComponent 具体组件
type ConcreteComponent struct {
	name string
}

func (c *ConcreteComponent) Operation() string {
	return fmt.Sprintf("ConcreteComponent: %s", c.name)
}

// Decorator 装饰器基类
type Decorator struct {
	component Component
}

func (d *Decorator) Operation() string {
	if d.component != nil {
		return d.component.Operation()
	}
	return ""
}

// ConcreteDecoratorA 具体装饰器A
type ConcreteDecoratorA struct {
	Decorator
}

func NewConcreteDecoratorA(component Component) *ConcreteDecoratorA {
	return &ConcreteDecoratorA{
		Decorator: Decorator{component: component},
	}
}

func (d *ConcreteDecoratorA) Operation() string {
	return fmt.Sprintf("DecoratorA(%s)", d.Decorator.Operation())
}

// ConcreteDecoratorB 具体装饰器B
type ConcreteDecoratorB struct {
	Decorator
}

func NewConcreteDecoratorB(component Component) *ConcreteDecoratorB {
	return &ConcreteDecoratorB{
		Decorator: Decorator{component: component},
	}
}

func (d *ConcreteDecoratorB) Operation() string {
	return fmt.Sprintf("DecoratorB[%s]", d.Decorator.Operation())
}

// ========== 实际应用：HTTP处理器装饰器 ==========

// Handler HTTP处理器接口
type Handler interface {
	Handle(request string) string
}

// BaseHandler 基础处理器
type BaseHandler struct{}

func (h *BaseHandler) Handle(request string) string {
	return fmt.Sprintf("Processing: %s", request)
}

// LoggingDecorator 日志装饰器
type LoggingDecorator struct {
	handler Handler
}

func NewLoggingDecorator(handler Handler) *LoggingDecorator {
	return &LoggingDecorator{handler: handler}
}

func (d *LoggingDecorator) Handle(request string) string {
	fmt.Printf("[LOG] Request received: %s at %s\n", request, time.Now().Format(time.RFC3339))
	result := d.handler.Handle(request)
	fmt.Printf("[LOG] Request processed: %s\n", result)
	return result
}

// AuthDecorator 认证装饰器
type AuthDecorator struct {
	handler Handler
	token   string
}

func NewAuthDecorator(handler Handler, token string) *AuthDecorator {
	return &AuthDecorator{
		handler: handler,
		token:   token,
	}
}

func (d *AuthDecorator) Handle(request string) string {
	if d.token == "" {
		return "Error: Authentication required"
	}
	fmt.Printf("[AUTH] Token validated: %s\n", d.token)
	return d.handler.Handle(request)
}

// RateLimitDecorator 限流装饰器
type RateLimitDecorator struct {
	handler   Handler
	maxRate   int
	current   int
	resetTime time.Time
}

func NewRateLimitDecorator(handler Handler, maxRate int) *RateLimitDecorator {
	return &RateLimitDecorator{
		handler:   handler,
		maxRate:   maxRate,
		resetTime: time.Now().Add(time.Minute),
	}
}

func (d *RateLimitDecorator) Handle(request string) string {
	now := time.Now()
	if now.After(d.resetTime) {
		d.current = 0
		d.resetTime = now.Add(time.Minute)
	}
	
	if d.current >= d.maxRate {
		return "Error: Rate limit exceeded"
	}
	
	d.current++
	fmt.Printf("[RATE_LIMIT] Request count: %d/%d\n", d.current, d.maxRate)
	return d.handler.Handle(request)
}

// CacheDecorator 缓存装饰器
type CacheDecorator struct {
	handler Handler
	cache   map[string]string
}

func NewCacheDecorator(handler Handler) *CacheDecorator {
	return &CacheDecorator{
		handler: handler,
		cache:   make(map[string]string),
	}
}

func (d *CacheDecorator) Handle(request string) string {
	// 检查缓存
	if result, ok := d.cache[request]; ok {
		fmt.Printf("[CACHE] Cache hit for: %s\n", request)
		return result
	}
	
	// 缓存未命中，调用实际处理器
	fmt.Printf("[CACHE] Cache miss for: %s\n", request)
	result := d.handler.Handle(request)
	d.cache[request] = result
	
	return result
}

// ========== 实际应用：数据流处理装饰器 ==========

// DataProcessor 数据处理器接口
type DataProcessor interface {
	Process(data string) string
}

// BaseDataProcessor 基础数据处理器
type BaseDataProcessor struct{}

func (p *BaseDataProcessor) Process(data string) string {
	return data
}

// EncryptionDecorator 加密装饰器
type EncryptionDecorator struct {
	processor DataProcessor
}

func NewEncryptionDecorator(processor DataProcessor) *EncryptionDecorator {
	return &EncryptionDecorator{processor: processor}
}

func (d *EncryptionDecorator) Process(data string) string {
	processed := d.processor.Process(data)
	// 模拟加密
	return fmt.Sprintf("ENCRYPTED(%s)", processed)
}

// CompressionDecorator 压缩装饰器
type CompressionDecorator struct {
	processor DataProcessor
}

func NewCompressionDecorator(processor DataProcessor) *CompressionDecorator {
	return &CompressionDecorator{processor: processor}
}

func (d *CompressionDecorator) Process(data string) string {
	processed := d.processor.Process(data)
	// 模拟压缩
	return fmt.Sprintf("COMPRESSED(%s)", processed)
}

// ValidationDecorator 验证装饰器
type ValidationDecorator struct {
	processor DataProcessor
}

func NewValidationDecorator(processor DataProcessor) *ValidationDecorator {
	return &ValidationDecorator{processor: processor}
}

func (d *ValidationDecorator) Process(data string) string {
	// 模拟验证
	if data == "" {
		return "Error: empty data"
	}
	
	fmt.Println("[VALIDATION] Data validated")
	return d.processor.Process(data)
}

// ========== 函数式装饰器（Go风格）==========

// ProcessFunc 处理函数类型
type ProcessFunc func(string) string

// WithLogging 日志装饰器（函数式）
func WithLogging(next ProcessFunc) ProcessFunc {
	return func(s string) string {
		fmt.Printf("[LOG] Input: %s\n", s)
		result := next(s)
		fmt.Printf("[LOG] Output: %s\n", result)
		return result
	}
}

// WithTiming 计时装饰器（函数式）
func WithTiming(next ProcessFunc) ProcessFunc {
	return func(s string) string {
		start := time.Now()
		result := next(s)
		fmt.Printf("[TIMING] Duration: %v\n", time.Since(start))
		return result
	}
}

// WithRetry 重试装饰器（函数式）
func WithRetry(maxRetries int) func(ProcessFunc) ProcessFunc {
	return func(next ProcessFunc) ProcessFunc {
		return func(s string) string {
			var result string
			for i := 0; i < maxRetries; i++ {
				result = next(s)
				if result != "" {
					return result
				}
				fmt.Printf("[RETRY] Attempt %d failed\n", i+1)
			}
			return result
		}
	}
}

// ========== 中间件模式（HTTP装饰器）==========

// Middleware HTTP中间件类型
type Middleware func(Handler) Handler

// ChainMiddleware 链接多个中间件
func ChainMiddleware(handler Handler, middlewares ...Middleware) Handler {
	// 从后向前应用中间件
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware(next Handler) Handler {
	return HandlerFunc(func(request string) string {
		fmt.Printf("[MIDDLEWARE] Logging: %s\n", request)
		return next.Handle(request)
	})
}

// AuthMiddleware 认证中间件
func AuthMiddleware(token string) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(request string) string {
			if token == "" {
				return "Error: No token"
			}
			fmt.Printf("[MIDDLEWARE] Auth passed\n")
			return next.Handle(request)
		})
	}
}

// HandlerFunc 函数适配器
type HandlerFunc func(string) string

func (f HandlerFunc) Handle(request string) string {
	return f(request)
}

// ========== 实际应用：数据库连接装饰器 ==========

// DBConnection 数据库连接接口
type DBConnection interface {
	Query(sql string) (interface{}, error)
}

// BasicDBConnection 基础数据库连接
type BasicDBConnection struct {
	name string
}

func (c *BasicDBConnection) Query(sql string) (interface{}, error) {
	return fmt.Sprintf("Query result for: %s", sql), nil
}

// PooledDBConnection 连接池装饰器
type PooledDBConnection struct {
	conn     DBConnection
	poolSize int
}

func NewPooledDBConnection(conn DBConnection, poolSize int) *PooledDBConnection {
	return &PooledDBConnection{
		conn:     conn,
		poolSize: poolSize,
	}
}

func (c *PooledDBConnection) Query(sql string) (interface{}, error) {
	fmt.Printf("[POOL] Using connection from pool (size: %d)\n", c.poolSize)
	return c.conn.Query(sql)
}

// CachedDBConnection 缓存查询装饰器
type CachedDBConnection struct {
	conn  DBConnection
	cache map[string]interface{}
}

func NewCachedDBConnection(conn DBConnection) *CachedDBConnection {
	return &CachedDBConnection{
		conn:  conn,
		cache: make(map[string]interface{}),
	}
}

func (c *CachedDBConnection) Query(sql string) (interface{}, error) {
	if result, ok := c.cache[sql]; ok {
		fmt.Println("[DB_CACHE] Cache hit")
		return result, nil
	}
	
	fmt.Println("[DB_CACHE] Cache miss")
	result, err := c.conn.Query(sql)
	if err == nil {
		c.cache[sql] = result
	}
	
	return result, err
}

// RetryDBConnection 重试装饰器
type RetryDBConnection struct {
	conn       DBConnection
	maxRetries int
}

func NewRetryDBConnection(conn DBConnection, maxRetries int) *RetryDBConnection {
	return &RetryDBConnection{
		conn:       conn,
		maxRetries: maxRetries,
	}
}

func (c *RetryDBConnection) Query(sql string) (interface{}, error) {
	var result interface{}
	var err error
	
	for i := 0; i < c.maxRetries; i++ {
		result, err = c.conn.Query(sql)
		if err == nil {
			return result, nil
		}
		fmt.Printf("[DB_RETRY] Attempt %d failed\n", i+1)
	}
	
	return nil, err
}

