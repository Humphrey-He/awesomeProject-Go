package design_patterns

import (
	"fmt"
	"sync"
)

// ========== 基础观察者模式 ==========

// Observer 观察者接口
type Observer interface {
	Update(data interface{})
}

// Subject 主题接口
type Subject interface {
	Register(observer Observer)
	Remove(observer Observer)
	Notify(data interface{})
}

// ConcreteSubject 具体主题
type ConcreteSubject struct {
	observers []Observer
	mu        sync.RWMutex
}

func NewConcreteSubject() *ConcreteSubject {
	return &ConcreteSubject{
		observers: make([]Observer, 0),
	}
}

func (s *ConcreteSubject) Register(observer Observer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

func (s *ConcreteSubject) Remove(observer Observer) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, obs := range s.observers {
		if obs == observer {
			s.observers = append(s.observers[:i], s.observers[i+1:]...)
			break
		}
	}
}

func (s *ConcreteSubject) Notify(data interface{}) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, observer := range s.observers {
		observer.Update(data)
	}
}

// ConcreteObserver 具体观察者
type ConcreteObserver struct {
	id   string
	data interface{}
}

func NewConcreteObserver(id string) *ConcreteObserver {
	return &ConcreteObserver{id: id}
}

func (o *ConcreteObserver) Update(data interface{}) {
	o.data = data
	fmt.Printf("[Observer %s] Received: %v\n", o.id, data)
}

// ========== 实际应用：事件发布订阅系统 ==========

// Event 事件
type Event struct {
	Name string
	Data interface{}
}

// EventListener 事件监听器
type EventListener interface {
	OnEvent(event Event)
	GetID() string
}

// EventPublisher 事件发布者
type EventPublisher struct {
	listeners map[string][]EventListener
	mu        sync.RWMutex
}

func NewEventPublisher() *EventPublisher {
	return &EventPublisher{
		listeners: make(map[string][]EventListener),
	}
}

func (p *EventPublisher) Subscribe(eventName string, listener EventListener) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.listeners[eventName] = append(p.listeners[eventName], listener)
	fmt.Printf("[EVENT] %s subscribed to: %s\n", listener.GetID(), eventName)
}

func (p *EventPublisher) Unsubscribe(eventName string, listener EventListener) {
	p.mu.Lock()
	defer p.mu.Unlock()

	listeners := p.listeners[eventName]
	for i, l := range listeners {
		if l.GetID() == listener.GetID() {
			p.listeners[eventName] = append(listeners[:i], listeners[i+1:]...)
			fmt.Printf("[EVENT] %s unsubscribed from: %s\n", listener.GetID(), eventName)
			break
		}
	}
}

func (p *EventPublisher) Publish(event Event) {
	p.mu.RLock()
	listeners := p.listeners[event.Name]
	p.mu.RUnlock()

	fmt.Printf("[EVENT] Publishing: %s\n", event.Name)
	for _, listener := range listeners {
		listener.OnEvent(event)
	}
}

// UserEventListener 用户事件监听器
type UserEventListener struct {
	id   string
	name string
}

func NewUserEventListener(id, name string) *UserEventListener {
	return &UserEventListener{id: id, name: name}
}

func (l *UserEventListener) OnEvent(event Event) {
	fmt.Printf("[%s] %s received event '%s': %v\n", l.id, l.name, event.Name, event.Data)
}

func (l *UserEventListener) GetID() string {
	return l.id
}

// ========== 实际应用：消息通知系统 ==========

// NotificationType 通知类型
type NotificationType string

const (
	NotificationEmail    NotificationType = "email"
	NotificationSMS      NotificationType = "sms"
	NotificationPush     NotificationType = "push"
	NotificationInApp    NotificationType = "in_app"
	NotificationWebhook  NotificationType = "webhook"
)

// Notification 通知
type Notification struct {
	Type    NotificationType
	Title   string
	Content string
	To      string
}

// NotificationListener 通知监听器
type NotificationListener interface {
	Send(notification Notification) error
	GetType() NotificationType
}

// NotificationManager 通知管理器
type NotificationManager struct {
	listeners map[NotificationType][]NotificationListener
	mu        sync.RWMutex
}

func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		listeners: make(map[NotificationType][]NotificationListener),
	}
}

func (m *NotificationManager) Register(listener NotificationListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	nt := listener.GetType()
	m.listeners[nt] = append(m.listeners[nt], listener)
	fmt.Printf("[NOTIFICATION] Registered: %s\n", nt)
}

func (m *NotificationManager) Send(notification Notification) {
	m.mu.RLock()
	listeners := m.listeners[notification.Type]
	m.mu.RUnlock()

	fmt.Printf("[NOTIFICATION] Sending %s to: %s\n", notification.Type, notification.To)
	for _, listener := range listeners {
		if err := listener.Send(notification); err != nil {
			fmt.Printf("[NOTIFICATION] Error: %v\n", err)
		}
	}
}

// EmailNotifier 邮件通知器
type EmailNotifier struct {
	smtpServer string
}

func NewEmailNotifier(smtpServer string) *EmailNotifier {
	return &EmailNotifier{smtpServer: smtpServer}
}

func (n *EmailNotifier) Send(notification Notification) error {
	fmt.Printf("[EMAIL] Sending to %s via %s: %s\n", notification.To, n.smtpServer, notification.Title)
	return nil
}

func (n *EmailNotifier) GetType() NotificationType {
	return NotificationEmail
}

// SMSNotifier 短信通知器
type SMSNotifier struct {
	provider string
}

func NewSMSNotifier(provider string) *SMSNotifier {
	return &SMSNotifier{provider: provider}
}

func (n *SMSNotifier) Send(notification Notification) error {
	fmt.Printf("[SMS] Sending to %s via %s: %s\n", notification.To, n.provider, notification.Content)
	return nil
}

func (n *SMSNotifier) GetType() NotificationType {
	return NotificationSMS
}

// PushNotifier 推送通知器
type PushNotifier struct {
	platform string
}

func NewPushNotifier(platform string) *PushNotifier {
	return &PushNotifier{platform: platform}
}

func (n *PushNotifier) Send(notification Notification) error {
	fmt.Printf("[PUSH] Sending to %s on %s: %s\n", notification.To, n.platform, notification.Title)
	return nil
}

func (n *PushNotifier) GetType() NotificationType {
	return NotificationPush
}

// ========== 实际应用：配置变更通知 ==========

// ConfigChangeCallback 配置变更回调
type ConfigChangeCallback func(key string, oldValue, newValue interface{})

// ConfigObserver 配置观察者
type ConfigObserver struct {
	name     string
	callback ConfigChangeCallback
}

func NewConfigObserver(name string, callback ConfigChangeCallback) *ConfigObserver {
	return &ConfigObserver{
		name:     name,
		callback: callback,
	}
}

func (o *ConfigObserver) OnConfigChange(key string, oldValue, newValue interface{}) {
	fmt.Printf("[CONFIG_OBSERVER] %s notified of change: %s\n", o.name, key)
	o.callback(key, oldValue, newValue)
}

// ObservableConfig 可观察的配置
type ObservableConfig struct {
	data      map[string]interface{}
	observers []*ConfigObserver
	mu        sync.RWMutex
}

func NewObservableConfig() *ObservableConfig {
	return &ObservableConfig{
		data:      make(map[string]interface{}),
		observers: make([]*ConfigObserver, 0),
	}
}

func (c *ObservableConfig) AddObserver(observer *ConfigObserver) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.observers = append(c.observers, observer)
}

func (c *ObservableConfig) Set(key string, value interface{}) {
	c.mu.Lock()
	oldValue := c.data[key]
	c.data[key] = value
	observers := make([]*ConfigObserver, len(c.observers))
	copy(observers, c.observers)
	c.mu.Unlock()

	// 通知观察者
	for _, observer := range observers {
		observer.OnConfigChange(key, oldValue, value)
	}
}

func (c *ObservableConfig) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data[key]
}

// ========== 实际应用：数据变化监听 ==========

// DataChangeListener 数据变化监听器
type DataChangeListener interface {
	OnInsert(collection string, document interface{})
	OnUpdate(collection string, id interface{}, changes map[string]interface{})
	OnDelete(collection string, id interface{})
}

// ObservableDataStore 可观察的数据存储
type ObservableDataStore struct {
	data      map[string]map[interface{}]interface{}
	listeners []DataChangeListener
	mu        sync.RWMutex
}

func NewObservableDataStore() *ObservableDataStore {
	return &ObservableDataStore{
		data:      make(map[string]map[interface{}]interface{}),
		listeners: make([]DataChangeListener, 0),
	}
}

func (s *ObservableDataStore) AddListener(listener DataChangeListener) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listeners = append(s.listeners, listener)
}

func (s *ObservableDataStore) Insert(collection string, id, document interface{}) {
	s.mu.Lock()
	if s.data[collection] == nil {
		s.data[collection] = make(map[interface{}]interface{})
	}
	s.data[collection][id] = document
	listeners := make([]DataChangeListener, len(s.listeners))
	copy(listeners, s.listeners)
	s.mu.Unlock()

	for _, listener := range listeners {
		listener.OnInsert(collection, document)
	}
}

func (s *ObservableDataStore) Update(collection string, id interface{}, changes map[string]interface{}) {
	s.mu.RLock()
	listeners := make([]DataChangeListener, len(s.listeners))
	copy(listeners, s.listeners)
	s.mu.RUnlock()

	for _, listener := range listeners {
		listener.OnUpdate(collection, id, changes)
	}
}

func (s *ObservableDataStore) Delete(collection string, id interface{}) {
	s.mu.Lock()
	delete(s.data[collection], id)
	listeners := make([]DataChangeListener, len(s.listeners))
	copy(listeners, s.listeners)
	s.mu.Unlock()

	for _, listener := range listeners {
		listener.OnDelete(collection, id)
	}
}

// LogDataListener 日志数据监听器
type LogDataListener struct{}

func (l *LogDataListener) OnInsert(collection string, document interface{}) {
	fmt.Printf("[DATA_LOG] Insert in %s: %v\n", collection, document)
}

func (l *LogDataListener) OnUpdate(collection string, id interface{}, changes map[string]interface{}) {
	fmt.Printf("[DATA_LOG] Update in %s, id: %v, changes: %v\n", collection, id, changes)
}

func (l *LogDataListener) OnDelete(collection string, id interface{}) {
	fmt.Printf("[DATA_LOG] Delete in %s, id: %v\n", collection, id)
}

// ========== 实际应用：股票价格监控 ==========

// StockPriceListener 股票价格监听器
type StockPriceListener interface {
	OnPriceChange(symbol string, oldPrice, newPrice float64)
}

// StockTicker 股票行情
type StockTicker struct {
	symbol    string
	price     float64
	listeners []StockPriceListener
	mu        sync.RWMutex
}

func NewStockTicker(symbol string, initialPrice float64) *StockTicker {
	return &StockTicker{
		symbol:    symbol,
		price:     initialPrice,
		listeners: make([]StockPriceListener, 0),
	}
}

func (t *StockTicker) AddListener(listener StockPriceListener) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.listeners = append(t.listeners, listener)
}

func (t *StockTicker) SetPrice(newPrice float64) {
	t.mu.Lock()
	oldPrice := t.price
	t.price = newPrice
	listeners := make([]StockPriceListener, len(t.listeners))
	copy(listeners, t.listeners)
	t.mu.Unlock()

	if oldPrice != newPrice {
		fmt.Printf("[STOCK] %s: %.2f -> %.2f\n", t.symbol, oldPrice, newPrice)
		for _, listener := range listeners {
			listener.OnPriceChange(t.symbol, oldPrice, newPrice)
		}
	}
}

func (t *StockTicker) GetPrice() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.price
}

// PriceAlertListener 价格预警监听器
type PriceAlertListener struct {
	name      string
	threshold float64
	above     bool
}

func NewPriceAlertListener(name string, threshold float64, above bool) *PriceAlertListener {
	return &PriceAlertListener{
		name:      name,
		threshold: threshold,
		above:     above,
	}
}

func (l *PriceAlertListener) OnPriceChange(symbol string, oldPrice, newPrice float64) {
	if l.above && newPrice > l.threshold && oldPrice <= l.threshold {
		fmt.Printf("[ALERT] %s: %s 价格 %.2f 超过阈值 %.2f!\n", l.name, symbol, newPrice, l.threshold)
	} else if !l.above && newPrice < l.threshold && oldPrice >= l.threshold {
		fmt.Printf("[ALERT] %s: %s 价格 %.2f 低于阈值 %.2f!\n", l.name, symbol, newPrice, l.threshold)
	}
}

// ========== 实际应用：任务状态监控 ==========

// TaskStatus 任务状态
type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
)

// TaskState 任务状态变化
type TaskState struct {
	TaskID    string
	OldStatus TaskStatus
	NewStatus TaskStatus
	Progress  int
	Message   string
}

// TaskStatusListener 任务状态监听器
type TaskStatusListener interface {
	OnStatusChange(state TaskState)
}

// TaskMonitor 任务监控器
type TaskMonitor struct {
	taskID    string
	status    TaskStatus
	progress  int
	listeners []TaskStatusListener
	mu        sync.RWMutex
}

func NewTaskMonitor(taskID string) *TaskMonitor {
	return &TaskMonitor{
		taskID:    taskID,
		status:    StatusPending,
		progress:  0,
		listeners: make([]TaskStatusListener, 0),
	}
}

func (m *TaskMonitor) AddListener(listener TaskStatusListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, listener)
}

func (m *TaskMonitor) UpdateStatus(newStatus TaskStatus, progress int, message string) {
	m.mu.Lock()
	oldStatus := m.status
	m.status = newStatus
	m.progress = progress
	listeners := make([]TaskStatusListener, len(m.listeners))
	copy(listeners, m.listeners)
	m.mu.Unlock()

	state := TaskState{
		TaskID:    m.taskID,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Progress:  progress,
		Message:   message,
	}

	for _, listener := range listeners {
		listener.OnStatusChange(state)
	}
}

// LoggingTaskListener 日志任务监听器
type LoggingTaskListener struct{}

func (l *LoggingTaskListener) OnStatusChange(state TaskState) {
	fmt.Printf("[TASK] %s: %s -> %s (%d%%) %s\n",
		state.TaskID, state.OldStatus, state.NewStatus, state.Progress, state.Message)
}

// WebhookTaskListener Webhook任务监听器
type WebhookTaskListener struct {
	webhookURL string
}

func NewWebhookTaskListener(webhookURL string) *WebhookTaskListener {
	return &WebhookTaskListener{webhookURL: webhookURL}
}

func (l *WebhookTaskListener) OnStatusChange(state TaskState) {
	fmt.Printf("[WEBHOOK] Would POST to %s: Task %s status change\n", l.webhookURL, state.TaskID)
}

// ========== 泛型观察者模式 ==========

// GenericObserver 泛型观察者接口
type GenericObserver[T any] interface {
	OnUpdate(data T)
}

// GenericSubject 泛型主题
type GenericSubject[T any] struct {
	observers []GenericObserver[T]
	mu        sync.RWMutex
}

func NewGenericSubject[T any]() *GenericSubject[T] {
	return &GenericSubject[T]{
		observers: make([]GenericObserver[T], 0),
	}
}

func (s *GenericSubject[T]) Subscribe(observer GenericObserver[T]) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, observer)
}

func (s *GenericSubject[T]) Notify(data T) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, observer := range s.observers {
		observer.OnUpdate(data)
	}
}
