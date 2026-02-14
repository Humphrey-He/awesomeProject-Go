# Singleflight 实战说明

本文补充 `singleflight/singleflight.go` 的核心机制与常见坑，重点覆盖：

- 击穿保护时序图
- `shared` 标志语义
- 超时路径返回差异（`ErrTimeout` vs `context deadline exceeded`）

---

## 1. 击穿保护时序图

场景：多个并发请求同时访问同一个 key（例如缓存 miss）。

```text
Caller A          Caller B          Caller C          Singleflight
   |                 |                 |                    |
   | Do("user:1")    |                 |                    |
   |---------------->|                 |                    |
   |                 |                 | create call        |
   |                 |                 | run fn()           |
   |                 | Do("user:1")    |                    |
   |                 |---------------->| find inflight call |
   |                 |                 | wait (wg.Wait)     |
   |                 |                 |                    |
   |                 |                 | Do("user:1")       |
   |                 |                 |------------------->|
   |                 |                 | find inflight call |
   |                 |                 | wait (wg.Wait)     |
   |                 |                 |                    |
   | <------ result ------ fn done ----|                    |
   |                 |                 |                    |
   | <------ shared result ------------|                    |
   |                 | <---- shared result -----------------|
```

核心点：同 key 只有一个实际执行者，其余并发调用等待并复用结果，避免“缓存击穿”时对下游的并发放大。

---

## 2. `shared` 标志语义

### 在本实现中的语义

- 返回值 `(val, err, shared)` 中：
  - `shared=true`：当前调用不是独立完成，结果发生了“共享/合并”
  - `shared=false`：当前调用走了独立路径（通常是首个执行者）

### 注意

在并发非常接近时，首个执行者若检测到存在重复等待者，也可能返回 `shared=true`。  
因此更稳妥的理解是：**“这个结果是否参与了请求合并”**，而不是机械理解成“我是不是第一个请求”。

---

## 3. 超时路径返回差异（常见坑）

### 你当前代码中的两条超时路径

1. `DoContext(...)`
   - 当调用方 context 先超时/取消时，直接返回 `ctx.Err()`
   - 常见值：`context deadline exceeded` 或 `context canceled`

2. `DoWithTimeout(...)`
   - 内部用 `context.WithTimeout(...)` 包装，并在内部 select 里返回 `ErrTimeout`
   - 但在某些竞态时序下，外层可能先观察到 `ctx.Err()`，导致调用方拿到 `context deadline exceeded`

### 结果

同为“超时”，调用方可能看到两类错误：

- `ErrTimeout`
- `context.DeadlineExceeded`

### 建议做法

业务层判断超时时不要写死一种：

```go
if errors.Is(err, ErrTimeout) || errors.Is(err, context.DeadlineExceeded) {
    // timeout handling
}
```

---

## 4. 典型接入模式（缓存加载）

`CacheLoader.Load` 的流程：

1. 先查缓存（命中且未过期直接返回）
2. miss 或过期时进入 `sf.Do(key, loader)`
3. 只有一个请求执行 `loader()`，其他并发请求等待共享
4. 成功后写回缓存（带 TTL）

---

## 5. 常见坑清单

1. **key 设计过粗**
   - 不同业务请求被错误合并，返回了不应共享的数据
2. **key 设计过细**
   - 合并率低，起不到保护作用
3. **fn 内部不可重入副作用**
   - singleflight 只能合并“同 key”，不能替你保证副作用幂等
4. **超时处理只判断一种错误**
   - 需同时兼容 `ErrTimeout` 与 `context.DeadlineExceeded`
5. **缓存层未设置合理 TTL/失效策略**
   - singleflight 是击穿保护，不是缓存策略替代品

---

## 6. 如何看打印过程

你已经有 `singleflight/demo_output_test.go`，可直接运行：

```bash
go test ./singleflight -run TestDemoOutput -v
```

你会看到：

- 各 caller 的 `shared` 值
- 底层函数真实执行次数
- 缓存命中与过期后的重新加载次数


