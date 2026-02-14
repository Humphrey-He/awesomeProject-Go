# Docker/K8s 核心部分 MVP（Go）

这个目录是一个 **K8s 核心调度/部署能力的最小可运行版本**，用于理解核心机制，而不是替代真实 K8s。

## MVP 覆盖范围

- Node 资源模型（CPU/内存）
- Pod 规格与状态（Running/Pending）
- Deployment（声明副本数 + 模板）
- 调度器：过滤 + 打分（Ready/NodeSelector/Taint-Toleration）
- Pending Reconcile（资源恢复后重调度）
- Deployment 状态统计（类似 ready/desired 概念）
- 滚动更新（`RolloutImage`，含 `maxUnavailable/maxSurge` 简化语义）
- Service 选择器与轮询路由（Round Robin）

## 核心映射（概念对应）

- `Node` -> k8s Node
- `PodSpec/Pod` -> Pod template / Pod
- `Deployment` -> Deployment controller（简化）
- `ExposeService/Route` -> Service + kube-proxy（简化）

## 关键行为

1. `Deploy`：
   - 根据过滤规则和资源打分挑选 Node
   - 不可调度时 Pod 标记为 `Pending`
2. `ReconcilePending`：
   - 节点恢复/扩容后可重新尝试调度 Pending Pods
3. `Scale`：
   - 支持扩缩容
4. `RolloutImage`：
   - 按 `maxUnavailable/maxSurge` 执行简化滚动发布
5. `ExposeService` + `Route`：
   - 根据 labels 选择 running pod
   - 请求按轮询分发到 endpoint

## 适用场景

- 面试/学习：快速理解 orchestrator 基本闭环
- 教学：对比真实 K8s 组件职责
- 原型：在不引入复杂依赖下验证调度策略想法

## 运行

```bash
go test ./orchestrator_mvp -v
go test ./orchestrator_mvp -run TestDemoOutput -v
```


