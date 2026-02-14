# Go 多级缓存（L1 + L2）MVP

## 目标

实现一个可落地的多级缓存示例：

- L1：进程内内存缓存（LRU + TTL）
- L2：后端缓存（示例用内存 backend，可替换 Redis/Memcached）
- 防击穿：同 key 并发 miss 使用 singleflight 合并请求

## 设计

- `MultiLevelCache.Get`
  1. 先查 L1
  2. miss 后进入 singleflight
  3. 在 singleflight 里查 L2
  4. 命中后回填 L1
- `Set`：写透到 L1 + L2
- `Delete`：双删 L1 + L2

## 关键点

- **防击穿**：并发下同 key 只会有一次 L2 加载
- **TTL**：L1 与 L2 都支持过期
- **容量控制**：L1 超容量时按 LRU 淘汰
- **可扩展性**：`Backend` 接口可直接对接真实缓存

## 运行测试

```bash
go test ./multi_level_cache -v
```


