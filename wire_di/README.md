# Google Wire 依赖注入

## 概述

本项目讲解 Google Wire 依赖注入框架的原理和使用，展示如何通过编译时代码生成实现依赖注入。

## 核心内容

### 1. 依赖注入基础

- 传统方式（紧耦合）：内部创建依赖
- 依赖注入（松耦合）：外部注入依赖

### 2. Wire Provider

```go
// Provider 函数
func ProvideMySQL() *MySQL {
    return NewMySQL("localhost", 3306)
}

// Wire Set
var UserSet = wire.NewSet(
    ProvideMySQL,
    ProvideUserService,
)
```

### 3. 接口绑定

```go
var LogSet = wire.NewSet(
    NewConsoleLogger,
    wire.Bind(new(Logger), new(*ConsoleLogger)),
)
```

### 4. 清理函数

```go
func ProvideResource() (*Resource, func()) {
    resource := NewResource()
    cleanup := func() { resource.Close() }
    return resource, cleanup
}
```

## Wire 优势

- **编译时生成**：无运行时反射开销
- **类型安全**：编译时检查
- **简单易用**：学习曲线低

## 安装

```bash
go install github.com/google/wire/cmd/wire@latest
```

## 使用步骤

1. 创建 wire.go 文件
```go
//go:build wireinject
func InitializeApp() *AppService {
    wire.Build(AppSet)
    return nil
}
```

2. 运行 wire 生成代码
```bash
wire
```

## 对比其他 DI 框架

| 特性 | Wire | Google Guice | Spring |
|------|------|--------------|--------|
| 运行时 | 编译时生成 | 运行时 | 运行时 |
| 性能 | 无开销 | 反射开销 | 反射开销 |
| 类型安全 | 编译时 | 运行时 | 运行时 |

## 关联项目

- [testify_practices](../testify_practices) - 测试框架
- [mutex_advanced](../mutex_advanced) - 并发控制
