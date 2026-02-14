package ring_buffer

import (
	"errors"
	"sync"
)

var (
	ErrBufferFull  = errors.New("ring buffer is full")
	ErrBufferEmpty = errors.New("ring buffer is empty")
)

// RingBuffer 环形缓冲区（线程安全版本）
type RingBuffer struct {
	buffer []interface{}
	size   int  // 缓冲区容量
	head   int  // 读指针
	tail   int  // 写指针
	count  int  // 当前元素数量
	mu     sync.Mutex
}

// NewRingBuffer 创建一个新的环形缓冲区
func NewRingBuffer(size int) *RingBuffer {
	if size <= 0 {
		size = 10
	}
	return &RingBuffer{
		buffer: make([]interface{}, size),
		size:   size,
		head:   0,
		tail:   0,
		count:  0,
	}
}

// Write 写入一个元素到缓冲区
func (rb *RingBuffer) Write(item interface{}) error {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	
	if rb.count == rb.size {
		return ErrBufferFull
	}
	
	rb.buffer[rb.tail] = item
	rb.tail = (rb.tail + 1) % rb.size
	rb.count++
	
	return nil
}

// Read 从缓冲区读取一个元素
func (rb *RingBuffer) Read() (interface{}, error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	
	if rb.count == 0 {
		return nil, ErrBufferEmpty
	}
	
	item := rb.buffer[rb.head]
	rb.buffer[rb.head] = nil // 防止内存泄漏
	rb.head = (rb.head + 1) % rb.size
	rb.count--
	
	return item, nil
}

// WriteOverwrite 写入元素，如果缓冲区满则覆盖最旧的元素
func (rb *RingBuffer) WriteOverwrite(item interface{}) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	
	rb.buffer[rb.tail] = item
	rb.tail = (rb.tail + 1) % rb.size
	
	if rb.count == rb.size {
		// 缓冲区满，移动head指针
		rb.head = (rb.head + 1) % rb.size
	} else {
		rb.count++
	}
}

// Peek 查看下一个要读取的元素，但不移除它
func (rb *RingBuffer) Peek() (interface{}, error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	
	if rb.count == 0 {
		return nil, ErrBufferEmpty
	}
	
	return rb.buffer[rb.head], nil
}

// Len 返回当前缓冲区中的元素数量
func (rb *RingBuffer) Len() int {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.count
}

// Cap 返回缓冲区的容量
func (rb *RingBuffer) Cap() int {
	return rb.size
}

// IsEmpty 检查缓冲区是否为空
func (rb *RingBuffer) IsEmpty() bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.count == 0
}

// IsFull 检查缓冲区是否已满
func (rb *RingBuffer) IsFull() bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.count == rb.size
}

// Clear 清空缓冲区
func (rb *RingBuffer) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	
	// 清空引用，帮助GC
	for i := 0; i < rb.size; i++ {
		rb.buffer[i] = nil
	}
	
	rb.head = 0
	rb.tail = 0
	rb.count = 0
}

// ToSlice 将缓冲区内容转换为切片（按读取顺序）
func (rb *RingBuffer) ToSlice() []interface{} {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	
	if rb.count == 0 {
		return nil
	}
	
	result := make([]interface{}, rb.count)
	idx := rb.head
	
	for i := 0; i < rb.count; i++ {
		result[i] = rb.buffer[idx]
		idx = (idx + 1) % rb.size
	}
	
	return result
}

// ========== 泛型版本（Go 1.18+）==========

// RingBufferGeneric 泛型环形缓冲区
type RingBufferGeneric[T any] struct {
	buffer []T
	size   int
	head   int
	tail   int
	count  int
	mu     sync.Mutex
}

// NewRingBufferGeneric 创建一个新的泛型环形缓冲区
func NewRingBufferGeneric[T any](size int) *RingBufferGeneric[T] {
	if size <= 0 {
		size = 10
	}
	return &RingBufferGeneric[T]{
		buffer: make([]T, size),
		size:   size,
		head:   0,
		tail:   0,
		count:  0,
	}
}

// Write 写入一个元素到缓冲区
func (rb *RingBufferGeneric[T]) Write(item T) error {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	
	if rb.count == rb.size {
		return ErrBufferFull
	}
	
	rb.buffer[rb.tail] = item
	rb.tail = (rb.tail + 1) % rb.size
	rb.count++
	
	return nil
}

// Read 从缓冲区读取一个元素
func (rb *RingBufferGeneric[T]) Read() (T, error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	
	var zero T
	if rb.count == 0 {
		return zero, ErrBufferEmpty
	}
	
	item := rb.buffer[rb.head]
	rb.buffer[rb.head] = zero // 清空
	rb.head = (rb.head + 1) % rb.size
	rb.count--
	
	return item, nil
}

// WriteOverwrite 写入元素，如果缓冲区满则覆盖最旧的元素
func (rb *RingBufferGeneric[T]) WriteOverwrite(item T) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	
	rb.buffer[rb.tail] = item
	rb.tail = (rb.tail + 1) % rb.size
	
	if rb.count == rb.size {
		rb.head = (rb.head + 1) % rb.size
	} else {
		rb.count++
	}
}

// Peek 查看下一个要读取的元素，但不移除它
func (rb *RingBufferGeneric[T]) Peek() (T, error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	
	var zero T
	if rb.count == 0 {
		return zero, ErrBufferEmpty
	}
	
	return rb.buffer[rb.head], nil
}

// Len 返回当前缓冲区中的元素数量
func (rb *RingBufferGeneric[T]) Len() int {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.count
}

// Cap 返回缓冲区的容量
func (rb *RingBufferGeneric[T]) Cap() int {
	return rb.size
}

// IsEmpty 检查缓冲区是否为空
func (rb *RingBufferGeneric[T]) IsEmpty() bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.count == 0
}

// IsFull 检查缓冲区是否已满
func (rb *RingBufferGeneric[T]) IsFull() bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.count == rb.size
}

// Clear 清空缓冲区
func (rb *RingBufferGeneric[T]) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	
	var zero T
	for i := 0; i < rb.size; i++ {
		rb.buffer[i] = zero
	}
	
	rb.head = 0
	rb.tail = 0
	rb.count = 0
}

// ToSlice 将缓冲区内容转换为切片（按读取顺序）
func (rb *RingBufferGeneric[T]) ToSlice() []T {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	
	if rb.count == 0 {
		return nil
	}
	
	result := make([]T, rb.count)
	idx := rb.head
	
	for i := 0; i < rb.count; i++ {
		result[i] = rb.buffer[idx]
		idx = (idx + 1) % rb.size
	}
	
	return result
}

