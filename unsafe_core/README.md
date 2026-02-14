# unsafe 包核心与实战案例

## 核心能力

- `unsafe.Sizeof`：对象大小
- `unsafe.Alignof`：对齐要求
- `unsafe.Offsetof`：字段偏移
- `unsafe.Pointer`：在不同指针类型之间转换

## 本项目实战

- 结构体内存布局分析（`StructLayoutDemo`）
- `[]byte <-> string` 零拷贝转换（高性能但高风险）
- `float64 <-> uint64` 位级重解释
- 原生字节序读取 `uint32`

## 风险与边界

- `unsafe` 会绕过 Go 类型系统和部分内存安全保证
- 零拷贝 `string -> []byte` 后绝不能写入
- 跨平台时要考虑字节序、对齐和实现细节差异
- 建议：仅在性能瓶颈明确、可控范围内使用，并配套测试


