# `if err != nil` 最佳实践

## 实践原则

- 尽早返回（early return），缩短正常路径
- 在边界做错误包装：`fmt.Errorf("op: %w", err)`
- 用 `errors.Is/As` 做错误分类，不依赖字符串匹配
- 保留上下文信息（哪个操作、哪个输入）
- 清理逻辑错误不要吞掉，至少可观测

## 本目录实现

- `ParseUserID`：输入校验 + 错误包装
- `OpError`：操作级错误封装（支持 `Unwrap`）
- `FindUserName`：`ErrNotFound` 哨兵错误 + `errors.Is`
- `RunWithCleanup`：主流程 + cleanup 错误处理策略

## 一句话指南

`if err != nil` 不是模板动作，关键是“错误被谁看见、带什么上下文、能否被程序正确分类处理”。


