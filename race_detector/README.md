# Go Race 检测

## 概述

本项目讲解 Go race 检测器的原理和使用，展示数据竞争的产生和解决方法。

## 核心内容

### 1. 数据竞争

- 写-写竞争
- 读-写竞争
- 竞争条件类型

### 2. Race 检测器

- 编译时插桩
- 阴影变量
- 运行时检测

### 3. 解决方案

- Mutex 互斥锁
- RWMutex 读写锁
- atomic 原子操作
- channel 通道
- sync.Once

### 4. 检测方法

```bash
go run -race main.go
go test -race ./...
```

## 工作原理

```
┌─────────────────────────────────────────┐
│           原始代码                       │
│  goroutine A: x = 1                    │
│  goroutine B: y = x + 1                │
├─────────────────────────────────────────┤
│           插桩后                         │
│  goroutine A: racefunclog(&x, Write)   │
│  goroutine B: racefunclog(&x, Read)    │
└─────────────────────────────────────────┘
```

## 性能特点

- 开发时启用
- 生产关闭
- 内存 2-10 倍，CPU 5-10 倍开销

## 关联项目

- [mutex_advanced](../mutex_advanced) - Mutex 原理
- [goroutine_practices](../goroutine_practices) - 并发编程
