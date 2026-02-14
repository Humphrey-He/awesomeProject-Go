# Go 配置中心 + 注册中心 MVP

## 能力范围

### 配置中心（ConfigCenter）

- 配置项 `Set/Get`
- 自动版本号（`Version`）
- key 级 Watch（订阅变更事件）
- 快照导出（`Snapshot`）

### 注册中心（Registry）

- 服务实例 `Register/Discover`
- 心跳续租（`Heartbeat`）
- 主动下线（`Deregister`）
- 过期自动剔除（后台 cleaner）

## 典型流程

1. 配置中心：
   - 服务启动时拉取快照
   - 订阅关键 key（数据库地址、开关项）
   - 收到变更后热更新
2. 注册中心：
   - 实例启动注册 + 周期心跳
   - 调用方通过 `Discover(service)` 获取可用实例

## 运行

```bash
go test ./config_registry -v
```


