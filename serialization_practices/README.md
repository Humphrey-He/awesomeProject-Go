# Go 序列化与反序列化最佳实践

## 核心原则

- **协议稳定**：明确 JSON tag，避免隐式字段名漂移
- **严格解码**：`Decoder.DisallowUnknownFields()` 防止脏字段悄悄通过
- **时间统一**：统一使用 UTC + RFC3339 / RFC3339Nano
- **流式处理**：大数据量优先 `Encoder/Decoder`，避免一次性读入
- **错误可追踪**：`fmt.Errorf("...: %w", err)` 做错误链包装
- **脱敏输出**：日志或落盘前先脱敏（如邮箱、手机号）

## 本目录实现

- `UserDTO` + `CustomTime`：展示时间字段自定义编解码
- `SerializeJSON` / `DeserializeJSONStrict`
- `StreamEncodeJSONL` / `StreamDecodeJSONL`（JSON Lines）
- `ToGob` / `FromGob`：内部二进制传输示例
- `RedactEmail`：落日志前脱敏

## 结论

API 场景优先 JSON（可读、跨语言）；内部高性能场景可考虑 gob/protobuf。
关键不是“选哪个格式”，而是“可演进协议 + 严格校验 + 可观测错误 + 安全输出”。


