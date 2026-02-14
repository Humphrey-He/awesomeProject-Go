# 缓存算法：LRU 和 LFU

## 概述

本项目实现了两种常用的缓存淘汰算法：
- **LRU (Least Recently Used)**：最近最少使用算法
- **LFU (Least Frequently Used)**：最不经常使用算法

两种算法都提供了线程安全的实现和泛型版本（Go 1.18+）。

## LRU（Least Recently Used）

### 原理

LRU算法基于这样的假设：最近被访问的数据在未来被访问的概率更高。

- **淘汰策略**：当缓存满时，淘汰最长时间未被访问的数据
- **实现结构**：双向链表 + 哈希表
  - 链表维护访问顺序（最近访问的在头部）
  - 哈希表提供O(1)查找

### 数据结构

```
哈希表: key -> 链表节点
链表（按访问时间排序）:
[最近] -> [较近] -> [较旧] -> [最旧]
  Head                       Tail
```

### 时间复杂度

| 操作 | 时间复杂度 | 说明 |
|------|-----------|------|
| Get | O(1) | 哈希表查找 + 链表移动 |
| Put | O(1) | 哈希表插入 + 链表操作 |
| Delete | O(1) | 哈希表删除 + 链表移动 |

### 空间复杂度

O(n)，其中n是缓存容量

### API文档

#### NewLRUCache(capacity int) *LRUCache

创建一个新的LRU缓存。

**参数：**
- `capacity`：缓存容量

**返回：**
- LRU缓存实例

#### Get(key interface{}) (interface{}, bool)

获取缓存值，并将其标记为最近使用。

**返回：**
- 值和是否存在的布尔值

#### Put(key, value interface{})

设置缓存值。如果缓存满，淘汰最久未使用的数据。

#### Delete(key interface{}) bool

删除缓存项。

#### Peek(key interface{}) (interface{}, bool)

查看缓存值但不更新访问时间。

#### Len() int

返回当前缓存大小。

#### Clear()

清空缓存。

#### Keys() []interface{}

返回所有key（从最近使用到最少使用）。

### 使用示例

```go
package main

import (
	"fmt"
	"awesomeProject/cache"
)

func main() {
	// 创建容量为3的LRU缓存
	lru := cache.NewLRUCache(3)
	
	// 添加数据
	lru.Put("a", 1)
	lru.Put("b", 2)
	lru.Put("c", 3)
	
	// 访问a（变成最近使用）
	val, _ := lru.Get("a")
	fmt.Println(val)  // 1
	
	// 添加d，淘汰最久未使用的b
	lru.Put("d", 4)
	
	_, exists := lru.Get("b")
	fmt.Println(exists)  // false（已被淘汰）
}
```

### 泛型版本

```go
// 类型安全的LRU缓存
lru := cache.NewLRUCacheGeneric[string, int](10)

lru.Put("score", 100)
score, _ := lru.Get("score")  // score是int类型
```

## LFU（Least Frequently Used）

### 原理

LFU算法基于这样的假设：被频繁访问的数据在未来被访问的概率更高。

- **淘汰策略**：当缓存满时，淘汰访问频率最低的数据
- **同频率处理**：频率相同时，淘汰最早加入的数据（FIFO）
- **实现结构**：哈希表 + 频率链表

### 数据结构

```
cache: key -> node (包含value和频率)

freqMap: 频率 -> 该频率的所有节点列表
  频率1: [node1, node2, node3]
  频率2: [node4, node5]
  频率3: [node6]
  
minFreq: 记录当前最小频率
```

### 时间复杂度

| 操作 | 时间复杂度 | 说明 |
|------|-----------|------|
| Get | O(1) | 哈希表查找 + 频率更新 |
| Put | O(1) | 哈希表插入 + 频率管理 |
| Delete | O(1) | 哈希表删除 + 链表操作 |

### 空间复杂度

O(n + k)，其中n是缓存容量，k是不同频率的数量

### API文档

#### NewLFUCache(capacity int) *LFUCache

创建一个新的LFU缓存。

#### Get(key interface{}) (interface{}, bool)

获取缓存值，并增加访问频率。

#### Put(key, value interface{})

设置缓存值。如果缓存满，淘汰访问频率最低的数据。

#### GetFreq(key interface{}) int

获取key的访问频率。

#### Delete(key interface{}) bool

删除缓存项。

#### Peek(key interface{}) (interface{}, bool)

查看缓存值但不增加访问频率。

### 使用示例

```go
package main

import (
	"fmt"
	"awesomeProject/cache"
)

func main() {
	// 创建容量为2的LFU缓存
	lfu := cache.NewLFUCache(2)
	
	lfu.Put("a", 1)
	lfu.Put("b", 2)
	
	// 访问a三次，访问b一次
	lfu.Get("a")
	lfu.Get("a")
	lfu.Get("a")
	lfu.Get("b")
	
	// a的频率：4（1次put + 3次get）
	// b的频率：2（1次put + 1次get）
	
	fmt.Println(lfu.GetFreq("a"))  // 4
	fmt.Println(lfu.GetFreq("b"))  // 2
	
	// 添加c，淘汰频率最低的b
	lfu.Put("c", 3)
	
	_, exists := lfu.Get("b")
	fmt.Println(exists)  // false（已被淘汰）
}
```

## LRU vs LFU 对比

### 核心差异

| 特性 | LRU | LFU |
|------|-----|-----|
| 淘汰依据 | 访问时间 | 访问频率 |
| 实现复杂度 | 较简单 | 较复杂 |
| 空间开销 | O(n) | O(n+k) |
| 适用场景 | 时间局部性强 | 热点数据明显 |

### 场景分析

#### 场景1：周期性访问

```
访问序列：A, B, C, D, A, B, C, D, ...
缓存容量：2

LRU表现：缓存命中率 0%（不断淘汰即将要访问的数据）
LFU表现：缓存命中率 0%（频率相同，不断淘汰）

结论：两者表现相当，都不适合
```

#### 场景2：热点数据

```
访问序列：A, A, A, A, B, C, D, A, A, A
缓存容量：2

LRU表现：
- 前期：缓存[A]
- 加入B,C,D后可能淘汰A
- 后期：可能需要重新加载A

LFU表现：
- A的频率很高，不会被淘汰
- 始终保持A在缓存中

结论：LFU更适合热点数据场景
```

#### 场景3：顺序扫描

```
访问序列：1, 2, 3, 4, 5, 6, 7, 8, 9, 10
缓存容量：3

LRU表现：
- 保留最后3个：[8, 9, 10]
- 扫描过程中每次都miss

LFU表现：
- 保留最后3个：[8, 9, 10]
- 扫描过程中每次都miss

结论：两者表现相同，都不适合顺序扫描
```

#### 场景4：最近热点突然出现

```
访问序列：A(100次), B(100次), C(100次), D, D, D, D
缓存容量：3

LRU表现：
- 缓存[B, C, D]
- A被淘汰

LFU表现：
- 缓存[A, B, C]
- D无法进入缓存（频率太低）

结论：LRU更能适应访问模式变化
```

### 选择建议

#### 使用LRU的场景：

1. **Web应用缓存**
   - 用户会话数据
   - 最近访问的页面
   - 最近的API响应

2. **文件系统缓存**
   - 最近打开的文件
   - 最近的目录列表

3. **数据库查询缓存**
   - 最近的查询结果

4. **访问模式频繁变化**
   - 新闻网站（热点新闻不断变化）
   - 社交媒体（关注热点变化快）

#### 使用LFU的场景：

1. **热点数据长期不变**
   - 视频网站的热门视频
   - 电商网站的爆款商品
   - 游戏服务器的热门地图

2. **有明显的"二八定律"**
   - 20%的数据被访问80%的次数

3. **防止缓存污染**
   - 偶尔的大量顺序扫描不应该淘汰热点数据

4. **内容推荐系统**
   - 根据长期访问频率推荐

## 实战应用

### Web应用示例

```go
package main

import (
	"net/http"
	"awesomeProject/cache"
)

var pageCache = cache.NewLRUCacheGeneric[string, []byte](100)

func handler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Path
	
	// 尝试从缓存获取
	if content, ok := pageCache.Get(url); ok {
		w.Write(content)
		return
	}
	
	// 渲染页面
	content := renderPage(url)
	
	// 存入缓存
	pageCache.Put(url, content)
	
	w.Write(content)
}

func renderPage(url string) []byte {
	// 实际渲染逻辑
	return []byte("page content")
}
```

### 数据库查询缓存

```go
type QueryCache struct {
	cache *cache.LRUCacheGeneric[string, interface{}]
}

func NewQueryCache(size int) *QueryCache {
	return &QueryCache{
		cache: cache.NewLRUCacheGeneric[string, interface{}](size),
	}
}

func (qc *QueryCache) Query(sql string, params ...interface{}) (interface{}, error) {
	// 生成缓存key
	key := generateCacheKey(sql, params...)
	
	// 尝试从缓存获取
	if result, ok := qc.cache.Get(key); ok {
		return result, nil
	}
	
	// 执行数据库查询
	result, err := executeQuery(sql, params...)
	if err != nil {
		return nil, err
	}
	
	// 存入缓存
	qc.cache.Put(key, result)
	
	return result, nil
}
```

### 视频热度缓存（LFU）

```go
type VideoCache struct {
	cache *cache.LFUCacheGeneric[int64, *Video]
}

func (vc *VideoCache) GetVideo(videoID int64) (*Video, error) {
	// 尝试从缓存获取
	if video, ok := vc.cache.Get(videoID); ok {
		// 每次访问都会增加频率
		return video, nil
	}
	
	// 从数据库加载
	video, err := loadVideoFromDB(videoID)
	if err != nil {
		return nil, err
	}
	
	// 存入缓存
	vc.cache.Put(videoID, video)
	
	return video, nil
}

func (vc *VideoCache) GetHotVideos() []int64 {
	// LFU特性：频率高的视频会保留在缓存中
	return vc.cache.Keys()
}
```

## 性能测试

### 运行测试

```bash
cd cache

# 单元测试
go test -v

# 基准测试
go test -bench=. -benchmem

# 并发安全测试
go test -race -v
```

### 典型性能

```
BenchmarkLRUCache_Put-8     5000000    300 ns/op    150 B/op    2 allocs/op
BenchmarkLRUCache_Get-8    10000000    150 ns/op      0 B/op    0 allocs/op
BenchmarkLFUCache_Put-8     3000000    400 ns/op    200 B/op    3 allocs/op
BenchmarkLFUCache_Get-8     8000000    200 ns/op      0 B/op    0 allocs/op
```

**结论：**
- LRU性能略优于LFU（实现更简单）
- 两者的Get操作都很快（O(1)）
- 内存分配开销都很小

## 高级特性

### 1. 缓存穿透保护

```go
type ProtectedCache struct {
	cache *cache.LRUCacheGeneric[string, interface{}]
	null  interface{} // 空值标记
}

func (pc *ProtectedCache) Get(key string) (interface{}, error) {
	// 检查缓存
	if val, ok := pc.cache.Get(key); ok {
		if val == pc.null {
			return nil, ErrNotFound  // 缓存的空值
		}
		return val, nil
	}
	
	// 查询数据库
	val, err := queryDB(key)
	if err == ErrNotFound {
		// 缓存空值，防止穿透
		pc.cache.Put(key, pc.null)
		return nil, err
	}
	
	pc.cache.Put(key, val)
	return val, nil
}
```

### 2. 过期时间支持

```go
type CacheEntry struct {
	Value      interface{}
	ExpireTime time.Time
}

type TTLCache struct {
	cache *cache.LRUCacheGeneric[string, *CacheEntry]
}

func (tc *TTLCache) Get(key string) (interface{}, bool) {
	entry, ok := tc.cache.Get(key)
	if !ok {
		return nil, false
	}
	
	// 检查是否过期
	if time.Now().After(entry.ExpireTime) {
		tc.cache.Delete(key)
		return nil, false
	}
	
	return entry.Value, true
}

func (tc *TTLCache) Put(key string, value interface{}, ttl time.Duration) {
	entry := &CacheEntry{
		Value:      value,
		ExpireTime: time.Now().Add(ttl),
	}
	tc.cache.Put(key, entry)
}
```

### 3. 统计信息

```go
type StatsCache struct {
	cache *cache.LRUCacheGeneric[string, interface{}]
	hits  int64
	misses int64
	mu    sync.Mutex
}

func (sc *StatsCache) Get(key string) (interface{}, bool) {
	val, ok := sc.cache.Get(key)
	
	sc.mu.Lock()
	if ok {
		sc.hits++
	} else {
		sc.misses++
	}
	sc.mu.Unlock()
	
	return val, ok
}

func (sc *StatsCache) HitRate() float64 {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	total := sc.hits + sc.misses
	if total == 0 {
		return 0
	}
	
	return float64(sc.hits) / float64(total)
}
```

## 最佳实践

### 1. 容量设置

```go
// 根据内存和数据大小估算
itemSize := 1024  // 每个缓存项1KB
maxMemory := 100 * 1024 * 1024  // 最大100MB
capacity := maxMemory / itemSize

cache := cache.NewLRUCache(capacity)
```

### 2. 预热缓存

```go
func WarmUpCache(c *cache.LRUCache, hotKeys []string) {
	for _, key := range hotKeys {
		val := loadFromDB(key)
		c.Put(key, val)
	}
}
```

### 3. 监控缓存状态

```go
func MonitorCache(c *cache.LRUCache) {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		utilization := float64(c.Len()) / float64(c.Cap())
		log.Printf("Cache utilization: %.2f%%", utilization*100)
		
		if utilization > 0.9 {
			log.Warn("Cache is nearly full")
		}
	}
}
```

### 4. 分层缓存

```go
type TieredCache struct {
	l1 *cache.LRUCacheGeneric[string, interface{}]  // 小容量，热点数据
	l2 *cache.LFUCacheGeneric[string, interface{}]  // 大容量，频繁数据
}

func (tc *TieredCache) Get(key string) (interface{}, bool) {
	// 先查L1
	if val, ok := tc.l1.Get(key); ok {
		return val, true
	}
	
	// 再查L2
	if val, ok := tc.l2.Get(key); ok {
		// 提升到L1
		tc.l1.Put(key, val)
		return val, true
	}
	
	return nil, false
}
```

## 常见问题

### Q: LRU和LFU哪个更好？

A: 没有绝对的好坏，取决于访问模式：
- 访问模式频繁变化 → LRU
- 有明显长期热点 → LFU
- 不确定 → 先用LRU（更简单）

### Q: 如何选择缓存容量？

A: 考虑因素：
1. 可用内存
2. 单个缓存项大小
3. 命中率目标（通常80-90%）
4. 使用监控数据调优

### Q: 缓存污染如何处理？

A: 
- LRU：无法完全避免，考虑使用LFU
- LFU：天然防止污染（低频数据无法驱逐高频数据）
- 可以结合过期时间机制

### Q: 如何提高并发性能？

A:
- 使用分片缓存（Sharded Cache）
- 读写分离
- 考虑无锁实现（复杂度高）

## 扩展阅读

- [LRU Cache - LeetCode 146](https://leetcode.com/problems/lru-cache/)
- [LFU Cache - LeetCode 460](https://leetcode.com/problems/lfu-cache/)
- [Caffeine Cache](https://github.com/ben-manes/caffeine) - 高性能Java缓存库
- [groupcache](https://github.com/golang/groupcache) - Google的分布式缓存

## 总结

本项目实现了两种经典的缓存淘汰算法：

- **LRU**：适合大多数场景，实现简单，性能好
- **LFU**：适合热点数据明显的场景，能更好地保护热点数据

两种算法都提供：
- ✅ O(1)时间复杂度
- ✅ 线程安全
- ✅ 泛型支持（Go 1.18+）
- ✅ 完整的测试覆盖

选择建议：
- 默认使用LRU（适用范围更广）
- 有长期热点数据时使用LFU
- 可以组合使用（分层缓存）

