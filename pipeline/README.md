# Pipeline

多阶段流水线处理，适合流式数据处理和阶段化转换。

## 场景

- 日志清洗 -> 转换 -> 聚合
- 数据 ETL
- 流式任务分段处理

## 实现要点

- 每个阶段单独 goroutine（或阶段内多 worker）
- 阶段间通过 channel 串联
- 通过有界缓冲控制背压
- `context` 统一取消，避免 goroutine 泄漏

## 示例拓扑

`Source -> MapStage(parallel) -> FilterStage -> Sink`


