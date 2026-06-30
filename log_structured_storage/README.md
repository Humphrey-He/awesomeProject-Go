# Log-Structured Storage / 日志结构化存储

## 概述
该子项目实现一个最小化的 Append-only 日志存储，使用内存分段与索引指向最新记录，并提供 Compaction 回收旧数据。

## Overview
A minimal append-only log store with in-memory segments and an index pointing to the latest record per key. Compaction collapses old segments into a clean segment.

## 目标
- 理解追加写入与 tombstone 删除
- 理解索引如何定位最新值
- 理解 Compaction 的空间回收

## Goals
- Model append-only writes and tombstones
- Understand index lookups for latest values
- Practice compaction to reclaim space

## 测试
```bash
go test ./log_structured_storage -v
```

## Tests
```bash
go test ./log_structured_storage -v
```

## 扩展方向
- 加入 WAL 校验与恢复
- 使用 mmap-backed 段文件
- 增加基于大小/时间的分段滚动

## Extension Ideas
- Add WAL checksums and recovery
- Use mmap-backed segments
- Implement size/time-based segment rollover
