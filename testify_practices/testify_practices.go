package testify_practices

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ========== Go Testify 测试框架深入理解 ==========

/*
本项目讲解 Go 测试框架 Testify 的使用，包括：

一、断言库 (assert/require)
1. 基础断言
2. 比较断言
3. 类型断言
4. 错误断言

二、Suite 测试套件
1. Suite 结构
2. Setup/Teardown
3. 子测试

三、Mock 模拟对象
1. Mock 定义
2. Expectation 期望
3. Stub 实现
*/

// ========== 1. 断言库 ==========

/*
assert vs require：

- assert: 失败后继续执行
- require: 失败后立即停止

常用断言：
- assert.Equal
- assert.NotEqual
- assert.Nil
- assert.NotNil
- assert.True
- assert.False
- assert.Contains
- assert.Len
- assert.NoError
- assert.Error
*/

// ========== 1.1 基础断言示例 ==========

// Add 简单加法函数
func Add(a, b int) int {
	return a + b
}

// Subtract 减法函数
func Subtract(a, b int) int {
	return a - b
}

// Multiply 乘法函数
func Multiply(a, b int) int {
	return a * b
}

// TestBasicAssertions 基础断言示例
func TestBasicAssertions(t *testing.T) {
	// Equal 断言
	assert.Equal(t, 4, Add(2, 2), "2 + 2 should equal 4")
	assert.Equal(t, 1, Subtract(3, 2), "3 - 2 should equal 1")
	assert.Equal(t, 6, Multiply(2, 3), "2 * 3 should equal 6")
	
	// NotEqual 断言
	assert.NotEqual(t, 5, Add(2, 2), "2 + 2 should not equal 5")
	
	// 使用 require（失败后停止）
	require.Equal(t, 4, Add(2, 2))
}

// TestNilAssertions nil 断言示例
func TestNilAssertions(t *testing.T) {
	var err error
	assert.Nil(t, err, "error should be nil")
	
	err = errors.New("test error")
	assert.NotNil(t, err, "error should not be nil")
}

// TestBoolAssertions 布尔断言示例
func TestBoolAssertions(t *testing.T) {
	assert.True(t, true, "should be true")
	assert.False(t, false, "should be false")
	
	// 条件断言
	assert.True(t, Add(1, 1) == 2)
}

// ========== 1.2 比较断言示例 ==========

// TestComparisonAssertions 比较断言示例
func TestComparisonAssertions(t *testing.T) {
	// 大于/小于
	assert.Greater(t, 10, 5, "10 should be greater than 5")
	assert.GreaterOrEqual(t, 10, 10, "10 should be greater or equal to 10")
	assert.Less(t, 5, 10, "5 should be less than 10")
	assert.LessOrEqual(t, 5, 5, "5 should be less or equal to 5")
	
	// 浮点数比较
	assert.InDelta(t, 1.0, 1.001, 0.01, "values should be within delta")
	assert.InEpsilon(t, 1.0, 1.01, 0.02, "values should be within epsilon")
}

// TestContainsAssertions 包含断言示例
func TestContainsAssertions(t *testing.T) {
	// 字符串包含
	assert.Contains(t, "Hello World", "World")
	assert.NotContains(t, "Hello World", "Go")
	
	// 切片包含
	assert.Contains(t, []int{1, 2, 3}, 2)
	
	// Map 包含
	assert.Contains(t, map[string]int{"a": 1}, "a")
}

// TestLenAssertions 长度断言示例
func TestLenAssertions(t *testing.T) {
	assert.Len(t, "hello", 5)
	assert.Len(t, []int{1, 2, 3}, 3)
	assert.Len(t, map[string]int{"a": 1, "b": 2}, 2)
}

// ========== 1.3 错误断言示例 ==========

// Divide 除法（可能返回错误）
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

// TestErrorAssertions 错误断言示例
func TestErrorAssertions(t *testing.T) {
	// 无错误
	result, err := Divide(10, 2)
	assert.NoError(t, err)
	assert.Equal(t, 5.0, result)
	
	// 有错误
	_, err = Divide(10, 0)
	assert.Error(t, err)
	assert.Equal(t, "division by zero", err.Error())
	assert.ErrorContains(t, err, "zero")
}

// ========== 1.4 类型断言示例 ==========

// TestTypeAssertions 类型断言示例
func TestTypeAssertions(t *testing.T) {
	var iface interface{} = "hello"
	
	// 类型检查
	assert.IsType(t, "hello", iface)
	
	// 类型断言
	str, ok := iface.(string)
	assert.True(t, ok)
	assert.Equal(t, "hello", str)
	
	// 实现检查
	var _ fmt.Stringer = (*StringerImpl)(nil)
}

// StringerImpl 实现 fmt.Stringer
type StringerImpl struct {
	value string
}

func (s *StringerImpl) String() string {
	return s.value
}

// ========== 2. Suite 测试套件 ==========

/*
Suite 特点：
- 类似面向对象测试
- Setup/Teardown 钩子
- 子测试支持
- 共享状态
*/

// UserService 用户服务
type UserService struct {
	users map[string]string
}

// NewUserService 创建服务
func NewUserService() *UserService {
	return &UserService{
		users: make(map[string]string),
	}
}

// AddUser 添加用户
func (s *UserService) AddUser(id, name string) error {
	if id == "" {
		return errors.New("id is empty")
	}
	s.users[id] = name
	return nil
}

// GetUser 获取用户
func (s *UserService) GetUser(id string) (string, error) {
	name, ok := s.users[id]
	if !ok {
		return "", errors.New("user not found")
	}
	return name, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id string) {
	delete(s.users, id)
}

// Count 用户数量
func (s *UserService) Count() int {
	return len(s.users)
}

// UserSuite 测试套件
type UserSuite struct {
	suite.Suite
	service *UserService
}

// SetupSuite 套件级别 Setup（所有测试前执行一次）
func (s *UserSuite) SetupSuite() {
	s.T().Log("SetupSuite: 初始化测试套件")
}

// TeardownSuite 套件级别 Teardown（所有测试后执行一次）
func (s *UserSuite) TeardownSuite() {
	s.T().Log("TeardownSuite: 清理测试套件")
}

// SetupTest 测试级别 Setup（每个测试前执行）
func (s *UserSuite) SetupTest() {
	s.service = NewUserService()
	s.T().Log("SetupTest: 创建新的 UserService")
}

// TearDownTest 测试级别 Teardown（每个测试后执行）
func (s *UserSuite) TearDownTest() {
	s.service = nil
	s.T().Log("TearDownTest: 清理 UserService")
}

// TestAddUser 测试添加用户
func (s *UserSuite) TestAddUser() {
	err := s.service.AddUser("1", "Alice")
	s.NoError(err)
	s.Equal(1, s.service.Count())
	
	name, err := s.service.GetUser("1")
	s.NoError(err)
	s.Equal("Alice", name)
}

// TestAddUserEmptyID 测试空ID
func (s *UserSuite) TestAddUserEmptyID() {
	err := s.service.AddUser("", "Alice")
	s.Error(err)
	s.Equal("id is empty", err.Error())
}

// TestGetUserNotFound 测试用户不存在
func (s *UserSuite) TestGetUserNotFound() {
	_, err := s.service.GetUser("nonexistent")
	s.Error(err)
	s.Equal("user not found", err.Error())
}

// TestDeleteUser 测试删除用户
func (s *UserSuite) TestDeleteUser() {
	s.service.AddUser("1", "Alice")
	s.Equal(1, s.service.Count())
	
	s.service.DeleteUser("1")
	s.Equal(0, s.service.Count())
	
	_, err := s.service.GetUser("1")
	s.Error(err)
}

// TestUserSuite 运行 Suite
func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserSuite))
}

// ========== 3. Mock 示例 ==========

// Database 接口
type Database interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Delete(key string) error
}

// MockDatabase Mock 实现
type MockDatabase struct {
	data    map[string]string
	GetErr  error
	SetErr  error
	DelErr  error
}

// NewMockDatabase 创建 Mock
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		data: make(map[string]string),
	}
}

func (m *MockDatabase) Get(key string) (string, error) {
	if m.GetErr != nil {
		return "", m.GetErr
	}
	return m.data[key], nil
}

func (m *MockDatabase) Set(key, value string) error {
	if m.SetErr != nil {
		return m.SetErr
	}
	m.data[key] = value
	return nil
}

func (m *MockDatabase) Delete(key string) error {
	if m.DelErr != nil {
		return m.DelErr
	}
	delete(m.data, key)
	return nil
}

// TestWithMock 使用 Mock 测试
func TestWithMock(t *testing.T) {
	mockDB := NewMockDatabase()
	
	// 测试 Set
	err := mockDB.Set("key", "value")
	assert.NoError(t, err)
	
	// 测试 Get
	val, err := mockDB.Get("key")
	assert.NoError(t, err)
	assert.Equal(t, "value", val)
	
	// 测试 Delete
	err = mockDB.Delete("key")
	assert.NoError(t, err)
	
	// 测试删除后获取
	val, err = mockDB.Get("key")
	assert.NoError(t, err)
	assert.Equal(t, "", val)
}

// TestMockError 测试 Mock 返回错误
func TestMockError(t *testing.T) {
	mockDB := NewMockDatabase()
	mockDB.GetErr = errors.New("connection error")
	
	_, err := mockDB.Get("key")
	assert.Error(t, err)
	assert.Equal(t, "connection error", err.Error())
}

// ========== 4. Stub 示例 ==========

// StubDatabase Stub 实现
type StubDatabase struct {
	GetFunc    func(key string) (string, error)
	SetFunc    func(key, value string) error
	DeleteFunc func(key string) error
}

func (s *StubDatabase) Get(key string) (string, error) {
	if s.GetFunc != nil {
		return s.GetFunc(key)
	}
	return "", nil
}

func (s *StubDatabase) Set(key, value string) error {
	if s.SetFunc != nil {
		return s.SetFunc(key, value)
	}
	return nil
}

func (s *StubDatabase) Delete(key string) error {
	if s.DeleteFunc != nil {
		return s.DeleteFunc(key)
	}
	return nil
}

// TestWithStub 使用 Stub 测试
func TestWithStub(t *testing.T) {
	stubDB := &StubDatabase{
		GetFunc: func(key string) (string, error) {
			return "stubbed:" + key, nil
		},
	}
	
	val, err := stubDB.Get("test")
	assert.NoError(t, err)
	assert.Equal(t, "stubbed:test", val)
}

// ========== 5. 表驱动测试 ==========

// TestAddTableDriven 表驱动测试示例
func TestAddTableDriven(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"positive", 2, 3, 5},
		{"negative", -1, -1, -2},
		{"zero", 0, 0, 0},
		{"mixed", -1, 1, 0},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Add(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestDivideTableDriven 除法表驱动测试
func TestDivideTableDriven(t *testing.T) {
	tests := []struct {
		name      string
		a, b      float64
		expected  float64
		expectErr bool
	}{
		{"normal", 10, 2, 5, false},
		{"zero divisor", 10, 0, 0, true},
		{"negative", -10, 2, -5, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Divide(tt.a, tt.b)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// ========== 6. 性能测试 ==========

// BenchmarkAdd 性能测试示例
func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Add(1, 2)
	}
}

// BenchmarkAddParallel 并行性能测试
func BenchmarkAddParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Add(1, 2)
		}
	})
}

// ========== 7. 断言方法演示 ==========

// DemonstrateAssertions 断言方法演示
func DemonstrateAssertions() {
	fmt.Println("=== Testify 断言方法 ===")
	fmt.Println("\n基础断言:")
	fmt.Println("  assert.Equal(t, expected, actual)")
	fmt.Println("  assert.NotEqual(t, expected, actual)")
	fmt.Println("  assert.Nil(t, object)")
	fmt.Println("  assert.NotNil(t, object)")
	fmt.Println("  assert.True(t, condition)")
	fmt.Println("  assert.False(t, condition)")
	
	fmt.Println("\n比较断言:")
	fmt.Println("  assert.Greater(t, a, b)")
	fmt.Println("  assert.Less(t, a, b)")
	fmt.Println("  assert.InDelta(t, a, b, delta)")
	
	fmt.Println("\n容器断言:")
	fmt.Println("  assert.Contains(t, container, element)")
	fmt.Println("  assert.Len(t, container, length)")
	fmt.Println("  assert.Empty(t, container)")
	
	fmt.Println("\n错误断言:")
	fmt.Println("  assert.NoError(t, err)")
	fmt.Println("  assert.Error(t, err)")
	fmt.Println("  assert.ErrorIs(t, err, target)")
	fmt.Println("  assert.ErrorContains(t, err, substring)")
}

// CompleteExample 完整示例
func CompleteExample() {
	DemonstrateAssertions()
}
