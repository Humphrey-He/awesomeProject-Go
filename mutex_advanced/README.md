# Go Mutex/RWMutex 进阶

## 概述

本项目深入讲解 Go 同步原语的底层实现原理，包括 Mutex、RWMutex、WaitGroup、Once 等。

## 核心内容

### 1. Mutex 互斥锁

```
状态位：
- bit 0: locked (1=已锁定)
- bit 1: woken (1=唤醒状态)
- 高位: waiters count (等待者数量)
```

- 锁状态与等待队列
- 公平锁 vs 非公平锁
- 自旋机制

### 2. RWMutex 读写锁

| 操作 | 说明 |
|------|------|
| RLock | readerCount++ |
| RUnlock | readerCount-- |
| Lock | 获取互斥锁，等待读者为0 |
| Unlock | 释放互斥锁 |

### 3. WaitGroup 原理

```go
type WaitGroup struct {
    state atomic.Uint64 // 高32位:counter, 低32位:waiters
}
```

- Add(delta): 增加计数器
- Done(): 减少计数器
- Wait(): 等待计数器为0

### 4. Once 原理

```go
type Once struct {
    m    Mutex
    done uint32
}
```

- 保证函数只执行一次
- 使用双重检查锁定

## 性能优化

### 锁粒度控制

```go
// ❌ 错误：锁内做耗时操作
mu.Lock()
data := expensiveOperation()
mu.Unlock()

// ✅ 正确：锁外处理，锁内只更新
data := expensiveOperation()
mu.Lock()
sharedData = data
mu.Unlock()
```

### 死锁避免

- 固定顺序获取锁
- 使用超时
- 避免嵌套锁

### 分片优化

```go
// 分散锁竞争
type OptimizedCounter struct {
    counters [10]int64
}
```

## 面试要点

| 问题 | 答案 |
|------|------|
| Go Mutex 是公平锁吗？ | 默认非公平，新请求可能插队 |
| Mutex vs RWMutex | Mutex 独占，RWMutex 读写分离 |
| 自旋锁优点 | 避免 goroutine 切换开销 |
| 如何避免死锁 | 固定顺序、超时、避免嵌套 |

## 关联项目

- [cond_practices](../cond_practices) - 条件变量
- [race_detector](../race_detector) - 竞态检测
- [goroutine_practices](../goroutine_practices) - 并发编程
