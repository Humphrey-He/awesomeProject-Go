# Go 打印最佳实践 & 时间打印最佳实践

## 打印最佳实践

- **开发调试**：`fmt.Printf` / `%+v`（结构体可读）
- **生产日志**：`log.Logger`（建议带时间、微秒、文件行号）
- **稳定输出格式**：使用 `key=value`，键名固定，便于检索
- **错误日志最少要素**：`op`（操作名）+ `err`（错误内容）+ `trace_id`（链路）
- **避免噪声**：不要在热路径打印巨量对象

## 时间打印最佳实践

- **跨系统传输**：UTC + `RFC3339Nano`
- **面向人类阅读**：按业务时区输出并显式带时区缩写
- **耗时打印**：统一格式（如 `us/ms/s`），避免同系统多种风格混用
- **日志字段建议**：
  - `time`：事件时间
  - `cost`：耗时
  - `start_time` / `end_time`（必要时）

## 本目录实现

- `FormatKV`：稳定 key=value 输出
- `PrintError`：统一错误打印模板
- `NewLogger`：标准 logger 初始化
- `FormatTimeUTC` / `FormatTimeInLocation`
- `HumanDuration`：统一耗时可读化


