# Memory Model Practices / 内存模型实践

## 概述
该子项目聚焦安全发布（safe publication）与一次性执行语义，通过 `atomic.Pointer` 与原子布尔实现轻量同步。

## Overview
This subproject focuses on safe publication and one-time execution semantics via `atomic.Pointer` and atomic booleans.

## 目标
- 使用原子指针安全发布不可变配置
- 理解 happens-before 的可见性保证
- 实现一次性执行的轻量原语

## Goals
- Publish immutable configs safely with atomic pointers
- Understand happens-before visibility guarantees
- Implement a lightweight one-time execution primitive

## 备注
- `Store` 后的 `Load` 对已发布对象建立可见性关系

## Notes
- A `Store` followed by a `Load` establishes visibility for the published object

## 测试
```bash
go test ./memory_model_practices -v
```

## Tests
```bash
go test ./memory_model_practices -v
```

## 扩展方向
- 基于版本号的读多写少配置快照
- 对比 `atomic.Value` 与 `atomic.Pointer`
- 基于 Release/Acquire 构建无锁环形队列

## Extension Ideas
- Read-mostly snapshots with versioning
- Compare `atomic.Value` vs `atomic.Pointer`
- Build a lock-free ring buffer with release/acquire semantics
