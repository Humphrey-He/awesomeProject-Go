package reflect_internals

import (
	"fmt"
	"reflect"
	"unsafe"
)

// ========== Go Reflect 内部原理与实战用法 ==========

/*
本文件深入讲解Go reflect包的内部原理与实战用法，包括：

一、Reflect 内部原理
1. reflect.Type 与 reflect.Value 接口
2. 运行时类型信息结构（rtype）
3. 指针和-interface的内存布局
4. 方法访问与动态调用
5. 指针运算与不安全操作

二、Type 接口层次
1. Type 接口定义
2. 常见类型：Int, String, Struct, Slice, Map, Chan, Ptr, Array, Func
3. 字段遍历与方法获取

三、Value 接口层次
1. Value 接口定义
2. 值操作：Elem, Set, Int, String
3. 动态创建对象
4. 动态调用方法

四、实战用法
1. 序列化/反序列化
2. 动态配置绑定
3. ORM 字段映射
4. 深拷贝
5. 测试用例动态生成

注意：这是教学性质的模拟实现，Go运行时的真实实现在 runtime/
*/

// ========== 1. Reflect 核心数据结构模拟 ==========

/*
Go Reflect 核心原理：

┌─────────────────────────────────────────────────────────────┐
│                      reflect.Type                          │
├─────────────────────────────────────────────────────────────┤
│  type rtype struct {                                        │
│      size       uintptr  // 类型大小                        │
│      ptrdata    uintptr  // 指针数据大小                    │
│      hash       uint32   // 类型哈希                        │
│      tflag      tflag    // 类型标志                        │
│      align      uint8    // 对齐                            │
│      fieldAlign uint8    // 字段对齐                        │
│      kind       uint8    // 基础类型(kind)                  │
│      equal      func(unsafe.Pointer, unsafe.Pointer) bool   │
│      gcdata     *byte    // GC信息                          │
│      str        nameOff  // 类型名偏移                       │
│      ptrToThis typeOff  // 指向该类型的指针                  │
│  }                                                          │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    reflect.Value                           │
├─────────────────────────────────────────────────────────────┤
│  type Value struct {                                       │
│      typ *rtype     // 指向类型信息                         │
│      ptr unsafe.Pointer // 指向数据的指针                   │
│      flag  uintptr  // 标志位(是否可寻址,是否为指针等)      │
│  }                                                          │
└─────────────────────────────────────────────────────────────┘

内存布局：

┌─────────────────────────────────────────────────────────────┐
│  空接口 (interface{}) 内存布局                              │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┬─────────────┐                             │
│  │   _type *   │   data *    │  共16字节                    │
│  │   (类型指针) │  (数据指针)  │                             │
│  └─────────────┴─────────────┘                             │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│  非空接口 (interface{Method()}) 内存布局                    │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┬─────────────┬─────────────┐               │
│  │  itab *     │   data *    │             │  共16字节      │
│  │ (类型+方法表)│  (数据指针)  │             │               │
│  └─────────────┴─────────────┴─────────────┘               │
└─────────────────────────────────────────────────────────────┘
*/

// rtype 模拟 Go 运行时的类型结构（简化版）
// 真实的 runtime.rtype 在 src/runtime/type.go
type rtype struct {
	size       uintptr  // 类型大小
	ptrdata    uintptr  // 指针数据大小
	hash       uint32   // 类型哈希
	tflag      uint8    // 类型标志
	align      uint8    // 对齐
	fieldAlign uint8    // 字段对齐
	kind       uint8    // 基础类型(kind)
	equal      unsafe.Pointer // 相等函数
	gcdata     *byte    // GC信息
	str        int32    // 类型名偏移
	ptrToThis int32    // 指向该类型的指针
}

// tflag 类型标志位
type tflag uint8

const (
	tflagUncommon tflag = 1 << 0  // 是否有非导出方法
	tflagExported         = 1 << 1  // 是否导出
	tflagEmbedInterfacet  = 1 << 2  // 嵌入接口
)

// Kind Go 基础类型
// 与 reflect.Kind 对应
type Kind uint8

const (
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	Array
	Chan
	Func
	Interface
	Map
	Ptr
	Slice
	String
	Struct
	UnsafePointer
)

// nameOff 类型名偏移
type nameOff int32

// textOff 方法表偏移
type textOff int32

// uncommonType 未导出方法类型信息
type uncommonType struct {
	pkgPath nameOff  // 包路径
	mcount  uint16  // 方法数量
	xcount  uint16  // 导出方法数量
	moff    uint32 // 方法表偏移
}

// Method 方法信息
type Method struct {
	Name string // 方法名
	Type *rtype // 方法类型
	Func uintptr // 方法函数指针
}

// imethod 接口方法信息
type imethod struct {
	name nameOff // 方法名偏移
	typ  textOff // 方法类型偏移
}

// itab 接口运行时表示
type itab struct {
	inter *rtype // 接口类型
	_type *rtype // 实体类型
	hash  uint32 // _type.hash 副本
	_     [4]byte
	fun   [1]uintptr // 方法表（可变长度）
}

// ========== 2. 模拟 Type 接口 ==========

// TypeInterface 模拟 reflect.Type 接口
type TypeInterface interface {
	String() string           // 返回类型名
	Kind() Kind               // 返回基础类型
	Size() uintptr            // 返回类型大小
	Align() int               // 返回对齐
	FieldAlign() int         // 返回字段对齐
	NumMethod() int          // 返回方法数量
	Method(int) Method       // 返回方法
	MethodByName(string) (Method, bool) // 按名字查找方法
	NumField() int           // 返回字段数量
	Field(int) StructField   // 返回字段
	FieldByName(string) (StructField, bool) // 按名字查找字段
	NumIn() int              // 返回输入参数数量
	In(int) TypeInterface             // 返回输入参数类型
	NumOut() int             // 返回输出参数数量
	Out(int) TypeInterface            // 返回输出参数类型
	Elem() TypeInterface              // 返回元素类型
	Key() TypeInterface               // 返回Map的Key类型
	ChanDir() ChanDir        // 返回Channel方向
	SliceOf() TypeInterface           // 返回切片类型
	PtrTo() TypeInterface             // 返回指针类型
}

// StructField 字段信息
type StructField struct {
	Name      string    // 字段名
	PkgPath   string    // 包路径
	Type      TypeInterface // 字段类型
	Tag       string    // 字段标签
	Index     []int     // 字段索引
	Anonymous bool     // 是否匿名
}

// ChanDir Channel方向
type ChanDir int

const (
	RecvDir ChanDir = 1 << iota // 接收
	SendDir                     // 发送
	BothDir = RecvDir | SendDir // 双向
)

// SimulatedType 模拟 Type 实现
type SimulatedType struct {
	name    string
	kind    Kind
	size    uintptr
	align   uint8
	methods []Method
	fields  []StructField
	elem    *SimulatedType // 元素类型（用于Ptr, Slice, Chan, Map）
	key     *SimulatedType // Map key类型
	dir     ChanDir        // Channel方向
	in      []*SimulatedType // 输入参数
	out     []*SimulatedType // 输出参数
}

func (t *SimulatedType) String() string   { return t.name }
func (t *SimulatedType) Kind() Kind       { return t.kind }
func (t *SimulatedType) Size() uintptr    { return t.size }
func (t *SimulatedType) Align() int        { return int(t.align) }
func (t *SimulatedType) FieldAlign() int   { return int(t.align) }
func (t *SimulatedType) Elem() TypeInterface {
	if t.elem != nil {
		return t.elem
	}
	return nil
}
func (t *SimulatedType) Key() TypeInterface {
	if t.key != nil {
		return t.key
	}
	return nil
}
func (t *SimulatedType) ChanDir() ChanDir { return t.dir }
func (t *SimulatedType) SliceOf() TypeInterface {
	return &SimulatedType{name: "[]" + t.name, kind: Slice, elem: t}
}
func (t *SimulatedType) PtrTo() TypeInterface {
	return &SimulatedType{name: "*" + t.name, kind: Ptr, elem: t}
}

func (t *SimulatedType) NumMethod() int        { return len(t.methods) }
func (t *SimulatedType) Method(i int) Method   { return t.methods[i] }
func (t *SimulatedType) MethodByName(name string) (Method, bool) {
	for _, m := range t.methods {
		if m.Name == name {
			return m, true
		}
	}
	return Method{}, false
}
func (t *SimulatedType) NumField() int           { return len(t.fields) }
func (t *SimulatedType) Field(i int) StructField { return t.fields[i] }
func (t *SimulatedType) FieldByName(name string) (StructField, bool) {
	for _, f := range t.fields {
		if f.Name == name {
			return f, true
		}
	}
	return StructField{}, false
}
func (t *SimulatedType) NumIn() int  { return len(t.in) }
func (t *SimulatedType) In(i int) TypeInterface { return t.in[i] }
func (t *SimulatedType) NumOut() int { return len(t.out) }
func (t *SimulatedType) Out(i int) TypeInterface { return t.out[i] }

// ========== 3. 模拟 Value 接口 ==========

/*
Value 结构解析：

type Value struct {
    typ *rtype     // 类型指针
    ptr unsafe.Pointer // 数据指针
    flag uintptr   // 标志位
}

flag 标志位：
- 0-3: kind (类型)
- 4:   flagStickyRO (不可写，由未导出字段设置)
- 5:   flagEmbedRO (嵌入的未导出字段)
- 6:   flagIndir (ptr是间接引用)
- 7:   flagAddr (可寻址)
- 8:   flagMayBeNil (可能为nil)
- 9+:  保留
*/

// ValueFlag Value 标志位
type ValueFlag uintptr

const (
	flagKindShift   = 0  // kind偏移
	flagKindMask    = 0x0F
	flagStickyRO    = 1 << 4 // 只读标志
	flagEmbedRO     = 1 << 5 // 嵌入只读
	flagIndir       = 1 << 6 // 间接引用
	flagAddr        = 1 << 7 // 可寻址
	flagMayBeNil    = 1 << 8 // 可能为nil
	flagMethod      = 1 << 9 // 方法
)

// ValueInterface 模拟 reflect.Value 接口
type ValueInterface struct {
	typ  *rtype
	ptr  unsafe.Pointer
	flag ValueFlag
}

// NewValue 创建新的Value
func NewValue(t *rtype, p unsafe.Pointer, flag ValueFlag) *ValueInterface {
	return &ValueInterface{typ: t, ptr: p, flag: flag}
}

// IsValid Value是否有效
func (v *ValueInterface) IsValid() bool {
	return v.typ != nil && v.ptr != nil
}

// Kind 返回值的类型
func (v *ValueInterface) Kind() Kind {
	return Kind(v.flag & flagKindMask)
}

// Type 返回值的Type
func (v *ValueInterface) Type() TypeInterface {
	if v.typ == nil {
		return nil
	}
	// 简化实现
	return &SimulatedType{kind: Kind(v.typ.kind)}
}

// Elem 返回指针/接口/map的Element
func (v *ValueInterface) Elem() *ValueInterface {
	switch v.Kind() {
	case Ptr, Interface:
		// 解引用获取实际值
		// 实际实现会检查ptr是否为nil
		return &ValueInterface{
			typ:  v.typ, // 简化：应该取elem type
			ptr:  *(*unsafe.Pointer)(v.ptr),
			flag: v.flag &^ flagAddr,
		}
	}
	return nil
}

// CanSet 值是否可设置
func (v *ValueInterface) CanSet() bool {
	// 不可寻址则不可设置
	if v.flag&flagAddr == 0 {
		return false
	}
	// 只读标志
	if v.flag&(flagStickyRO|flagEmbedRO) != 0 {
		return false
	}
	return true
}

// Set 设置值
func (v *ValueInterface) Set(x *ValueInterface) {
	if !v.CanSet() {
		panic("reflect: call of Value.Set")
	}
	// 实际会检查类型是否匹配
	// 然后执行内存拷贝
	v.ptr = x.ptr
}

// Int 获取int值
func (v *ValueInterface) Int() int64 {
	if v.Kind() != Int {
		panic("reflect: call of Value.Int")
	}
	// 如果是间接引用，需要取数据
	if v.flag&flagIndir != 0 {
		return *(*int64)(v.ptr)
	}
	return int64(*(*int)(v.ptr))
}

// SetInt 设置int值
func (v *ValueInterface) SetInt(x int64) {
	if !v.CanSet() {
		panic("reflect: call of Value.SetInt")
	}
	if v.Kind() != Int {
		panic("reflect: call of Value.SetInt")
	}
	// 写入数据
	if v.flag&flagIndir != 0 {
		*(*int64)(v.ptr) = x
	} else {
		*(*int)(v.ptr) = int(x)
	}
}

// String 获取string值
func (v *ValueInterface) String() string {
	if v.Kind() != String {
		panic("reflect: call of Value.String")
	}
	if v.flag&flagIndir != 0 {
		return *(*string)(v.ptr)
	}
	return *(*string)(v.ptr)
}

// Interface 将Value转为interface{}
func (v *ValueInterface) Interface() interface{} {
	if v.flag&flagMayBeNil != 0 {
		return nil
	}
	// 实际实现会构造空接口返回
	// return *(interface{})(unsafe.Pointer(&v))
	return nil
}

// ========== 4. Type 与 Value 关系图 ==========

/*
Type 与 Value 关系：

         ┌─────────────────┐
         │  reflect.Type  │
         │   (接口)        │
         └────────┬────────┘
                  │ 实现
                  ▼
         ┌─────────────────┐
         │   *rtype        │  ── 运行时类型信息
         │   (底层结构)     │
         └────────┬────────┘
                  │
                  ▼
         ┌─────────────────┐
         │ 内存布局信息     │
         │ - size          │
         │ - align         │
         │ - kind          │
         │ - method table  │
         └─────────────────┘

         ┌─────────────────┐
         │ reflect.Value  │
         │   (接口)        │
         └────────┬────────┘
                  │ 实现
                  ▼
         ┌─────────────────┐
         │   Value         │
         │   (底层结构)     │
         ├─────────────────┤
         │ typ *rtype      │ ── 指向类型信息
         │ ptr unsafe.Ptr  │ ── 指向实际数据
         │ flag uintptr   │ ── 标志位
         └─────────────────┘

获取流程：

1. 获取 Type:
   t := reflect.TypeOf(x)

2. 获取 Value:
   v := reflect.ValueOf(x)

3. Type → Value:
   v := reflect.ValueOf(x)
   t := v.Type()

4. Value → Type:
   t := reflect.TypeOf(x)
   v := reflect.ValueOf(x)

5. interface{} → Value:
   v := reflect.ValueOf(i)

6. Value → interface{}:
   i := v.Interface()
*/

// ========== 5. 反射核心原理说明 ==========

/*
Go Reflect 核心原理详解：

📌 为什么需要 Reflect？

1. 动态类型：运行时才知道类型
2. 动态操作：运行时才能调用方法
3. 通用框架：序列化、ORM、DI等框架需要

🔍 TypeOf vs ValueOf：

1. reflect.TypeOf(x)
   - 返回 reflect.Type 接口
   - 静态信息：类型名、大小、方法、字段
   - 原理：直接从 interface{} 提取 _type 指针

2. reflect.ValueOf(x)
   - 返回 reflect.Value 接口
   - 动态信息：值本身、可读写性
   - 原理：构造 Value 结构体，包装数据指针

🔑 内存布局关键点：

1. interface{} 存储：
   - _type *：类型信息指针
   - data *：数据指针

2. 指针类型：
   - 存储的是指向数据的指针
   - 通过 Elem() 获取实际值

3. 切片/Map/Channel：
   - 头部包含指向底层数据的指针
   - 通过 Elem() 可获取内部结构

⚠️ 可寻址性 (Addressability)：

值可寻址条件：
1. 可变变量（&x 获取的）
2. 切片元素
3. 数组元素
4. 字段（结构体可寻址时）

不可寻址：
1. 字面量
2. map元素
3. 临时值

❌ 常见误区：

1. ValueOf 返回的是副本
   v := reflect.ValueOf(x)
   v.SetInt(100) // 除非 x 是指针，否则 panic

2. 类型不匹配会 panic
   v := reflect.ValueOf(&x).Elem()
   v.SetString("abc") // 如果 x 不是 string，panic

3. nil interface 的 Type 为 nil
   var i interface{}
   reflect.TypeOf(i) // nil

🔄 性能特点：

- 反射调用有额外开销（约 10x）
- 建议缓存 Type/Value
- 避免频繁反射操作
*/

// ========== 6. 实战用法示例 ==========

// Person 示例结构体
type Person struct {
	Name string `json:"name" db:"person_name"`
	Age  int    `json:"age" db:"person_age"`
}

func (p Person) SayHello(msg string) string {
	return fmt.Sprintf("%s says: %s", p.Name, msg)
}

func (p *Person) SetName(name string) {
	p.Name = name
}

// ========== 6.1 类型检查与转换 ==========

// TypeCheckDemo 类型检查演示
func TypeCheckDemo() string {
	var i interface{} = 42
	var s string = "hello"
	var p Person = Person{Name: "Tom"}

	// 使用 reflect.TypeOf 获取类型
	t1 := reflect.TypeOf(i)
	t2 := reflect.TypeOf(s)
	t3 := reflect.TypeOf(p)

	result := fmt.Sprintf(`TypeOf examples:
  i (int):    %v, Kind: %v
  s (string): %v, Kind: %v
  p (Person): %v, Kind: %v`,
		t1, t1.Kind(),
		t2, t2.Kind(),
		t3, t3.Kind(),
	)
	return result
}

// TypeAssertionDemo 类型断言（不使用反射）
func TypeAssertionDemo() (int, bool) {
	var i interface{} = 42
	// 方式1: 直接断言
	if v, ok := i.(int); ok {
		return v, true
	}
	return 0, false
}

// TypeSwitchDemo 类型-switch（不使用反射）
func TypeSwitchDemo(i interface{}) string {
	switch v := i.(type) {
	case int:
		return fmt.Sprintf("int: %d", v)
	case string:
		return fmt.Sprintf("string: %s", v)
	case Person:
		return fmt.Sprintf("Person: %s", v.Name)
	default:
		return "unknown"
	}
}

// ========== 6.2 动态访问字段 ==========

// GetFieldByName 通过反射获取字段值
func GetFieldByName(obj interface{}, fieldName string) (interface{}, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("not a struct")
	}
	
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return nil, fmt.Errorf("field %s not found", fieldName)
	}
	
	return field.Interface(), nil
}

// SetFieldByName 通过反射设置字段值
func SetFieldByName(obj interface{}, fieldName string, value interface{}) error {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("not a struct")
	}
	
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("field %s not found", fieldName)
	}
	
	if !field.CanSet() {
		return fmt.Errorf("field %s cannot be set", fieldName)
	}
	
	fieldValue := reflect.ValueOf(value)
	if !fieldValue.Type().AssignableTo(field.Type()) {
		return fmt.Errorf("type mismatch")
	}
	
	field.Set(fieldValue)
	return nil
}

// ========== 6.3 动态调用方法 ==========

// CallMethodByName 动态调用方法
func CallMethodByName(obj interface{}, methodName string, args ...interface{}) ([]interface{}, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	method := v.MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("method %s not found", methodName)
	}
	
	// 构建参数
	argsVals := make([]reflect.Value, len(args))
	for i, arg := range args {
		argsVals[i] = reflect.ValueOf(arg)
	}
	
	// 调用方法
	results := method.Call(argsVals)
	
	// 转换结果
	resultVals := make([]interface{}, len(results))
	for i, r := range results {
		resultVals[i] = r.Interface()
	}
	
	return resultVals, nil
}

// ========== 6.4 动态创建对象 ==========

// CreateInstance 动态创建实例
func CreateInstance(typ reflect.Type) (reflect.Value, error) {
	if typ.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("not a struct type")
	}
	
	// 创建新实例
	v := reflect.New(typ).Elem()
	return v, nil
}

// CreateSlice 动态创建切片
func CreateSlice(elemType reflect.Type, length, capacity int) reflect.Value {
	return reflect.MakeSlice(reflect.SliceOf(elemType), length, capacity)
}

// CreateMap 动态创建Map
func CreateMap(keyType, valueType reflect.Type) reflect.Value {
	return reflect.MakeMap(reflect.MapOf(keyType, valueType))
}

// CreateChan 动态创建Channel
func CreateChan(elemType reflect.Type, buffer int) reflect.Value {
	return reflect.MakeChan(reflect.ChanOf(reflect.BothDir, elemType), buffer)
}

// ========== 6.5 深拷贝 ==========

// DeepCopy 深拷贝
func DeepCopy(src interface{}) (interface{}, error) {
	if src == nil {
		return nil, nil
	}
	
	v := reflect.ValueOf(src)
	vCopy := reflect.New(v.Type()).Elem()
	
	if err := deepCopyValue(v, vCopy); err != nil {
		return nil, err
	}
	
	return vCopy.Interface(), nil
}

func deepCopyValue(src, dst reflect.Value) error {
	switch src.Kind() {
	case reflect.Ptr:
		if src.IsNil() {
			return nil
		}
		dst.Set(reflect.New(src.Elem().Type()))
		return deepCopyValue(src.Elem(), dst.Elem())
	
	case reflect.Interface:
		if src.IsNil() {
			return nil
		}
		dst.Set(src)
		return nil
	
	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			if err := deepCopyValue(src.Field(i), dst.Field(i)); err != nil {
				return err
			}
		}
		return nil
	
	case reflect.Slice:
		dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			if err := deepCopyValue(src.Index(i), dst.Index(i)); err != nil {
				return err
			}
		}
		return nil
	
	case reflect.Map:
		dst.Set(reflect.MakeMap(src.Type()))
		for _, k := range src.MapKeys() {
			newVal := reflect.New(src.MapIndex(k).Type()).Elem()
			if err := deepCopyValue(src.MapIndex(k), newVal); err != nil {
				return err
			}
			dst.SetMapIndex(k, newVal)
		}
		return nil
	
	default:
		dst.Set(src)
		return nil
	}
}

// ========== 6.6 标签解析 ==========

// TagDemo 标签解析演示
func TagDemo() string {
	t := reflect.TypeOf(Person{})
	
	var result string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		jsonTag := f.Tag.Get("json")
		dbTag := f.Tag.Get("db")
		result += fmt.Sprintf("Field: %s, json: %s, db: %s\n",
			f.Name, jsonTag, dbTag)
	}
	return result
}

// GetTagValue 获取指定标签的值
func GetTagValue(typ interface{}, fieldName, tagName string) string {
	t := reflect.TypeOf(typ)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	
	field, ok := t.FieldByName(fieldName)
	if !ok {
		return ""
	}
	
	return field.Tag.Get(tagName)
}

// ========== 6.7 序列化/反序列化模拟 ==========

// ToMap 将结构体转为Map（使用反射）
func ToMap(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	if v.Kind() != reflect.Struct {
		return result
	}
	
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)
		
		// 获取json标签作为key
		key := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			key = jsonTag
		}
		
		result[key] = fieldValue.Interface()
	}
	
	return result
}

// StructToMap 结构体转Map（包含嵌套）
func StructToMap(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	toMapRecursive(reflect.ValueOf(obj), "", result)
	return result
}

func toMapRecursive(v reflect.Value, prefix string, result map[string]interface{}) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	
	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldValue := v.Field(i)
			
			key := field.Name
			if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
				key = jsonTag
			}
			if prefix != "" {
				key = prefix + "." + key
			}
			
			if fieldValue.Kind() == reflect.Struct && field.Type != reflect.TypeOf((*interface{})(nil)) {
				toMapRecursive(fieldValue, key, result)
			} else {
				result[key] = fieldValue.Interface()
			}
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			key := fmt.Sprintf("%s[%v]", prefix, k)
			toMapRecursive(v.MapIndex(k), key, result)
		}
	}
}

// ========== 6.8 零值与空值判断 ==========

// IsZero 判断是否为零值
func IsZero(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Slice, reflect.Array, reflect.Map:
		return v.Len() == 0
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !IsZero(v.Field(i)) {
				return false
			}
		}
		return true
	}
	return false
}

// IsEmpty 判断是否为空（与Zero不同）
func IsEmpty(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
		return v.Len() == 0
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !IsEmpty(v.Field(i)) {
				return false
			}
		}
		return true
	}
	return false
}

// ========== 6.9 性能优化示例 ==========

/*
反射性能优化策略：

1. 缓存 Type 信息
   var typeCache = make(map[string]reflect.Type)
   
2. 缓存 Method/Field 信息
   type FieldInfo struct {
       Index int
       Type  reflect.Type
   }
   var fieldCache = make(map[string][]FieldInfo)

3. 使用 reflect.ValueOf 替代 TypeOf
   // 少一次类型展开

4. 避免在热路径使用反射
   // 循环中使用反射很慢

5. 预先计算偏移量
   type StructField struct {
       Offset uintptr
       Type   reflect.Type
   }
*/

// TypeCache 类型缓存
var TypeCache = make(map[string]reflect.Type)

// GetCachedType 获取缓存的类型
func GetCachedType(typ string) reflect.Type {
	if t, ok := TypeCache[typ]; ok {
		return t
	}
	return nil
}

// CacheType 缓存类型
func CacheType(typ string, t reflect.Type) {
	TypeCache[typ] = t
}

// FieldInfo 字段缓存信息
type FieldInfo struct {
	Index int
	Type  reflect.Type
}

// FieldCache 字段缓存
var FieldCache = make(map[string][]FieldInfo)

// GetCachedFields 获取缓存的字段信息
func GetCachedFields(typ string) []FieldInfo {
	if fields, ok := FieldCache[typ]; ok {
		return fields
	}
	return nil
}

// CacheFields 缓存字段信息
func CacheFields(typ string, fields []FieldInfo) {
	FieldCache[typ] = fields
}

// ========== 7. 面试要点总结 ==========

/*
🔬 Go Reflect 面试要点：

Q: reflect.TypeOf 和 reflect.ValueOf 的区别？
A: TypeOf 获取类型信息，ValueOf 获取值信息。Type 是静态的，Value 是动态的。

Q: 什么是可寻址性？
A: 值可以通过 & 获取地址。可寻址才能调用 Set 方法。临时值、map 元素不可寻址。

Q: 如何通过反射修改结构体字段？
A: 需要传入指针，获取 Elem 后设置 CanSet=true 的字段。

Q: 反射的性能问题？
A: 反射调用有 10x 开销，应避免在热路径使用，建议缓存 Type/Value。

Q: 如何判断 interface{} 是否为 nil？
A: (interface{})(nil) 既是 nil Type 也是 nil Value，需要用 reflect.ValueOf(i).IsValid() 判断。

Q: struct tag 的作用？
A: 用于元数据，如 JSON 序列化、ORM 映射等，通过 Tag.Get("name") 获取。

Q: 反射如何调用方法？
A: 通过 MethodByName 获取方法，用 Call 调用，注意参数类型匹配。

Q: Elem 的作用？
A: 获取指针/接口/Map 引用的实际值。

🎯 最佳实践：

1. 避免频繁反射 → 缓存 Type 信息
2. 使用指针操作结构体
3. 检查 CanSet 再设置值
4. 注意空接口与指针接口的区别
5. 优先使用具体类型而非反射
*/

// ========== 8. 完整示例 ==========

// CompleteExample 完整示例
func CompleteExample() {
	// 1. 获取类型
	p := Person{Name: "Tom", Age: 20}
	t := reflect.TypeOf(p)
	v := reflect.ValueOf(p)
	
	fmt.Println("Type:", t)
	fmt.Println("Value:", v)
	
	// 2. 遍历字段
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fmt.Printf("Field: %s, Type: %v, Tag: %s\n", 
			f.Name, f.Type, f.Tag.Get("json"))
	}
	
	// 3. 动态调用方法
	result, _ := CallMethodByName(p, "SayHello", "Hello")
	fmt.Println("Method result:", result)
	
	// 4. 动态修改字段
	vp := reflect.ValueOf(&p)
	vp.Elem().FieldByName("Age").SetInt(25)
	fmt.Println("After set:", p)
	
	// 5. 深拷贝
	pCopy, _ := DeepCopy(p)
	fmt.Println("Copy:", pCopy)
	
	// 6. 转Map
	m := ToMap(p)
	fmt.Println("ToMap:", m)
}
