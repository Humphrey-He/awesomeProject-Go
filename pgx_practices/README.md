# PostgreSQL 数据库操作实践项目

基于 [pgx](https://github.com/jackc/pgx) 驱动，配合 [ent](https://entgo.io/) 和 [gorm](https://gorm.io/) ORM 框架实现的 PostgreSQL 数据库操作示例与最佳实践。

## 项目概述

本项目演示了 Go 语言中 PostgreSQL 数据库的常用操作，涵盖原生驱动、GORM ORM 和 Ent ORM 三种方式的完整实现，并提供了详细的最佳实践指南。

## 技术栈

| 组件 | 说明 |
|------|------|
| pgx/v5 | 高性能 PostgreSQL 驱动 |
| gorm | 成熟的 Go ORM 框架 |
| ent | Facebook 开源的实体框架 |
| postgres driver | GORM 的 PostgreSQL 驱动 |

## 核心功能

### 1. PGX 原生驱动

PGX 是 Go 中性能最好的 PostgreSQL 驱动，核心特性：

- **连接池** - 内置高性能连接池
- **Prepared Statement 缓存** - 自动缓存预编译语句
- **批量操作** - 支持高效的批量插入/更新
- **COPY 协议** - 大数据量快速导入导出
- **通知监听** - LISTEN/NOTIFY 支持

```go
// 创建连接池
pool, err := pgxpool.New(ctx, config.DSN())
if err != nil {
    log.Fatal(err)
}
defer pool.Close()

// 查询数据
rows, err := pool.Query(ctx, "SELECT id, name FROM users WHERE status = $1", "active")
defer rows.Close()

for rows.Next() {
    var id int64
    var name string
    rows.Scan(&id, &name)
    fmt.Printf("User: %d, %s\n", id, name)
}
```

### 2. GORM ORM 操作

GORM 是 Go 中最流行的 ORM 框架，功能丰富：

#### 基础 CRUD

```go
// 创建
user := &User{Name: "张三", Email: "zhangsan@example.com"}
result := db.Create(user)

// 查询
var user User
db.First(&user, 1)  // 根据 ID 查询
db.Where("status = ?", "active").Find(&users)  // 条件查询

// 更新
db.Model(&user).Updates(User{Name: "新名称", Age: 30})

// 删除
db.Delete(&user, 1)  // 软删除
db.Unscoped().Delete(&user, 1)  // 硬删除
```

#### 关联查询

```go
// 预加载关联
db.Preload("Orders").Preload("Orders.Items").First(&user, 1)

// 关联创建
user := User{
    Name: "张三",
    Profile: Profile{Bio: "个人简介"},
}
db.Create(&user)
```

#### 事务处理

```go
// 自动事务
err := db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&order).Error; err != nil {
        return err
    }
    return tx.Create(&items).Error
})

// 手动事务
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()
// ... 操作
tx.Commit()
```

#### 查询构建器

```go
// 链式查询
users, total, err := NewUserQueryBuilder().
    WithStatus("active").
    WithAgeRange(18, 60).
    WithNameLike("张").
    WithPagination(1, 20).
    WithOrderBy("created_at DESC").
    Build()
```

### 3. Ent ORM 操作

Ent 是类型安全的实体框架：

#### 代码生成

```go
// 定义 Schema (ent/schema/user.go)
func (User) Fields() []ent.Field {
    return []ent.Field{
        field.String("name").NotEmpty(),
        field.String("email").Unique(),
        field.Int("age").Positive().Optional(),
    }
}

func (User) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("orders", Order.Type),
    }
}

// 生成代码
// go generate ./ent
```

#### CRUD 操作

```go
// 创建
user, err := client.User.
    Create().
    SetName("张三").
    SetEmail("zhangsan@example.com").
    SetAge(25).
    Save(ctx)

// 查询
user, err := client.User.Get(ctx, id)
users, err := client.User.Query().
    Where(user.StatusEQ("active")).
    All(ctx)

// 更新
err := client.User.
    UpdateOneID(id).
    SetName("新名称").
    SetAge(30).
    Exec(ctx)

// 删除
err := client.User.DeleteOneID(id).Exec(ctx)
```

#### 复杂查询

```go
// 多条件查询
users, err := client.User.Query().
    Where(
        user.AgeGTE(18),
        user.AgeLTE(60),
        user.Or(
            user.NameContains("张"),
            user.EmailContains("example"),
        ),
    ).
    Order(ent.Asc(user.FieldAge)).
    Limit(20).
    Offset(0).
    All(ctx)
```

## 配置说明

### DBConfig 数据库配置

| 参数 | 默认值 | 说明 |
|------|--------|------|
| Host | localhost | 数据库主机 |
| Port | 5432 | 数据库端口 |
| User | postgres | 用户名 |
| Password | postgres | 密码 |
| Database | testdb | 数据库名 |
| SSLMode | disable | SSL 模式 |
| MaxOpenConns | 25 | 最大打开连接数 |
| MaxIdleConns | 10 | 最大空闲连接数 |
| ConnMaxLifetime | 5m | 连接最大生命周期 |
| ConnMaxIdleTime | 10m | 空闲连接最大生命周期 |

### 连接池配置建议

```go
// 默认配置 - 适合一般应用
DefaultPoolConfig()

// 高并发配置 - 适合高并发场景
HighConcurrencyPoolConfig()
```

## 最佳实践

### 1. 连接管理

```go
// 推荐使用连接池
config := DefaultDBConfig()
pool, err := pgxpool.New(ctx, config.DSN())
if err != nil {
    log.Fatal(err)
}
defer pool.Close()

// 设置连接池参数
config.MaxOpenConns = 25      // 最大连接数
config.MaxIdleConns = 10      // 空闲连接数
config.ConnMaxLifetime = 5 * time.Minute  // 连接生命周期
```

### 2. 事务处理

```go
// 事务隔离级别
// PostgreSQL 默认: READ COMMITTED
// 常用级别:
// - READ COMMITTED: 读已提交
// - REPEATABLE READ: 可重复读
// - SERIALIZABLE: 可串行化

// 正确的事务模式
err := db.Transaction(func(tx *gorm.DB) error {
    // 业务逻辑
    if err := tx.Create(&order).Error; err != nil {
        return err  // 自动回滚
    }
    return nil  // 自动提交
})
```

### 3. 批量操作

```go
// GORM 批量插入
users := []User{{Name: "用户1"}, {Name: "用户2"}}
db.CreateInBatches(users, 100)  // 每批 100 条

// PGX 批量操作
batch := &pgx.Batch{}
for _, user := range users {
    batch.Queue("INSERT INTO users (name) VALUES ($1)", user.Name)
}
results := pool.SendBatch(ctx, batch)
defer results.Close()
```

### 4. 预加载优化

```go
// 不推荐：N+1 查询问题
users, _ := db.Find(&users).Rows()
for _, user := range users {
    db.Model(&user).Association("Orders").Find(&user.Orders)
}

// 推荐：使用预加载
db.Preload("Orders").Preload("Orders.Items").Find(&users)
```

### 5. 软删除处理

```go
// GORM 软删除模型
type User struct {
    ID        uint           `gorm:"primaryKey"`
    Name      string
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 查询时自动过滤已删除记录
db.Find(&users)  // 不包含已删除

// 包含已删除记录
db.Unscoped().Find(&users)

// 永久删除
db.Unscoped().Delete(&user)
```

### 6. 健康检查

```go
// 简单健康检查
func HealthCheck(pool *pgxpool.Pool) error {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()
    return pool.Ping(ctx)
}

// 带重试的健康检查
checker := NewHealthChecker(config)
err := checker.CheckWithRetry(ctx, 3)
```

## 数据模型

### 用户模型

```go
type User struct {
    ID        uint           `gorm:"primaryKey"`
    Name      string         `gorm:"size:100;not null"`
    Email     string         `gorm:"size:255;uniqueIndex"`
    Age       int            `gorm:"default:0"`
    Status    string         `gorm:"size:20;default:'active'"`
    CreatedAt time.Time      `gorm:"autoCreateTime"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime"`
    DeletedAt gorm.DeletedAt `gorm:"index"`
    Profile   Profile        `gorm:"foreignKey:UserID"`
    Orders    []Order        `gorm:"foreignKey:UserID"`
}
```

### 订单模型

```go
type Order struct {
    ID          uint        `gorm:"primaryKey"`
    UserID      uint        `gorm:"index;not null"`
    OrderNo     string      `gorm:"size:50;uniqueIndex"`
    TotalAmount float64     `gorm:"type:decimal(10,2)"`
    Status      string      `gorm:"size:20;default:'pending'"`
    CreatedAt   time.Time   `gorm:"autoCreateTime"`
    Items       []OrderItem `gorm:"foreignKey:OrderID"`
}
```

## Docker 快速启动

```bash
# 启动 PostgreSQL
docker run -d --name postgres \
    -p 5432:5432 \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=postgres \
    -e POSTGRES_DB=testdb \
    postgres:latest

# 连接数据库
docker exec -it postgres psql -U postgres -d testdb
```

## 数据库迁移

### GORM 自动迁移

```go
db.AutoMigrate(
    &User{},
    &Profile{},
    &Order{},
    &OrderItem{},
)
```

### Ent 迁移

```go
// 自动创建表
client.Schema.Create(ctx)

// 如果不存在才创建
client.Schema.CreateIfNotExists(ctx)
```

### 手动建表 SQL

```sql
-- 参见 BuildCreateTableSQL() 函数
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    age INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
```

## 性能对比

| 操作 | PGX | GORM | Ent |
|------|-----|------|-----|
| 简单查询 | 最快 | 较快 | 较快 |
| 批量插入 | 最快 | 中等 | 较快 |
| 复杂关联 | 中等 | 较慢 | 较快 |
| 类型安全 | 无 | 弱 | 强 |
| 开发效率 | 低 | 高 | 高 |

## 测试

```bash
# 运行测试（需要本地 PostgreSQL）
go test -v

# 跳过需要数据库的测试
go test -v -short

# 基准测试
go test -bench=.
```

## 依赖

```go
require (
    github.com/jackc/pgx/v5 v5.8.0      // PGX 驱动
    gorm.io/gorm v1.31.1                // GORM ORM
    gorm.io/driver/postgres v1.6.0      // PostgreSQL 驱动
    entgo.io/ent v0.14.5                // Ent ORM
)
```

## 参考资料

- [pgx 官方文档](https://github.com/jackc/pgx)
- [GORM 官方文档](https://gorm.io/docs/)
- [Ent 官方文档](https://entgo.io/docs/getting-started)
- [PostgreSQL 官方文档](https://www.postgresql.org/docs/)
