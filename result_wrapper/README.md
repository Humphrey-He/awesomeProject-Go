# Go 泛型优雅错误处理：Result Wrapper

这个项目按“成熟业界标准”实现了一套可落地的 `Result[T]` 模式，用来在 Go 的**显式错误处理**和**代码简洁度**之间取得平衡。

---

## 核心思想（与你的总结一致）

### 1) 结构化抽象（Encapsulation）

传统返回值：

```go
user, err := repo.GetUser(id)
if err != nil { ... }
```

统一容器后：

```go
r := repo.GetUserR(id) // Result[User]
```

- API 形状统一：所有函数都返回 `Result[T]`
- 类型安全：无 `interface{}`，编译期可校验
- 语义明确：成功/失败状态与值同处一个对象

### 2) 行为链式化（Monadic Thinking）

- `Map`：成功时变换值，失败时短路
- `Bind`：串联“返回 Result”的步骤
- `Tap` / `TapError`：副作用观测（日志/指标）
- `Recover`：失败兜底为成功值

这把垂直散落的 `if err != nil` 收敛为水平流转的步骤组合。

### 3) 零值自动管理

`Failure[T](err)` 自动推导 `T` 零值，不再手写：

- `return User{}, err`
- `return nil, err`
- `return 0, err`

统一由泛型封装完成，减少样板代码和遗漏风险。

---

## API 速览

- `Success[T](v)` / `Failure[T](err)`
- `From(v, err)`（兼容传统 `(T, error)`）
- `Unpack()`（回退到传统 `(T, error)`）
- `Map` / `Bind` / `Recover`
- `Tap` / `TapError`
- `WrapError`（加上下文）
- `Must`（仅限测试/启动阶段）

---

## 什么时候推荐用？

适合：

- 服务内有大量“多步骤依赖”流程（A 成功后才做 B）
- 需要统一错误处理策略（日志、指标、错误包装）
- 团队接受函数式链式风格，且愿意约定代码规范

不必强上：

- 简短函数只做一次 IO 时，传统 `(T, error)` 更直接
- 团队暂未形成统一风格时，混用成本可能高于收益

---

## 开发规范建议（关键）

- 边界层（handler/repo/外部调用）务必 `WrapError` 增加上下文
- 领域层优先返回 `Result[T]`，边界层 `Unpack` 适配旧接口
- `Must` 禁止进入业务路径，只用于测试或启动硬失败场景
- 与 `errors.Is/As` 配合保留错误语义（不要只拼字符串）

---

## 一个简化链式示例

```go
res := Bind(
    Map(parseUserID(input), func(id int) int { return id * 10 }),
    func(v int) Result[Profile] { return loadProfile(v) },
)

profile, err := res.Unpack()
```

- 任一步失败会自动短路
- 成功路径保持线性可读

---

## 一句话总结

`Result[T]` 不是要取代 Go 的显式错误处理，而是把显式错误处理**结构化、类型化、可组合化**，让复杂流程既安全又更简洁。


