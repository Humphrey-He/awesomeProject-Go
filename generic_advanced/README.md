# Go 泛型进阶

## 概述

本项目深入讲解 Go 1.18+ 泛型的进阶特性，包括泛型函数、泛型结构体、泛型接口、泛型容器等内容。

## 核心内容

### 1. 泛型基础

- **类型约束**：`Number`, `Comparable`, `Stringable`
- **泛型函数**：`Sum`, `Max`, `Min`, `Contains`, `Reverse`

### 2. 泛型结构体

- **泛型容器**：`Stack[T]`, `Queue[T]`
- **泛型 Map/Set**：`Map[K,V]`, `Set[T]`

### 3. 泛型方法

- `Filter()`: 过滤元素
- `Map()`: 转换元素
- `Transform()`: 类型转换

### 4. 泛型接口

- `Adder[T]`: 可相加接口
- `Iterable[T]`: 可迭代接口

### 5. 泛型算法

- `QuickSort`: 快速排序
- `BinarySearch`: 二分查找

## 性能特点

✅ **优势**：
- 编译时类型展开，无运行时类型开销
- 比 `interface{}` 更高效
- 内存布局更紧凑

⚠️ **注意**：
- 过多泛型实例化可能导致二进制膨胀

## 最佳实践

- 使用泛型实现数据结构：栈、队列、Map、Set
- 使用泛型实现通用算法：排序、搜索
- 使用内置约束 `comparable`

## 关联项目

- [generic_practices](../generic_practices) - 泛型基础
