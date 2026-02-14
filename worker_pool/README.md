# Worker Pool

固定数量 worker 处理大量任务，避免 goroutine 无限膨胀。

## 场景

- IO 密集：批量请求、批量文件处理
- CPU 密集：受控并行计算
- 需要明确并发上限的任务系统

## 实现要点

- `taskCh`：任务队列（可配置缓冲）
- 固定数量 worker goroutine
- `Submit(ctx, task)`：支持提交超时/取消
- `Shutdown(ctx)`：优雅关闭，等待已接收任务完成
- `Stop()`：立即停止，取消运行中的任务

## 快速使用

```go
p := worker_pool.New(8, 256)
defer p.Stop()

resCh, err := p.Submit(context.Background(), func(ctx context.Context) (any, error) {
    return "ok", nil
})
if err != nil {
    panic(err)
}
res := <-resCh
_ = res
```


