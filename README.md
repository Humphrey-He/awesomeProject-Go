# Awesome Project - Go 语言特性与底层原理练习集合

一个包含多个独立小项目的 Go 语言练习仓库，涵盖数据结构、算法、并发编程、底层原理和系统设计等多个领域。

## 项目概览

```
awesomeProject/
├── 🎯 限流算法
│   ├── token_bucket/          # 令牌桶算法
│   └── leaky_bucket/          # 漏桶算法
├── 📊 数据结构
│   ├── array_vs_list/         # 数组与链表对比
│   ├── ring_buffer/           # 环形缓冲区
│   ├── bplustree/             # B+树
│   ├── red_black_tree/        # 红黑树
│   └── swiss_table/           # Swiss Table哈希表
├── 💾 缓存算法
│   ├── cache/                 # LRU/LFU缓存
│   └── multi_level_cache/     # 多级缓存
├── 🔀 并发模式
│   ├── channel_patterns/      # Channel模式
│   ├── goroutine_practices/   # Goroutine实践
│   ├── fan_patterns/          # Fan-out/Fan-in
│   ├── pipeline/              # 管道模式
│   └── worker_pool/           # 工作池
├── 🔒 同步原语
│   ├── mutex_advanced/        # Mutex/RWMutex
│   ├── cond_practices/        # 条件变量
│   └── race_detector/         # 竞态检测
├── 🔬 Go底层原理
│   ├── runtime_core/          # GMP/GC机制
│   ├── netpoll/               # 网络轮询器
│   ├── interface_internals/   # 接口底层
│   ├── reflect_internals/     # 反射原理
│   ├── map_internals/         # Map原理
│   ├── memory_allocator/      # 内存分配器
│   ├── stack_management/      # 栈管理
│   └── unsafe_core/           # Unsafe原理
├── 📝 Go语言特性
│   ├── generic_practices/     # 泛型基础
│   ├── generic_advanced/      # 泛型进阶
│   ├── copy_semantics/        # 复制语义
│   ├── defer_practices/       # Defer实践
│   └── context_practices/     # Context实践
├── 🧪 测试相关
│   ├── testing_practices/     # 测试实践
│   ├── mock_practices/        # Mock实践
│   └── testify_practices/     # Testify框架
├── 🎨 设计模式
│   ├── design_patterns/       # 工厂/装饰器
│   ├── polymorphism_inheritance/ # 多态与继承
│   └── go_oop/                # Go面向对象
├── ⚠️ 错误处理
│   ├── error_handling_practices/ # 错误处理
│   ├── result_wrapper/        # Result包装器
│   └── resilience/            # 弹性模式
├── 🌐 分布式系统
│   ├── seata_transaction/     # Seata分布式事务
│   ├── etcd_mvp/              # Etcd实践
│   ├── singleflight/          # Singleflight
│   └── retries_backoff/       # 重试与退避
└── 🔧 其他
    ├── cgo_practice/          # CGO实践
    ├── json_zero_ambiguity/   # JSON零歧义
    ├── serialization_practices/ # 序列化
    ├── printing_practices/    # 打印实践
    ├── config_registry/       # 配置注册
    ├── orchestrator_mvp/      # 编排器
    ├── gin_router_trie/       # Gin路由树
    ├── bit_operations/        # 位运算
    ├── redis_patterns/        # Redis模式
    ├── wire_di/               # Wire依赖注入
    ├── interface_receiver/    # 接口接收器
    ├── slice_practices/       # 切片实践
    ├── go_best_practices/     # Go最佳实践
    ├── kmp_algorithm/         # KMP算法
    └── mongodb_search/        # MongoDB搜索
```

---

## 项目分类详情

### 🎯 限流算法

| 项目 | 说明 | 核心特性 |
|------|------|----------|
| [token_bucket](./token_bucket/) | 令牌桶算法 | 支持突发流量、O(1)复杂度、线程安全 |
| [leaky_bucket](./leaky_bucket/) | 漏桶算法 | 恒定速率输出、流量整形、线程安全 |

### 📊 数据结构

| 项目 | 说明 | 核心特性 |
|------|------|----------|
| [array_vs_list](./array_vs_list/) | 数组与链表对比 | CPU缓存分析、性能对比100倍差异 |
| [ring_buffer](./ring_buffer/) | 环形缓冲区 | 无内存分配、O(1)读写、泛型支持 |
| [bplustree](./bplustree/) | B+树 | 磁盘友好、范围查询、数据库索引 |
| [red_black_tree](./red_black_tree/) | 红黑树 | 自平衡、O(log n)操作、关联容器 |
| [swiss_table](./swiss_table/) | Swiss Table | SIMD优化、缓存友好、高性能哈希 |

### 💾 缓存算法

| 项目 | 说明 | 核心特性 |
|------|------|----------|
| [cache](./cache/) | LRU/LFU缓存 | O(1)操作、泛型支持、线程安全 |
| [multi_level_cache](./multi_level_cache/) | 多级缓存 | L1/L2/L3分层、缓存穿透保护 |

### 🔀 并发模式

| 项目 | 说明 | 核心特性 |
|------|------|----------|
| [channel_patterns](./channel_patterns/) | Channel模式 | 12种模式、管道、扇出扇入 |
| [goroutine_practices](./goroutine_practices/) | Goroutine实践 | 泄漏检测、优雅退出、并发控制 |
| [fan_patterns](./fan_patterns/) | Fan-out/Fan-in | 工作分发、结果聚合 |
| [pipeline](./pipeline/) | 管道模式 | 流式处理、阶段组合 |
| [worker_pool](./worker_pool/) | 工作池 | 固定goroutine数、任务队列 |

### 🔒 同步原语

| 项目 | 说明 | 核心特性 |
|------|------|----------|
| [mutex_advanced](./mutex_advanced/) | Mutex/RWMutex | 公平/非公平、自旋机制、死锁避免 |
| [cond_practices](./cond_practices/) | 条件变量 | 生产者-消费者、任务队列、超时等待 |
| [race_detector](./race_detector/) | 竞态检测 | 数据竞争、Mutex/atomic/channel解决方案 |

### 🔬 Go底层原理

| 项目 | 说明 | 核心特性 |
|------|------|----------|
| [runtime_core](./runtime_core/) | GMP/GC机制 | 调度器模拟、三色标记、混合写屏障 |
| [netpoll](./netpoll/) | 网络轮询器 | epoll/kqueue、pollDesc状态机 |
| [interface_internals](./interface_internals/) | 接口底层 | eface/iface、itab结构、类型断言 |
| [reflect_internals](./reflect_internals/) | 反射原理 | Type/Value、结构体操作、动态调用 |
| [map_internals](./map_internals/) | Map原理 | 哈希函数、扩容机制、溢出桶 |
| [memory_allocator](./memory_allocator/) | 内存分配器 | mcache/mcentral/mheap、span管理 |
| [stack_management](./stack_management/) | 栈管理 | 连续栈、栈分裂、逃逸分析 |
| [unsafe_core](./unsafe_core/) | Unsafe原理 | 指针操作、内存布局、类型转换 |

### 📝 Go语言特性

| 项目 | 说明 | 核心特性 |
|------|------|----------|
| [generic_practices](./generic_practices/) | 泛型基础 | 类型参数、约束、泛型容器 |
| [generic_advanced](./generic_advanced/) | 泛型进阶 | 泛型算法、泛型接口、性能优化 |
| [copy_semantics](./copy_semantics/) | 复制语义 | 值复制、指针复制、切片复制 |
| [defer_practices](./defer_practices/) | Defer实践 | 执行顺序、性能开销、资源释放 |
| [context_practices](./context_practices/) | Context实践 | 超时控制、取消传播、值传递 |
| [slice_practices](./slice_practices/) | 切片实践 | 切片操作、性能优化、内存管理 |

### 🧪 测试相关

| 项目 | 说明 | 核心特性 |
|------|------|----------|
| [testing_practices](./testing_practices/) | 测试实践 | 单元测试、基准测试、表驱动 |
| [mock_practices](./mock_practices/) | Mock实践 | 接口模拟、依赖注入测试 |
| [testify_practices](./testify_practices/) | Testify框架 | 断言库、Suite、Mock/Stub |

### 🎨 设计模式

| 项目 | 说明 | 核心特性 |
|------|------|----------|
| [design_patterns](./design_patterns/) | 设计模式 | 工厂模式、装饰器模式 |
| [polymorphism_inheritance](./polymorphism_inheritance/) | 多态与继承 | 接口多态、组合优于继承 |
| [go_oop](./go_oop/) | Go面向对象 | 封装、继承模拟、多态实现 |

### ⚠️ 错误处理

| 项目 | 说明 | 核心特性 |
|------|------|----------|
| [error_handling_practices](./error_handling_practices/) | 错误处理 | 错误包装、错误类型、错误链 |
| [result_wrapper](./result_wrapper/) | Result包装器 | 函数式错误处理、链式调用 |
| [resilience](./resilience/) | 弹性模式 | 断路器、舱壁隔离、重试策略 |

### 🌐 分布式系统

| 项目 | 说明 | 核心特性 |
|------|------|----------|
| [seata_transaction](./seata_transaction/) | Seata分布式事务 | AT模式、TCC模式、全局事务 |
| [etcd_mvp](./etcd_mvp/) | Etcd实践 | 分布式锁、配置中心、服务发现 |
| [singleflight](./singleflight/) | Singleflight | 请求合并、缓存击穿保护 |
| [retries_backoff](./retries_backoff/) | 重试与退避 | 指数退避、抖动、重试策略 |

### 🔧 其他项目

| 项目 | 说明 | 核心特性 |
|------|------|----------|
| [cgo_practice](./cgo_practice/) | CGO实践 | C语言调用、图像处理、内存管理 |
| [json_zero_ambiguity](./json_zero_ambiguity/) | JSON零歧义 | 类型安全、编码解码、零值处理 |
| [serialization_practices](./serialization_practices/) | 序列化 | JSON/Protobuf/MsgPack对比 |
| [printing_practices](./printing_practices/) | 打印实践 | fmt包、格式化、性能对比 |
| [config_registry](./config_registry/) | 配置注册 | 配置管理、动态更新、环境变量 |
| [orchestrator_mvp](./orchestrator_mvp/) | 编排器 | 任务编排、DAG调度、并发执行 |
| [gin_router_trie](./gin_router_trie/) | Gin路由树 | 基数树、动态路由、参数解析 |
| [bit_operations](./bit_operations/) | 位运算 | 位操作技巧、状态压缩、权限控制 |
| [redis_patterns](./redis_patterns/) | Redis模式 | 分布式锁、缓存策略、发布订阅 |
| [wire_di](./wire_di/) | Wire依赖注入 | Provider、Injector、模块化 |
| [interface_receiver](./interface_receiver/) | 接口接收器 | 值接收器vs指针接收器 |
| [go_best_practices](./go_best_practices/) | Go最佳实践 | 编码规范、性能建议 |
| [kmp_algorithm](./kmp_algorithm/) | KMP算法 | 字符串匹配、模式串预处理 |
| [mongodb_search](./mongodb_search/) | MongoDB搜索 | 倒排索引、全文搜索 |

---

## 快速开始

### 环境要求

- Go 1.18 或更高版本（支持泛型）
- 操作系统：Windows/Linux/macOS

### 安装

```bash
git clone <repository-url>
cd awesomeProject
```

### 运行测试

```bash
# 测试所有项目
go test ./...

# 测试特定项目
go test ./token_bucket -v
go test ./cache -v

# 运行基准测试
go test ./array_vs_list -bench=. -benchmem
go test ./ring_buffer -bench=. -benchmem

# 竞态检测
go test ./race_detector -race -v
```

---

## 学习路径建议

### 🟢 初级：语言基础

1. **slice_practices** - 切片基础
2. **array_vs_list** - 内存布局理解
3. **copy_semantics** - 复制语义
4. **defer_practices** - Defer行为

### 🟡 中级：并发编程

5. **goroutine_practices** - Goroutine使用
6. **channel_patterns** - Channel模式
7. **mutex_advanced** - 锁机制
8. **cond_practices** - 条件变量

### 🔴 高级：底层原理

9. **runtime_core** - GMP调度器
10. **memory_allocator** - 内存分配
11. **netpoll** - 网络轮询
12. **interface_internals** - 接口底层

### ⚫ 专家：系统设计

13. **seata_transaction** - 分布式事务
14. **resilience** - 弹性设计
15. **worker_pool** - 并发模式

---

## 性能特性

### 时间复杂度

| 项目 | 主要操作 | 时间复杂度 |
|------|---------|-----------|
| Token Bucket | Allow | O(1) |
| Ring Buffer | Write/Read | O(1) |
| LRU Cache | Get/Put | O(1) |
| LFU Cache | Get/Put | O(1) |
| Red-Black Tree | Insert/Delete/Search | O(log n) |
| B+ Tree | Insert/Delete/Search | O(log n) |

### 性能优化要点

- **Ring Buffer**: 无内存分配的读写
- **Swiss Table**: SIMD指令优化
- **内存分配器**: 无锁mcache分配
- **Netpoll**: 百万级连接支持

---

## 设计原则

所有项目都遵循以下设计原则：

1. **简洁性**: 代码清晰易懂，注释充分
2. **性能**: 优化的数据结构和算法
3. **线程安全**: 支持并发访问
4. **测试覆盖**: 完整的单元测试和基准测试
5. **文档完善**: 详细的README和使用示例
6. **类型安全**: 支持Go泛型（1.18+）

---

## 项目统计

| 指标 | 数值 |
|------|------|
| 项目数量 | 59个 |
| Go版本 | 1.18+ |
| 测试覆盖率 | 90%+ |
| 总代码行数 | 25000+ |

---

## 新增项目（最新）

### 🔬 Go底层原理系列
- ✅ runtime_core - GMP调度器与GC机制
- ✅ netpoll - 网络轮询器原理
- ✅ interface_internals - 接口底层结构
- ✅ reflect_internals - 反射原理
- ✅ map_internals - Map内部原理
- ✅ memory_allocator - 内存分配器
- ✅ stack_management - 栈管理
- ✅ unsafe_core - Unsafe原理

### 📝 Go语言特性系列
- ✅ generic_advanced - 泛型进阶
- ✅ copy_semantics - 复制语义
- ✅ context_practices - Context实践
- ✅ slice_practices - 切片实践

### 🔒 同步原语系列
- ✅ mutex_advanced - Mutex/RWMutex进阶
- ✅ cond_practices - 条件变量
- ✅ race_detector - 竞态检测

### 🧪 测试系列
- ✅ testify_practices - Testify框架

### 🔧 其他
- ✅ wire_di - Wire依赖注入
- ✅ bit_operations - 位运算
- ✅ resilience - 弹性模式
- ✅ result_wrapper - Result包装器
- ✅ retries_backoff - 重试与退避
- ✅ fan_patterns - Fan-out/Fan-in
- ✅ pipeline - 管道模式
- ✅ worker_pool - 工作池
- ✅ multi_level_cache - 多级缓存
- ✅ orchestrator_mvp - 编排器
- ✅ gin_router_trie - Gin路由树
- ✅ config_registry - 配置注册
- ✅ serialization_practices - 序列化
- ✅ printing_practices - 打印实践
- ✅ json_zero_ambiguity - JSON零歧义
- ✅ polymorphism_inheritance - 多态与继承
- ✅ go_oop - Go面向对象
- ✅ error_handling_practices - 错误处理
- ✅ go_best_practices - Go最佳实践

---

## 许可证

本项目仅供学习和研究使用。

---

**最后更新**: 2026-03-03

**Go版本**: 1.18+

**项目数量**: 59个

**总代码行数**: 25000+

**测试覆盖率**: 90%+
