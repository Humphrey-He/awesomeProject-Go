//go:build cgo
// +build cgo

package cgo_practice

import (
	"fmt"
	"testing"
	"time"
)

// ========== 1. 功能测试 ==========

func TestGaussianBlur_Consistency(t *testing.T) {
	img := GenerateTestImage(100, 100)

	resultC := img.GaussianBlurC(2.0)
	resultGo := img.GaussianBlurGo(2.0)

	// 检查结果一致性（允许小的浮点误差）
	diff := 0
	maxDiff := 0
	for i := 0; i < len(resultC.Data); i++ {
		d := int(resultC.Data[i]) - int(resultGo.Data[i])
		if d < 0 {
			d = -d
		}
		diff += d
		if d > maxDiff {
			maxDiff = d
		}
	}

	avgDiff := float64(diff) / float64(len(resultC.Data))
	t.Logf("Average difference: %.2f, Max difference: %d", avgDiff, maxDiff)

	// 允许平均误差小于2（浮点计算差异）
	if avgDiff > 2.0 {
		t.Errorf("Results too different: %.2f", avgDiff)
	}
}

func TestEdgeDetection_Consistency(t *testing.T) {
	img := GenerateTestImage(100, 100)

	resultC := img.EdgeDetectionC()
	resultGo := img.EdgeDetectionGo()

	diff := 0
	for i := 0; i < len(resultC.Data); i++ {
		d := int(resultC.Data[i]) - int(resultGo.Data[i])
		if d < 0 {
			d = -d
		}
		diff += d
	}

	avgDiff := float64(diff) / float64(len(resultC.Data))
	t.Logf("Average difference: %.2f", avgDiff)

	if avgDiff > 2.0 {
		t.Errorf("Results too different: %.2f", avgDiff)
	}
}

func TestAdjustBrightness_Consistency(t *testing.T) {
	img := GenerateTestImage(100, 100)

	resultC := img.AdjustBrightnessC(1.5)
	resultGo := img.AdjustBrightnessGo(1.5)

	diff := 0
	for i := 0; i < len(resultC.Data); i++ {
		d := int(resultC.Data[i]) - int(resultGo.Data[i])
		if d < 0 {
			d = -d
		}
		diff += d
	}

	avgDiff := float64(diff) / float64(len(resultC.Data))
	t.Logf("Average difference: %.2f", avgDiff)

	if avgDiff > 1.0 {
		t.Errorf("Results too different: %.2f", avgDiff)
	}
}

func TestReverseString_Consistency(t *testing.T) {
	tests := []string{
		"hello",
		"world",
		"Go CGO",
		"12345",
		"a",
		"",
	}

	for _, tt := range tests {
		resultC := ReverseStringC(tt)
		resultGo := ReverseStringGo(tt)

		if resultC != resultGo {
			t.Errorf("ReverseString(%q): C=%q, Go=%q", tt, resultC, resultGo)
		}
	}
}

func TestMatrixMultiply_Consistency(t *testing.T) {
	N := 10
	A := make([][]float32, N)
	B := make([][]float32, N)

	for i := 0; i < N; i++ {
		A[i] = make([]float32, N)
		B[i] = make([]float32, N)
		for j := 0; j < N; j++ {
			A[i][j] = float32(i + j)
			B[i][j] = float32(i - j)
		}
	}

	resultC := MatrixMultiplyC(A, B)
	resultGo := MatrixMultiplyGo(A, B)

	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			diff := resultC[i][j] - resultGo[i][j]
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.001 {
				t.Errorf("Matrix[%d][%d]: C=%f, Go=%f", i, j, resultC[i][j], resultGo[i][j])
			}
		}
	}
}

// ========== 2. 性能基准测试 ==========

func BenchmarkGaussianBlur_C(b *testing.B) {
	img := GenerateTestImage(512, 512)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img.GaussianBlurC(2.0)
	}
}

func BenchmarkGaussianBlur_Go(b *testing.B) {
	img := GenerateTestImage(512, 512)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img.GaussianBlurGo(2.0)
	}
}

func BenchmarkEdgeDetection_C(b *testing.B) {
	img := GenerateTestImage(512, 512)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img.EdgeDetectionC()
	}
}

func BenchmarkEdgeDetection_Go(b *testing.B) {
	img := GenerateTestImage(512, 512)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img.EdgeDetectionGo()
	}
}

func BenchmarkAdjustBrightness_C(b *testing.B) {
	img := GenerateTestImage(1024, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img.AdjustBrightnessC(1.5)
	}
}

func BenchmarkAdjustBrightness_Go(b *testing.B) {
	img := GenerateTestImage(1024, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img.AdjustBrightnessGo(1.5)
	}
}

func BenchmarkHistogramEqualization_C(b *testing.B) {
	img := GenerateTestImage(512, 512)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img.HistogramEqualizationC()
	}
}

func BenchmarkHistogramEqualization_Go(b *testing.B) {
	img := GenerateTestImage(512, 512)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img.HistogramEqualizationGo()
	}
}

func BenchmarkReverseString_C(b *testing.B) {
	s := "Hello, World! This is a test string for benchmarking."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ReverseStringC(s)
	}
}

func BenchmarkReverseString_Go(b *testing.B) {
	s := "Hello, World! This is a test string for benchmarking."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ReverseStringGo(s)
	}
}

func BenchmarkMatrixMultiply_C_Small(b *testing.B) {
	N := 32
	A := make([][]float32, N)
	B := make([][]float32, N)

	for i := 0; i < N; i++ {
		A[i] = make([]float32, N)
		B[i] = make([]float32, N)
		for j := 0; j < N; j++ {
			A[i][j] = float32(i + j)
			B[i][j] = float32(i - j)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MatrixMultiplyC(A, B)
	}
}

func BenchmarkMatrixMultiply_Go_Small(b *testing.B) {
	N := 32
	A := make([][]float32, N)
	B := make([][]float32, N)

	for i := 0; i < N; i++ {
		A[i] = make([]float32, N)
		B[i] = make([]float32, N)
		for j := 0; j < N; j++ {
			A[i][j] = float32(i + j)
			B[i][j] = float32(i - j)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MatrixMultiplyGo(A, B)
	}
}

// ========== 3. 性能对比测试 ==========

func TestPerformanceComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance comparison in short mode")
	}

	fmt.Println("\n========== CGO Performance Comparison ==========")

	// 图像大小测试
	sizes := []int{256, 512, 1024}

	for _, size := range sizes {
		fmt.Printf("\n--- Image Size: %dx%d ---\n", size, size)
		img := GenerateTestImage(size, size)

		// 高斯模糊
		startC := time.Now()
		img.GaussianBlurC(2.0)
		durationC := time.Since(startC)

		startGo := time.Now()
		img.GaussianBlurGo(2.0)
		durationGo := time.Since(startGo)

		speedup := float64(durationGo) / float64(durationC)
		fmt.Printf("Gaussian Blur:\n")
		fmt.Printf("  C:  %v\n", durationC)
		fmt.Printf("  Go: %v\n", durationGo)
		fmt.Printf("  Speedup: %.2fx\n", speedup)

		// 边缘检测
		startC = time.Now()
		img.EdgeDetectionC()
		durationC = time.Since(startC)

		startGo = time.Now()
		img.EdgeDetectionGo()
		durationGo = time.Since(startGo)

		speedup = float64(durationGo) / float64(durationC)
		fmt.Printf("Edge Detection:\n")
		fmt.Printf("  C:  %v\n", durationC)
		fmt.Printf("  Go: %v\n", durationGo)
		fmt.Printf("  Speedup: %.2fx\n", speedup)

		// 亮度调整
		startC = time.Now()
		img.AdjustBrightnessC(1.5)
		durationC = time.Since(startC)

		startGo = time.Now()
		img.AdjustBrightnessGo(1.5)
		durationGo = time.Since(startGo)

		speedup = float64(durationGo) / float64(durationC)
		fmt.Printf("Adjust Brightness:\n")
		fmt.Printf("  C:  %v\n", durationC)
		fmt.Printf("  Go: %v\n", durationGo)
		fmt.Printf("  Speedup: %.2fx\n", speedup)
	}

	// 矩阵乘法
	fmt.Println("\n--- Matrix Multiplication ---")
	matrixSizes := []int{32, 64, 128}

	for _, N := range matrixSizes {
		A := make([][]float32, N)
		B := make([][]float32, N)

		for i := 0; i < N; i++ {
			A[i] = make([]float32, N)
			B[i] = make([]float32, N)
			for j := 0; j < N; j++ {
				A[i][j] = float32(i + j)
				B[i][j] = float32(i - j)
			}
		}

		startC := time.Now()
		MatrixMultiplyC(A, B)
		durationC := time.Since(startC)

		startGo := time.Now()
		MatrixMultiplyGo(A, B)
		durationGo := time.Since(startGo)

		speedup := float64(durationGo) / float64(durationC)
		fmt.Printf("Matrix %dx%d:\n", N, N)
		fmt.Printf("  C:  %v\n", durationC)
		fmt.Printf("  Go: %v\n", durationGo)
		fmt.Printf("  Speedup: %.2fx\n", speedup)
	}

	fmt.Println("\n================================================")
}

// ========== 4. 示例测试 ==========

func ExampleGrayImage_GaussianBlurC() {
	img := GenerateTestImage(10, 10)
	fmt.Println("Original image created:", img.Width, "x", img.Height)

	blurred := img.GaussianBlurC(1.0)
	fmt.Println("Blurred image:", blurred.Width, "x", blurred.Height)

	// Output:
	// Original image created: 10 x 10
	// Blurred image: 10 x 10
}

func ExampleReverseStringC() {
	original := "Hello"
	reversed := ReverseStringC(original)
	fmt.Println(reversed)

	// Output:
	// olleH
}

func ExampleMatrixMultiplyC() {
	A := [][]float32{
		{1, 2},
		{3, 4},
	}
	B := [][]float32{
		{5, 6},
		{7, 8},
	}

	C := MatrixMultiplyC(A, B)
	fmt.Printf("Result[0][0]: %.0f\n", C[0][0])
	fmt.Printf("Result[1][1]: %.0f\n", C[1][1])

	// Output:
	// Result[0][0]: 19
	// Result[1][1]: 50
}

// ========== 5. 内存泄漏测试 ==========

func TestMemoryLeak(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory leak test in short mode")
	}

	// 这个测试重复调用CGO函数，确保没有内存泄漏
	// 实际使用时可以配合 -memprofile 检查

	img := GenerateTestImage(512, 512)

	for i := 0; i < 1000; i++ {
		img.GaussianBlurC(2.0)
		img.EdgeDetectionC()
		img.AdjustBrightnessC(1.5)

		// 字符串操作
		_ = ReverseStringC("test string for memory leak detection")

		if i%100 == 0 {
			t.Logf("Iteration %d/1000", i)
		}
	}

	t.Log("Memory leak test completed")
}
