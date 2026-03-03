# runtime 包核心与实战案例

## runtime 的核心能力

- CPU 与并行度：`NumCPU` / `GOMAXPROCS`
- 调度观察：`NumGoroutine` / `Gosched`
- 内存与 GC：`ReadMemStats` / `GC`

## GMP 调度器与 GC 机制

本模块深入讲解 Go 运行时的核心机制：

### GMP 调度器

- **G (Goroutine)**：用户态轻量级协程，初始栈 2KB
- **M (Machine)**：操作系统线程，绑定 P 执行 G
- **P (Processor)**：调度器上下文，数量 = GOMAXPROCS
- **Work Stealing**：本地队列空时从其他 P 偷取
- **抢占式调度**：每个 G 最长运行 10ms

### GC 垃圾回收

- **三色标记清除**：白→灰→黑→清除白色
- **混合写屏障**：保证三色不变式
- **增量式 GC**：并发标记 + 并发清除
- **STW**：仅在开始和结束时短暂停止

## 本项目实战函数

- `TakeSnapshot()`：快速查看运行时快照
- `TuneForCPUBound()`：CPU 密集任务一键调优并行度
- `ForceGCAndReadStats()`：触发 GC 并对比前后指标
- `GoroutineMonitor`：goroutine 泄漏早期预警
- `DemoGMP()`：演示 GMP 调度器工作原理
- `DemoGC()`：演示 GC 垃圾回收机制

## 生产实践建议

- CPU 密集型服务可结合压测动态评估 `GOMAXPROCS`
- 对延迟敏感服务监控 `NumGoroutine` 和 GC 次数
- 泄漏排查：`NumGoroutine` 趋势 + pprof 联合定位
- GC 调优：调整 GOGC 环境变量
