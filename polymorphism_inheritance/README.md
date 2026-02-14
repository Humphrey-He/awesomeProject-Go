# Go 的多态与“继承”实现

Go 没有传统 class 继承，但可以通过以下方式实现同等能力：

- 组合（embedding）复用通用字段和方法
- 方法重写（同名方法覆盖嵌入类型方法）
- interface 做运行期多态分发

本项目示例：

- `BaseAnimal`：共同行为
- `Dog` / `Bird`：组合 BaseAnimal 并重写部分方法
- `Animal` 接口：统一抽象
- `Zoo`：按接口集合执行多态调用


