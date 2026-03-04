// pgx_practices.go - PostgreSQL 数据库操作示例与最佳实践
// 使用 pgx 驱动，配合 ent 和 gorm ORM 框架实现

package pgx_practices

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// ==================== 数据库配置 ====================

// DBConfig 数据库连接配置
type DBConfig struct {
	Host            string        // 数据库主机地址
	Port            int           // 数据库端口
	User            string        // 用户名
	Password        string        // 密码
	Database        string        // 数据库名
	SSLMode         string        // SSL 模式 (disable/require/verify-ca/verify-full)
	MaxOpenConns    int           // 最大打开连接数
	MaxIdleConns    int           // 最大空闲连接数
	ConnMaxLifetime time.Duration // 连接最大生命周期
	ConnMaxIdleTime time.Duration // 空闲连接最大生命周期
}

// DefaultDBConfig 返回默认数据库配置
func DefaultDBConfig() *DBConfig {
	return &DBConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		Database:        "testdb",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
	}
}

// DSN 返回数据源连接字符串
func (c *DBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)
}

// ==================== PGX 原生驱动操作 ====================

// PGXDriver PGX 原生驱动封装
// PGX 是 Go 中性能最好的 PostgreSQL 驱动，支持：
// - 连接池
// - Prepared Statement 缓存
// - 批量操作
// - COPY 协议
// - 通知监听
type PGXDriver struct {
	config *DBConfig
	// 实际使用时需要 *pgxpool.Pool 或 *sql.DB
}

// NewPGXDriver 创建 PGX 驱动实例
func NewPGXDriver(config *DBConfig) *PGXDriver {
	if config == nil {
		config = DefaultDBConfig()
	}
	return &PGXDriver{config: config}
}

// Connect 连接数据库示例
// 实际使用代码：
// pool, err := pgxpool.New(ctx, config.DSN())
func (p *PGXDriver) Connect(ctx context.Context) error {
	log.Printf("PGX 连接数据库: %s", p.config.DSN())
	// 实际项目中使用：
	// pool, err := pgxpool.New(ctx, dsn)
	// if err != nil { return err }
	// defer pool.Close()
	return nil
}

// QueryExample 查询示例
// 使用 pgxpool 查询数据
func (p *PGXDriver) QueryExample(ctx context.Context) error {
	// 实际使用代码：
	// rows, err := pool.Query(ctx, "SELECT id, name, email FROM users WHERE status = $1", "active")
	// if err != nil { return err }
	// defer rows.Close()
	// 
	// for rows.Next() {
	//     var id int64
	//     var name, email string
	//     err := rows.Scan(&id, &name, &email)
	//     if err != nil { return err }
	//     fmt.Printf("User: %d, %s, %s\n", id, name, email)
	// }
	// return rows.Err()
	return nil
}

// BatchInsertExample 批量插入示例
// PGX 支持 batch 操作，性能优异
func (p *PGXDriver) BatchInsertExample(ctx context.Context, users []User) error {
	// 实际使用代码：
	// batch := &pgx.Batch{}
	// for _, user := range users {
	//     batch.Queue("INSERT INTO users (name, email, created_at) VALUES ($1, $2, $3)", 
	//         user.Name, user.Email, user.CreatedAt)
	// }
	// results := pool.SendBatch(ctx, batch)
	// defer results.Close()
	// for i := 0; i < len(users); i++ {
	//     _, err := results.Exec()
	//     if err != nil { return err }
	// }
	return nil
}

// TransactionExample 事务示例
func (p *PGXDriver) TransactionExample(ctx context.Context) error {
	// 实际使用代码：
	// tx, err := pool.Begin(ctx)
	// if err != nil { return err }
	// defer tx.Rollback(ctx)
	// 
	// _, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", 100, 1)
	// if err != nil { return err }
	// 
	// _, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", 100, 2)
	// if err != nil { return err }
	// 
	// return tx.Commit(ctx)
	return nil
}

// ==================== GORM ORM 操作 ====================

// User 用户模型（GORM 使用）
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	Email     string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Age       int            `gorm:"default:0" json:"age"`
	Status    string         `gorm:"size:20;default:'active'" json:"status"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Profile   Profile        `gorm:"foreignKey:UserID" json:"profile"`
	Orders    []Order        `gorm:"foreignKey:UserID" json:"orders"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// Profile 用户资料模型
type Profile struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	Bio       string    `gorm:"type:text" json:"bio"`
	Avatar    string    `gorm:"size:255" json:"avatar"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (Profile) TableName() string {
	return "profiles"
}

// Order 订单模型
type Order struct {
	ID          uint          `gorm:"primaryKey" json:"id"`
	UserID      uint          `gorm:"index;not null" json:"user_id"`
	OrderNo     string        `gorm:"size:50;uniqueIndex;not null" json:"order_no"`
	TotalAmount float64       `gorm:"type:decimal(10,2)" json:"total_amount"`
	Status      string        `gorm:"size:20;default:'pending'" json:"status"`
	CreatedAt   time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
	Items       []OrderItem   `gorm:"foreignKey:OrderID" json:"items"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

// OrderItem 订单项模型
type OrderItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OrderID   uint      `gorm:"index;not null" json:"order_id"`
	ProductID uint      `gorm:"not null" json:"product_id"`
	Quantity  int       `gorm:"default:1" json:"quantity"`
	Price     float64   `gorm:"type:decimal(10,2)" json:"price"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName 指定表名
func (OrderItem) TableName() string {
	return "order_items"
}

// DeletedAt 软删除字段类型（简化版，实际使用 gorm.DeletedAt）
type DeletedAt = time.Time

// GORMRepository GORM 仓储封装
type GORMRepository struct {
	// db *gorm.DB - 实际使用时需要 gorm.DB 实例
	config *DBConfig
}

// NewGORMRepository 创建 GORM 仓储
func NewGORMRepository(config *DBConfig) *GORMRepository {
	if config == nil {
		config = DefaultDBConfig()
	}
	return &GORMRepository{config: config}
}

// Connect 连接数据库
// 实际使用：
// dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
//     config.Host, config.User, config.Password, config.Database, config.Port, config.SSLMode)
// db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
func (r *GORMRepository) Connect() error {
	log.Printf("GORM 连接数据库: %s", r.config.DSN())
	return nil
}

// ==================== GORM CRUD 操作 ====================

// CreateUser 创建用户
func (r *GORMRepository) CreateUser(ctx context.Context, user *User) error {
	// 实际使用：
	// return r.db.WithContext(ctx).Create(user).Error
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	log.Printf("GORM 创建用户: %+v", user)
	return nil
}

// GetUserByID 根据 ID 查询用户
func (r *GORMRepository) GetUserByID(ctx context.Context, id uint) (*User, error) {
	// 实际使用：
	// var user User
	// err := r.db.WithContext(ctx).First(&user, id).Error
	// return &user, err
	return &User{ID: id, Name: "测试用户", Email: "test@example.com"}, nil
}

// UpdateUser 更新用户
func (r *GORMRepository) UpdateUser(ctx context.Context, user *User) error {
	// 实际使用：
	// return r.db.WithContext(ctx).Save(user).Error
	user.UpdatedAt = time.Now()
	log.Printf("GORM 更新用户: %+v", user)
	return nil
}

// DeleteUser 删除用户（软删除）
func (r *GORMRepository) DeleteUser(ctx context.Context, id uint) error {
	// 实际使用：
	// return r.db.WithContext(ctx).Delete(&User{}, id).Error
	log.Printf("GORM 删除用户: %d", id)
	return nil
}

// HardDeleteUser 硬删除用户
func (r *GORMRepository) HardDeleteUser(ctx context.Context, id uint) error {
	// 实际使用：
	// return r.db.WithContext(ctx).Unscoped().Delete(&User{}, id).Error
	log.Printf("GORM 硬删除用户: %d", id)
	return nil
}

// ListUsers 分页查询用户列表
func (r *GORMRepository) ListUsers(ctx context.Context, page, pageSize int, query string) ([]User, int64, error) {
	// 实际使用：
	// var users []User
	// var total int64
	// offset := (page - 1) * pageSize
	// 
	// db := r.db.WithContext(ctx).Model(&User{})
	// if query != "" {
	//     db = db.Where("name LIKE ? OR email LIKE ?", "%"+query+"%", "%"+query+"%")
	// }
	// 
	// db.Count(&total)
	// err := db.Offset(offset).Limit(pageSize).Find(&users).Error
	// return users, total, err

	users := []User{
		{ID: 1, Name: "用户1", Email: "user1@example.com"},
		{ID: 2, Name: "用户2", Email: "user2@example.com"},
	}
	return users, 2, nil
}

// GetUserWithOrders 获取用户及其订单（预加载关联）
func (r *GORMRepository) GetUserWithOrders(ctx context.Context, id uint) (*User, error) {
	// 实际使用：
	// var user User
	// err := r.db.WithContext(ctx).
	//     Preload("Orders").
	//     Preload("Orders.Items").
	//     First(&user, id).Error
	// return &user, err

	user := &User{
		ID:    id,
		Name:  "测试用户",
		Email: "test@example.com",
		Orders: []Order{
			{ID: 1, OrderNo: "ORD001", TotalAmount: 99.99, Status: "completed"},
		},
	}
	return user, nil
}

// CreateOrderWithItems 创建订单及订单项（事务）
func (r *GORMRepository) CreateOrderWithItems(ctx context.Context, order *Order, items []OrderItem) error {
	// 实际使用：
	// return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
	//     if err := tx.Create(order).Error; err != nil {
	//         return err
	//     }
	//     for i := range items {
	//         items[i].OrderID = order.ID
	//     }
	//     return tx.Create(&items).Error
	// })
	log.Printf("GORM 创建订单: %+v, 订单项: %+v", order, items)
	return nil
}

// BatchCreateUsers 批量创建用户
func (r *GORMRepository) BatchCreateUsers(ctx context.Context, users []User) error {
	// 实际使用：
	// return r.db.WithContext(ctx).CreateInBatches(users, 100).Error
	log.Printf("GORM 批量创建用户: %d 条", len(users))
	return nil
}

// ==================== GORM 查询构建器 ====================

// UserQueryBuilder 用户查询构建器
type UserQueryBuilder struct {
	// db *gorm.DB
	status   string
	minAge   int
	maxAge   int
	nameLike string
	orderBy  string
	page     int
	pageSize int
}

// NewUserQueryBuilder 创建用户查询构建器
func NewUserQueryBuilder() *UserQueryBuilder {
	return &UserQueryBuilder{
		page:     1,
		pageSize: 20,
		orderBy:  "created_at DESC",
	}
}

// WithStatus 设置状态过滤
func (b *UserQueryBuilder) WithStatus(status string) *UserQueryBuilder {
	b.status = status
	return b
}

// WithAgeRange 设置年龄范围
func (b *UserQueryBuilder) WithAgeRange(min, max int) *UserQueryBuilder {
	b.minAge = min
	b.maxAge = max
	return b
}

// WithNameLike 设置名称模糊查询
func (b *UserQueryBuilder) WithNameLike(name string) *UserQueryBuilder {
	b.nameLike = name
	return b
}

// WithPagination 设置分页
func (b *UserQueryBuilder) WithPagination(page, pageSize int) *UserQueryBuilder {
	b.page = page
	b.pageSize = pageSize
	return b
}

// WithOrderBy 设置排序
func (b *UserQueryBuilder) WithOrderBy(orderBy string) *UserQueryBuilder {
	b.orderBy = orderBy
	return b
}

// Build 构建查询
func (b *UserQueryBuilder) Build() ([]User, int64, error) {
	// 实际使用：
	// var users []User
	// var total int64
	// 
	// db := r.db.Model(&User{})
	// 
	// if b.status != "" {
	//     db = db.Where("status = ?", b.status)
	// }
	// if b.minAge > 0 {
	//     db = db.Where("age >= ?", b.minAge)
	// }
	// if b.maxAge > 0 {
	//     db = db.Where("age <= ?", b.maxAge)
	// }
	// if b.nameLike != "" {
	//     db = db.Where("name LIKE ?", "%"+b.nameLike+"%")
	// }
	// 
	// db.Count(&total)
	// 
	// offset := (b.page - 1) * b.pageSize
	// err := db.Order(b.orderBy).Offset(offset).Limit(b.pageSize).Find(&users).Error
	// 
	// return users, total, err

	return []User{}, 0, nil
}

// ==================== ENT ORM 操作（概念性实现）====================

// EntRepository Ent ORM 仓储封装
// Ent 是 Facebook 开源的 Go 实体框架，特点：
// - 代码生成
// - 类型安全
// - GraphQL 集成
// - Schema 迁移
type EntRepository struct {
	config *DBConfig
	// client *ent.Client - 实际使用时需要 ent.Client 实例
}

// NewEntRepository 创建 Ent 仓储
func NewEntRepository(config *DBConfig) *EntRepository {
	if config == nil {
		config = DefaultDBConfig()
	}
	return &EntRepository{config: config}
}

// Connect 连接数据库
// 实际使用：
// client, err := ent.Open("postgres", config.DSN())
// if err != nil { return err }
func (r *EntRepository) Connect() error {
	log.Printf("Ent 连接数据库: %s", r.config.DSN())
	return nil
}

// ==================== ENT CRUD 操作 ====================

// CreateUser 创建用户
func (r *EntRepository) CreateUser(ctx context.Context, name, email string, age int) (*User, error) {
	// 实际使用：
	// user, err := r.client.User.
	//     Create().
	//     SetName(name).
	//     SetEmail(email).
	//     SetAge(age).
	//     Save(ctx)
	// return user, err

	user := &User{
		ID:        1,
		Name:      name,
		Email:     email,
		Age:       age,
		Status:    "active",
		CreatedAt: time.Now(),
	}
	log.Printf("Ent 创建用户: %+v", user)
	return user, nil
}

// GetUserByID 根据 ID 查询用户
func (r *EntRepository) GetUserByID(ctx context.Context, id int) (*User, error) {
	// 实际使用：
	// user, err := r.client.User.Get(ctx, id)
	// return user, err
	return &User{ID: uint(id), Name: "Ent用户", Email: "ent@example.com"}, nil
}

// UpdateUser 更新用户
func (r *EntRepository) UpdateUser(ctx context.Context, id int, name string, age int) error {
	// 实际使用：
	// _, err := r.client.User.
	//     UpdateOneID(id).
	//     SetName(name).
	//     SetAge(age).
	//     Save(ctx)
	// return err
	log.Printf("Ent 更新用户: id=%d, name=%s, age=%d", id, name, age)
	return nil
}

// DeleteUser 删除用户
func (r *EntRepository) DeleteUser(ctx context.Context, id int) error {
	// 实际使用：
	// return r.client.User.DeleteOneID(id).Exec(ctx)
	log.Printf("Ent 删除用户: %d", id)
	return nil
}

// ListUsers 分页查询用户
func (r *EntRepository) ListUsers(ctx context.Context, page, pageSize int) ([]User, int, error) {
	// 实际使用：
	// total, err := r.client.User.Query().Count(ctx)
	// if err != nil { return nil, 0, err }
	// 
	// offset := (page - 1) * pageSize
	// users, err := r.client.User.
	//     Query().
	//     Offset(offset).
	//     Limit(pageSize).
	//     All(ctx)
	// return users, total, err

	users := []User{
		{ID: 1, Name: "Ent用户1", Email: "ent1@example.com"},
		{ID: 2, Name: "Ent用户2", Email: "ent2@example.com"},
	}
	return users, 2, nil
}

// QueryUsersWithPredicate 条件查询
func (r *EntRepository) QueryUsersWithPredicate(ctx context.Context, minAge int, status string) ([]User, error) {
	// 实际使用：
	// users, err := r.client.User.
	//     Query().
	//     Where(
	//         user.AgeGTE(minAge),
	//         user.StatusEQ(status),
	//     ).
	//     Order(ent.Asc(user.FieldAge)).
	//     All(ctx)
	// return users, err

	return []User{}, nil
}

// CreateOrderWithItems 创建订单（事务）
func (r *EntRepository) CreateOrderWithItems(ctx context.Context, userID int, orderNo string, items []OrderItemInput) error {
	// 实际使用：
	// tx, err := r.client.Tx(ctx)
	// if err != nil { return err }
	// defer tx.Rollback()
	// 
	// order, err := tx.Order.
	//     Create().
	//     SetUserID(userID).
	//     SetOrderNo(orderNo).
	//     SetStatus("pending").
	//     Save(ctx)
	// if err != nil { return err }
	// 
	// bulk := make([]*ent.OrderItemCreate, len(items))
	// for i, item := range items {
	//     bulk[i] = tx.OrderItem.
	//         Create().
	//         SetOrderID(order.ID).
	//         SetProductID(item.ProductID).
	//         SetQuantity(item.Quantity).
	//         SetPrice(item.Price)
	// }
	// _, err = tx.OrderItem.CreateBulk(bulk...).Save(ctx)
	// if err != nil { return err }
	// 
	// return tx.Commit()

	log.Printf("Ent 创建订单: userID=%d, orderNo=%s", userID, orderNo)
	return nil
}

// OrderItemInput 订单项输入
type OrderItemInput struct {
	ProductID uint
	Quantity  int
	Price     float64
}

// ==================== 数据库迁移 ====================

// MigrationManager 迁移管理器
type MigrationManager struct {
	config *DBConfig
}

// NewMigrationManager 创建迁移管理器
func NewMigrationManager(config *DBConfig) *MigrationManager {
	return &MigrationManager{config: config}
}

// AutoMigrateGORM GORM 自动迁移
func (m *MigrationManager) AutoMigrateGORM() error {
	// 实际使用：
	// return db.AutoMigrate(
	//     &User{},
	//     &Profile{},
	//     &Order{},
	//     &OrderItem{},
	// )
	log.Println("GORM 自动迁移数据库表结构")
	return nil
}

// AutoMigrateEnt Ent 自动迁移
func (m *MigrationManager) AutoMigrateEnt(ctx context.Context) error {
	// 实际使用：
	// return client.Schema.Create(ctx)
	// 或者：
	// return client.Schema.CreateIfNotExists(ctx)
	log.Println("Ent 自动迁移数据库表结构")
	return nil
}

// ==================== 连接池最佳实践 ====================

// PoolConfig 连接池配置
type PoolConfig struct {
	MaxOpenConns    int           // 最大打开连接数
	MaxIdleConns    int           // 最大空闲连接数
	ConnMaxLifetime time.Duration // 连接最大生命周期
	ConnMaxIdleTime time.Duration // 空闲连接最大生命周期
}

// DefaultPoolConfig 返回默认连接池配置
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
	}
}

// HighConcurrencyPoolConfig 高并发场景连接池配置
func HighConcurrencyPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxOpenConns:    100,
		MaxIdleConns:    25,
		ConnMaxLifetime: 10 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

// ==================== 事务隔离级别 ====================

// IsolationLevel 事务隔离级别
type IsolationLevel int

const (
	LevelDefault IsolationLevel = iota
	LevelReadUncommitted
	LevelReadCommitted
	LevelRepeatableRead
	LevelSerializable
)

// GetIsolationLevelName 获取隔离级别名称
func GetIsolationLevelName(level IsolationLevel) string {
	names := map[IsolationLevel]string{
		LevelDefault:          "DEFAULT",
		LevelReadUncommitted:  "READ UNCOMMITTED",
		LevelReadCommitted:    "READ COMMITTED",
		LevelRepeatableRead:   "REPEATABLE READ",
		LevelSerializable:     "SERIALIZABLE",
	}
	return names[level]
}

// ==================== 健康检查 ====================

// HealthChecker 健康检查器
type HealthChecker struct {
	config *DBConfig
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(config *DBConfig) *HealthChecker {
	return &HealthChecker{config: config}
}

// Check 执行健康检查
func (h *HealthChecker) Check(ctx context.Context) error {
	// 实际使用：
	// ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	// defer cancel()
	// 
	// return pool.Ping(ctx)
	return nil
}

// CheckWithRetry 带重试的健康检查
func (h *HealthChecker) CheckWithRetry(ctx context.Context, maxRetries int) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := h.Check(ctx); err == nil {
			return nil
		} else {
			lastErr = err
			time.Sleep(time.Second * time.Duration(i+1))
		}
	}
	return fmt.Errorf("健康检查失败，重试 %d 次后仍然失败: %w", maxRetries, lastErr)
}

// ==================== 工具函数 ====================

// FormatDSN 格式化 DSN
func FormatDSN(host string, port int, user, password, database, sslMode string) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, database, sslMode)
}

// ValidateConfig 验证数据库配置
func ValidateConfig(config *DBConfig) error {
	if config.Host == "" {
		return fmt.Errorf("host 不能为空")
	}
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("port 必须在 1-65535 之间")
	}
	if config.User == "" {
		return fmt.Errorf("user 不能为空")
	}
	if config.Database == "" {
		return fmt.Errorf("database 不能为空")
	}
	return nil
}

// BuildCreateTableSQL 构建建表 SQL（用于演示）
func BuildCreateTableSQL() string {
	return `
-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    age INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- 用户资料表
CREATE TABLE IF NOT EXISTS profiles (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    bio TEXT,
    avatar VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 订单表
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    order_no VARCHAR(50) UNIQUE NOT NULL,
    total_amount DECIMAL(10,2),
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 订单项表
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL,
    quantity INTEGER DEFAULT 1,
    price DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
`
}
