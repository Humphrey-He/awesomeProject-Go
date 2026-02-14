# Go 泛型最佳实践

- 只在“能减少重复且保持可读性”时使用泛型
- 约束要最小化（如 `comparable` / `Ordered`），避免过度复杂
- 通用算法优先泛型：`Map/Filter/Reduce/Min/Max`
- 容器场景最适合泛型：`Set[T]`、`SafeCache[K,V]`
- 避免为了泛型而泛型：业务语义不清时保留具体类型更好


