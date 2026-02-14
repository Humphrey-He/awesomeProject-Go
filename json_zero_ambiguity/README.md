# JSON 处理与零值模糊性最佳实践

## 问题

普通结构体反序列化时，常无法区分：

- 字段缺失（客户端没传）
- 字段显式传 `null`
- 字段传了零值（如 `0`、`false`、`""`）

这会让 PATCH/部分更新逻辑出错。

## 方案

用 `Optional[T]` 三态模型：

- `Set=false`：字段缺失
- `Set=true && Null=true`：显式 `null`
- `Set=true && Null=false`：有值（包括零值）

## 本目录实现

- `Optional[T]` 自定义 JSON 反序列化
- `PatchUserRequest` 三态字段
- `ApplyPatch` 只更新已传字段
- `DecodePatchStrict` 启用未知字段校验


