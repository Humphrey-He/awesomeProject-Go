package design_patterns

import (
	"fmt"
	"time"
)

// ========== 基础建造者模式 ==========

// ProductForBuilder 产品
type ProductForBuilder struct {
	PartA string
	PartB string
	PartC string
}

// Builder 建造者接口
type Builder interface {
	BuildPartA()
	BuildPartB()
	BuildPartC()
	GetResult() *ProductForBuilder
}

// ConcreteBuilder 具体建造者
type ConcreteBuilder struct {
	product *ProductForBuilder
}

func NewConcreteBuilder() *ConcreteBuilder {
	return &ConcreteBuilder{
		product: &ProductForBuilder{},
	}
}

func (b *ConcreteBuilder) BuildPartA() {
	b.product.PartA = "Part A built"
}

func (b *ConcreteBuilder) BuildPartB() {
	b.product.PartB = "Part B built"
}

func (b *ConcreteBuilder) BuildPartC() {
	b.product.PartC = "Part C built"
}

func (b *ConcreteBuilder) GetResult() *ProductForBuilder {
	return b.product
}

// Director 指挥者
type Director struct {
	builder Builder
}

func NewDirector(builder Builder) *Director {
	return &Director{builder: builder}
}

func (d *Director) Construct() *ProductForBuilder {
	d.builder.BuildPartA()
	d.builder.BuildPartB()
	d.builder.BuildPartC()
	return d.builder.GetResult()
}

// ========== 实际应用：HTTP请求构建器 ==========

// HTTPRequest HTTP请求
type HTTPRequest struct {
	Method      string
	URL         string
	Headers     map[string]string
	QueryParams map[string]string
	Body        string
	Timeout     time.Duration
}

// HTTPRequestBuilder HTTP请求构建器
type HTTPRequestBuilder struct {
	request *HTTPRequest
}

func NewHTTPRequestBuilder() *HTTPRequestBuilder {
	return &HTTPRequestBuilder{
		request: &HTTPRequest{
			Headers:     make(map[string]string),
			QueryParams: make(map[string]string),
			Timeout:     30 * time.Second,
		},
	}
}

func (b *HTTPRequestBuilder) WithMethod(method string) *HTTPRequestBuilder {
	b.request.Method = method
	return b
}

func (b *HTTPRequestBuilder) WithURL(url string) *HTTPRequestBuilder {
	b.request.URL = url
	return b
}

func (b *HTTPRequestBuilder) WithHeader(key, value string) *HTTPRequestBuilder {
	b.request.Headers[key] = value
	return b
}

func (b *HTTPRequestBuilder) WithQueryParam(key, value string) *HTTPRequestBuilder {
	b.request.QueryParams[key] = value
	return b
}

func (b *HTTPRequestBuilder) WithBody(body string) *HTTPRequestBuilder {
	b.request.Body = body
	return b
}

func (b *HTTPRequestBuilder) WithTimeout(timeout time.Duration) *HTTPRequestBuilder {
	b.request.Timeout = timeout
	return b
}

func (b *HTTPRequestBuilder) Build() *HTTPRequest {
	if b.request.Method == "" {
		b.request.Method = "GET"
	}
	return b.request
}

// ========== 实际应用：数据库连接配置构建器 ==========

// DBConfig 数据库配置
type DBConfig struct {
	Driver          string
	Host            string
	Port            int
	Database        string
	Username        string
	Password        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	SSLMode         string
	Timezone        string
}

// DBConfigBuilder 数据库配置构建器
type DBConfigBuilder struct {
	config *DBConfig
}

func NewDBConfigBuilder() *DBConfigBuilder {
	return &DBConfigBuilder{
		config: &DBConfig{
			Port:            3306,
			MaxOpenConns:    100,
			MaxIdleConns:    10,
			ConnMaxLifetime: time.Hour,
		},
	}
}

func (b *DBConfigBuilder) WithDriver(driver string) *DBConfigBuilder {
	b.config.Driver = driver
	return b
}

func (b *DBConfigBuilder) WithHost(host string) *DBConfigBuilder {
	b.config.Host = host
	return b
}

func (b *DBConfigBuilder) WithPort(port int) *DBConfigBuilder {
	b.config.Port = port
	return b
}

func (b *DBConfigBuilder) WithDatabase(database string) *DBConfigBuilder {
	b.config.Database = database
	return b
}

func (b *DBConfigBuilder) WithCredentials(username, password string) *DBConfigBuilder {
	b.config.Username = username
	b.config.Password = password
	return b
}

func (b *DBConfigBuilder) WithPoolConfig(maxOpen, maxIdle int, maxLifetime time.Duration) *DBConfigBuilder {
	b.config.MaxOpenConns = maxOpen
	b.config.MaxIdleConns = maxIdle
	b.config.ConnMaxLifetime = maxLifetime
	return b
}

func (b *DBConfigBuilder) WithSSL(sslMode string) *DBConfigBuilder {
	b.config.SSLMode = sslMode
	return b
}

func (b *DBConfigBuilder) WithTimezone(tz string) *DBConfigBuilder {
	b.config.Timezone = tz
	return b
}

func (b *DBConfigBuilder) Build() *DBConfig {
	return b.config
}

func (c *DBConfig) DSN() string {
	switch c.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", c.Username, c.Password, c.Host, c.Port, c.Database)
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode)
	default:
		return ""
	}
}

// ========== 实际应用：响应构建器 ==========

// APIResponse API响应
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Errors  []Error     `json:"errors,omitempty"`
}

// Meta 元数据
type Meta struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
}

// Error 错误
type Error struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ResponseBuilder 响应构建器
type ResponseBuilder struct {
	response *APIResponse
}

func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{
		response: &APIResponse{
			Code:    200,
			Message: "success",
		},
	}
}

func (b *ResponseBuilder) WithCode(code int) *ResponseBuilder {
	b.response.Code = code
	return b
}

func (b *ResponseBuilder) WithMessage(message string) *ResponseBuilder {
	b.response.Message = message
	return b
}

func (b *ResponseBuilder) WithData(data interface{}) *ResponseBuilder {
	b.response.Data = data
	return b
}

func (b *ResponseBuilder) WithPagination(page, pageSize, total int) *ResponseBuilder {
	b.response.Meta = &Meta{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}
	return b
}

func (b *ResponseBuilder) WithError(field, message string) *ResponseBuilder {
	b.response.Errors = append(b.response.Errors, Error{
		Field:   field,
		Message: message,
	})
	return b
}

func (b *ResponseBuilder) Success(data interface{}) *APIResponse {
	return b.WithCode(200).WithMessage("success").WithData(data).Build()
}

func (b *ResponseBuilder) Fail(code int, message string) *APIResponse {
	return b.WithCode(code).WithMessage(message).Build()
}

func (b *ResponseBuilder) Build() *APIResponse {
	return b.response
}

// ========== 实际应用：查询构建器 ==========

// QueryConfig 查询配置
type QueryConfig struct {
	Table    string
	Selects  []string
	Wheres   []WhereClause
	OrderBy  []OrderByClause
	Limit    int
	Offset   int
	Joins    []JoinClause
	GroupBy  []string
	Having   []WhereClause
	Distinct bool
}

// WhereClause Where条件
type WhereClause struct {
	Field    string
	Operator string
	Value    interface{}
}

// OrderByClause OrderBy条件
type OrderByClause struct {
	Field string
	ASC   bool
}

// JoinClause Join条件
type JoinClause struct {
	Type  string // INNER, LEFT, RIGHT
	Table string
	On    string
}

// QueryBuilder 查询构建器
type QueryBuilder struct {
	config *QueryConfig
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilderBuilder{
		config: &QueryConfig{},
	}
}

type QueryBuilderBuilder struct {
	config *QueryConfig
}

func NewQueryBuilderBuilder(table string) *QueryBuilderBuilder {
	return &QueryBuilderBuilder{
		config: &QueryConfig{
			Table: table,
		},
	}
}

func (b *QueryBuilderBuilder) Select(fields ...string) *QueryBuilderBuilder {
	b.config.Selects = append(b.config.Selects, fields...)
	return b
}

func (b *QueryBuilderBuilder) Where(field, operator string, value interface{}) *QueryBuilderBuilder {
	b.config.Wheres = append(b.config.Wheres, WhereClause{
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return b
}

func (b *QueryBuilderBuilder) OrderBy(field string, asc bool) *QueryBuilderBuilder {
	b.config.OrderBy = append(b.config.OrderBy, OrderByClause{
		Field: field,
		ASC:   asc,
	})
	return b
}

func (b *QueryBuilderBuilder) Limit(limit int) *QueryBuilderBuilder {
	b.config.Limit = limit
	return b
}

func (b *QueryBuilderBuilder) Offset(offset int) *QueryBuilderBuilder {
	b.config.Offset = offset
	return b
}

func (b *QueryBuilderBuilder) Join(joinType, table, on string) *QueryBuilderBuilder {
	b.config.Joins = append(b.config.Joins, JoinClause{
		Type:  joinType,
		Table: table,
		On:    on,
	})
	return b
}

func (b *QueryBuilderBuilder) GroupBy(fields ...string) *QueryBuilderBuilder {
	b.config.GroupBy = append(b.config.GroupBy, fields...)
	return b
}

func (b *QueryBuilderBuilder) Distinct() *QueryBuilderBuilder {
	b.config.Distinct = true
	return b
}

func (b *QueryBuilderBuilder) Build() *QueryConfig {
	return b.config
}

func (c *QueryConfig) ToSQL() string {
	sql := "SELECT "

	if c.Distinct {
		sql += "DISTINCT "
	}

	if len(c.Selects) == 0 {
		sql += "*"
	} else {
		for i, s := range c.Selects {
			if i > 0 {
				sql += ", "
			}
			sql += s
		}
	}

	sql += " FROM " + c.Table

	for _, join := range c.Joins {
		sql += fmt.Sprintf(" %s JOIN %s ON %s", join.Type, join.Table, join.On)
	}

	if len(c.Wheres) > 0 {
		sql += " WHERE "
		for i, w := range c.Wheres {
			if i > 0 {
				sql += " AND "
			}
			sql += fmt.Sprintf("%s %s %v", w.Field, w.Operator, w.Value)
		}
	}

	if len(c.GroupBy) > 0 {
		sql += " GROUP BY "
		for i, g := range c.GroupBy {
			if i > 0 {
				sql += ", "
			}
			sql += g
		}
	}

	if len(c.OrderBy) > 0 {
		sql += " ORDER BY "
		for i, o := range c.OrderBy {
			if i > 0 {
				sql += ", "
			}
			order := "ASC"
			if !o.ASC {
				order = "DESC"
			}
			sql += fmt.Sprintf("%s %s", o.Field, order)
		}
	}

	if c.Limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", c.Limit)
	}

	if c.Offset > 0 {
		sql += fmt.Sprintf(" OFFSET %d", c.Offset)
	}

	return sql
}

// ========== 实际应用：用户实体构建器 ==========

// User 用户实体
type User struct {
	ID        string
	Username  string
	Email     string
	Password  string
	Nickname  string
	Avatar    string
	Status    int
	Role      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserBuilder 用户构建器
type UserBuilder struct {
	user *User
}

func NewUserBuilder() *UserBuilder {
	return &UserBuilder{
		user: &User{
			Status:    1,
			Role:      "user",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

func (b *UserBuilder) WithID(id string) *UserBuilder {
	b.user.ID = id
	return b
}

func (b *UserBuilder) WithUsername(username string) *UserBuilder {
	b.user.Username = username
	return b
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.user.Email = email
	return b
}

func (b *UserBuilder) WithPassword(password string) *UserBuilder {
	b.user.Password = password
	return b
}

func (b *UserBuilder) WithNickname(nickname string) *UserBuilder {
	b.user.Nickname = nickname
	return b
}

func (b *UserBuilder) WithAvatar(avatar string) *UserBuilder {
	b.user.Avatar = avatar
	return b
}

func (b *UserBuilder) WithStatus(status int) *UserBuilder {
	b.user.Status = status
	return b
}

func (b *UserBuilder) WithRole(role string) *UserBuilder {
	b.user.Role = role
	return b
}

func (b *UserBuilder) Build() *User {
	return b.user
}

// ========== 实际应用：邮件构建器 ==========

// Email 邮件
type Email struct {
	From        string
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	Body        string
	IsHTML      bool
	Attachments []string
}

// EmailBuilder 邮件构建器
type EmailBuilder struct {
	email *Email
}

func NewEmailBuilder() *EmailBuilder {
	return &EmailBuilder{
		email: &Email{
			To:          make([]string, 0),
			CC:          make([]string, 0),
			BCC:         make([]string, 0),
			Attachments: make([]string, 0),
		},
	}
}

func (b *EmailBuilder) From(from string) *EmailBuilder {
	b.email.From = from
	return b
}

func (b *EmailBuilder) To(to ...string) *EmailBuilder {
	b.email.To = append(b.email.To, to...)
	return b
}

func (b *EmailBuilder) CC(cc ...string) *EmailBuilder {
	b.email.CC = append(b.email.CC, cc...)
	return b
}

func (b *EmailBuilder) BCC(bcc ...string) *EmailBuilder {
	b.email.BCC = append(b.email.BCC, bcc...)
	return b
}

func (b *EmailBuilder) Subject(subject string) *EmailBuilder {
	b.email.Subject = subject
	return b
}

func (b *EmailBuilder) Body(body string) *EmailBuilder {
	b.email.Body = body
	return b
}

func (b *EmailBuilder) HTML() *EmailBuilder {
	b.email.IsHTML = true
	return b
}

func (b *EmailBuilder) Attachment(path string) *EmailBuilder {
	b.email.Attachments = append(b.email.Attachments, path)
	return b
}

func (b *EmailBuilder) Build() *Email {
	return b.email
}

// ========== 函数式选项模式（Go风格建造者）==========

// ServerConfig 服务器配置
type ServerConfig struct {
	Host    string
	Port    int
	Timeout time.Duration
	Debug   bool
}

// ServerOption 服务器选项函数
type ServerOption func(*ServerConfig)

func WithHost(host string) ServerOption {
	return func(c *ServerConfig) {
		c.Host = host
	}
}

func WithPort(port int) ServerOption {
	return func(c *ServerConfig) {
		c.Port = port
	}
}

func WithTimeout(timeout time.Duration) ServerOption {
	return func(c *ServerConfig) {
		c.Timeout = timeout
	}
}

func WithDebug(debug bool) ServerOption {
	return func(c *ServerConfig) {
		c.Debug = debug
	}
}

// NewServerConfig 创建服务器配置（函数式选项）
func NewServerConfig(opts ...ServerOption) *ServerConfig {
	config := &ServerConfig{
		Host:    "0.0.0.0",
		Port:    8080,
		Timeout: 30 * time.Second,
		Debug:   false,
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}
