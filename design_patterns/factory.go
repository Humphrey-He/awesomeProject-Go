package design_patterns

import (
	"fmt"
)

// ========== 简单工厂模式 ==========

// ProductForBuilder 产品接口
type Product interface {
	Use() string
}

// ConcreteProductA 具体产品A
type ConcreteProductA struct {
	name string
}

func (p *ConcreteProductA) Use() string {
	return fmt.Sprintf("Using ProductForBuilder A: %s", p.name)
}

// ConcreteProductB 具体产品B
type ConcreteProductB struct {
	name string
}

func (p *ConcreteProductB) Use() string {
	return fmt.Sprintf("Using ProductForBuilder B: %s", p.name)
}

// ProductType 产品类型
type ProductType string

const (
	ProductTypeA ProductType = "A"
	ProductTypeB ProductType = "B"
)

// SimpleFactory 简单工厂
type SimpleFactory struct{}

// NewSimpleFactory 创建简单工厂
func NewSimpleFactory() *SimpleFactory {
	return &SimpleFactory{}
}

// CreateProduct 创建产品
func (f *SimpleFactory) CreateProduct(productType ProductType) (Product, error) {
	switch productType {
	case ProductTypeA:
		return &ConcreteProductA{name: "Default A"}, nil
	case ProductTypeB:
		return &ConcreteProductB{name: "Default B"}, nil
	default:
		return nil, fmt.Errorf("unknown product type: %s", productType)
	}
}

// ========== 工厂方法模式 ==========

// Factory 工厂接口
type Factory interface {
	CreateProduct() Product
}

// FactoryA 工厂A
type FactoryA struct{}

func (f *FactoryA) CreateProduct() Product {
	return &ConcreteProductA{name: "Factory A ProductForBuilder"}
}

// FactoryB 工厂B
type FactoryB struct{}

func (f *FactoryB) CreateProduct() Product {
	return &ConcreteProductB{name: "Factory B ProductForBuilder"}
}

// ========== 抽象工厂模式 ==========

// AbstractFactory 抽象工厂接口
type AbstractFactory interface {
	CreateButton() Button
	CreateTextBox() TextBox
}

// UI组件接口
type Button interface {
	Render() string
}

type TextBox interface {
	Render() string
}

// Windows风格组件
type WindowsButton struct{}

func (b *WindowsButton) Render() string {
	return "Rendering Windows Button"
}

type WindowsTextBox struct{}

func (t *WindowsTextBox) Render() string {
	return "Rendering Windows TextBox"
}

// Mac风格组件
type MacButton struct{}

func (b *MacButton) Render() string {
	return "Rendering Mac Button"
}

type MacTextBox struct{}

func (t *MacTextBox) Render() string {
	return "Rendering Mac TextBox"
}

// Windows工厂
type WindowsFactory struct{}

func (f *WindowsFactory) CreateButton() Button {
	return &WindowsButton{}
}

func (f *WindowsFactory) CreateTextBox() TextBox {
	return &WindowsTextBox{}
}

// Mac工厂
type MacFactory struct{}

func (f *MacFactory) CreateButton() Button {
	return &MacButton{}
}

func (f *MacFactory) CreateTextBox() TextBox {
	return &MacTextBox{}
}

// ========== 实际应用：数据库连接工厂 ==========

// Database 数据库接口
type Database interface {
	Connect() error
	Query(sql string) (interface{}, error)
	Close() error
	Type() string
}

// MySQLDB MySQL数据库
type MySQLDB struct {
	host string
	port int
}

func (db *MySQLDB) Connect() error {
	fmt.Printf("Connecting to MySQL at %s:%d\n", db.host, db.port)
	return nil
}

func (db *MySQLDB) Query(sql string) (interface{}, error) {
	return fmt.Sprintf("MySQL query: %s", sql), nil
}

func (db *MySQLDB) Close() error {
	fmt.Println("Closing MySQL connection")
	return nil
}

func (db *MySQLDB) Type() string {
	return "MySQL"
}

// PostgreSQLDatabase PostgreSQL数据库
type PostgreSQLDatabase struct {
	host string
	port int
}

func (db *PostgreSQLDatabase) Connect() error {
	fmt.Printf("Connecting to PostgreSQL at %s:%d\n", db.host, db.port)
	return nil
}

func (db *PostgreSQLDatabase) Query(sql string) (interface{}, error) {
	return fmt.Sprintf("PostgreSQL query: %s", sql), nil
}

func (db *PostgreSQLDatabase) Close() error {
	fmt.Println("Closing PostgreSQL connection")
	return nil
}

func (db *PostgreSQLDatabase) Type() string {
	return "PostgreSQL"
}

// DatabaseType 数据库类型
type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgresql"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type DatabaseType
	Host string
	Port int
}

// DatabaseFactory 数据库工厂
type DatabaseFactory struct{}

// NewDatabaseFactory 创建数据库工厂
func NewDatabaseFactory() *DatabaseFactory {
	return &DatabaseFactory{}
}

// CreateDatabase 创建数据库连接
func (f *DatabaseFactory) CreateDatabase(config DatabaseConfig) (Database, error) {
	switch config.Type {
	case MySQL:
		return &MySQLDB{
			host: config.Host,
			port: config.Port,
		}, nil
	case PostgreSQL:
		return &PostgreSQLDatabase{
			host: config.Host,
			port: config.Port,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}

// ========== 实际应用：HTTP客户端工厂 ==========

// HTTPClient HTTP客户端接口
type HTTPClient interface {
	Get(url string) (string, error)
	Post(url string, data interface{}) (string, error)
	Name() string
}

// StandardHTTPClient 标准HTTP客户端
type StandardHTTPClient struct {
	timeout int
}

func (c *StandardHTTPClient) Get(url string) (string, error) {
	return fmt.Sprintf("Standard GET: %s (timeout: %ds)", url, c.timeout), nil
}

func (c *StandardHTTPClient) Post(url string, data interface{}) (string, error) {
	return fmt.Sprintf("Standard POST: %s with data: %v", url, data), nil
}

func (c *StandardHTTPClient) Name() string {
	return "StandardHTTP"
}

// RetryHTTPClient 带重试的HTTP客户端
type RetryHTTPClient struct {
	timeout    int
	maxRetries int
}

func (c *RetryHTTPClient) Get(url string) (string, error) {
	return fmt.Sprintf("Retry GET: %s (retries: %d)", url, c.maxRetries), nil
}

func (c *RetryHTTPClient) Post(url string, data interface{}) (string, error) {
	return fmt.Sprintf("Retry POST: %s with data: %v (retries: %d)", url, data, c.maxRetries), nil
}

func (c *RetryHTTPClient) Name() string {
	return "RetryHTTP"
}

// HTTPClientType 客户端类型
type HTTPClientType string

const (
	StandardClient HTTPClientType = "standard"
	RetryClient    HTTPClientType = "retry"
)

// HTTPClientConfig 客户端配置
type HTTPClientConfig struct {
	Type       HTTPClientType
	Timeout    int
	MaxRetries int
}

// HTTPClientFactory HTTP客户端工厂
type HTTPClientFactory struct{}

// NewHTTPClientFactory 创建HTTP客户端工厂
func NewHTTPClientFactory() *HTTPClientFactory {
	return &HTTPClientFactory{}
}

// CreateClient 创建HTTP客户端
func (f *HTTPClientFactory) CreateClient(config HTTPClientConfig) (HTTPClient, error) {
	switch config.Type {
	case StandardClient:
		return &StandardHTTPClient{
			timeout: config.Timeout,
		}, nil
	case RetryClient:
		return &RetryHTTPClient{
			timeout:    config.Timeout,
			maxRetries: config.MaxRetries,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported client type: %s", config.Type)
	}
}

// ========== 注册表模式（可扩展工厂）==========

// Creator 创建器函数类型
type Creator func() Product

// Registry 注册表工厂
type Registry struct {
	creators map[string]Creator
}

// NewRegistry 创建注册表
func NewRegistry() *Registry {
	return &Registry{
		creators: make(map[string]Creator),
	}
}

// Register 注册产品创建器
func (r *Registry) Register(name string, creator Creator) {
	r.creators[name] = creator
}

// Create 创建产品
func (r *Registry) Create(name string) (Product, error) {
	creator, ok := r.creators[name]
	if !ok {
		return nil, fmt.Errorf("no creator registered for: %s", name)
	}
	return creator(), nil
}

// GetRegisteredTypes 获取所有已注册的类型
func (r *Registry) GetRegisteredTypes() []string {
	types := make([]string, 0, len(r.creators))
	for t := range r.creators {
		types = append(types, t)
	}
	return types
}
