# Retries / Backoff

带抖动的指数退避重试，避免瞬时失败放大与重试风暴。

## 场景

- 临时网络故障
- 下游短暂不可用
- 限流后的稍后重试

## 实现要点

- 指数退避：`initial * multiplier^attempt`
- 上限保护：`max`
- 抖动（jitter）：降低同一时刻重试洪峰
- 支持 `context` 取消
- 提供 `Retry` 与 `RetryValue`

## 说明

生产环境可替换为：

- `github.com/cenkalti/backoff`


