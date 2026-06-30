# DDD Practices / 领域驱动设计实践

## 概述
该子项目以订单聚合为核心，展示实体、值对象、领域事件与仓储接口的协作关系。

## Overview
A DDD example centered on an Order aggregate, showing entities, value objects, domain events, and repositories.

## 目标
- 在聚合边界内维护业务不变量
- 在状态迁移时产生领域事件
- 将持久化隐藏在仓储接口之后

## Goals
- Enforce invariants inside the aggregate boundary
- Emit domain events on state transitions
- Keep persistence behind a repository interface

## 设计要点
- `Order` 为聚合根
- `Money` 为值对象，确保币种一致
- 事件用于集成或审计

## Design Highlights
- `Order` is the aggregate root
- `Money` is a value object with currency safety
- Events are captured for integration or auditing

## 测试
```bash
go test ./ddd_practices -v
```

## Tests
```bash
go test ./ddd_practices -v
```

## 扩展方向
- 增加应用服务（Use Case 层）
- 引入折扣规则的策略对象
- 集成 Outbox 进行事件发布

## Extension Ideas
- Add an application service (use case layer)
- Add policy objects for discount rules
- Integrate an outbox for event publishing
