# Event Sourcing MVP / 事件溯源 MVP

## 概述
该子项目实现一个最小化事件溯源聚合，包含内存事件仓储与基于版本的乐观并发控制。

## Overview
A minimal event-sourced aggregate with an in-memory event store and optimistic concurrency control.

## 目标
- 使用不可变事件表示状态变更
- 通过事件重放还原聚合状态
- 使用版本号进行并发控制

## Goals
- Capture state changes as immutable events
- Rebuild aggregate state by replaying events
- Enforce optimistic concurrency via stream version checks

## 测试
```bash
go test ./event_sourcing_mvp -v
```

## Tests
```bash
go test ./event_sourcing_mvp -v
```

## 扩展方向
- 增加快照以加速重放
- 引入 Outbox 发布事件
- 构建读模型/投影以优化查询

## Extension Ideas
- Add snapshots for faster rehydration
- Introduce an outbox for event publishing
- Add projections/read models for query optimization
