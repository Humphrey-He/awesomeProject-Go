# LSM Tree / LSM 树

## 概述
该子项目实现一个最小化 LSM Tree，包含 Memtable、SSTable 与 Compaction 的核心路径，强调写放大与读取顺序。

## Overview
A compact LSM Tree with memtable, SSTables, and compaction to illustrate the write/read paths and key-version ordering.

## 目标
- 模拟写路径：memtable -> flush -> SSTable
- 模拟读路径：memtable 优先，SSTable 从新到旧
- 理解 Compaction 去重与回收

## Goals
- Model write path: memtable -> flush -> SSTable
- Model read path: memtable first, then newest SSTables
- Understand compaction and key deduplication

## 设计要点
- Memtable 使用 map，达到阈值后 Flush
- SSTable 使用有序数组表示
- Compaction 保留最新版本

## Design Highlights
- Memtable as a map with size threshold
- SSTable as a sorted array
- Compaction keeps the newest version per key

## 测试
```bash
go test ./lsm_tree -v
```

## Tests
```bash
go test ./lsm_tree -v
```

## 扩展方向
- 为 SSTable 加入 Bloom Filter
- 实现分层 Compaction 策略
- 增加 WAL 以支持崩溃恢复

## Extension Ideas
- Add bloom filters per SSTable
- Implement leveled compaction
- Add WAL for durability
