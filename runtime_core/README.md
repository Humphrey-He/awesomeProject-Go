# runtime 包核心与实战案例

## runtime 的核心能力

- CPU 与并行度：`NumCPU` / `GOMAXPROCS`
- 调度观察：`NumGoroutine` / `Gosched`
- 内存与 GC：`ReadMemStats` / `GC`

## 本项目实战函数

- `TakeSnapshot()`：快速查看运行时快照
- `TuneForCPUBound()`：CPU 密集任务一键调优并行度
- `ForceGCAndReadStats()`：触发 GC 并对比前后指标
- `GoroutineMonitor`：goroutine 泄漏早期预警
- `RunCPUTasks()`：并行 CPU 任务压测辅助

## 生产实践建议

- CPU 密集型服务可结合压测动态评估 `GOMAXPROCS`
- 对延迟敏感服务监控 `NumGoroutine` 和 GC 次数
- 泄漏排查：`NumGoroutine` 趋势 + pprof 联合定位


