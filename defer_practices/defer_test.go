package defer_practices

import (
	"testing"
)

// ========== Defer 行为测试 ==========

func TestDeferReturnValue(t *testing.T) {
	// Test case 1: 不影响未命名返回值
	result1 := DeferReturnValue1()
	if result1 != 0 {
		t.Errorf("Expected 0, got %d", result1)
	}

	// Test case 2: 影响命名返回值
	result2 := DeferReturnValue2()
	if result2 != 1 {
		t.Errorf("Expected 1, got %d", result2)
	}
}

func TestSafeFunction(t *testing.T) {
	// 测试不会panic
	err := SafeFunction()
	if err != nil {
		t.Logf("Error (expected): %v", err)
	}
}

func TestReadFileErrorHandling(t *testing.T) {
	// 测试不存在的文件
	_, err := ReadFile("nonexistent.txt")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

// ========== 性能基准测试 ==========

func BenchmarkNoDefer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NoDefer()
	}
}

func BenchmarkWithDefer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = WithDefer()
	}
}

func BenchmarkMultipleDefer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = MultipleDefer()
	}
}

// 测试defer开销
func BenchmarkDeferOverhead(b *testing.B) {
	b.Run("NoDefer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			x := 1
			x++
			_ = x
		}
	})

	b.Run("WithSimpleDefer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			x := 1
			defer func() {}()
			x++
			_ = x
		}
	})

	b.Run("WithComplexDefer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			x := 1
			defer func() {
				x++
			}()
			x++
			_ = x
		}
	})
}

// ========== 示例测试 ==========

func ExampleDeferExecutionOrder() {
	DeferExecutionOrder()
	// Output:
	// Start
	// End
	// Defer 3
	// Defer 2
	// Defer 1
}

func ExampleDeferArgumentEvaluation() {
	DeferArgumentEvaluation()
	// Output:
	// Current x: 2
	// Deferred x: 1
}

func ExampleDeferWithClosure() {
	DeferWithClosure()
	// Output:
	// Current x: 2
	// Deferred x (closure): 2
}
