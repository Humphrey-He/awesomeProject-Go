package testing_practices

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// ========== 1. 基本单元测试 ==========

func TestCalculator_Add(t *testing.T) {
	calc := NewCalculator(2)
	result := calc.Add(1, 2)
	
	if result != 3 {
		t.Errorf("Add(1, 2) = %v, want 3", result)
	}
}

// ========== 2. 表驱动测试（Table-Driven Tests）==========

func TestCalculator_Divide(t *testing.T) {
	calc := NewCalculator(2)
	
	tests := []struct {
		name    string
		a, b    float64
		want    float64
		wantErr bool
	}{
		{"normal division", 10, 2, 5, false},
		{"division by zero", 10, 0, 0, true},
		{"negative numbers", -10, 2, -5, false},
		{"floating point", 7, 2, 3.5, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calc.Divide(tt.a, tt.b)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Divide() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && got != tt.want {
				t.Errorf("Divide(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// ========== 3. 子测试（Subtests）==========

func TestUserValidation(t *testing.T) {
	t.Run("valid user", func(t *testing.T) {
		user := &User{ID: 1, Name: "Alice", Email: "alice@example.com", Age: 30}
		if err := user.Validate(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	
	t.Run("missing name", func(t *testing.T) {
		user := &User{ID: 1, Email: "alice@example.com", Age: 30}
		if err := user.Validate(); err == nil {
			t.Error("expected error, got nil")
		}
	})
	
	t.Run("missing email", func(t *testing.T) {
		user := &User{ID: 1, Name: "Alice", Age: 30}
		if err := user.Validate(); err == nil {
			t.Error("expected error, got nil")
		}
	})
	
	t.Run("invalid age", func(t *testing.T) {
		user := &User{ID: 1, Name: "Alice", Email: "alice@example.com", Age: -1}
		if err := user.Validate(); err == nil {
			t.Error("expected error, got nil")
		}
	})
}

// ========== 4. 测试辅助函数（Test Helpers）==========

func TestWithHelper(t *testing.T) {
	helper := NewTestHelper(t)
	
	calc := NewCalculator(2)
	result := calc.Add(1, 2)
	
	helper.AssertEqual(result, 3.0)
	
	_, err := calc.Divide(10, 0)
	helper.AssertError(err, "division by zero")
}

// ========== 5. Setup 和 Teardown ==========

func TestWithSetupTeardown(t *testing.T) {
	// Setup
	fixture := SetupTestFixture()
	defer fixture.Teardown() // Teardown
	
	// 测试
	user, err := fixture.service.GetUser(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	if user.Name != "Alice" {
		t.Errorf("got %v, want Alice", user.Name)
	}
}

// TestMain 用于全局 setup/teardown
// func TestMain(m *testing.M) {
// 	// Global setup
// 	fmt.Println("Setting up tests...")
// 	
// 	code := m.Run() // 运行所有测试
// 	
// 	// Global teardown
// 	fmt.Println("Tearing down tests...")
// 	
// 	os.Exit(code)
// }

// ========== 6. 并行测试 ==========

func TestParallel(t *testing.T) {
	tests := []struct {
		name string
		a, b float64
		want float64
	}{
		{"case1", 1, 2, 3},
		{"case2", 5, 3, 8},
		{"case3", 10, 20, 30},
	}
	
	for _, tt := range tests {
		tt := tt // 捕获循环变量
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // 标记为并行测试
			
			calc := NewCalculator(2)
			got := calc.Add(tt.a, tt.b)
			
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// ========== 7. HTTP Handler 测试 ==========

func TestUserHandler_CreateUser(t *testing.T) {
	service := NewUserService()
	handler := NewUserHandler(service)
	
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "valid user",
			body:       `{"id":1,"name":"Alice","email":"alice@example.com","age":30}`,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid json",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing name",
			body:       `{"id":1,"email":"alice@example.com","age":30}`,
			wantStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/users", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			
			handler.CreateUser(rec, req)
			
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %v, want %v", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	service := NewUserService()
	user := &User{ID: 1, Name: "Alice", Email: "alice@example.com", Age: 30}
	service.CreateUser(user)
	
	handler := NewUserHandler(service)
	
	req := httptest.NewRequest("GET", "/users/1", nil)
	rec := httptest.NewRecorder()
	
	handler.GetUser(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %v, want %v", rec.Code, http.StatusOK)
	}
	
	var got User
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	
	if got.Name != user.Name {
		t.Errorf("got name %v, want %v", got.Name, user.Name)
	}
}

// ========== 8. 文件I/O测试 ==========

func TestFileProcessor(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir) // 清理
	
	processor := NewFileProcessor(tmpDir)
	
	t.Run("write and read", func(t *testing.T) {
		content := "Hello, World!"
		filename := "test.txt"
		
		// 写入
		if err := processor.WriteFile(filename, content); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
		
		// 读取
		got, err := processor.ReadFile(filename)
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		
		if got != content {
			t.Errorf("got %q, want %q", got, content)
		}
	})
}

// ========== 9. 基准测试（Benchmarks）==========

func BenchmarkCalculator_Add(b *testing.B) {
	calc := NewCalculator(2)
	
	b.ResetTimer() // 重置计时器，排除setup时间
	for i := 0; i < b.N; i++ {
		calc.Add(1, 2)
	}
}

func BenchmarkStringReverse(b *testing.B) {
	utils := &StringUtils{}
	str := "Hello, World! This is a longer string for benchmarking."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		utils.Reverse(str)
	}
}

// 基准测试对比
func BenchmarkUserService(b *testing.B) {
	b.Run("CreateUser", func(b *testing.B) {
		service := NewUserService()
		user := &User{ID: 1, Name: "Alice", Email: "alice@example.com", Age: 30}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			user.ID = i
			service.CreateUser(user)
		}
	})
	
	b.Run("GetUser", func(b *testing.B) {
		service := NewUserService()
		user := &User{ID: 1, Name: "Alice", Email: "alice@example.com", Age: 30}
		service.CreateUser(user)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			service.GetUser(1)
		}
	})
}

// ========== 10. 示例测试（Examples）==========

func ExampleCalculator_Add() {
	calc := NewCalculator(2)
	result := calc.Add(1, 2)
	fmt.Println(result)
	// Output: 3
}

func ExampleStringUtils_Reverse() {
	utils := &StringUtils{}
	result := utils.Reverse("Hello")
	fmt.Println(result)
	// Output: olleH
}

func ExampleUser_Validate() {
	user := &User{ID: 1, Name: "Alice", Email: "alice@example.com", Age: 30}
	err := user.Validate()
	fmt.Println(err == nil)
	// Output: true
}

// ========== 11. 错误测试模式 ==========

func TestErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		user    *User
		wantErr string
	}{
		{
			name:    "empty name",
			user:    &User{ID: 1, Email: "test@example.com", Age: 30},
			wantErr: "name is required",
		},
		{
			name:    "empty email",
			user:    &User{ID: 1, Name: "Alice", Age: 30},
			wantErr: "email is required",
		},
		{
			name:    "invalid age negative",
			user:    &User{ID: 1, Name: "Alice", Email: "test@example.com", Age: -1},
			wantErr: "invalid age",
		},
		{
			name:    "invalid age too large",
			user:    &User{ID: 1, Name: "Alice", Email: "test@example.com", Age: 200},
			wantErr: "invalid age",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// ========== 12. 测试跳过 ==========

func TestSkipExample(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	
	// 耗时的测试...
}

func TestConditionalSkip(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("skipping integration test")
	}
	
	// 集成测试...
}

// ========== 13. 清理函数 ==========

func TestWithCleanup(t *testing.T) {
	// 创建资源
	tmpFile := "test_file.txt"
	os.WriteFile(tmpFile, []byte("test"), 0644)
	
	// 注册清理函数
	t.Cleanup(func() {
		os.Remove(tmpFile)
	})
	
	// 测试逻辑...
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	
	if string(content) != "test" {
		t.Errorf("got %q, want %q", content, "test")
	}
	
	// Cleanup会在测试结束时自动调用
}

// ========== 14. HTTP Server 测试 ==========

func TestHTTPServer(t *testing.T) {
	service := NewUserService()
	handler := NewUserHandler(service)
	
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(handler.CreateUser))
	defer server.Close()
	
	// 发送请求
	user := &User{ID: 1, Name: "Alice", Email: "alice@example.com", Age: 30}
	body, _ := json.Marshal(user)
	
	resp, err := http.Post(server.URL, "application/json", strings.NewReader(string(body)))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %v, want %v", resp.StatusCode, http.StatusCreated)
	}
}

// ========== 15. 数据驱动测试（更复杂的场景）==========

func TestComplexScenarios(t *testing.T) {
	type testCase struct {
		name       string
		setup      func(*UserService)
		action     func(*UserService) error
		validate   func(*testing.T, *UserService)
		wantErr    bool
	}
	
	tests := []testCase{
		{
			name: "create and retrieve user",
			setup: func(s *UserService) {},
			action: func(s *UserService) error {
				return s.CreateUser(&User{ID: 1, Name: "Alice", Email: "alice@example.com", Age: 30})
			},
			validate: func(t *testing.T, s *UserService) {
				user, _ := s.GetUser(1)
				if user.Name != "Alice" {
					t.Errorf("got %v, want Alice", user.Name)
				}
			},
			wantErr: false,
		},
		{
			name: "update existing user",
			setup: func(s *UserService) {
				s.CreateUser(&User{ID: 1, Name: "Alice", Email: "alice@example.com", Age: 30})
			},
			action: func(s *UserService) error {
				return s.UpdateUser(&User{ID: 1, Name: "Bob", Email: "bob@example.com", Age: 25})
			},
			validate: func(t *testing.T, s *UserService) {
				user, _ := s.GetUser(1)
				if user.Name != "Bob" {
					t.Errorf("got %v, want Bob", user.Name)
				}
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewUserService()
			tt.setup(service)
			
			err := tt.action(service)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if tt.validate != nil {
				tt.validate(t, service)
			}
		})
	}
}

