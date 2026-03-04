// pgx_practices_test.go - PostgreSQL 数据库操作测试

package pgx_practices

import (
	"context"
	"testing"
	"time"
)

// 注意：这些测试需要本地运行 PostgreSQL 服务
// 可以使用 Docker 启动: docker run -d --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:latest

// TestDBConfig 测试数据库配置
func TestDBConfig(t *testing.T) {
	config := DefaultDBConfig()

	if config.Host != "localhost" {
		t.Errorf("默认 Host 应该是 localhost，实际: %s", config.Host)
	}
	if config.Port != 5432 {
		t.Errorf("默认 Port 应该是 5432，实际: %d", config.Port)
	}
	if config.MaxOpenConns != 25 {
		t.Errorf("默认 MaxOpenConns 应该是 25，实际: %d", config.MaxOpenConns)
	}
}

// TestDSN 测试 DSN 生成
func TestDSN(t *testing.T) {
	config := &DBConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "secret",
		Database: "testdb",
		SSLMode:  "disable",
	}

	dsn := config.DSN()
	expected := "host=localhost port=5432 user=postgres password=secret dbname=testdb sslmode=disable"

	if dsn != expected {
		t.Errorf("DSN 不正确\n期望: %s\n实际: %s", expected, dsn)
	}
}

// TestValidateConfig 测试配置验证
func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *DBConfig
		wantErr bool
	}{
		{
			name:    "有效配置",
			config:  DefaultDBConfig(),
			wantErr: false,
		},
		{
			name: "缺少 Host",
			config: &DBConfig{
				Port:     5432,
				User:     "postgres",
				Database: "testdb",
			},
			wantErr: true,
		},
		{
			name: "无效端口",
			config: &DBConfig{
				Host:     "localhost",
				Port:     0,
				User:     "postgres",
				Database: "testdb",
			},
			wantErr: true,
		},
		{
			name: "端口超出范围",
			config: &DBConfig{
				Host:     "localhost",
				Port:     70000,
				User:     "postgres",
				Database: "testdb",
			},
			wantErr: true,
		},
		{
			name: "缺少 User",
			config: &DBConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
			},
			wantErr: true,
		},
		{
			name: "缺少 Database",
			config: &DBConfig{
				Host: "localhost",
				Port: 5432,
				User: "postgres",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGORMRepository 测试 GORM 仓储
func TestGORMRepository(t *testing.T) {
	t.Skip("跳过：需要本地 PostgreSQL 环境")

	repo := NewGORMRepository(nil)
	ctx := context.Background()

	// 测试创建用户
	user := &User{
		Name:  "测试用户",
		Email: "test@example.com",
		Age:   25,
	}
	err := repo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	// 测试查询用户
	found, err := repo.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("查询用户失败: %v", err)
	}
	if found.Name != user.Name {
		t.Errorf("用户名称不匹配: 期望 %s, 实际 %s", user.Name, found.Name)
	}

	// 测试更新用户
	user.Name = "更新后的名称"
	err = repo.UpdateUser(ctx, user)
	if err != nil {
		t.Fatalf("更新用户失败: %v", err)
	}

	// 测试删除用户
	err = repo.DeleteUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("删除用户失败: %v", err)
	}
}

// TestEntRepository 测试 Ent 仓储
func TestEntRepository(t *testing.T) {
	t.Skip("跳过：需要本地 PostgreSQL 环境")

	repo := NewEntRepository(nil)
	ctx := context.Background()

	// 测试创建用户
	user, err := repo.CreateUser(ctx, "Ent用户", "ent@example.com", 30)
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	// 测试查询用户
	found, err := repo.GetUserByID(ctx, int(user.ID))
	if err != nil {
		t.Fatalf("查询用户失败: %v", err)
	}
	if found.Name != "Ent用户" {
		t.Errorf("用户名称不匹配")
	}

	// 测试更新用户
	err = repo.UpdateUser(ctx, int(user.ID), "新名称", 31)
	if err != nil {
		t.Fatalf("更新用户失败: %v", err)
	}

	// 测试删除用户
	err = repo.DeleteUser(ctx, int(user.ID))
	if err != nil {
		t.Fatalf("删除用户失败: %v", err)
	}
}

// TestUserQueryBuilder 测试用户查询构建器
func TestUserQueryBuilder(t *testing.T) {
	builder := NewUserQueryBuilder().
		WithStatus("active").
		WithAgeRange(18, 60).
		WithNameLike("张").
		WithPagination(1, 20).
		WithOrderBy("created_at DESC")

	if builder.status != "active" {
		t.Errorf("状态应该是 active")
	}
	if builder.minAge != 18 {
		t.Errorf("最小年龄应该是 18")
	}
	if builder.maxAge != 60 {
		t.Errorf("最大年龄应该是 60")
	}
	if builder.nameLike != "张" {
		t.Errorf("名称查询应该是 张")
	}
	if builder.page != 1 {
		t.Errorf("页码应该是 1")
	}
	if builder.pageSize != 20 {
		t.Errorf("每页大小应该是 20")
	}
}

// TestListUsers 测试分页查询
func TestListUsers(t *testing.T) {
	t.Skip("跳过：需要本地 PostgreSQL 环境")

	repo := NewGORMRepository(nil)
	ctx := context.Background()

	users, total, err := repo.ListUsers(ctx, 1, 10, "测试")
	if err != nil {
		t.Fatalf("查询用户列表失败: %v", err)
	}

	if total < 0 {
		t.Errorf("总数不能为负数")
	}

	if len(users) > 10 {
		t.Errorf("返回数量不应超过每页大小")
	}
}

// TestCreateOrderWithItems 测试创建订单
func TestCreateOrderWithItems(t *testing.T) {
	t.Skip("跳过：需要本地 PostgreSQL 环境")

	repo := NewGORMRepository(nil)
	ctx := context.Background()

	order := &Order{
		UserID:      1,
		OrderNo:     "ORD" + time.Now().Format("20060102150405"),
		TotalAmount: 199.99,
		Status:      "pending",
	}

	items := []OrderItem{
		{ProductID: 1, Quantity: 2, Price: 99.99},
		{ProductID: 2, Quantity: 1, Price: 49.99},
	}

	err := repo.CreateOrderWithItems(ctx, order, items)
	if err != nil {
		t.Fatalf("创建订单失败: %v", err)
	}
}

// TestPoolConfig 测试连接池配置
func TestPoolConfig(t *testing.T) {
	// 默认配置
	defaultConfig := DefaultPoolConfig()
	if defaultConfig.MaxOpenConns != 25 {
		t.Errorf("默认 MaxOpenConns 应该是 25")
	}

	// 高并发配置
	highConfig := HighConcurrencyPoolConfig()
	if highConfig.MaxOpenConns != 100 {
		t.Errorf("高并发 MaxOpenConns 应该是 100")
	}
}

// TestIsolationLevel 测试事务隔离级别
func TestIsolationLevel(t *testing.T) {
	tests := []struct {
		level    IsolationLevel
		expected string
	}{
		{LevelDefault, "DEFAULT"},
		{LevelReadUncommitted, "READ UNCOMMITTED"},
		{LevelReadCommitted, "READ COMMITTED"},
		{LevelRepeatableRead, "REPEATABLE READ"},
		{LevelSerializable, "SERIALIZABLE"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			name := GetIsolationLevelName(tt.level)
			if name != tt.expected {
				t.Errorf("隔离级别名称错误: 期望 %s, 实际 %s", tt.expected, name)
			}
		})
	}
}

// TestHealthChecker 测试健康检查
func TestHealthChecker(t *testing.T) {
	t.Skip("跳过：需要本地 PostgreSQL 环境")

	checker := NewHealthChecker(DefaultDBConfig())
	ctx := context.Background()

	err := checker.Check(ctx)
	if err != nil {
		t.Errorf("健康检查失败: %v", err)
	}
}

// TestFormatDSN 测试 DSN 格式化
func TestFormatDSN(t *testing.T) {
	dsn := FormatDSN("localhost", 5432, "user", "pass", "db", "disable")
	expected := "host=localhost port=5432 user=user password=pass dbname=db sslmode=disable"

	if dsn != expected {
		t.Errorf("DSN 格式错误\n期望: %s\n实际: %s", expected, dsn)
	}
}

// TestBuildCreateTableSQL 测试建表 SQL
func TestBuildCreateTableSQL(t *testing.T) {
	sql := BuildCreateTableSQL()

	if sql == "" {
		t.Error("建表 SQL 不应为空")
	}

	// 检查是否包含必要的表
	tables := []string{"users", "profiles", "orders", "order_items"}
	for _, table := range tables {
		if !contains(sql, table) {
			t.Errorf("SQL 应包含表 %s", table)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// BenchmarkUserOperations 用户操作基准测试
func BenchmarkUserOperations(b *testing.B) {
	b.Skip("跳过：需要本地 PostgreSQL 环境")

	repo := NewGORMRepository(nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &User{
			Name:  "基准测试用户",
			Email: "bench@example.com",
			Age:   25,
		}
		repo.CreateUser(ctx, user)
	}
}

// ExampleDBConfig 数据库配置示例
func ExampleDBConfig() {
	config := &DBConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		Database:        "testdb",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
	}

	fmt.Println(config.DSN())
	// Output: host=localhost port=5432 user=postgres password=postgres dbname=testdb sslmode=disable
}

// ExampleUserQueryBuilder 用户查询构建器示例
func ExampleUserQueryBuilder() {
	builder := NewUserQueryBuilder().
		WithStatus("active").
		WithAgeRange(18, 60).
		WithPagination(1, 20)

	fmt.Printf("状态: %s, 年龄范围: %d-%d", builder.status, builder.minAge, builder.maxAge)
	// Output: 状态: active, 年龄范围: 18-60
}
