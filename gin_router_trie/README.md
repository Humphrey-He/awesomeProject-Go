# Gin 风格高效路由分发与前缀树匹配（MVP）

## 能力

- 按 HTTP Method 分树（`GET`/`POST`... 独立）
- Trie 前缀树匹配
- 路由类型：
  - 静态路由：`/users/new`
  - 参数路由：`/users/:id`
  - 通配符路由：`/assets/*filepath`
- 匹配优先级：**静态 > 参数 > 通配符**

## 核心设计

- `Router.Add(method, path, handler)` 注册路由
- `Router.Match(method, path)` 解析并回填 `Params`
- `Router.Dispatch(method, path)` 执行分发并返回 handler 输出

## 为什么高效

- 前缀树将路径按段切分，匹配复杂度接近路径长度
- method 分树减少无关分支匹配
- 固定优先级避免歧义，提高命中确定性

## 运行

```bash
go test ./gin_router_trie -v
go test ./gin_router_trie -run TestDemoOutput -v
```


