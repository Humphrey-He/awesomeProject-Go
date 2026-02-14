# Leaky Bucket（漏桶）

## 概述

漏桶是一种流量整形（Traffic Shaping）算法，用于控制数据的传输速率。它的核心思想是：无论流入速率如何变化，流出速率始终保持恒定。漏桶算法强制限制数据的传输速率，使其以恒定的速率流出。

## 算法原理

1. **初始化**：创建一个固定容量的桶，初始时桶是空的
2. **请求进入**：
   - 当请求到来时，尝试将其加入桶中
   - 如果桶未满，请求进入桶中等待处理
   - 如果桶已满，请求被拒绝（溢出）
3. **漏水处理**：以固定的速率从桶底漏出请求（处理请求）

## 算法特点

### 优点
- **流量整形效果好**：输出速率恒定，适合需要平滑流量的场景
- **简单直观**：算法实现简单，易于理解
- **防止突发**：强制限制流量速率，保护下游系统

### 缺点
- **不支持突发流量**：即使系统有处理能力，也无法处理突发请求
- **可能浪费资源**：在系统空闲时无法充分利用资源
- **响应延迟**：请求需要在桶中等待

## 使用方法

### 基本用法

```go
package main

import (
	"fmt"
	"awesomeProject/leaky_bucket"
)

func main() {
	// 创建容量为 10，每秒漏出 5 个请求的漏桶
	lb := leaky_bucket.NewLeakyBucket(10, 5)
	
	// 尝试添加一个请求
	if lb.Allow() {
		fmt.Println("请求进入桶中")
	} else {
		fmt.Println("桶已满，请求被拒绝")
	}
	
	// 查看当前桶中的请求数
	fmt.Printf("当前水量: %d\n", lb.CurrentWater())
}
```

### 批量添加请求

```go
// 一次性添加 5 个请求
if lb.AllowN(5) {
	fmt.Println("批量请求进入桶中")
} else {
	fmt.Println("桶空间不足，请求被拒绝")
}
```

### 查询等待时间

```go
waitTime := lb.WaitTime()
if waitTime > 0 {
	fmt.Printf("需要等待 %v\n", waitTime)
} else {
	fmt.Println("可以立即处理")
}
```

### 流量整形示例

```go
package main

import (
	"fmt"
	"time"
	"awesomeProject/leaky_bucket"
)

func main() {
	lb := leaky_bucket.NewLeakyBucket(5, 2) // 容量 5，每秒处理 2 个请求
	
	// 模拟突发流量
	for i := 0; i < 10; i++ {
		if lb.Allow() {
			fmt.Printf("时间 %s: 请求 %d 进入桶\n", time.Now().Format("15:04:05.000"), i+1)
		} else {
			fmt.Printf("时间 %s: 请求 %d 被拒绝（桶满）\n", time.Now().Format("15:04:05.000"), i+1)
		}
		time.Sleep(100 * time.Millisecond)
	}
	
	// 等待桶慢慢漏空
	fmt.Println("\n等待桶漏空...")
	for {
		water := lb.CurrentWater()
		if water == 0 {
			break
		}
		fmt.Printf("当前水量: %d, 可用空间: %d\n", water, lb.AvailableSpace())
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Println("桶已空")
}
```

## API 文档

### NewLeakyBucket(capacity, leakRate int64) *LeakyBucket

创建一个新的漏桶。

**参数：**
- `capacity`：桶的容量（最大请求数）
- `leakRate`：漏水速率（请求数/秒）

**返回：**
- 漏桶实例

### Allow() bool

尝试添加一个请求到桶中。

**返回：**
- `true`：添加成功
- `false`：桶已满，添加失败

### AllowN(n int64) bool

尝试添加 n 个请求到桶中。

**参数：**
- `n`：需要添加的请求数

**返回：**
- `true`：添加成功
- `false`：桶空间不足，添加失败

### CurrentWater() int64

返回当前桶中的水量（待处理的请求数）。

**返回：**
- 当前水量

### AvailableSpace() int64

返回桶中的可用空间。

**返回：**
- 可用空间大小

### Capacity() int64

返回桶的容量。

**返回：**
- 桶的容量

### LeakRate() int64

返回漏水速率。

**返回：**
- 漏水速率（请求数/秒）

### WaitTime() time.Duration

返回当前请求需要等待的时间（如果桶满了）。

**返回：**
- 等待时间（Duration）

## 应用场景

1. **流量整形**：控制向下游发送数据的速率
2. **流媒体传输**：保证视频、音频数据以恒定速率传输
3. **网络设备**：路由器、交换机中的流量控制
4. **API 限流**：严格控制 API 调用速率
5. **消息队列消费**：以固定速率消费消息

## 与令牌桶算法的对比

| 特性 | 漏桶 | 令牌桶 |
|------|------|--------|
| 流出速率 | 恒定 | 可变 |
| 突发流量 | 不支持 | 支持 |
| 流量整形 | 强 | 弱 |
| 灵活性 | 低 | 高 |
| 实现复杂度 | 简单 | 中等 |
| 适用场景 | 流量整形、保护下游 | API 限流、突发处理 |

### 形象比喻

**漏桶**：像一个底部有小孔的水桶，无论上面倒水多快，水都只能以固定速度从底部流出。

**令牌桶**：像一个银行账户，定期存入固定金额，可以一次性取出所有余额（突发消费）。

## 算法变种

### 1. 计数器漏桶
简化版本，使用计数器而不是队列，适合只关心限流不关心排队的场景。

### 2. 队列漏桶
使用真实队列存储请求，按 FIFO 顺序处理。

### 3. 分层漏桶
多个漏桶级联，实现多级流量控制。

## 性能测试

运行基准测试：

```bash
go test -bench=. -benchmem
```

示例输出：
```
BenchmarkLeakyBucket_Allow-8         5000000    250 ns/op    0 B/op    0 allocs/op
BenchmarkLeakyBucket_AllowN-8        5000000    260 ns/op    0 B/op    0 allocs/op
BenchmarkLeakyBucket_CurrentWater-8  5000000    245 ns/op    0 B/op    0 allocs/op
```

## 线程安全

本实现使用 `sync.Mutex` 保证并发安全，可以在多个 goroutine 中安全使用。

## 测试

运行单元测试：

```bash
go test -v
```

运行特定测试：

```bash
go test -v -run TestLeakyBucket_Leak
```

测试覆盖率：

```bash
go test -cover
```

生成覆盖率报告：

```bash
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 最佳实践

1. **选择合适的容量**：容量太小会频繁拒绝请求，太大会占用过多内存
2. **合理设置漏水速率**：根据下游系统的处理能力设置
3. **监控指标**：定期监控桶的水位、拒绝率等指标
4. **降级策略**：配合熔断器使用，在持续拒绝时触发降级
5. **日志记录**：记录被拒绝的请求，便于分析和调优

## 注意事项

1. **时钟敏感**：依赖系统时间，时钟回拨可能导致异常
2. **内存占用**：需要维护桶的状态
3. **不适合突发**：如果业务需要支持突发流量，建议使用令牌桶
4. **公平性**：先到先得，可能不够公平，考虑使用优先级队列

## 扩展阅读

- [令牌桶算法](../token_bucket/README.md)
- [滑动窗口限流](https://en.wikipedia.org/wiki/Sliding_window_protocol)
- [TCP 流量控制](https://en.wikipedia.org/wiki/TCP_congestion_control)

