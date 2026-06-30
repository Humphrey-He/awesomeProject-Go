# Raft Practices / Raft 实践

## 概述
该子项目实现一个精简的 Raft 内核，聚焦任期推进、投票规则与日志匹配。网络、定时器、快照等被刻意省略以突出核心机制。

## Overview
A compact Raft core focusing on term management, voting rules, and log matching. Networking, timers, and snapshots are omitted to keep the mechanics clear.

## 目标
- 理解任期推进与 Leader 选举
- 理解日志新旧判断与投票准则
- 理解 AppendEntries 的匹配与追加规则

## Goals
- Understand term advancement and leader election
- Validate log up-to-date checks for voting
- Apply AppendEntries log matching rules

## 测试
```bash
go test ./raft_practices -v
```

## Tests
```bash
go test ./raft_practices -v
```

## 扩展方向
- 增加心跳与随机选举超时
- 实现快照与日志压缩
- 加入内存传输层模拟集群

## Extension Ideas
- Add heartbeat and randomized election timeouts
- Implement snapshots and log compaction
- Add an in-memory transport to simulate a cluster
