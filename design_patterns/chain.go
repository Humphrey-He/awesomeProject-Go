package design_patterns

import (
	"fmt"
	"strings"
)

// ========== 基础责任链模式 ==========

// Handler 处理者接口
type Handler interface {
	SetNext(handler Handler) Handler
	Handle(request string) string
}

// AbstractHandler 抽象处理者
type AbstractHandler struct {
	next Handler
}

func (h *AbstractHandler) SetNext(handler Handler) Handler {
	h.next = handler
	return handler
}

func (h *AbstractHandler) Handle(request string) string {
	if h.next != nil {
		return h.next.Handle(request)
	}
	return ""
}

// ConcreteHandlerA 具体处理者A
type ConcreteHandlerA struct {
	AbstractHandler
}

func (h *ConcreteHandlerA) Handle(request string) string {
	if strings.Contains(request, "A") {
		return fmt.Sprintf("Handler A processed: %s", request)
	}
	return h.AbstractHandler.Handle(request)
}

// ConcreteHandlerB 具体处理者B
type ConcreteHandlerB struct {
	AbstractHandler
}

func (h *ConcreteHandlerB) Handle(request string) string {
	if strings.Contains(request, "B") {
		return fmt.Sprintf("Handler B processed: %s", request)
	}
	return h.AbstractHandler.Handle(request)
}

// ConcreteHandlerC 具体处理者C
type ConcreteHandlerC struct {
	AbstractHandler
}

func (h *ConcreteHandlerC) Handle(request string) string {
	if strings.Contains(request, "C") {
		return fmt.Sprintf("Handler C processed: %s", request)
	}
	return h.AbstractHandler.Handle(request)
}

// ========== 实际应用：HTTP中间件链 ==========

// HTTPRequest HTTP请求
type HTTPRequest struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    string
}

// HTTPResponse HTTP响应
type HTTPResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       string
}

// Middleware 中间件接口
type Middleware interface {
	Next(handler Middleware) Middleware
	Process(request *HTTPRequest) (*HTTPResponse, error)
}

// BaseMiddleware 基础中间件
type BaseMiddleware struct {
	next Middleware
}

func (m *BaseMiddleware) Next(handler Middleware) Middleware {
	m.next = handler
	return handler
}

func (m *BaseMiddleware) Process(request *HTTPRequest) (*HTTPResponse, error) {
	if m.next != nil {
		return m.next.Process(request)
	}
	return &HTTPResponse{
		StatusCode: 200,
		Body:       "OK",
	}, nil
}

// LoggingMiddleware 日志中间件
type LoggingMiddleware struct {
	BaseMiddleware
}

func (m *LoggingMiddleware) Process(request *HTTPRequest) (*HTTPResponse, error) {
	fmt.Printf("[MIDDLEWARE] Logging: %s %s\n", request.Method, request.Path)
	return m.BaseMiddleware.Process(request)
}

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	BaseMiddleware
	token string
}

func NewAuthMiddleware(token string) *AuthMiddleware {
	return &AuthMiddleware{token: token}
}

func (m *AuthMiddleware) Process(request *HTTPRequest) (*HTTPResponse, error) {
	token, ok := request.Headers["Authorization"]
	if !ok || token != m.token {
		return &HTTPResponse{
			StatusCode: 401,
			Body:       "Unauthorized",
		}, nil
	}
	fmt.Println("[MIDDLEWARE] Auth: Token validated")
	return m.BaseMiddleware.Process(request)
}

// CORSMiddleware CORS中间件
type CORSMiddleware struct {
	BaseMiddleware
	allowedOrigins []string
}

func NewCORSMiddleware(origins []string) *CORSMiddleware {
	return &CORSMiddleware{allowedOrigins: origins}
}

func (m *CORSMiddleware) Process(request *HTTPRequest) (*HTTPResponse, error) {
	origin := request.Headers["Origin"]
	allowed := false
	for _, o := range m.allowedOrigins {
		if o == "*" || o == origin {
			allowed = true
			break
		}
	}

	if !allowed {
		return &HTTPResponse{
			StatusCode: 403,
			Body:       "Origin not allowed",
		}, nil
	}

	fmt.Printf("[MIDDLEWARE] CORS: Origin %s allowed\n", origin)
	return m.BaseMiddleware.Process(request)
}

// RateLimitMiddleware 限流中间件
type RateLimitMiddleware struct {
	BaseMiddleware
	maxRequests int
	secondCount int
}

func NewRateLimitMiddleware(maxRequests int) *RateLimitMiddleware {
	return &RateLimitMiddleware{maxRequests: maxRequests}
}

func (m *RateLimitMiddleware) Process(request *HTTPRequest) (*HTTPResponse, error) {
	if m.secondCount >= m.maxRequests {
		return &HTTPResponse{
			StatusCode: 429,
			Body:       "Too Many Requests",
		}, nil
	}
	m.secondCount++
	fmt.Printf("[MIDDLEWARE] RateLimit: %d/%d\n", m.secondCount, m.maxRequests)
	return m.BaseMiddleware.Process(request)
}

// ========== 实际应用：审批流程 ==========

// ApprovalRequest 审批请求
type ApprovalRequest struct {
	ID          string
	Title       string
	Amount      float64
	Applicant   string
	ApprovedBy  []string
	Status      string
}

// Approver 审批人接口
type Approver interface {
	SetSuccessor(successor Approver) Approver
	ProcessRequest(request *ApprovalRequest) *ApprovalRequest
}

// BaseApprover 基础审批人
type BaseApprover struct {
	name     string
	limit    float64
	successor Approver
}

func (a *BaseApprover) SetSuccessor(successor Approver) Approver {
	a.successor = successor
	return successor
}

func (a *BaseApprover) ProcessRequest(request *ApprovalRequest) *ApprovalRequest {
	if a.successor != nil {
		return a.successor.ProcessRequest(request)
	}
	return request
}

// TeamLeader 团队组长
type TeamLeader struct {
	BaseApprover
}

func NewTeamLeader(name string, limit float64) *TeamLeader {
	return &TeamLeader{
		BaseApprover: BaseApprover{name: name, limit: limit},
	}
}

func (a *TeamLeader) ProcessRequest(request *ApprovalRequest) *ApprovalRequest {
	if request.Amount <= a.limit {
		request.Status = "Approved"
		request.ApprovedBy = append(request.ApprovedBy, a.name)
		fmt.Printf("[APPROVAL] TeamLeader %s approved: %s (amount: %.2f)\n", a.name, request.Title, request.Amount)
		return request
	}
	fmt.Printf("[APPROVAL] TeamLeader %s: forwarding to next level\n", a.name)
	return a.BaseApprover.ProcessRequest(request)
}

// Manager 经理
type Manager struct {
	BaseApprover
}

func NewManager(name string, limit float64) *Manager {
	return &Manager{
		BaseApprover: BaseApprover{name: name, limit: limit},
	}
}

func (a *Manager) ProcessRequest(request *ApprovalRequest) *ApprovalRequest {
	if request.Amount <= a.limit {
		request.Status = "Approved"
		request.ApprovedBy = append(request.ApprovedBy, a.name)
		fmt.Printf("[APPROVAL] Manager %s approved: %s (amount: %.2f)\n", a.name, request.Title, request.Amount)
		return request
	}
	fmt.Printf("[APPROVAL] Manager %s: forwarding to next level\n", a.name)
	return a.BaseApprover.ProcessRequest(request)
}

// Director 总监
type Director struct {
	BaseApprover
}

func NewDirector(name string, limit float64) *Director {
	return &Director{
		BaseApprover: BaseApprover{name: name, limit: limit},
	}
}

func (a *Director) ProcessRequest(request *ApprovalRequest) *ApprovalRequest {
	if request.Amount <= a.limit {
		request.Status = "Approved"
		request.ApprovedBy = append(request.ApprovedBy, a.name)
		fmt.Printf("[APPROVAL] Director %s approved: %s (amount: %.2f)\n", a.name, request.Title, request.Amount)
		return request
	}
	request.Status = "Rejected"
	fmt.Printf("[APPROVAL] Director %s: Amount exceeds limit, rejected\n", a.name)
	return request
}

// ========== 实际应用：异常处理链 ==========

// ErrorLevel 错误级别
type ErrorLevel int

const (
	ErrorLevelInfo ErrorLevel = iota
	ErrorLevelWarning
	ErrorLevelError
	ErrorLevelCritical
)

// AppError 应用错误
type AppError struct {
	Level   ErrorLevel
	Message string
	Code    int
	Context map[string]interface{}
}

// ErrorHandler 错误处理器接口
type ErrorHandler interface {
	SetNext(handler ErrorHandler) ErrorHandler
	Handle(err *AppError) bool
}

// BaseErrorHandler 基础错误处理器
type BaseErrorHandler struct {
	next ErrorHandler
}

func (h *BaseErrorHandler) SetNext(handler ErrorHandler) ErrorHandler {
	h.next = handler
	return handler
}

func (h *BaseErrorHandler) Handle(err *AppError) bool {
	if h.next != nil {
		return h.next.Handle(err)
	}
	return false
}

// LogErrorHandler 日志错误处理器
type LogErrorHandler struct {
	BaseErrorHandler
}

func (h *LogErrorHandler) Handle(err *AppError) bool {
	fmt.Printf("[ERROR_HANDLER] Logging error: Level=%d, Message=%s\n", err.Level, err.Message)
	return h.BaseErrorHandler.Handle(err)
}

// AlertErrorHandler 告警错误处理器
type AlertErrorHandler struct {
	BaseErrorHandler
	alertLevel ErrorLevel
}

func NewAlertErrorHandler(alertLevel ErrorLevel) *AlertErrorHandler {
	return &AlertErrorHandler{alertLevel: alertLevel}
}

func (h *AlertErrorHandler) Handle(err *AppError) bool {
	if err.Level >= h.alertLevel {
		fmt.Printf("[ERROR_HANDLER] Sending alert for: %s\n", err.Message)
	}
	return h.BaseErrorHandler.Handle(err)
}

// RetryErrorHandler 重试错误处理器
type RetryErrorHandler struct {
	BaseErrorHandler
	maxRetries int
}

func NewRetryErrorHandler(maxRetries int) *RetryErrorHandler {
	return &RetryErrorHandler{maxRetries: maxRetries}
}

func (h *RetryErrorHandler) Handle(err *AppError) bool {
	if err.Code >= 500 && err.Code < 600 {
		fmt.Printf("[ERROR_HANDLER] Scheduling retry for error code: %d\n", err.Code)
	}
	return h.BaseErrorHandler.Handle(err)
}

// FallbackErrorHandler 降级错误处理器
type FallbackErrorHandler struct {
	BaseErrorHandler
}

func (h *FallbackErrorHandler) Handle(err *AppError) bool {
	if err.Level == ErrorLevelCritical {
		fmt.Printf("[ERROR_HANDLER] Activating fallback for critical error: %s\n", err.Message)
		return true
	}
	return h.BaseErrorHandler.Handle(err)
}

// ========== 实际应用：请求处理管道 ==========

// RequestContext 请求上下文
type RequestContext struct {
	RequestID  string
	UserID     string
	Data       map[string]interface{}
	Trace      []string
	Aborted    bool
	StatusCode int
}

// PipelineHandler 管道处理器接口
type PipelineHandler interface {
	SetNext(handler PipelineHandler) PipelineHandler
	Handle(ctx *RequestContext)
}

// BasePipelineHandler 基础管道处理器
type BasePipelineHandler struct {
	next PipelineHandler
}

func (h *BasePipelineHandler) SetNext(handler PipelineHandler) PipelineHandler {
	h.next = handler
	return handler
}

func (h *BasePipelineHandler) Handle(ctx *RequestContext) {
	if !ctx.Aborted && h.next != nil {
		h.next.Handle(ctx)
	}
}

// ValidationHandler 验证处理器
type ValidationHandler struct {
	BasePipelineHandler
}

func (h *ValidationHandler) Handle(ctx *RequestContext) {
	ctx.Trace = append(ctx.Trace, "Validation")

	// 模拟验证
	if ctx.UserID == "" {
		ctx.Aborted = true
		ctx.StatusCode = 400
		fmt.Println("[PIPELINE] Validation failed: missing UserID")
		return
	}

	fmt.Println("[PIPELINE] Validation passed")
	h.BasePipelineHandler.Handle(ctx)
}

// AuthenticationHandler 认证处理器
type AuthenticationHandler struct {
	BasePipelineHandler
}

func (h *AuthenticationHandler) Handle(ctx *RequestContext) {
	ctx.Trace = append(ctx.Trace, "Authentication")

	// 模拟认证
	fmt.Printf("[PIPELINE] Authenticated user: %s\n", ctx.UserID)
	h.BasePipelineHandler.Handle(ctx)
}

// AuthorizationHandler 授权处理器
type AuthorizationHandler struct {
	BasePipelineHandler
	requiredRole string
}

func NewAuthorizationHandler(role string) *AuthorizationHandler {
	return &AuthorizationHandler{requiredRole: role}
}

func (h *AuthorizationHandler) Handle(ctx *RequestContext) {
	ctx.Trace = append(ctx.Trace, "Authorization")

	// 模拟授权检查
	fmt.Printf("[PIPELINE] Authorized for role: %s\n", h.requiredRole)
	h.BasePipelineHandler.Handle(ctx)
}

// CachingHandler 缓存处理器
type CachingHandler struct {
	BasePipelineHandler
	cache map[string]interface{}
}

func NewCachingHandler() *CachingHandler {
	return &CachingHandler{cache: make(map[string]interface{})}
}

func (h *CachingHandler) Handle(ctx *RequestContext) {
	ctx.Trace = append(ctx.Trace, "Caching")

	// 检查缓存
	if data, ok := h.cache[ctx.RequestID]; ok {
		fmt.Printf("[PIPELINE] Cache hit for: %s\n", ctx.RequestID)
		ctx.Data["cached"] = data
		return
	}

	fmt.Println("[PIPELINE] Cache miss")
	h.BasePipelineHandler.Handle(ctx)

	// 存入缓存
	if !ctx.Aborted {
		h.cache[ctx.RequestID] = ctx.Data
	}
}

// RateLimitHandler 限流处理器
type RateLimitHandler struct {
	BasePipelineHandler
	maxRequests int
	counters    map[string]int
}

func NewRateLimitHandler(maxRequests int) *RateLimitHandler {
	return &RateLimitHandler{
		maxRequests: maxRequests,
		counters:    make(map[string]int),
	}
}

func (h *RateLimitHandler) Handle(ctx *RequestContext) {
	ctx.Trace = append(ctx.Trace, "RateLimit")

	h.counters[ctx.UserID]++
	if h.counters[ctx.UserID] > h.maxRequests {
		ctx.Aborted = true
		ctx.StatusCode = 429
		fmt.Printf("[PIPELINE] Rate limit exceeded for: %s\n", ctx.UserID)
		return
	}

	fmt.Printf("[PIPELINE] Rate limit check passed: %d/%d\n", h.counters[ctx.UserID], h.maxRequests)
	h.BasePipelineHandler.Handle(ctx)
}

// BusinessLogicHandler 业务逻辑处理器
type BusinessLogicHandler struct {
	BasePipelineHandler
}

func (h *BusinessLogicHandler) Handle(ctx *RequestContext) {
	ctx.Trace = append(ctx.Trace, "BusinessLogic")

	// 模拟业务处理
	fmt.Printf("[PIPELINE] Processing business logic for: %s\n", ctx.RequestID)
	ctx.Data["result"] = "processed"
	ctx.StatusCode = 200

	h.BasePipelineHandler.Handle(ctx)
}

// ========== 实际应用：过滤器链 ==========

// Filter 过滤器接口
type Filter interface {
	DoFilter(data string) (string, bool)
}

// FilterChain 过滤器链
type FilterChain struct {
	filters []Filter
}

func NewFilterChain() *FilterChain {
	return &FilterChain{
		filters: make([]Filter, 0),
	}
}

func (c *FilterChain) AddFilter(filter Filter) {
	c.filters = append(c.filters, filter)
}

func (c *FilterChain) DoFilter(data string) (string, bool) {
	for _, filter := range c.filters {
		var passed bool
		data, passed = filter.DoFilter(data)
		if !passed {
			return data, false
		}
	}
	return data, true
}

// XSSFilter XSS过滤器
type XSSFilter struct{}

func (f *XSSFilter) DoFilter(data string) (string, bool) {
	// 简单的XSS过滤
	data = strings.ReplaceAll(data, "<script>", "&lt;script&gt;")
	data = strings.ReplaceAll(data, "</script>", "&lt;/script&gt;")
	fmt.Println("[FILTER] XSS filtering applied")
	return data, true
}

// SQLInjectionFilter SQL注入过滤器
type SQLInjectionFilter struct{}

func (f *SQLInjectionFilter) DoFilter(data string) (string, bool) {
	// 简单的SQL注入检查
	dangerous := []string{"DROP", "DELETE", "TRUNCATE", "INSERT", "UPDATE"}
	upperData := strings.ToUpper(data)
	for _, d := range dangerous {
		if strings.Contains(upperData, d) {
			fmt.Printf("[FILTER] SQL injection detected: %s\n", d)
			return "", false
		}
	}
	fmt.Println("[FILTER] SQL injection check passed")
	return data, true
}

// SensitiveWordFilter 敏感词过滤器
type SensitiveWordFilter struct {
	words []string
}

func NewSensitiveWordFilter(words []string) *SensitiveWordFilter {
	return &SensitiveWordFilter{words: words}
}

func (f *SensitiveWordFilter) DoFilter(data string) (string, bool) {
	for _, word := range f.words {
		if strings.Contains(data, word) {
			data = strings.ReplaceAll(data, word, "***")
		}
	}
	fmt.Println("[FILTER] Sensitive word filtering applied")
	return data, true
}

// LengthFilter 长度过滤器
type LengthFilter struct {
	maxLength int
}

func NewLengthFilter(maxLength int) *LengthFilter {
	return &LengthFilter{maxLength: maxLength}
}

func (f *LengthFilter) DoFilter(data string) (string, bool) {
	if len(data) > f.maxLength {
		fmt.Printf("[FILTER] Data too long: %d > %d\n", len(data), f.maxLength)
		return data[:f.maxLength], true
	}
	fmt.Println("[FILTER] Length check passed")
	return data, true
}

// ========== 实际应用：事件处理链 ==========

// EventProcessor 事件处理器接口
type EventProcessor interface {
	SetNext(processor EventProcessor) EventProcessor
	Process(eventType string, data interface{}) interface{}
}

// BaseEventProcessor 基础事件处理器
type BaseEventProcessor struct {
	next EventProcessor
}

func (p *BaseEventProcessor) SetNext(processor EventProcessor) EventProcessor {
	p.next = processor
	return processor
}

func (p *BaseEventProcessor) Process(eventType string, data interface{}) interface{} {
	if p.next != nil {
		return p.next.Process(eventType, data)
	}
	return data
}

// MetricsEventProcessor 指标事件处理器
type MetricsEventProcessor struct {
	BaseEventProcessor
}

func (p *MetricsEventProcessor) Process(eventType string, data interface{}) interface{} {
	fmt.Printf("[EVENT] Recording metric for: %s\n", eventType)
	return p.BaseEventProcessor.Process(eventType, data)
}

// LogEventProcessor 日志事件处理器
type LogEventProcessor struct {
	BaseEventProcessor
}

func (p *LogEventProcessor) Process(eventType string, data interface{}) interface{} {
	fmt.Printf("[EVENT] Logging event: %s, data: %v\n", eventType, data)
	return p.BaseEventProcessor.Process(eventType, data)
}

// PersistEventProcessor 持久化事件处理器
type PersistEventProcessor struct {
	BaseEventProcessor
}

func (p *PersistEventProcessor) Process(eventType string, data interface{}) interface{} {
	fmt.Printf("[EVENT] Persisting event: %s\n", eventType)
	return p.BaseEventProcessor.Process(eventType, data)
}
