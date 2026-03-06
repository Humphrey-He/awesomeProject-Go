package design_patterns

import (
	"fmt"
	"sort"
)

// ========== 基础策略模式 ==========

// Strategy 策略接口
type Strategy interface {
	Execute(a, b int) int
}

// AddStrategy 加法策略
type AddStrategy struct{}

func (s *AddStrategy) Execute(a, b int) int {
	return a + b
}

// SubtractStrategy 减法策略
type SubtractStrategy struct{}

func (s *SubtractStrategy) Execute(a, b int) int {
	return a - b
}

// MultiplyStrategy 乘法策略
type MultiplyStrategy struct{}

func (s *MultiplyStrategy) Execute(a, b int) int {
	return a * b
}

// Context 上下文
type Context struct {
	strategy Strategy
}

// NewContext 创建上下文
func NewContext(strategy Strategy) *Context {
	return &Context{strategy: strategy}
}

// SetStrategy 设置策略
func (c *Context) SetStrategy(strategy Strategy) {
	c.strategy = strategy
}

// Execute 执行策略
func (c *Context) Execute(a, b int) int {
	return c.strategy.Execute(a, b)
}

// ========== 实际应用：支付策略 ==========

// PaymentStrategy 支付策略接口
type PaymentStrategy interface {
	Pay(amount float64) string
	GetName() string
}

// AlipayStrategy 支付宝支付
type AlipayStrategy struct {
	account string
}

func NewAlipayStrategy(account string) *AlipayStrategy {
	return &AlipayStrategy{account: account}
}

func (s *AlipayStrategy) Pay(amount float64) string {
	return fmt.Sprintf("支付宝支付 %.2f 元, 账户: %s", amount, s.account)
}

func (s *AlipayStrategy) GetName() string {
	return "支付宝"
}

// WeChatPayStrategy 微信支付
type WeChatPayStrategy struct {
	openID string
}

func NewWeChatPayStrategy(openID string) *WeChatPayStrategy {
	return &WeChatPayStrategy{openID: openID}
}

func (s *WeChatPayStrategy) Pay(amount float64) string {
	return fmt.Sprintf("微信支付 %.2f 元, OpenID: %s", amount, s.openID)
}

func (s *WeChatPayStrategy) GetName() string {
	return "微信支付"
}

// CreditCardStrategy 信用卡支付
type CreditCardStrategy struct {
	cardNumber string
	cvv        string
	expiry     string
}

func NewCreditCardStrategy(cardNumber, cvv, expiry string) *CreditCardStrategy {
	return &CreditCardStrategy{
		cardNumber: cardNumber,
		cvv:        cvv,
		expiry:     expiry,
	}
}

func (s *CreditCardStrategy) Pay(amount float64) string {
	return fmt.Sprintf("信用卡支付 %.2f 元, 卡号: ****%s", amount, s.cardNumber[len(s.cardNumber)-4:])
}

func (s *CreditCardStrategy) GetName() string {
	return "信用卡"
}

// PaymentContext 支付上下文
type PaymentContext struct {
	strategy PaymentStrategy
}

func NewPaymentContext(strategy PaymentStrategy) *PaymentContext {
	return &PaymentContext{strategy: strategy}
}

func (c *PaymentContext) SetPaymentStrategy(strategy PaymentStrategy) {
	c.strategy = strategy
}

func (c *PaymentContext) Pay(amount float64) string {
	fmt.Printf("[PAYMENT] 使用 %s 支付\n", c.strategy.GetName())
	return c.strategy.Pay(amount)
}

// ========== 实际应用：排序策略 ==========

// SortStrategy 排序策略接口
type SortStrategy interface {
	Sort(data []int) []int
	GetName() string
}

// BubbleSortStrategy 冒泡排序
type BubbleSortStrategy struct{}

func (s *BubbleSortStrategy) Sort(data []int) []int {
	result := make([]int, len(data))
	copy(result, data)

	n := len(result)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if result[j] > result[j+1] {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}
	return result
}

func (s *BubbleSortStrategy) GetName() string {
	return "冒泡排序"
}

// QuickSortStrategy 快速排序
type QuickSortStrategy struct{}

func (s *QuickSortStrategy) Sort(data []int) []int {
	result := make([]int, len(data))
	copy(result, data)
	s.quickSort(result, 0, len(result)-1)
	return result
}

func (s *QuickSortStrategy) quickSort(arr []int, low, high int) {
	if low < high {
		pi := s.partition(arr, low, high)
		s.quickSort(arr, low, pi-1)
		s.quickSort(arr, pi+1, high)
	}
}

func (s *QuickSortStrategy) partition(arr []int, low, high int) int {
	pivot := arr[high]
	i := low - 1

	for j := low; j < high; j++ {
		if arr[j] < pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	arr[i+1], arr[high] = arr[high], arr[i+1]
	return i + 1
}

func (s *QuickSortStrategy) GetName() string {
	return "快速排序"
}

// BuiltInSortStrategy 内置排序
type BuiltInSortStrategy struct{}

func (s *BuiltInSortStrategy) Sort(data []int) []int {
	result := make([]int, len(data))
	copy(result, data)
	sort.Ints(result)
	return result
}

func (s *BuiltInSortStrategy) GetName() string {
	return "内置排序"
}

// SortContext 排序上下文
type SortContext struct {
	strategy SortStrategy
}

func NewSortContext(strategy SortStrategy) *SortContext {
	return &SortContext{strategy: strategy}
}

func (c *SortContext) SetStrategy(strategy SortStrategy) {
	c.strategy = strategy
}

func (c *SortContext) Sort(data []int) []int {
	fmt.Printf("[SORT] 使用 %s\n", c.strategy.GetName())
	return c.strategy.Sort(data)
}

// ========== 实际应用：压缩策略 ==========

// CompressionStrategy 压缩策略接口
type CompressionStrategy interface {
	Compress(data string) string
	Decompress(data string) string
	GetName() string
}

// ZipCompressionStrategy ZIP压缩
type ZipCompressionStrategy struct{}

func (s *ZipCompressionStrategy) Compress(data string) string {
	return fmt.Sprintf("ZIP(%s)", data)
}

func (s *ZipCompressionStrategy) Decompress(data string) string {
	return fmt.Sprintf("UNZIP(%s)", data)
}

func (s *ZipCompressionStrategy) GetName() string {
	return "ZIP"
}

// GzipCompressionStrategy GZIP压缩
type GzipCompressionStrategy struct{}

func (s *GzipCompressionStrategy) Compress(data string) string {
	return fmt.Sprintf("GZIP(%s)", data)
}

func (s *GzipCompressionStrategy) Decompress(data string) string {
	return fmt.Sprintf("UNGZIP(%s)", data)
}

func (s *GzipCompressionStrategy) GetName() string {
	return "GZIP"
}

// LZ4CompressionStrategy LZ4压缩
type LZ4CompressionStrategy struct{}

func (s *LZ4CompressionStrategy) Compress(data string) string {
	return fmt.Sprintf("LZ4(%s)", data)
}

func (s *LZ4CompressionStrategy) Decompress(data string) string {
	return fmt.Sprintf("UNLZ4(%s)", data)
}

func (s *LZ4CompressionStrategy) GetName() string {
	return "LZ4"
}

// CompressionContext 压缩上下文
type CompressionContext struct {
	strategy CompressionStrategy
}

func NewCompressionContext(strategy CompressionStrategy) *CompressionContext {
	return &CompressionContext{strategy: strategy}
}

func (c *CompressionContext) Compress(data string) string {
	fmt.Printf("[COMPRESSION] 使用 %s 压缩\n", c.strategy.GetName())
	return c.strategy.Compress(data)
}

func (c *CompressionContext) Decompress(data string) string {
	return c.strategy.Decompress(data)
}

// ========== 实际应用：验证策略 ==========

// ValidationStrategy 验证策略接口
type ValidationStrategy interface {
	Validate(data string) bool
	GetErrorMessage() string
}

// EmailValidationStrategy 邮箱验证
type EmailValidationStrategy struct {
	errorMessage string
}

func (s *EmailValidationStrategy) Validate(data string) bool {
	// 简化的邮箱验证
	hasAt := false
	for _, c := range data {
		if c == '@' {
			hasAt = true
			break
		}
	}
	if !hasAt {
		s.errorMessage = "邮箱格式不正确，缺少@符号"
		return false
	}
	s.errorMessage = ""
	return true
}

func (s *EmailValidationStrategy) GetErrorMessage() string {
	return s.errorMessage
}

// PhoneValidationStrategy 手机号验证
type PhoneValidationStrategy struct {
	errorMessage string
}

func (s *PhoneValidationStrategy) Validate(data string) bool {
	if len(data) != 11 {
		s.errorMessage = "手机号长度必须为11位"
		return false
	}
	if data[0] != '1' {
		s.errorMessage = "手机号必须以1开头"
		return false
	}
	s.errorMessage = ""
	return true
}

func (s *PhoneValidationStrategy) GetErrorMessage() string {
	return s.errorMessage
}

// PasswordValidationStrategy 密码验证
type PasswordValidationStrategy struct {
	minLength    int
	requireUpper bool
	errorMessage string
}

func NewPasswordValidationStrategy(minLength int, requireUpper bool) *PasswordValidationStrategy {
	return &PasswordValidationStrategy{
		minLength:    minLength,
		requireUpper: requireUpper,
	}
}

func (s *PasswordValidationStrategy) Validate(data string) bool {
	if len(data) < s.minLength {
		s.errorMessage = fmt.Sprintf("密码长度至少为%d位", s.minLength)
		return false
	}

	if s.requireUpper {
		hasUpper := false
		for _, c := range data {
			if c >= 'A' && c <= 'Z' {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			s.errorMessage = "密码必须包含大写字母"
			return false
		}
	}

	s.errorMessage = ""
	return true
}

func (s *PasswordValidationStrategy) GetErrorMessage() string {
	return s.errorMessage
}

// Validator 验证器
type Validator struct {
	strategy ValidationStrategy
}

func NewValidator(strategy ValidationStrategy) *Validator {
	return &Validator{strategy: strategy}
}

func (v *Validator) SetStrategy(strategy ValidationStrategy) {
	v.strategy = strategy
}

func (v *Validator) Validate(data string) bool {
	result := v.strategy.Validate(data)
	if !result {
		fmt.Printf("[VALIDATION] 验证失败: %s\n", v.strategy.GetErrorMessage())
	} else {
		fmt.Println("[VALIDATION] 验证通过")
	}
	return result
}

// ========== 实际应用：路由策略 ==========

// RouteStrategy 路由策略接口
type RouteStrategy interface {
	CalculateRoute(from, to string) string
	GetName() string
}

// FastestRouteStrategy 最快路线
type FastestRouteStrategy struct{}

func (s *FastestRouteStrategy) CalculateRoute(from, to string) string {
	return fmt.Sprintf("最快路线: %s -> 高速公路 -> %s", from, to)
}

func (s *FastestRouteStrategy) GetName() string {
	return "最快路线"
}

// ShortestRouteStrategy 最短路线
type ShortestRouteStrategy struct{}

func (s *ShortestRouteStrategy) CalculateRoute(from, to string) string {
	return fmt.Sprintf("最短路线: %s -> 直线距离 -> %s", from, to)
}

func (s *ShortestRouteStrategy) GetName() string {
	return "最短路线"
}

// AvoidTollsRouteStrategy 避开收费
type AvoidTollsRouteStrategy struct{}

func (s *AvoidTollsRouteStrategy) CalculateRoute(from, to string) string {
	return fmt.Sprintf("避开收费路线: %s -> 国道 -> %s", from, to)
}

func (s *AvoidTollsRouteStrategy) GetName() string {
	return "避开收费"
}

// NavigationApp 导航应用
type NavigationApp struct {
	routeStrategy RouteStrategy
}

func NewNavigationApp(strategy RouteStrategy) *NavigationApp {
	return &NavigationApp{routeStrategy: strategy}
}

func (n *NavigationApp) SetRouteStrategy(strategy RouteStrategy) {
	n.routeStrategy = strategy
}

func (n *NavigationApp) Navigate(from, to string) string {
	fmt.Printf("[NAVIGATION] 使用策略: %s\n", n.routeStrategy.GetName())
	return n.routeStrategy.CalculateRoute(from, to)
}

// ========== 函数式策略（Go风格）==========

// CompareFunc 比较函数类型
type CompareFunc func(a, b int) bool

// SortWithStrategy 使用函数式策略排序
func SortWithStrategy(data []int, compare CompareFunc) []int {
	result := make([]int, len(data))
	copy(result, data)

	for i := 0; i < len(result)-1; i++ {
		for j := 0; j < len(result)-i-1; j++ {
			if compare(result[j], result[j+1]) {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}
	return result
}

// Ascending 升序比较
func Ascending(a, b int) bool {
	return a > b
}

// Descending 降序比较
func Descending(a, b int) bool {
	return a < b
}
