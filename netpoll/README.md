# Go Netpoll 网络轮询器

## 概述

本项目深入讲解 Go 运行时网络轮询器（Netpoll）的实现原理，展示 Go 如何实现高效的异步网络 IO。

## 核心内容

### 1. Netpoll 架构

```
┌─────────────────────────────────────────────────────────────┐
│                      Go Netpoll 架构                        │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────┐    ┌──────────┐    ┌──────────┐            │
│  │ Goroutine│    │ Goroutine│    │ Goroutine│            │
│  └────┬─────┘    └────┬─────┘    └────┬─────┘            │
│       ▼               ▼               ▼                   │
│  ┌─────────────────────────────────────────────┐          │
│  │              pollDesc (per fd)              │          │
│  │  - rg: 读等待的 goroutine                   │          │
│  │  - wg: 写等待的 goroutine                   │          │
│  └────────────────────┬──────────────────────┘          │
│                       ▼                                  │
│  ┌─────────────────────────────────────────────┐          │
│  │            Netpoll (epoll/kqueue)           │          │
│  └────────────────────┬──────────────────────┘          │
│                       ▼                                  │
│  ┌─────────────────────────────────────────────┐          │
│  │              OS Kernel                       │          │
│  └─────────────────────────────────────────────┘          │
└─────────────────────────────────────────────────────────────┘
```

### 2. Poll Descriptor 状态

| 状态 | 值 | 说明 |
|------|-----|------|
| pdNil | 0 | 初始状态，无等待 |
| pdReady | 1 | IO就绪通知待消费 |
| pdWait | 2 | goroutine 准备等待 |

### 3. 核心组件

- **PollDesc**: 网络描述符，管理读写等待状态
- **Uintptr**: 原子操作封装
- **GoroutineList**: 就绪 goroutine 列表
- **PollError**: 错误码定义

### 4. 关键函数

| 函数 | 说明 |
|------|------|
| PollReset | 重置 poll 描述符 |
| PollWait | 等待 IO 就绪 |
| NetpollReady | IO 就绪通知 |
| SetDeadline | 设置截止时间 |

### 5. Goroutine IO 流程

1. 调用 Read/Write
2. PollWait 等待 IO
3. netpollBlock 设置 pdWait
4. gopark 挂起 goroutine
5. IO 就绪 → goready 唤醒

## 平台支持

| 平台 | 机制 |
|------|------|
| Linux | epoll |
| macOS/BSD | kqueue |
| Windows | IOCP |

## 性能特点

- **非阻塞 IO**: 不阻塞 OS 线程
- **百万连接**: 支持高并发连接
- **边沿触发**: 高效的事件通知
- **GMP 集成**: 与调度器完美配合

## 面试要点

- Go 网络模型与传统阻塞模型的区别
- pollDesc 状态转换过程
- 边沿触发 vs 水平触发
- netpoll 与 GMP 调度的交互

## 关联项目

- [runtime_core](../runtime_core) - GMP/GC
- [goroutine_practices](../goroutine_practices) - 并发编程
- [channel_patterns](../channel_patterns) - 通道模式
