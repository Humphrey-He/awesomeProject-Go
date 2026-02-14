package testing_practices

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
)

// ========== Go Test 最佳实践 ==========

/*
本文件展示Go测试的各种最佳实践，包括：
1. 单元测试编写规范
2. 表驱动测试
3. 子测试
4. 测试辅助函数
5. Setup/Teardown
6. 测试覆盖率
7. 基准测试
8. 示例测试
9. HTTP测试
10. 文件I/O测试
*/

// ========== 1. 被测试的代码示例 ==========

// Calculator 计算器
type Calculator struct {
	precision int
}

func NewCalculator(precision int) *Calculator {
	return &Calculator{precision: precision}
}

func (c *Calculator) Add(a, b float64) float64 {
	return a + b
}

func (c *Calculator) Subtract(a, b float64) float64 {
	return a - b
}

func (c *Calculator) Multiply(a, b float64) float64 {
	return a * b
}

func (c *Calculator) Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return a / b, nil
}

// StringUtils 字符串工具
type StringUtils struct{}

func (s *StringUtils) Reverse(str string) string {
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func (s *StringUtils) IsPalindrome(str string) bool {
	str = strings.ToLower(str)
	return str == s.Reverse(str)
}

// User 用户模型
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func (u *User) Validate() error {
	if u.Name == "" {
		return fmt.Errorf("name is required")
	}
	if u.Email == "" {
		return fmt.Errorf("email is required")
	}
	if u.Age < 0 || u.Age > 150 {
		return fmt.Errorf("invalid age: %d", u.Age)
	}
	return nil
}

// UserService 用户服务
type UserService struct {
	users map[int]*User
}

func NewUserService() *UserService {
	return &UserService{
		users: make(map[int]*User),
	}
}

func (s *UserService) CreateUser(user *User) error {
	if err := user.Validate(); err != nil {
		return err
	}
	s.users[user.ID] = user
	return nil
}

func (s *UserService) GetUser(id int) (*User, error) {
	user, ok := s.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found: %d", id)
	}
	return user, nil
}

func (s *UserService) UpdateUser(user *User) error {
	if _, ok := s.users[user.ID]; !ok {
		return fmt.Errorf("user not found: %d", user.ID)
	}
	if err := user.Validate(); err != nil {
		return err
	}
	s.users[user.ID] = user
	return nil
}

func (s *UserService) DeleteUser(id int) error {
	if _, ok := s.users[id]; !ok {
		return fmt.Errorf("user not found: %d", id)
	}
	delete(s.users, id)
	return nil
}

// ========== HTTP Handler示例 ==========

type UserHandler struct {
	service *UserService
}

func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.CreateUser(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// 简化实现，实际应该从URL参数获取ID
	id := 1

	user, err := h.service.GetUser(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// ========== 文件操作示例 ==========

type FileProcessor struct {
	basePath string
}

func NewFileProcessor(basePath string) *FileProcessor {
	return &FileProcessor{basePath: basePath}
}

func (f *FileProcessor) ReadFile(filename string) (string, error) {
	data, err := os.ReadFile(f.basePath + "/" + filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (f *FileProcessor) WriteFile(filename, content string) error {
	return os.WriteFile(f.basePath+"/"+filename, []byte(content), 0644)
}

// ========== 测试最佳实践示例代码 ==========

/*
以下是测试代码的最佳实践示例，实际测试代码在 testing_best_practices_test.go 中
*/

// TestHelper 测试辅助函数模式
type TestHelper struct {
	t *testing.T
}

func NewTestHelper(t *testing.T) *TestHelper {
	t.Helper() // 标记为helper，错误会报告在调用处
	return &TestHelper{t: t}
}

func (h *TestHelper) AssertEqual(got, want interface{}) {
	h.t.Helper()
	if got != want {
		h.t.Errorf("got %v, want %v", got, want)
	}
}

func (h *TestHelper) AssertNoError(err error) {
	h.t.Helper()
	if err != nil {
		h.t.Fatalf("unexpected error: %v", err)
	}
}

func (h *TestHelper) AssertError(err error, msg string) {
	h.t.Helper()
	if err == nil {
		h.t.Fatalf("expected error containing %q, got nil", msg)
	}
	if !strings.Contains(err.Error(), msg) {
		h.t.Fatalf("expected error containing %q, got %q", msg, err.Error())
	}
}

// ========== 测试夹具（Test Fixtures）==========

type TestFixture struct {
	service *UserService
	users   []*User
}

func SetupTestFixture() *TestFixture {
	service := NewUserService()
	users := []*User{
		{ID: 1, Name: "Alice", Email: "alice@example.com", Age: 30},
		{ID: 2, Name: "Bob", Email: "bob@example.com", Age: 25},
		{ID: 3, Name: "Charlie", Email: "charlie@example.com", Age: 35},
	}

	for _, user := range users {
		service.CreateUser(user)
	}

	return &TestFixture{
		service: service,
		users:   users,
	}
}

func (f *TestFixture) Teardown() {
	// 清理资源
	f.service = nil
	f.users = nil
}

// ========== Golden Files 模式 ==========

func SaveGoldenFile(t *testing.T, filename string, data []byte) {
	t.Helper()
	if err := os.WriteFile(filename, data, 0644); err != nil {
		t.Fatalf("failed to write golden file: %v", err)
	}
}

func LoadGoldenFile(t *testing.T, filename string) []byte {
	t.Helper()
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}
	return data
}

func CompareWithGolden(t *testing.T, got []byte, goldenFile string, update bool) {
	t.Helper()

	if update {
		SaveGoldenFile(t, goldenFile, got)
		return
	}

	want := LoadGoldenFile(t, goldenFile)
	if !bytes.Equal(got, want) {
		t.Errorf("output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// ========== 测试最佳实践总结 ==========

/*
Go Test 最佳实践清单：

✅ 1. 命名规范
   - 测试文件：xxx_test.go
   - 测试函数：TestXxx
   - 基准测试：BenchmarkXxx
   - 示例测试：ExampleXxx

✅ 2. 表驱动测试
   - 使用struct切片定义测试案例
   - 每个案例包含：name, input, want
   - 易于添加新案例
   - 清晰的测试逻辑

✅ 3. 子测试（Subtests）
   - 使用 t.Run() 创建子测试
   - 可以独立运行：go test -run TestName/SubtestName
   - 更好的错误定位

✅ 4. 测试辅助函数
   - 使用 t.Helper() 标记
   - 错误报告在调用处
   - 减少重复代码

✅ 5. Setup/Teardown
   - TestMain 用于全局setup/teardown
   - 子测试的 defer 用于局部teardown
   - 确保资源清理

✅ 6. 错误处理
   - t.Error() - 记录错误继续
   - t.Fatal() - 记录错误并停止
   - 使用 t.Helper() 改善错误位置

✅ 7. 并行测试
   - t.Parallel() 标记并行测试
   - 注意共享状态
   - 提高测试速度

✅ 8. HTTP测试
   - httptest.NewRecorder() 记录响应
   - httptest.NewServer() 模拟服务器
   - 测试handler逻辑

✅ 9. 基准测试
   - 使用 b.ResetTimer()
   - 使用 b.N 作为循环次数
   - 添加 -benchmem 查看内存分配

✅ 10. 示例测试
   - 作为文档
   - 可执行的示例
   - Output: 注释验证输出

❌ 避免的陷阱：

1. 全局状态
   - 避免测试间共享状态
   - 使用 setup/teardown

2. 外部依赖
   - 使用 mock/stub
   - 不依赖网络/数据库

3. 时间依赖
   - 使用可注入的时钟
   - 避免 time.Sleep()

4. 随机性
   - 使用固定的随机种子
   - 或使测试确定性

5. 顺序依赖
   - 测试应该独立
   - 可以任意顺序运行

⚡ 性能优化：

1. 并行测试
   - t.Parallel()
   - 充分利用多核

2. 缓存
   - 缓存昂贵的setup
   - sync.Once

3. 跳过慢速测试
   - testing.Short()
   - go test -short

🎯 测试策略：

1. 单元测试
   - 测试单个函数/方法
   - 快速、独立
   - 高覆盖率

2. 集成测试
   - 测试多个组件交互
   - 使用真实依赖或轻量mock
   - 较慢但更真实

3. 端到端测试
   - 测试完整流程
   - 最接近生产环境
   - 数量较少

4. 测试金字塔
   - 70% 单元测试
   - 20% 集成测试
   - 10% 端到端测试
*/
