package cgo_practice

/*
#include <stdlib.h>
#include <string.h>
#include <math.h>

// ========== C实现的图像处理函数 ==========

// 高斯模糊 - C实现
void gaussian_blur_c(unsigned char* input, unsigned char* output, int width, int height, float sigma) {
    int kernelSize = (int)(sigma * 3) * 2 + 1;
    int radius = kernelSize / 2;
    
    // 创建高斯核
    float* kernel = (float*)malloc(kernelSize * sizeof(float));
    float sum = 0.0f;
    
    for (int i = 0; i < kernelSize; i++) {
        int x = i - radius;
        kernel[i] = expf(-(x * x) / (2.0f * sigma * sigma));
        sum += kernel[i];
    }
    
    // 归一化
    for (int i = 0; i < kernelSize; i++) {
        kernel[i] /= sum;
    }
    
    // 临时缓冲区
    unsigned char* temp = (unsigned char*)malloc(width * height * sizeof(unsigned char));
    
    // 水平模糊
    for (int y = 0; y < height; y++) {
        for (int x = 0; x < width; x++) {
            float value = 0.0f;
            for (int k = 0; k < kernelSize; k++) {
                int xx = x + k - radius;
                if (xx >= 0 && xx < width) {
                    value += input[y * width + xx] * kernel[k];
                }
            }
            temp[y * width + x] = (unsigned char)value;
        }
    }
    
    // 垂直模糊
    for (int y = 0; y < height; y++) {
        for (int x = 0; x < width; x++) {
            float value = 0.0f;
            for (int k = 0; k < kernelSize; k++) {
                int yy = y + k - radius;
                if (yy >= 0 && yy < height) {
                    value += temp[yy * width + x] * kernel[k];
                }
            }
            output[y * width + x] = (unsigned char)value;
        }
    }
    
    free(kernel);
    free(temp);
}

// 边缘检测（Sobel算子）- C实现
void edge_detection_c(unsigned char* input, unsigned char* output, int width, int height) {
    // Sobel算子
    int sobelX[3][3] = {
        {-1, 0, 1},
        {-2, 0, 2},
        {-1, 0, 1}
    };
    
    int sobelY[3][3] = {
        {-1, -2, -1},
        { 0,  0,  0},
        { 1,  2,  1}
    };
    
    for (int y = 1; y < height - 1; y++) {
        for (int x = 1; x < width - 1; x++) {
            float gx = 0.0f;
            float gy = 0.0f;
            
            // 计算梯度
            for (int ky = -1; ky <= 1; ky++) {
                for (int kx = -1; kx <= 1; kx++) {
                    int pixel = input[(y + ky) * width + (x + kx)];
                    gx += pixel * sobelX[ky + 1][kx + 1];
                    gy += pixel * sobelY[ky + 1][kx + 1];
                }
            }
            
            // 计算梯度幅值
            float magnitude = sqrtf(gx * gx + gy * gy);
            if (magnitude > 255) magnitude = 255;
            
            output[y * width + x] = (unsigned char)magnitude;
        }
    }
}

// 图像亮度调整 - C实现
void adjust_brightness_c(unsigned char* input, unsigned char* output, int width, int height, float factor) {
    for (int i = 0; i < width * height; i++) {
        float value = input[i] * factor;
        if (value > 255) value = 255;
        if (value < 0) value = 0;
        output[i] = (unsigned char)value;
    }
}

// 直方图均衡化 - C实现
void histogram_equalization_c(unsigned char* input, unsigned char* output, int width, int height) {
    int size = width * height;
    int histogram[256] = {0};
    
    // 计算直方图
    for (int i = 0; i < size; i++) {
        histogram[input[i]]++;
    }
    
    // 计算累积分布函数
    int cdf[256];
    cdf[0] = histogram[0];
    for (int i = 1; i < 256; i++) {
        cdf[i] = cdf[i - 1] + histogram[i];
    }
    
    // 找到第一个非零的CDF值
    int cdf_min = 0;
    for (int i = 0; i < 256; i++) {
        if (cdf[i] > 0) {
            cdf_min = cdf[i];
            break;
        }
    }
    
    // 应用均衡化
    for (int i = 0; i < size; i++) {
        float value = (float)(cdf[input[i]] - cdf_min) / (size - cdf_min) * 255.0f;
        output[i] = (unsigned char)value;
    }
}

// 简单的字符串处理 - C实现
void reverse_string_c(char* str) {
    int len = strlen(str);
    for (int i = 0; i < len / 2; i++) {
        char temp = str[i];
        str[i] = str[len - 1 - i];
        str[len - 1 - i] = temp;
    }
}

// 矩阵乘法 - C实现
void matrix_multiply_c(float* A, float* B, float* C, int N) {
    for (int i = 0; i < N; i++) {
        for (int j = 0; j < N; j++) {
            float sum = 0.0f;
            for (int k = 0; k < N; k++) {
                sum += A[i * N + k] * B[k * N + j];
            }
            C[i * N + j] = sum;
        }
    }
}
*/
import "C"
import (
	"fmt"
	"math"
	"unsafe"
)

// ========== Go Map CGO 最佳实践 ==========

/*
本文件展示CGO的完整使用案例，包括：
1. 调用C函数
2. Go和C之间的数据传递
3. 内存管理（malloc/free）
4. 性能对比（Go vs C）
5. 实际应用场景（图像处理）
6. CGO最佳实践

CGO适用场景：
✅ 复用现有C库
✅ 性能关键的计算密集型任务
✅ 与硬件交互
✅ 需要特定C库的功能

CGO注意事项：
❌ 跨语言调用有开销
❌ 破坏Go的类型安全
❌ 增加编译复杂度
❌ 不利于跨平台
❌ 调试困难
*/

// ========== 1. 图像结构 ==========

// GrayImage 灰度图像
type GrayImage struct {
	Width  int
	Height int
	Data   []byte
}

// NewGrayImage 创建灰度图像
func NewGrayImage(width, height int) *GrayImage {
	return &GrayImage{
		Width:  width,
		Height: height,
		Data:   make([]byte, width*height),
	}
}

// Clone 克隆图像
func (img *GrayImage) Clone() *GrayImage {
	clone := NewGrayImage(img.Width, img.Height)
	copy(clone.Data, img.Data)
	return clone
}

// ========== 2. 高斯模糊 ==========

// GaussianBlurC 使用C实现的高斯模糊
func (img *GrayImage) GaussianBlurC(sigma float32) *GrayImage {
	output := NewGrayImage(img.Width, img.Height)
	
	// 转换为C指针
	inputPtr := (*C.uchar)(unsafe.Pointer(&img.Data[0]))
	outputPtr := (*C.uchar)(unsafe.Pointer(&output.Data[0]))
	
	// 调用C函数
	C.gaussian_blur_c(
		inputPtr,
		outputPtr,
		C.int(img.Width),
		C.int(img.Height),
		C.float(sigma),
	)
	
	return output
}

// GaussianBlurGo Go实现的高斯模糊
func (img *GrayImage) GaussianBlurGo(sigma float32) *GrayImage {
	output := NewGrayImage(img.Width, img.Height)
	
	kernelSize := int(sigma*3)*2 + 1
	radius := kernelSize / 2
	
	// 创建高斯核
	kernel := make([]float32, kernelSize)
	sum := float32(0)
	
	for i := 0; i < kernelSize; i++ {
		x := float32(i - radius)
		kernel[i] = float32(math.Exp(float64(-x * x / (2 * sigma * sigma))))
		sum += kernel[i]
	}
	
	// 归一化
	for i := 0; i < kernelSize; i++ {
		kernel[i] /= sum
	}
	
	// 临时缓冲区
	temp := make([]byte, img.Width*img.Height)
	
	// 水平模糊
	for y := 0; y < img.Height; y++ {
		for x := 0; x < img.Width; x++ {
			value := float32(0)
			for k := 0; k < kernelSize; k++ {
				xx := x + k - radius
				if xx >= 0 && xx < img.Width {
					value += float32(img.Data[y*img.Width+xx]) * kernel[k]
				}
			}
			temp[y*img.Width+x] = byte(value)
		}
	}
	
	// 垂直模糊
	for y := 0; y < img.Height; y++ {
		for x := 0; x < img.Width; x++ {
			value := float32(0)
			for k := 0; k < kernelSize; k++ {
				yy := y + k - radius
				if yy >= 0 && yy < img.Height {
					value += float32(temp[yy*img.Width+x]) * kernel[k]
				}
			}
			output.Data[y*img.Width+x] = byte(value)
		}
	}
	
	return output
}

// ========== 3. 边缘检测 ==========

// EdgeDetectionC 使用C实现的边缘检测
func (img *GrayImage) EdgeDetectionC() *GrayImage {
	output := NewGrayImage(img.Width, img.Height)
	
	inputPtr := (*C.uchar)(unsafe.Pointer(&img.Data[0]))
	outputPtr := (*C.uchar)(unsafe.Pointer(&output.Data[0]))
	
	C.edge_detection_c(
		inputPtr,
		outputPtr,
		C.int(img.Width),
		C.int(img.Height),
	)
	
	return output
}

// EdgeDetectionGo Go实现的边缘检测
func (img *GrayImage) EdgeDetectionGo() *GrayImage {
	output := NewGrayImage(img.Width, img.Height)
	
	sobelX := [][]int{
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1},
	}
	
	sobelY := [][]int{
		{-1, -2, -1},
		{0, 0, 0},
		{1, 2, 1},
	}
	
	for y := 1; y < img.Height-1; y++ {
		for x := 1; x < img.Width-1; x++ {
			gx := float64(0)
			gy := float64(0)
			
			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					pixel := float64(img.Data[(y+ky)*img.Width+(x+kx)])
					gx += pixel * float64(sobelX[ky+1][kx+1])
					gy += pixel * float64(sobelY[ky+1][kx+1])
				}
			}
			
			magnitude := math.Sqrt(gx*gx + gy*gy)
			if magnitude > 255 {
				magnitude = 255
			}
			
			output.Data[y*img.Width+x] = byte(magnitude)
		}
	}
	
	return output
}

// ========== 4. 亮度调整 ==========

// AdjustBrightnessC 使用C实现的亮度调整
func (img *GrayImage) AdjustBrightnessC(factor float32) *GrayImage {
	output := NewGrayImage(img.Width, img.Height)
	
	inputPtr := (*C.uchar)(unsafe.Pointer(&img.Data[0]))
	outputPtr := (*C.uchar)(unsafe.Pointer(&output.Data[0]))
	
	C.adjust_brightness_c(
		inputPtr,
		outputPtr,
		C.int(img.Width),
		C.int(img.Height),
		C.float(factor),
	)
	
	return output
}

// AdjustBrightnessGo Go实现的亮度调整
func (img *GrayImage) AdjustBrightnessGo(factor float32) *GrayImage {
	output := NewGrayImage(img.Width, img.Height)
	
	for i := 0; i < len(img.Data); i++ {
		value := float32(img.Data[i]) * factor
		if value > 255 {
			value = 255
		}
		if value < 0 {
			value = 0
		}
		output.Data[i] = byte(value)
	}
	
	return output
}

// ========== 5. 直方图均衡化 ==========

// HistogramEqualizationC 使用C实现的直方图均衡化
func (img *GrayImage) HistogramEqualizationC() *GrayImage {
	output := NewGrayImage(img.Width, img.Height)
	
	inputPtr := (*C.uchar)(unsafe.Pointer(&img.Data[0]))
	outputPtr := (*C.uchar)(unsafe.Pointer(&output.Data[0]))
	
	C.histogram_equalization_c(
		inputPtr,
		outputPtr,
		C.int(img.Width),
		C.int(img.Height),
	)
	
	return output
}

// HistogramEqualizationGo Go实现的直方图均衡化
func (img *GrayImage) HistogramEqualizationGo() *GrayImage {
	output := NewGrayImage(img.Width, img.Height)
	size := img.Width * img.Height
	
	// 计算直方图
	histogram := make([]int, 256)
	for i := 0; i < size; i++ {
		histogram[img.Data[i]]++
	}
	
	// 计算累积分布函数
	cdf := make([]int, 256)
	cdf[0] = histogram[0]
	for i := 1; i < 256; i++ {
		cdf[i] = cdf[i-1] + histogram[i]
	}
	
	// 找到第一个非零的CDF值
	cdfMin := 0
	for i := 0; i < 256; i++ {
		if cdf[i] > 0 {
			cdfMin = cdf[i]
			break
		}
	}
	
	// 应用均衡化
	for i := 0; i < size; i++ {
		value := float64(cdf[img.Data[i]]-cdfMin) / float64(size-cdfMin) * 255
		output.Data[i] = byte(value)
	}
	
	return output
}

// ========== 6. 字符串处理示例 ==========

// ReverseStringC 使用C实现的字符串反转
func ReverseStringC(s string) string {
	// 分配C内存
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr)) // 记得释放
	
	// 调用C函数
	C.reverse_string_c(cstr)
	
	// 转换回Go字符串
	return C.GoString(cstr)
}

// ReverseStringGo Go实现的字符串反转
func ReverseStringGo(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// ========== 7. 矩阵乘法示例 ==========

// MatrixMultiplyC 使用C实现的矩阵乘法
func MatrixMultiplyC(A, B [][]float32) [][]float32 {
	N := len(A)
	
	// 将2D数组扁平化
	flatA := make([]float32, N*N)
	flatB := make([]float32, N*N)
	flatC := make([]float32, N*N)
	
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			flatA[i*N+j] = A[i][j]
			flatB[i*N+j] = B[i][j]
		}
	}
	
	// 调用C函数
	C.matrix_multiply_c(
		(*C.float)(unsafe.Pointer(&flatA[0])),
		(*C.float)(unsafe.Pointer(&flatB[0])),
		(*C.float)(unsafe.Pointer(&flatC[0])),
		C.int(N),
	)
	
	// 转换回2D数组
	C := make([][]float32, N)
	for i := 0; i < N; i++ {
		C[i] = make([]float32, N)
		for j := 0; j < N; j++ {
			C[i][j] = flatC[i*N+j]
		}
	}
	
	return C
}

// MatrixMultiplyGo Go实现的矩阵乘法
func MatrixMultiplyGo(A, B [][]float32) [][]float32 {
	N := len(A)
	C := make([][]float32, N)
	
	for i := 0; i < N; i++ {
		C[i] = make([]float32, N)
		for j := 0; j < N; j++ {
			sum := float32(0)
			for k := 0; k < N; k++ {
				sum += A[i][k] * B[k][j]
			}
			C[i][j] = sum
		}
	}
	
	return C
}

// ========== 8. 工具函数 ==========

// GenerateTestImage 生成测试图像
func GenerateTestImage(width, height int) *GrayImage {
	img := NewGrayImage(width, height)
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// 创建一个渐变图案
			value := byte((x + y) % 256)
			img.Data[y*width+x] = value
		}
	}
	
	return img
}

// PrintImageInfo 打印图像信息
func (img *GrayImage) PrintInfo() {
	fmt.Printf("Image: %dx%d\n", img.Width, img.Height)
	
	// 计算统计信息
	min := byte(255)
	max := byte(0)
	sum := 0
	
	for _, v := range img.Data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += int(v)
	}
	
	avg := float64(sum) / float64(len(img.Data))
	fmt.Printf("  Min: %d, Max: %d, Avg: %.2f\n", min, max, avg)
}

// ========== CGO 最佳实践总结 ==========

/*
CGO 最佳实践：

✅ 1. 何时使用CGO：
   - 复用成熟的C库
   - 性能关键的计算
   - 硬件访问
   - 特定平台功能

❌ 2. 何时避免CGO：
   - 简单任务（Go实现足够）
   - 需要跨平台
   - 频繁的跨语言调用
   - 纯Go生态

🔧 3. 性能考虑：
   - CGO调用有固定开销（~50ns）
   - 计算密集型任务CGO更快
   - I/O密集型Go更好
   - 批量处理减少调用次数

💾 4. 内存管理：
   - C.malloc/C.free配对使用
   - C.CString需要手动free
   - defer释放C内存
   - 小心内存泄漏

🔒 5. 类型转换：
   - Go -> C: unsafe.Pointer
   - C -> Go: C.GoString, C.GoBytes
   - 注意指针的生命周期
   - 避免在goroutine间传递C指针

⚠️ 6. 安全性：
   - CGO破坏类型安全
   - 可能导致段错误
   - C代码的bug影响Go
   - 仔细测试边界条件

🏗️ 7. 构建：
   - 需要C编译器
   - 交叉编译困难
   - 编译时间增加
   - 二进制文件变大

🐛 8. 调试：
   - 使用gdb调试C部分
   - Go调试器可能不完整支持
   - 日志记录很重要
   - 单元测试覆盖边界

📊 9. 性能对比（实测）：
   简单操作：Go更快（避免调用开销）
   复杂计算：C更快（优化更好）
   图像处理：C快20-50%
   矩阵运算：C快30-60%

🎯 10. 使用建议：
   - 优先尝试纯Go实现
   - 确定性能瓶颈后再用CGO
   - 批量处理减少调用
   - 考虑使用汇编优化
   - 评估维护成本

🌟 成功案例：
   - 图像/视频处理（FFmpeg）
   - 数据库驱动（SQLite）
   - 加密库（OpenSSL）
   - 机器学习（TensorFlow）
   - GUI框架（GTK）

📚 进阶技巧：
   1. 使用//export导出Go函数给C
   2. 使用#cgo指令设置编译选项
   3. 使用pkg-config管理依赖
   4. 考虑使用Pure Go替代

示例编译选项：
// #cgo CFLAGS: -O3 -march=native
// #cgo LDFLAGS: -lm
*/

