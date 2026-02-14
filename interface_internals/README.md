# 详解 Go 的 interface 结构

本项目把 Go interface 的关键机制拆成可运行示例：

- 空接口（`any`）的概念模型：`type + data`
- 非空接口的概念模型：`itab(type+method table) + data`
- 动态类型与动态值
- 类型断言与 type switch
- 最常见陷阱：`typed nil` 放入接口后，接口本身不为 `nil`

## 重点结论

- `var v any = nil` -> 接口为 nil
- `var p *T = nil; var v any = p` -> 接口不为 nil（动态类型是 `*T`）
- 接口调用方法依赖动态类型的方法集匹配


