package mock_practices

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ========== Go Mock 最佳实践 ==========

/*
本文件展示Go Mock的各种最佳实践，包括：
1. 接口定义与Mock
2. 手动Mock实现
3. Mock数据库
4. Mock HTTP客户端
5. Mock时间
6. 测试替身模式（Stub, Spy, Fake）
7. Mock最佳实践
*/

// ========== 1. 接口定义（为Mock而设计）==========

// UserRepository 用户仓储接口
type UserRepository interface {
	GetUser(ctx context.Context, id int) (*User, error)
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id int) error
	ListUsers(ctx context.Context, limit, offset int) ([]*User, error)
}

// EmailService 邮件服务接口
type EmailService interface {
	SendEmail(to, subject, body string) error
	SendTemplateEmail(to, templateID string, data map[string]interface{}) error
}

// PaymentGateway 支付网关接口
type PaymentGateway interface {
	Charge(amount float64, currency string, cardToken string) (string, error)
	Refund(transactionID string, amount float64) error
	GetTransaction(transactionID string) (*Transaction, error)
}

// CacheService 缓存服务接口
type CacheService interface {
	Get(key string) (string, error)
	Set(key, value string, ttl time.Duration) error
	Delete(key string) error
	Exists(key string) bool
}

// ========== 2. 领域模型 ==========

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Transaction struct {
	ID       string
	Amount   float64
	Currency string
	Status   string
}

// ========== 3. 业务逻辑层（依赖接口）==========

// UserService 用户服务
type UserService struct {
	repo  UserRepository
	email EmailService
	cache CacheService
}

func NewUserService(repo UserRepository, email EmailService, cache CacheService) *UserService {
	return &UserService{
		repo:  repo,
		email: email,
		cache: cache,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, user *User) error {
	// 1. 创建用户
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	// 2. 发送欢迎邮件
	if err := s.email.SendEmail(user.Email, "Welcome!", "Welcome to our platform"); err != nil {
		// 邮件发送失败不影响注册
		fmt.Printf("failed to send email: %v\n", err)
	}
	
	// 3. 缓存用户信息
	cacheKey := fmt.Sprintf("user:%d", user.ID)
	s.cache.Set(cacheKey, user.Email, 1*time.Hour)
	
	return nil
}

func (s *UserService) GetUser(ctx context.Context, id int) (*User, error) {
	// 1. 尝试从缓存获取
	cacheKey := fmt.Sprintf("user:%d", id)
	if s.cache.Exists(cacheKey) {
		// 实际应该从缓存反序列化
		fmt.Println("Cache hit")
	}
	
	// 2. 从数据库获取
	user, err := s.repo.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// 3. 更新缓存
	s.cache.Set(cacheKey, user.Email, 1*time.Hour)
	
	return user, nil
}

// PaymentService 支付服务
type PaymentService struct {
	gateway PaymentGateway
	email   EmailService
}

func NewPaymentService(gateway PaymentGateway, email EmailService) *PaymentService {
	return &PaymentService{
		gateway: gateway,
		email:   email,
	}
}

func (s *PaymentService) ProcessPayment(userEmail string, amount float64, cardToken string) error {
	// 1. 执行支付
	transactionID, err := s.gateway.Charge(amount, "USD", cardToken)
	if err != nil {
		return fmt.Errorf("payment failed: %w", err)
	}
	
	// 2. 发送收据邮件
	subject := "Payment Receipt"
	body := fmt.Sprintf("Transaction ID: %s, Amount: $%.2f", transactionID, amount)
	if err := s.email.SendEmail(userEmail, subject, body); err != nil {
		// 邮件发送失败不影响支付
		fmt.Printf("failed to send receipt: %v\n", err)
	}
	
	return nil
}

// ========== 4. 手动Mock实现 ==========

// MockUserRepository 手动Mock实现
type MockUserRepository struct {
	users          map[int]*User
	GetUserFunc    func(ctx context.Context, id int) (*User, error)
	CreateUserFunc func(ctx context.Context, user *User) error
	UpdateUserFunc func(ctx context.Context, user *User) error
	DeleteUserFunc func(ctx context.Context, id int) error
	ListUsersFunc  func(ctx context.Context, limit, offset int) ([]*User, error)
	
	// 调用计数
	GetUserCalls    int
	CreateUserCalls int
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[int]*User),
	}
}

func (m *MockUserRepository) GetUser(ctx context.Context, id int) (*User, error) {
	m.GetUserCalls++
	if m.GetUserFunc != nil {
		return m.GetUserFunc(ctx, id)
	}
	user, ok := m.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *User) error {
	m.CreateUserCalls++
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, user)
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *User) error {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, user)
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id int) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, id)
	}
	delete(m.users, id)
	return nil
}

func (m *MockUserRepository) ListUsers(ctx context.Context, limit, offset int) ([]*User, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx, limit, offset)
	}
	users := make([]*User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

// MockEmailService Mock邮件服务
type MockEmailService struct {
	SendEmailFunc         func(to, subject, body string) error
	SendTemplateEmailFunc func(to, templateID string, data map[string]interface{}) error
	
	// 记录调用
	Calls []EmailCall
}

type EmailCall struct {
	To      string
	Subject string
	Body    string
}

func NewMockEmailService() *MockEmailService {
	return &MockEmailService{
		Calls: make([]EmailCall, 0),
	}
}

func (m *MockEmailService) SendEmail(to, subject, body string) error {
	m.Calls = append(m.Calls, EmailCall{To: to, Subject: subject, Body: body})
	if m.SendEmailFunc != nil {
		return m.SendEmailFunc(to, subject, body)
	}
	return nil
}

func (m *MockEmailService) SendTemplateEmail(to, templateID string, data map[string]interface{}) error {
	if m.SendTemplateEmailFunc != nil {
		return m.SendTemplateEmailFunc(to, templateID, data)
	}
	return nil
}

// MockCacheService Mock缓存服务
type MockCacheService struct {
	data      map[string]string
	GetFunc   func(key string) (string, error)
	SetFunc   func(key, value string, ttl time.Duration) error
	DeleteFunc func(key string) error
	ExistsFunc func(key string) bool
}

func NewMockCacheService() *MockCacheService {
	return &MockCacheService{
		data: make(map[string]string),
	}
}

func (m *MockCacheService) Get(key string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(key)
	}
	val, ok := m.data[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return val, nil
}

func (m *MockCacheService) Set(key, value string, ttl time.Duration) error {
	if m.SetFunc != nil {
		return m.SetFunc(key, value, ttl)
	}
	m.data[key] = value
	return nil
}

func (m *MockCacheService) Delete(key string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(key)
	}
	delete(m.data, key)
	return nil
}

func (m *MockCacheService) Exists(key string) bool {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(key)
	}
	_, ok := m.data[key]
	return ok
}

// MockPaymentGateway Mock支付网关
type MockPaymentGateway struct {
	ChargeFunc          func(amount float64, currency string, cardToken string) (string, error)
	RefundFunc          func(transactionID string, amount float64) error
	GetTransactionFunc  func(transactionID string) (*Transaction, error)
	
	// 记录支付
	Charges []ChargeRecord
}

type ChargeRecord struct {
	Amount    float64
	Currency  string
	CardToken string
	Success   bool
}

func NewMockPaymentGateway() *MockPaymentGateway {
	return &MockPaymentGateway{
		Charges: make([]ChargeRecord, 0),
	}
}

func (m *MockPaymentGateway) Charge(amount float64, currency string, cardToken string) (string, error) {
	record := ChargeRecord{Amount: amount, Currency: currency, CardToken: cardToken}
	
	if m.ChargeFunc != nil {
		txID, err := m.ChargeFunc(amount, currency, cardToken)
		record.Success = (err == nil)
		m.Charges = append(m.Charges, record)
		return txID, err
	}
	
	record.Success = true
	m.Charges = append(m.Charges, record)
	return "mock-tx-" + fmt.Sprint(len(m.Charges)), nil
}

func (m *MockPaymentGateway) Refund(transactionID string, amount float64) error {
	if m.RefundFunc != nil {
		return m.RefundFunc(transactionID, amount)
	}
	return nil
}

func (m *MockPaymentGateway) GetTransaction(transactionID string) (*Transaction, error) {
	if m.GetTransactionFunc != nil {
		return m.GetTransactionFunc(transactionID)
	}
	return &Transaction{
		ID:       transactionID,
		Amount:   100.0,
		Currency: "USD",
		Status:   "completed",
	}, nil
}

// ========== 5. Spy模式（记录调用）==========

type SpyEmailService struct {
	Calls    []EmailCall
	SendFunc func(to, subject, body string) error
}

func NewSpyEmailService() *SpyEmailService {
	return &SpyEmailService{
		Calls: make([]EmailCall, 0),
	}
}

func (s *SpyEmailService) SendEmail(to, subject, body string) error {
	s.Calls = append(s.Calls, EmailCall{To: to, Subject: subject, Body: body})
	if s.SendFunc != nil {
		return s.SendFunc(to, subject, body)
	}
	return nil
}

func (s *SpyEmailService) SendTemplateEmail(to, templateID string, data map[string]interface{}) error {
	return nil
}

func (s *SpyEmailService) WasCalled() bool {
	return len(s.Calls) > 0
}

func (s *SpyEmailService) CallCount() int {
	return len(s.Calls)
}

func (s *SpyEmailService) LastCall() *EmailCall {
	if len(s.Calls) == 0 {
		return nil
	}
	return &s.Calls[len(s.Calls)-1]
}

// ========== 6. Stub模式（预定义响应）==========

type StubUserRepository struct {
	users map[int]*User
}

func NewStubUserRepository() *StubUserRepository {
	return &StubUserRepository{
		users: map[int]*User{
			1: {ID: 1, Name: "Alice", Email: "alice@example.com"},
			2: {ID: 2, Name: "Bob", Email: "bob@example.com"},
		},
	}
}

func (s *StubUserRepository) GetUser(ctx context.Context, id int) (*User, error) {
	user, ok := s.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *StubUserRepository) CreateUser(ctx context.Context, user *User) error {
	return nil
}

func (s *StubUserRepository) UpdateUser(ctx context.Context, user *User) error {
	return nil
}

func (s *StubUserRepository) DeleteUser(ctx context.Context, id int) error {
	return nil
}

func (s *StubUserRepository) ListUsers(ctx context.Context, limit, offset int) ([]*User, error) {
	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users, nil
}

// ========== 7. Fake实现（轻量级真实实现）==========

type FakeUserRepository struct {
	users  map[int]*User
	nextID int
}

func NewFakeUserRepository() *FakeUserRepository {
	return &FakeUserRepository{
		users:  make(map[int]*User),
		nextID: 1,
	}
}

func (f *FakeUserRepository) GetUser(ctx context.Context, id int) (*User, error) {
	user, ok := f.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found: %d", id)
	}
	return user, nil
}

func (f *FakeUserRepository) CreateUser(ctx context.Context, user *User) error {
	if user.ID == 0 {
		user.ID = f.nextID
		f.nextID++
	}
	user.CreatedAt = time.Now()
	f.users[user.ID] = user
	return nil
}

func (f *FakeUserRepository) UpdateUser(ctx context.Context, user *User) error {
	if _, ok := f.users[user.ID]; !ok {
		return fmt.Errorf("user not found: %d", user.ID)
	}
	f.users[user.ID] = user
	return nil
}

func (f *FakeUserRepository) DeleteUser(ctx context.Context, id int) error {
	if _, ok := f.users[id]; !ok {
		return fmt.Errorf("user not found: %d", id)
	}
	delete(f.users, id)
	return nil
}

func (f *FakeUserRepository) ListUsers(ctx context.Context, limit, offset int) ([]*User, error) {
	users := make([]*User, 0, len(f.users))
	for _, user := range f.users {
		users = append(users, user)
	}
	return users, nil
}

// ========== Mock最佳实践总结 ==========

/*
Go Mock 最佳实践：

✅ 1. 接口设计
   - 定义小而精确的接口
   - 接口在使用方定义
   - 便于测试和Mock

✅ 2. Mock类型选择
   - Stub: 预定义响应，简单场景
   - Spy: 记录调用，验证行为
   - Mock: 完全可控，复杂验证
   - Fake: 轻量级真实实现

✅ 3. 手动Mock
   - 简单场景手动实现
   - 完全控制
   - 无额外依赖

✅ 4. Mock框架
   - testify/mock: 流行的Mock库
   - gomock: 官方推荐
   - mockery: 自动生成Mock

✅ 5. 测试模式
   - 依赖注入
   - 接口隔离
   - 单一职责

✅ 6. Mock数据
   - 使用合理的测试数据
   - 边界情况
   - 错误场景

✅ 7. 验证
   - 验证方法调用
   - 验证参数
   - 验证调用次数
   - 验证调用顺序

❌ 避免的陷阱：

1. 过度Mock
   - 只Mock外部依赖
   - 不Mock领域对象

2. Mock太复杂
   - 简化接口
   - 减少依赖

3. 脆弱的测试
   - 不过度验证实现细节
   - 关注行为而非实现

4. 不真实的Mock
   - Mock行为应该接近真实
   - 测试真实场景

5. 忽略错误路径
   - 测试失败场景
   - 测试边界条件

⚡ 性能考虑：

1. Fake vs Mock
   - Fake更快（真实逻辑）
   - Mock更灵活（完全控制）

2. 内存Mock
   - 避免真实I/O
   - 提高测试速度

3. 并行测试
   - Mock应该线程安全
   - 或使用独立实例

🎯 选择指南：

使用Stub when:
- 只需要返回固定值
- 不关心调用细节

使用Spy when:
- 需要验证调用
- 记录调用历史

使用Mock when:
- 需要复杂的行为控制
- 需要详细的验证

使用Fake when:
- 需要真实的业务逻辑
- 简化的实现即可满足
- 多个测试共享

📚 工具推荐：

1. testify/mock
   - 流行且简单
   - 良好的文档

2. gomock
   - Google官方
   - 强大的验证

3. mockery
   - 自动生成
   - 减少手动工作

4. 手动Mock
   - 简单场景首选
   - 完全可控
*/

