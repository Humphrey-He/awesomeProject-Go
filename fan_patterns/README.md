# Fan-Out / Fan-In

并行拆分子任务（fan-out）并在末端聚合（fan-in）。

## 场景

- 对一批任务并行计算
- 子查询并发执行后合并
- 批处理中的并行阶段

## 实现要点

- `FanOut`：固定 worker + `WaitGroup` + 结果 channel
- `FanIn`：多个输入 channel 合并为单输出 channel
- 使用 `context` 做取消传播


