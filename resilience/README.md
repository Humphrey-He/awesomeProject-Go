# Circuit Breaker / Rate Limiter

面向“过载保护 + 故障隔离”的基础韧性组件。

## 场景

- 下游服务不稳定时快速失败，防止雪崩
- 控制请求速率，保护本地和下游资源

## 实现要点

- `RateLimiter`：令牌桶（突发 + 持续速率）
- `CircuitBreaker`：`closed/open/half-open` 三态
- 失败阈值触发打开，冷却后进入半开探测
- 半开探测成功则关闭，失败则重新打开

## 说明

当前实现为纯 Go 版本，便于理解原理。生产可替换为：

- 限流：`golang.org/x/time/rate`
- 断路器：`github.com/sony/gobreaker`


