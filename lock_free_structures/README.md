# Lock-Free Structures / 无锁结构

## 概述
该子项目实现一个基于 CAS 的无锁栈，展示高并发场景下的原子更新模式与 LIFO 行为。

## Overview
A lock-free stack built with CAS loops and atomic pointers, a common building block for low-latency systems.

## 目标
- 练习 CAS 循环更新模式
- 理解无锁结构的 LIFO 行为
- 观察并发压测下的正确性

## Goals
- Practice CAS-based updates
- Understand LIFO behavior without locks
- Observe correctness under concurrency

## 测试
```bash
go test ./lock_free_structures -v
```

## Tests
```bash
go test ./lock_free_structures -v
```

## 扩展方向
- 添加消除退避（elimination backoff）
- 实现无锁队列（Michael-Scott）
- 加入 ABA 防护策略

## Extension Ideas
- Add elimination backoff
- Implement a lock-free queue (Michael-Scott)
- Add ABA prevention techniques
