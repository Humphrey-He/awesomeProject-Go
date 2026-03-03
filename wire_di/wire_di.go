package wire_di

import (
	"fmt"

	"github.com/google/wire"
)

// ========== Google Wire 依赖注入深入理解 ==========

/*
本项目讲解 Google Wire 依赖注入框架的原理和使用，包括：

一、依赖注入基础
1. 什么是依赖注入
2. 为什么需要依赖注入
3. 手动依赖注入

二、Wire 入门
1. Provider
2. Injector
3. 基本使用

三、Wire 进阶
1. 绑定接口
2. 约束
3. 清理函数

四、Wire 原理
1. 代码生成
2. 静态分析
*/

// ========== 1. 依赖注入基础 ==========

/*
依赖注入 (DI) 概念：

传统方式（紧耦合）：
type UserService struct {
    db *MySQL
}

func NewUserService() *UserService {
    return &UserService{db: NewMySQL()} // 内部创建依赖
}

依赖注入（松耦合）：
type UserService struct {
    db Database // 依赖抽象
}

func NewUserService(db Database) *UserService {
    return &UserService{db: db} // 外部注入
}

依赖注入优点：
1. 便于单元测试（可注入 mock）
2. 降低耦合
3. 提高可维护性
*/

// ========== 1.1 依赖示例 ==========

// Database 接口
type Database interface {
	Query(sql string) ([]string, error)
	Execute(sql string) error
}

// MySQL MySQL 实现
type MySQL struct {
	host string
	port int
}

// NewMySQL 创建 MySQL
func NewMySQL(host string, port int) *MySQL {
	return &MySQL{host: host, port: port}
}

func (m *MySQL) Query(sql string) ([]string, error) {
	return []string{"result"}, nil
}

func (m *MySQL) Execute(sql string) error {
	return nil
}

// UserService 用户服务
type UserService struct {
	db Database
}

// NewUserService 创建 UserService
func NewUserService(db Database) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetUser(id string) ([]string, error) {
	return s.db.Query("SELECT * FROM users WHERE id = " + id)
}

// ========== 2. Wire Provider ==========

/*
Wire 使用步骤：

1. 定义 Provider 函数
   func ProvideMySQL() *MySQL { ... }

2. 定义 Injector 函数
   func InitializeUserService() *UserService { ... }

3. 运行 wire generate

Wire 规则：
- Provider 函数必须返回至少一个值
- Injector 函数依赖由 Provider 提供
- Wire 生成代码调用 Provider
*/

// ProvideMySQL MySQL Provider
func ProvideMySQL() *MySQL {
	return NewMySQL("localhost", 3306)
}

// ProvideUserService UserService Provider
func ProvideUserService(db *MySQL) *UserService {
	return NewUserService(db)
}

// UserSet Wire Set（模块化）
var UserSet = wire.NewSet(
	ProvideMySQL,
	ProvideUserService,
)

// ========== 2.1 依赖图 ==========

/*
依赖关系：

MySQL -> UserService
  |
  v
Database (接口)

Wire 分析：
1. UserService 依赖 MySQL
2. MySQL 是具体类型，直接提供
3. Wire 自动组装
*/

// ========== 3. Wire 进阶 ==========

// ========== 3.1 接口绑定 ==========

// Logger 接口
type Logger interface {
	Log(msg string)
}

// ConsoleLogger 控制台日志
type ConsoleLogger struct{}

// NewConsoleLogger 创建控制台日志
func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{}
}

func (l *ConsoleLogger) Log(msg string) {
	fmt.Println("[LOG]", msg)
}

// FileLogger 文件日志
type FileLogger struct {
	filename string
}

// NewFileLogger 创建文件日志
func NewFileLogger(filename string) *FileLogger {
	return &FileLogger{filename: filename}
}

func (l *FileLogger) Log(msg string) {
	fmt.Println("[FILE]", l.filename, msg)
}

// ProvideLogger Logger Provider
func ProvideLogger() Logger {
	return NewConsoleLogger()
}

// LogSet 日志模块 Wire Set
var LogSet = wire.NewSet(
	NewConsoleLogger,
	wire.Bind(new(Logger), new(*ConsoleLogger)),
)

// ========== 3.2 结构体 Provider ==========

// Config 配置
type Config struct {
	Host string
	Port int
}

// NewConfig 创建配置
func NewConfig() *Config {
	return &Config{
		Host: "localhost",
		Port: 8080,
	}
}

// AppService 应用服务
type AppService struct {
	Config *Config
	DB     Database
	Logger Logger
}

// NewAppService 创建应用服务
func NewAppService(cfg *Config, db Database, logger Logger) *AppService {
	return &AppService{
		Config: cfg,
		DB:     db,
		Logger: logger,
	}
}

// AppSet 应用 Wire Set
var AppSet = wire.NewSet(
	NewConfig,
	UserSet,
	LogSet,
	NewAppService,
)

// ========== 3.3 清理函数 ==========

/*
清理函数：

func ProvideResource() (*Resource, func()) {
    r := NewResource()
    cleanup := func() {
        r.Close()
    }
    return r, cleanup
}

Wire 会：
1. 调用 Provider
2. 将清理函数保存
3. Injector 返回时调用
*/

// Resource 资源
type Resource struct {
	id int
}

// NewResource 创建资源
func NewResource(id int) *Resource {
	return &Resource{id: id}
}

func (r *Resource) Close() {
	fmt.Println("Resource closed:", r.id)
}

// ProvideResource 资源 Provider（带清理）
func ProvideResource(id int) (*Resource, func()) {
	resource := NewResource(id)
	cleanup := func() {
		resource.Close()
	}
	return resource, cleanup
}

// ========== 4. Wire 原理 ==========

/*
Wire 原理：

1. 代码生成
   - 分析 Provider 函数签名
   - 生成依赖组装代码
   - 静态类型检查

2. 编译时验证
   - 检查所有依赖是否提供
   - 检查类型匹配
   - 检查循环依赖

生成的代码示例：

func InitializeUserService() *UserService {
    mySQL := ProvideMySQL()
    return ProvideUserService(mySQL)
}
*/

// ========== 5. 手动组装示例 ==========

// ManualDI 手动依赖注入示例
func ManualDI() {
	fmt.Println("=== 手动依赖注入 ===")
	
	// 手动创建依赖
	config := NewConfig()
	db := NewMySQL(config.Host, config.Port)
	logger := NewConsoleLogger()
	
	// 注入依赖
	app := NewAppService(config, db, logger)
	
	fmt.Printf("AppService created: %+v\n", app)
}

// ========== 6. 测试中使用 ==========

/*
测试时替换 Provider：

1. 创建测试 Provider
   func ProvideMockDB() *MockDatabase { ... }

2. 在测试中覆盖
   userService := NewUserService(mockDB)
*/

// MockDatabase Mock 数据库
type MockDatabase struct {
	QueryResult []string
	QueryError  error
}

// NewMockDatabase 创建 Mock
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{}
}

func (m *MockDatabase) Query(sql string) ([]string, error) {
	return m.QueryResult, m.QueryError
}

func (m *MockDatabase) Execute(sql string) error {
	return nil
}

// TestWithMock 使用 Mock 测试
func TestWithMock() {
	fmt.Println("\n=== Mock 测试 ===")
	
	// 创建 mock
	mockDB := NewMockDatabase()
	mockDB.QueryResult = []string{"mock result"}
	
	// 使用 mock 创建 service
	service := NewUserService(mockDB)
	
	// 测试
	result, err := service.GetUser("1")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	
	fmt.Println("Result:", result)
}

// ========== 7. Wire 与其他 DI 对比 ==========

/*
Wire vs 其他 DI 框架：

| 特性 | Wire | Google Guice | Spring |
|------|------|--------------|--------|
| 运行时 | 编译时生成 | 运行时 | 运行时 |
| 性能 | 无开销 | 反射开销 | 反射开销 |
| 类型安全 | 编译时 | 运行时 | 运行时 |
| 学习曲线 | 低 | 中 | 高 |

Wire 优势：
1. 编译时错误检测
2. 无运行时反射
3. 简单易用
*/

// ========== 8. Wire 使用步骤 ==========

/*
Wire 使用步骤：

1. 安装 wire
   go install github.com/google/wire/cmd/wire@latest

2. 创建 wire.go 文件
   //go:build wireinject
   package main
   
   func InitializeApp() *AppService {
       wire.Build(AppSet)
       return nil
   }

3. 运行 wire 生成代码
   wire

4. 生成的代码
   func InitializeApp() *AppService {
       config := NewConfig()
       mysql := ProvideMySQL()
       logger := ProvideLogger()
       appService := NewAppService(config, mysql, logger)
       return appService
   }
*/

// ========== 9. 最佳实践 ==========

/*
最佳实践：

✅ 推荐：
1. 使用接口抽象依赖
2. 保持 Provider 简单
3. 模块化组织（使用 wire.NewSet）
4. 编写测试

❌ 避免：
1. 复杂 Provider 逻辑
2. 全局状态
3. 循环依赖
*/

// DemonstrateWire Wire 演示
func DemonstrateWire() {
	fmt.Println("=== Wire 依赖注入 ===")
	
	fmt.Println("\nWire 核心概念:")
	fmt.Println("  Provider: 提供依赖的函数")
	fmt.Println("  Injector: 组装依赖的函数")
	fmt.Println("  WireSet: 模块化的 Provider 集合")
	
	fmt.Println("\nWire 特点:")
	fmt.Println("  - 编译时代码生成")
	fmt.Println("  - 无运行时反射开销")
	fmt.Println("  - 类型安全")
	
	fmt.Println("\nWire 命令:")
	fmt.Println("  wire          # 生成 wire_gen.go")
	fmt.Println("  wire ./...    # 递归生成")
	fmt.Println("  wire check    # 检查但不生成")
}

// CompleteExample 完整示例
func CompleteExample() {
	DemonstrateWire()
	ManualDI()
	TestWithMock()
}
