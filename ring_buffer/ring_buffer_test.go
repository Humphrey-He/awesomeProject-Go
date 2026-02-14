package ring_buffer

import (
	"sync"
	"testing"
)

// ========== 基础功能测试 ==========

func TestNewRingBuffer(t *testing.T) {
	rb := NewRingBuffer(5)
	
	if rb.Cap() != 5 {
		t.Errorf("Expected capacity 5, got %d", rb.Cap())
	}
	
	if rb.Len() != 0 {
		t.Errorf("Expected length 0, got %d", rb.Len())
	}
	
	if !rb.IsEmpty() {
		t.Error("Expected buffer to be empty")
	}
}

func TestRingBuffer_WriteRead(t *testing.T) {
	rb := NewRingBuffer(3)
	
	// 写入元素
	if err := rb.Write(1); err != nil {
		t.Errorf("Write failed: %v", err)
	}
	if err := rb.Write(2); err != nil {
		t.Errorf("Write failed: %v", err)
	}
	
	if rb.Len() != 2 {
		t.Errorf("Expected length 2, got %d", rb.Len())
	}
	
	// 读取元素
	val, err := rb.Read()
	if err != nil {
		t.Errorf("Read failed: %v", err)
	}
	if val != 1 {
		t.Errorf("Expected 1, got %v", val)
	}
	
	val, err = rb.Read()
	if err != nil {
		t.Errorf("Read failed: %v", err)
	}
	if val != 2 {
		t.Errorf("Expected 2, got %v", val)
	}
	
	if !rb.IsEmpty() {
		t.Error("Expected buffer to be empty")
	}
}

func TestRingBuffer_Full(t *testing.T) {
	rb := NewRingBuffer(3)
	
	rb.Write(1)
	rb.Write(2)
	rb.Write(3)
	
	if !rb.IsFull() {
		t.Error("Expected buffer to be full")
	}
	
	// 尝试再写入应该失败
	err := rb.Write(4)
	if err != ErrBufferFull {
		t.Errorf("Expected ErrBufferFull, got %v", err)
	}
}

func TestRingBuffer_Empty(t *testing.T) {
	rb := NewRingBuffer(3)
	
	// 从空缓冲区读取应该失败
	_, err := rb.Read()
	if err != ErrBufferEmpty {
		t.Errorf("Expected ErrBufferEmpty, got %v", err)
	}
}

func TestRingBuffer_Wrap(t *testing.T) {
	rb := NewRingBuffer(3)
	
	// 填满缓冲区
	rb.Write(1)
	rb.Write(2)
	rb.Write(3)
	
	// 读取一些元素
	rb.Read()
	rb.Read()
	
	// 再写入，测试环绕
	rb.Write(4)
	rb.Write(5)
	
	// 验证顺序
	val, _ := rb.Read()
	if val != 3 {
		t.Errorf("Expected 3, got %v", val)
	}
	
	val, _ = rb.Read()
	if val != 4 {
		t.Errorf("Expected 4, got %v", val)
	}
	
	val, _ = rb.Read()
	if val != 5 {
		t.Errorf("Expected 5, got %v", val)
	}
}

func TestRingBuffer_WriteOverwrite(t *testing.T) {
	rb := NewRingBuffer(3)
	
	// 填满缓冲区
	rb.WriteOverwrite(1)
	rb.WriteOverwrite(2)
	rb.WriteOverwrite(3)
	
	// 继续写入，应该覆盖最旧的元素
	rb.WriteOverwrite(4)
	rb.WriteOverwrite(5)
	
	// 现在缓冲区应该包含 [3, 4, 5]
	val, _ := rb.Read()
	if val != 3 {
		t.Errorf("Expected 3, got %v", val)
	}
	
	val, _ = rb.Read()
	if val != 4 {
		t.Errorf("Expected 4, got %v", val)
	}
	
	val, _ = rb.Read()
	if val != 5 {
		t.Errorf("Expected 5, got %v", val)
	}
}

func TestRingBuffer_Peek(t *testing.T) {
	rb := NewRingBuffer(3)
	
	rb.Write(1)
	rb.Write(2)
	
	// Peek不应该移除元素
	val, err := rb.Peek()
	if err != nil {
		t.Errorf("Peek failed: %v", err)
	}
	if val != 1 {
		t.Errorf("Expected 1, got %v", val)
	}
	
	// 长度不应该改变
	if rb.Len() != 2 {
		t.Errorf("Expected length 2, got %d", rb.Len())
	}
	
	// 再次Peek应该返回相同的值
	val, _ = rb.Peek()
	if val != 1 {
		t.Errorf("Expected 1, got %v", val)
	}
}

func TestRingBuffer_Clear(t *testing.T) {
	rb := NewRingBuffer(3)
	
	rb.Write(1)
	rb.Write(2)
	rb.Write(3)
	
	rb.Clear()
	
	if !rb.IsEmpty() {
		t.Error("Expected buffer to be empty after clear")
	}
	
	if rb.Len() != 0 {
		t.Errorf("Expected length 0, got %d", rb.Len())
	}
}

func TestRingBuffer_ToSlice(t *testing.T) {
	rb := NewRingBuffer(5)
	
	rb.Write(1)
	rb.Write(2)
	rb.Write(3)
	
	slice := rb.ToSlice()
	
	if len(slice) != 3 {
		t.Errorf("Expected slice length 3, got %d", len(slice))
	}
	
	expected := []interface{}{1, 2, 3}
	for i, exp := range expected {
		if slice[i] != exp {
			t.Errorf("Expected %v at index %d, got %v", exp, i, slice[i])
		}
	}
}

func TestRingBuffer_ToSliceAfterWrap(t *testing.T) {
	rb := NewRingBuffer(3)
	
	rb.Write(1)
	rb.Write(2)
	rb.Write(3)
	rb.Read()
	rb.Read()
	rb.Write(4)
	rb.Write(5)
	
	// 现在缓冲区包含 [3, 4, 5]
	slice := rb.ToSlice()
	
	expected := []interface{}{3, 4, 5}
	for i, exp := range expected {
		if slice[i] != exp {
			t.Errorf("Expected %v at index %d, got %v", exp, i, slice[i])
		}
	}
}

// ========== 并发测试 ==========

func TestRingBuffer_ConcurrentWriteRead(t *testing.T) {
	rb := NewRingBuffer(100)
	
	var wg sync.WaitGroup
	writeCount := 1000
	
	// 启动写入goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < writeCount; i++ {
			for {
				err := rb.Write(i)
				if err == nil {
					break
				}
				// 缓冲区满，稍后重试
			}
		}
	}()
	
	// 启动读取goroutine
	wg.Add(1)
	readCount := 0
	go func() {
		defer wg.Done()
		for readCount < writeCount {
			_, err := rb.Read()
			if err == nil {
				readCount++
			}
			// 缓冲区空，稍后重试
		}
	}()
	
	wg.Wait()
	
	if readCount != writeCount {
		t.Errorf("Expected to read %d items, got %d", writeCount, readCount)
	}
}

func TestRingBuffer_MultipleProducersConsumers(t *testing.T) {
	rb := NewRingBuffer(50)
	
	var wg sync.WaitGroup
	producers := 5
	consumers := 5
	itemsPerProducer := 100
	
	// 启动生产者
	for i := 0; i < producers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < itemsPerProducer; j++ {
				for rb.Write(j) != nil {
					// 重试
				}
			}
		}()
	}
	
	// 启动消费者
	totalItems := producers * itemsPerProducer
	itemsRead := 0
	var mu sync.Mutex
	
	for i := 0; i < consumers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				mu.Lock()
				if itemsRead >= totalItems {
					mu.Unlock()
					return
				}
				mu.Unlock()
				
				_, err := rb.Read()
				if err == nil {
					mu.Lock()
					itemsRead++
					mu.Unlock()
				}
			}
		}()
	}
	
	wg.Wait()
	
	if itemsRead != totalItems {
		t.Errorf("Expected to read %d items, got %d", totalItems, itemsRead)
	}
}

// ========== 泛型版本测试 ==========

func TestRingBufferGeneric_Basic(t *testing.T) {
	rb := NewRingBufferGeneric[int](3)
	
	rb.Write(1)
	rb.Write(2)
	rb.Write(3)
	
	val, _ := rb.Read()
	if val != 1 {
		t.Errorf("Expected 1, got %d", val)
	}
	
	val, _ = rb.Read()
	if val != 2 {
		t.Errorf("Expected 2, got %d", val)
	}
}

func TestRingBufferGeneric_String(t *testing.T) {
	rb := NewRingBufferGeneric[string](3)
	
	rb.Write("hello")
	rb.Write("world")
	
	val, _ := rb.Read()
	if val != "hello" {
		t.Errorf("Expected 'hello', got '%s'", val)
	}
	
	val, _ = rb.Read()
	if val != "world" {
		t.Errorf("Expected 'world', got '%s'", val)
	}
}

func TestRingBufferGeneric_Struct(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}
	
	rb := NewRingBufferGeneric[Person](3)
	
	rb.Write(Person{"Alice", 30})
	rb.Write(Person{"Bob", 25})
	
	val, _ := rb.Read()
	if val.Name != "Alice" || val.Age != 30 {
		t.Errorf("Expected Alice/30, got %v/%v", val.Name, val.Age)
	}
}

// ========== 性能基准测试 ==========

func BenchmarkRingBuffer_Write(b *testing.B) {
	rb := NewRingBuffer(1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Write(i)
		if rb.IsFull() {
			rb.Read()
		}
	}
}

func BenchmarkRingBuffer_Read(b *testing.B) {
	rb := NewRingBuffer(1000)
	for i := 0; i < 1000; i++ {
		rb.Write(i)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Read()
		if rb.IsEmpty() {
			rb.Write(i)
		}
	}
}

func BenchmarkRingBuffer_WriteOverwrite(b *testing.B) {
	rb := NewRingBuffer(1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.WriteOverwrite(i)
	}
}

func BenchmarkRingBufferGeneric_Write(b *testing.B) {
	rb := NewRingBufferGeneric[int](1000)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Write(i)
		if rb.IsFull() {
			rb.Read()
		}
	}
}

func BenchmarkRingBuffer_ConcurrentAccess(b *testing.B) {
	rb := NewRingBuffer(100)
	
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%2 == 0 {
				rb.Write(i)
			} else {
				rb.Read()
			}
			i++
		}
	})
}

