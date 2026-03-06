package design_patterns

import (
	"fmt"
	"sync"
	"time"
)

// ========== 基础代理模式 ==========

// Subject 主题接口
type ProxySubject interface {
	Request() string
}

// RealSubject 真实主题
type RealSubject struct{}

func (s *RealSubject) Request() string {
	return "RealSubject: Handling request"
}

// Proxy 代理
type Proxy struct {
	realSubject *RealSubject
}

func NewProxy() *Proxy {
	return &Proxy{}
}

func (p *Proxy) Request() string {
	// 延迟初始化
	if p.realSubject == nil {
		p.realSubject = &RealSubject{}
	}
	fmt.Println("Proxy: Pre-processing")
	result := p.realSubject.Request()
	fmt.Println("Proxy: Post-processing")
	return result
}

// ========== 实际应用：远程代理 ==========

// RemoteService 远程服务接口
type RemoteService interface {
	GetData(id string) (string, error)
	SetData(id, data string) error
}

// RemoteServer 远程服务器（模拟）
type RemoteServer struct {
	data map[string]string
}

func NewRemoteServer() *RemoteServer {
	return &RemoteServer{
		data: make(map[string]string),
	}
}

func (s *RemoteServer) GetData(id string) (string, error) {
	return s.data[id], nil
}

func (s *RemoteServer) SetData(id, data string) error {
	s.data[id] = data
	return nil
}

// RemoteProxy 远程代理
type RemoteProxy struct {
	server   *RemoteServer
	cache    map[string]string
	cacheTTL time.Duration
	mu       sync.RWMutex
}

func NewRemoteProxy(server *RemoteServer) *RemoteProxy {
	return &RemoteProxy{
		server:   server,
		cache:    make(map[string]string),
		cacheTTL: 5 * time.Minute,
	}
}

func (p *RemoteProxy) GetData(id string) (string, error) {
	// 检查缓存
	p.mu.RLock()
	if data, ok := p.cache[id]; ok {
		p.mu.RUnlock()
		fmt.Printf("[REMOTE_PROXY] Cache hit for: %s\n", id)
		return data, nil
	}
	p.mu.RUnlock()

	// 从远程获取
	fmt.Printf("[REMOTE_PROXY] Fetching from remote: %s\n", id)
	data, err := p.server.GetData(id)
	if err != nil {
		return "", err
	}

	// 缓存结果
	p.mu.Lock()
	p.cache[id] = data
	p.mu.Unlock()

	return data, nil
}

func (p *RemoteProxy) SetData(id, data string) error {
	// 更新远程
	err := p.server.SetData(id, data)
	if err != nil {
		return err
	}

	// 更新缓存
	p.mu.Lock()
	p.cache[id] = data
	p.mu.Unlock()

	fmt.Printf("[REMOTE_PROXY] Set data for: %s\n", id)
	return nil
}

// ========== 实际应用：虚拟代理（延迟加载）==========

// Image 图像接口
type Image interface {
	Display() string
}

// RealImage 真实图像
type RealImage struct {
	filename string
	loaded   bool
}

func NewRealImage(filename string) *RealImage {
	return &RealImage{filename: filename}
}

func (i *RealImage) Load() {
	fmt.Printf("[IMAGE] Loading %s from disk...\n", i.filename)
	time.Sleep(100 * time.Millisecond) // 模拟加载时间
	i.loaded = true
}

func (i *RealImage) Display() string {
	if !i.loaded {
		i.Load()
	}
	return fmt.Sprintf("Displaying %s", i.filename)
}

// ImageProxy 图像代理（延迟加载）
type ImageProxy struct {
	realImage *RealImage
	filename  string
}

func NewImageProxy(filename string) *ImageProxy {
	return &ImageProxy{filename: filename}
}

func (p *ImageProxy) Display() string {
	// 延迟初始化真实对象
	if p.realImage == nil {
		fmt.Printf("[IMAGE_PROXY] Creating real image for %s\n", p.filename)
		p.realImage = NewRealImage(p.filename)
	}
	return p.realImage.Display()
}

// ========== 实际应用：保护代理 ==========

// SensitiveResource 敏感资源接口
type SensitiveResource interface {
	Read() string
	Write(data string) error
	Delete() error
}

// RealSensitiveResource 真实敏感资源
type RealSensitiveResource struct {
	data string
}

func (r *RealSensitiveResource) Read() string {
	return r.data
}

func (r *RealSensitiveResource) Write(data string) error {
	r.data = data
	return nil
}

func (r *RealSensitiveResource) Delete() error {
	r.data = ""
	return nil
}

// UserPermission 用户权限
type UserPermission int

const (
	PermissionRead UserPermission = 1 << iota
	PermissionWrite
	PermissionDelete
	PermissionAdmin = PermissionRead | PermissionWrite | PermissionDelete
)

// ProtectionProxy 保护代理
type ProtectionProxy struct {
	resource   *RealSensitiveResource
	permission UserPermission
	user       string
}

func NewProtectionProxy(resource *RealSensitiveResource, permission UserPermission, user string) *ProtectionProxy {
	return &ProtectionProxy{
		resource:   resource,
		permission: permission,
		user:       user,
	}
}

func (p *ProtectionProxy) Read() string {
	if p.permission&PermissionRead == 0 {
		return fmt.Sprintf("[PROTECTION_PROXY] %s has no read permission", p.user)
	}
	fmt.Printf("[PROTECTION_PROXY] %s reading data\n", p.user)
	return p.resource.Read()
}

func (p *ProtectionProxy) Write(data string) error {
	if p.permission&PermissionWrite == 0 {
		return fmt.Errorf("[PROTECTION_PROXY] %s has no write permission", p.user)
	}
	fmt.Printf("[PROTECTION_PROXY] %s writing data: %s\n", p.user, data)
	return p.resource.Write(data)
}

func (p *ProtectionProxy) Delete() error {
	if p.permission&PermissionDelete == 0 {
		return fmt.Errorf("[PROTECTION_PROXY] %s has no delete permission", p.user)
	}
	fmt.Printf("[PROTECTION_PROXY] %s deleting data\n", p.user)
	return p.resource.Delete()
}

// ========== 实际应用：智能代理（缓存、日志、限流）==========

// APIService API服务接口
type APIService interface {
	Call(method string, params map[string]interface{}) (interface{}, error)
}

// RealAPIService 真实API服务
type RealAPIService struct {
	name string
}

func (s *RealAPIService) Call(method string, params map[string]interface{}) (interface{}, error) {
	fmt.Printf("[API] Calling %s.%s with params: %v\n", s.name, method, params)
	return fmt.Sprintf("Result of %s", method), nil
}

// SmartProxy 智能代理
type SmartProxy struct {
	service   *RealAPIService
	cache     map[string]interface{}
	callCount map[string]int
	maxCalls  int
	mu        sync.RWMutex
}

func NewSmartProxy(service *RealAPIService, maxCalls int) *SmartProxy {
	return &SmartProxy{
		service:   service,
		cache:     make(map[string]interface{}),
		callCount: make(map[string]int),
		maxCalls:  maxCalls,
	}
}

func (p *SmartProxy) Call(method string, params map[string]interface{}) (interface{}, error) {
	// 日志
	start := time.Now()
	defer func() {
		fmt.Printf("[SMART_PROXY] Call %s took %v\n", method, time.Since(start))
	}()

	// 限流检查
	p.mu.Lock()
	if p.callCount[method] >= p.maxCalls {
		p.mu.Unlock()
		return nil, fmt.Errorf("[SMART_PROXY] Rate limit exceeded for %s", method)
	}
	p.callCount[method]++
	p.mu.Unlock()

	// 缓存检查
	cacheKey := fmt.Sprintf("%s_%v", method, params)
	p.mu.RLock()
	if result, ok := p.cache[cacheKey]; ok {
		p.mu.RUnlock()
		fmt.Printf("[SMART_PROXY] Cache hit for %s\n", method)
		return result, nil
	}
	p.mu.RUnlock()

	// 调用真实服务
	result, err := p.service.Call(method, params)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	p.mu.Lock()
	p.cache[cacheKey] = result
	p.mu.Unlock()

	return result, nil
}

// ========== 实际应用：数据库连接代理 ==========

// DBConnectionProxy 数据库连接代理
type DBConnectionProxy struct {
	realConn  *DBConnection
	dsn       string
	connected bool
	mu        sync.Mutex
}

func NewDBConnectionProxy(dsn string) *DBConnectionProxy {
	return &DBConnectionProxy{
		dsn: dsn,
	}
}

func (p *DBConnectionProxy) connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.connected {
		return nil
	}

	fmt.Printf("[DB_PROXY] Connecting to: %s\n", p.dsn)
	// 模拟连接
	time.Sleep(50 * time.Millisecond)
	p.connected = true
	return nil
}

func (p *DBConnectionProxy) Query(sql string) (interface{}, error) {
	if err := p.connect(); err != nil {
		return nil, err
	}

	fmt.Printf("[DB_PROXY] Executing: %s\n", sql)
	return fmt.Sprintf("Result: %s", sql), nil
}

func (p *DBConnectionProxy) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.connected {
		return nil
	}

	fmt.Println("[DB_PROXY] Closing connection")
	p.connected = false
	return nil
}

// ========== 实际应用：HTTP反向代理 ==========

// BackendServer 后端服务器
type BackendServer struct {
	Name    string
	Address string
	Healthy bool
}

// LoadBalancerProxy 负载均衡代理
type LoadBalancerProxy struct {
	servers   []*BackendServer
	current   int
	algorithm string // "round_robin", "random", "least_conn"
	mu        sync.Mutex
}

func NewLoadBalancerProxy(servers []*BackendServer, algorithm string) *LoadBalancerProxy {
	return &LoadBalancerProxy{
		servers:   servers,
		algorithm: algorithm,
	}
}

func (p *LoadBalancerProxy) SelectServer() *BackendServer {
	p.mu.Lock()
	defer p.mu.Unlock()

	switch p.algorithm {
	case "round_robin":
		for i := 0; i < len(p.servers); i++ {
			idx := (p.current + i) % len(p.servers)
			if p.servers[idx].Healthy {
				p.current = (idx + 1) % len(p.servers)
				return p.servers[idx]
			}
		}
	}

	return nil
}

func (p *LoadBalancerProxy) ForwardRequest(path string) (string, error) {
	server := p.SelectServer()
	if server == nil {
		return "", fmt.Errorf("[LB_PROXY] No healthy servers available")
	}

	fmt.Printf("[LB_PROXY] Forwarding %s to %s (%s)\n", path, server.Name, server.Address)
	return fmt.Sprintf("Response from %s", server.Name), nil
}

func (p *LoadBalancerProxy) MarkUnhealthy(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, s := range p.servers {
		if s.Name == name {
			s.Healthy = false
			fmt.Printf("[LB_PROXY] Marked %s as unhealthy\n", name)
			break
		}
	}
}

// ========== 实际应用：缓存代理 ==========

// ExpensiveComputation 昂贵计算接口
type ExpensiveComputation interface {
	Compute(input string) string
}

// RealComputation 真实计算
type RealComputation struct{}

func (c *RealComputation) Compute(input string) string {
	// 模拟耗时计算
	time.Sleep(100 * time.Millisecond)
	return fmt.Sprintf("Computed(%s)", input)
}

// CacheProxy 缓存代理
type CacheProxy struct {
	computation ExpensiveComputation
	cache      map[string]cacheEntry
	ttl        time.Duration
	mu         sync.RWMutex
}

type cacheEntry struct {
	value     string
	timestamp time.Time
}

func NewCacheProxy(computation ExpensiveComputation, ttl time.Duration) *CacheProxy {
	return &CacheProxy{
		computation: computation,
		cache:       make(map[string]cacheEntry),
		ttl:         ttl,
	}
}

func (p *CacheProxy) Compute(input string) string {
	// 检查缓存
	p.mu.RLock()
	if entry, ok := p.cache[input]; ok {
		if time.Since(entry.timestamp) < p.ttl {
			p.mu.RUnlock()
			fmt.Printf("[CACHE_PROXY] Cache hit for: %s\n", input)
			return entry.value
		}
	}
	p.mu.RUnlock()

	// 执行计算
	fmt.Printf("[CACHE_PROXY] Computing: %s\n", input)
	result := p.computation.Compute(input)

	// 存入缓存
	p.mu.Lock()
	p.cache[input] = cacheEntry{
		value:     result,
		timestamp: time.Now(),
	}
	p.mu.Unlock()

	return result
}

// ========== 实际应用：事务代理 ==========

// TransactionalService 事务服务接口
type TransactionalService interface {
	Begin() error
	Commit() error
	Rollback() error
	Execute(op string) error
}

// TransactionProxy 事务代理
type TransactionProxy struct {
	inTransaction bool
	operations    []string
}

func NewTransactionProxy() *TransactionProxy {
	return &TransactionProxy{
		operations: make([]string, 0),
	}
}

func (p *TransactionProxy) Begin() error {
	if p.inTransaction {
		return fmt.Errorf("[TX_PROXY] Already in transaction")
	}
	p.inTransaction = true
	p.operations = make([]string, 0)
	fmt.Println("[TX_PROXY] Transaction started")
	return nil
}

func (p *TransactionProxy) Commit() error {
	if !p.inTransaction {
		return fmt.Errorf("[TX_PROXY] No active transaction")
	}

	// 执行所有操作
	fmt.Printf("[TX_PROXY] Committing %d operations\n", len(p.operations))
	for _, op := range p.operations {
		fmt.Printf("[TX_PROXY] Executing: %s\n", op)
	}

	p.inTransaction = false
	p.operations = nil
	fmt.Println("[TX_PROXY] Transaction committed")
	return nil
}

func (p *TransactionProxy) Rollback() error {
	if !p.inTransaction {
		return fmt.Errorf("[TX_PROXY] No active transaction")
	}

	fmt.Printf("[TX_PROXY] Rolling back %d operations\n", len(p.operations))
	p.inTransaction = false
	p.operations = nil
	fmt.Println("[TX_PROXY] Transaction rolled back")
	return nil
}

func (p *TransactionProxy) Execute(op string) error {
	if !p.inTransaction {
		// 不在事务中，直接执行
		fmt.Printf("[TX_PROXY] Direct execution: %s\n", op)
		return nil
	}

	// 在事务中，加入队列
	p.operations = append(p.operations, op)
	fmt.Printf("[TX_PROXY] Queued: %s\n", op)
	return nil
}

// ========== 动态代理（反射实现）==========

// InvocationHandler 调用处理器
type InvocationHandler interface {
	Invoke(method string, args []interface{}) ([]interface{}, error)
}

// DynamicProxy 动态代理
type DynamicProxy struct {
	handler InvocationHandler
}

func NewDynamicProxy(handler InvocationHandler) *DynamicProxy {
	return &DynamicProxy{handler: handler}
}

func (p *DynamicProxy) Invoke(method string, args []interface{}) ([]interface{}, error) {
	fmt.Printf("[DYNAMIC_PROXY] Before method: %s\n", method)
	result, err := p.handler.Invoke(method, args)
	fmt.Printf("[DYNAMIC_PROXY] After method: %s\n", method)
	return result, err
}

// LoggingHandler 日志处理器
type LoggingHandler struct {
	target interface{}
}

func NewLoggingHandler(target interface{}) *LoggingHandler {
	return &LoggingHandler{target: target}
}

func (h *LoggingHandler) Invoke(method string, args []interface{}) ([]interface{}, error) {
	fmt.Printf("[LOG_HANDLER] Method: %s, Args: %v\n", method, args)
	// 这里可以通过反射调用真实对象的方法
	return []interface{}{fmt.Sprintf("Handled %s", method)}, nil
}
