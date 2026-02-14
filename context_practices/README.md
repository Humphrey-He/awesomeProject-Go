# Context 传递最佳实践与开发指南

## 核心规则

- `context.Context` 永远作为函数第一个参数
- 不把 Context 存进结构体字段（按调用链显式传递）
- 只传“请求级”数据（如 trace/request id），不要塞业务大对象
- 每个 `WithCancel/WithTimeout` 都要 `defer cancel()`
- 下游调用必须尊重 `<-ctx.Done()`

## 本目录实现

- `WithRequestID/RequestIDFrom`：类型安全的 Context Value
- `DoWork`：可取消任务
- `ProcessWithTimeout`：超时边界控制
- `PipelineStep/RunPipeline`：调用链透传与错误包装

## 常见反模式

- 把 context 作为可选参数放到最后
- 忘记 cancel 导致资源泄漏
- 在 context.Value 里塞数据库连接、巨大对象


