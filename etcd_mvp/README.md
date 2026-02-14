# etcd MVP（选举 + 心跳）

## 范围

这是一个简化版 etcd/raft 核心机制演示：

- 成员管理：`AddMember/RemoveMember`
- Leader 选举：超时后重选
- Leader 心跳租约：`Heartbeat`
- Leader-only 写路径：`Put`
- Key-Value 读取：`Get`

## 机制说明

- 使用 `heartbeatTimeout` 作为 leader 租约
- `Tick(now)` 检查 leader 是否超时
- 超时后触发选举（MVP 中采用确定性策略：按成员 ID 排序选最小）
- `Put` 只允许 leader 写，follower 返回 `ErrNotLeader`

> 注意：这是教学 MVP，不包含真实 Raft 的日志复制、投票多数派、持久化 WAL、快照、成员变更一致性协议等。

## 运行

```bash
go test ./etcd_mvp -v
go test ./etcd_mvp -run TestDemoOutput -v
```


