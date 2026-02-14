package interface_receiver

import (
	"fmt"
	"sync"
)

// ========== 接口定义最佳实践 ==========

// 1. 接口应该小而精确（Interface Segregation Principle）
// Bad: 大而全的接口
type BadAnimalInterface interface {
	Eat()
	Sleep()
	Fly()  // 不是所有动物都会飞
	Swim() // 不是所有动物都会游泳
	Speak()
}

// Good: 小接口组合
type Eater interface {
	Eat()
}

type Sleeper interface {
	Sleep()
}

type Flyer interface {
	Fly()
}

type Swimmer interface {
	Swim()
}

// 组合多个接口
type Bird interface {
	Eater
	Sleeper
	Flyer
}

// 2. 接口定义在使用方，而非实现方
// Bad: 在实现包中定义接口
// package animal
// type Animal interface { ... }

// Good: 在使用包中定义接口
// package zoo
// type Feeder interface {
//     Feed(animal Animal)
// }

// 3. 接受接口，返回具体类型
// Bad: 返回接口（不推荐）
// func ProcessBad() BadAnimalInterface {
//     return implementation
// }

// Good: 返回具体类型
func ProcessGood() *Dog {
	return &Dog{} // 返回具体类型
}

// 参数使用接口
func FeedAnimal(e Eater) {
	e.Eat()
}

// ========== 方法接收器：值 vs 指针 ==========

// Counter 示例结构体
type Counter struct {
	count int
	mu    sync.Mutex
}

// 值接收器示例（不推荐，因为包含Mutex）
// 注意：实际应该使用指针接收器，这里仅作演示
// Bad: 包含Mutex的结构体不应使用值接收器
// func (c Counter) GetValue() int {
//     return c.count
// }

// 指针接收器：正确的做法
func (c *Counter) GetValue() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

// 指针接收器：可以修改原始对象
func (c *Counter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
}

// 指针接收器：读取也建议用指针
func (c *Counter) Get() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

// ========== 接收器选择规则 ==========

/*
使用指针接收器的情况：
1. ✅ 方法需要修改接收器
2. ✅ 接收器是大结构体（避免复制）
3. ✅ 接收器包含互斥锁等不能复制的字段
4. ✅ 保持一致性（同一类型的方法都用指针）

使用值接收器的情况：
1. ✅ 方法不修改接收器
2. ✅ 接收器是小结构体（几个字节）
3. ✅ 接收器是基本类型或小数组
4. ✅ 需要值语义（如 time.Time）
*/

// 示例：大结构体应该用指针接收器
type LargeStruct struct {
	data [1024]int
	name string
}

// Bad: 值接收器会复制 1024 个 int
func (l LargeStruct) ProcessBad() {
	// 复制了 4KB+ 数据
	_ = l.data[0]
}

// Good: 指针接收器，不复制
func (l *LargeStruct) ProcessGood() {
	_ = l.data[0]
}

// 示例：小结构体可以用值接收器
type Point struct {
	X, Y int
}

// 值接收器：Point 很小，复制开销小
func (p Point) Distance() float64 {
	return float64(p.X*p.X + p.Y*p.Y)
}

// ========== 接口实现 ==========

// 值接收器和指针接收器实现接口的区别

type Speaker interface {
	Speak() string
}

type Dog struct {
	Name string
}

type Cat struct {
	Name string
}

// Dog 使用值接收器
func (d Dog) Speak() string {
	return "Woof! I'm " + d.Name
}

// Cat 使用指针接收器
func (c *Cat) Speak() string {
	return "Meow! I'm " + c.Name
}

// 演示接口实现的差异
func DemonstrateInterfaceReceiver() {
	dog := Dog{Name: "Buddy"}
	cat := Cat{Name: "Whiskers"}

	// Dog: 值接收器，值和指针都能满足接口
	var s1 Speaker = dog  // ✅ OK
	var s2 Speaker = &dog // ✅ OK

	// Cat: 指针接收器，只有指针能满足接口
	// var s3 Speaker = cat   // ❌ 编译错误：Cat 未实现 Speaker
	var s4 Speaker = &cat // ✅ OK

	fmt.Println(s1.Speak())
	fmt.Println(s2.Speak())
	fmt.Println(s4.Speak())
}

// ========== 方法集（Method Set）==========

/*
类型 T 的方法集：
- 值接收器的方法

类型 *T 的方法集：
- 值接收器的方法
- 指针接收器的方法

结论：
- 值类型只能调用值接收器方法
- 指针类型可以调用所有方法
- 接口满足：T 只能用值接收器方法，*T 可以用所有方法
*/

type Person struct {
	Name string
	Age  int
}

// 值接收器
func (p Person) IntroduceValue() string {
	return fmt.Sprintf("Hi, I'm %s (value)", p.Name)
}

// 指针接收器
func (p *Person) IntroducePointer() string {
	return fmt.Sprintf("Hi, I'm %s (pointer)", p.Name)
}

// 指针接收器修改
func (p *Person) SetAge(age int) {
	p.Age = age
}

func DemonstrateMethodSet() {
	// 值类型
	p1 := Person{Name: "Alice", Age: 30}
	p1.IntroduceValue()   // ✅ OK
	p1.IntroducePointer() // ✅ OK (Go 自动转换为 &p1)
	p1.SetAge(31)         // ✅ OK (Go 自动转换为 &p1)

	// 指针类型
	p2 := &Person{Name: "Bob", Age: 25}
	p2.IntroduceValue()   // ✅ OK (Go 自动解引用)
	p2.IntroducePointer() // ✅ OK
	p2.SetAge(26)         // ✅ OK

	// 但在接口中不会自动转换
	type Introducer interface {
		IntroducePointer() string
	}

	// var i1 Introducer = p1  // ❌ 编译错误
	var i2 Introducer = p2 // ✅ OK
	_ = i2
}

// ========== 空接口（interface{}）与 any ==========

// Go 1.18+ 推荐使用 any 代替 interface{}
func AcceptAnything(v any) {
	fmt.Printf("Type: %T, Value: %v\n", v, v)
}

// 类型断言
func TypeAssertion(v any) {
	// 单返回值（不安全）
	// s := v.(string) // panic if v is not string

	// 双返回值（安全）
	if s, ok := v.(string); ok {
		fmt.Println("String:", s)
	} else {
		fmt.Println("Not a string")
	}
}

// 类型开关（Type Switch）
func TypeSwitch(v any) {
	switch v := v.(type) {
	case int:
		fmt.Println("Integer:", v)
	case string:
		fmt.Println("String:", v)
	case []int:
		fmt.Println("Int slice:", v)
	case Speaker:
		fmt.Println("Speaker:", v.Speak())
	default:
		fmt.Printf("Unknown type: %T\n", v)
	}
}

// ========== 接口组合 ==========

type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

type Closer interface {
	Close() error
}

// 组合接口
type ReadWriter interface {
	Reader
	Writer
}

type ReadWriteCloser interface {
	Reader
	Writer
	Closer
}

// ========== 接口最佳实践示例 ==========

// 1. 依赖注入：使用接口解耦

type DataStore interface {
	Save(data string) error
	Load(id string) (string, error)
}

// 实现1：MySQL
type MySQLStore struct{}

func (m *MySQLStore) Save(data string) error {
	fmt.Println("Saving to MySQL:", data)
	return nil
}

func (m *MySQLStore) Load(id string) (string, error) {
	return "data from MySQL", nil
}

// 实现2：Redis
type RedisStore struct{}

func (r *RedisStore) Save(data string) error {
	fmt.Println("Saving to Redis:", data)
	return nil
}

func (r *RedisStore) Load(id string) (string, error) {
	return "data from Redis", nil
}

// Service 依赖接口，不依赖具体实现
type Service struct {
	store DataStore
}

func NewService(store DataStore) *Service {
	return &Service{store: store}
}

func (s *Service) ProcessData(data string) error {
	return s.store.Save(data)
}

// 使用示例
func DemonstrateDependencyInjection() {
	// 可以轻松切换实现
	service1 := NewService(&MySQLStore{})
	service1.ProcessData("hello")

	service2 := NewService(&RedisStore{})
	service2.ProcessData("world")
}

// 2. 测试友好：接口便于 mock

type UserRepository interface {
	GetUser(id int) (*User, error)
	SaveUser(user *User) error
}

type User struct {
	ID   int
	Name string
}

// 生产实现
type RealUserRepository struct {
	// db connection
}

func (r *RealUserRepository) GetUser(id int) (*User, error) {
	// 从数据库查询
	return &User{ID: id, Name: "Real User"}, nil
}

func (r *RealUserRepository) SaveUser(user *User) error {
	// 保存到数据库
	return nil
}

// Mock 实现（用于测试）
type MockUserRepository struct {
	users map[int]*User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[int]*User),
	}
}

func (m *MockUserRepository) GetUser(id int) (*User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (m *MockUserRepository) SaveUser(user *User) error {
	m.users[user.ID] = user
	return nil
}

// UserService 依赖接口
type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUserName(id int) (string, error) {
	user, err := s.repo.GetUser(id)
	if err != nil {
		return "", err
	}
	return user.Name, nil
}

// ========== 接口的零值 ==========

func DemonstrateNilInterface() {
	var s Speaker

	// nil 接口
	fmt.Printf("s == nil: %v\n", s == nil) // true

	// 非 nil 接口，但值为 nil
	var c *Cat
	s = c
	fmt.Printf("s == nil: %v\n", s == nil) // false!

	// 检查接口的值是否为 nil
	// 注意：这里 s != nil 总是 true（演示用）
	// 实际代码中应该做类型断言检查
	_, _ = s, c // 使用变量避免未使用警告
	fmt.Println("Interface is not nil, but value might be nil")
}

// 安全的接口检查
func SafeSpeak(s Speaker) {
	if s == nil {
		fmt.Println("Speaker is nil")
		return
	}

	// 类型断言检查具体值
	if c, ok := s.(*Cat); ok && c == nil {
		fmt.Println("Cat pointer is nil")
		return
	}

	fmt.Println(s.Speak())
}

// ========== 性能考虑 ==========

// 接口调用有一定开销（虚函数表查找）
type Calculator interface {
	Add(a, b int) int
}

type SimpleCalc struct{}

func (s SimpleCalc) Add(a, b int) int {
	return a + b
}

// 直接调用（快）
func DirectCall(c *SimpleCalc, a, b int) int {
	return c.Add(a, b)
}

// 接口调用（稍慢，但通常可以忽略）
func InterfaceCall(c Calculator, a, b int) int {
	return c.Add(a, b)
}

// ========== 常见陷阱 ==========

// 陷阱1：返回具体类型的 nil，而非接口的 nil
func GetSpeaker(condition bool) Speaker {
	if condition {
		return &Dog{Name: "Rex"}
	}
	// Bad: 返回具体类型的 nil
	var c *Cat
	return c // 接口不是 nil！
}

// Good: 返回真正的 nil
func GetSpeakerGood(condition bool) Speaker {
	if condition {
		return &Dog{Name: "Rex"}
	}
	return nil
}

// 陷阱2：接口比较
func CompareInterfaces() {
	var s1 Speaker = &Dog{Name: "A"}
	var s2 Speaker = &Dog{Name: "A"}

	// 比较的是接口的动态类型和动态值
	fmt.Println(s1 == s2) // false（不同的指针）

	// 值类型的比较
	var p1 Speaker = Dog{Name: "A"}
	var p2 Speaker = Dog{Name: "A"}
	fmt.Println(p1 == p2) // true（值相同）
}

// ========== 接口与方法接收器最佳实践总结 ==========

/*
接口设计原则：

✅ 1. 保持接口小而精确
   - 单一职责
   - 容易实现
   - 易于测试

✅ 2. 接受接口，返回具体类型
   - 参数使用接口（灵活）
   - 返回值使用具体类型（明确）

✅ 3. 在使用方定义接口
   - 不在实现包定义接口
   - 按需定义接口

✅ 4. 使用标准接口
   - io.Reader, io.Writer
   - fmt.Stringer
   - error

方法接收器选择：

✅ 1. 优先使用指针接收器
   - 需要修改接收器时
   - 接收器是大结构体
   - 包含不可复制的字段（如 sync.Mutex）
   - 保持一致性

✅ 2. 使用值接收器的场景
   - 小结构体（几个字段）
   - 不可变类型（如 time.Time）
   - 基本类型的别名

✅ 3. 理解方法集
   - 值类型：只包含值接收器方法
   - 指针类型：包含所有方法
   - 接口实现受方法集影响

✅ 4. 避免常见陷阱
   - nil 接口 vs 包含 nil 值的接口
   - 接口比较的语义
   - 返回 nil 接口的正确方式

✅ 5. 依赖注入
   - 使用接口解耦
   - 便于测试和替换实现

✅ 6. 性能考虑
   - 接口调用有微小开销
   - 通常可以忽略
   - 性能关键路径考虑使用具体类型
*/
