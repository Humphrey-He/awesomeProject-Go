package design_patterns

import (
	"fmt"
	"time"
)

// ========== 基础模板方法模式 ==========

// AbstractClass 抽象类
type AbstractClass interface {
	TemplateMethod()
	PrimitiveOperation1()
	PrimitiveOperation2()
}

// ConcreteClass 具体类
type ConcreteClass struct{}

func (c *ConcreteClass) TemplateMethod() {
	c.PrimitiveOperation1()
	c.PrimitiveOperation2()
}

func (c *ConcreteClass) PrimitiveOperation1() {
	fmt.Println("ConcreteClass: Operation 1")
}

func (c *ConcreteClass) PrimitiveOperation2() {
	fmt.Println("ConcreteClass: Operation 2")
}

// ========== 实际应用：数据处理器模板 ==========

// DataProcessorTemplate 数据处理器模板
type DataProcessorTemplate struct{}

// ProcessData 模板方法
func (t *DataProcessorTemplate) ProcessData(data []interface{}) []interface{} {
	data = t.Read(data)
	data = t.Validate(data)
	data = t.Transform(data)
	t.Save(data)
	return data
}

func (t *DataProcessorTemplate) Read(data []interface{}) []interface{} {
	fmt.Println("[TEMPLATE] Reading data...")
	return data
}

func (t *DataProcessorTemplate) Validate(data []interface{}) []interface{} {
	fmt.Println("[TEMPLATE] Validating data...")
	return data
}

func (t *DataProcessorTemplate) Transform(data []interface{}) []interface{} {
	// 默认实现，子类可以覆盖
	fmt.Println("[TEMPLATE] Transforming data (default)...")
	return data
}

func (t *DataProcessorTemplate) Save(data []interface{}) {
	fmt.Println("[TEMPLATE] Saving data...")
}

// CSVDataProcessor CSV数据处理器
type CSVDataProcessor struct {
	DataProcessorTemplate
	delimiter string
}

func NewCSVDataProcessor(delimiter string) *CSVDataProcessor {
	return &CSVDataProcessor{delimiter: delimiter}
}

func (p *CSVDataProcessor) Validate(data []interface{}) []interface{} {
	fmt.Printf("[CSV] Validating CSV data with delimiter '%s'...\n", p.delimiter)
	// CSV特定验证逻辑
	validated := make([]interface{}, 0)
	for _, item := range data {
		if item != nil && item != "" {
			validated = append(validated, item)
		}
	}
	return validated
}

func (p *CSVDataProcessor) Transform(data []interface{}) []interface{} {
	fmt.Println("[CSV] Transforming CSV data...")
	// CSV特定转换逻辑
	result := make([]interface{}, len(data))
	for i, item := range data {
		result[i] = fmt.Sprintf("CSV_TRANSFORMED(%v)", item)
	}
	return result
}

// JSONDataProcessor JSON数据处理器
type JSONDataProcessor struct {
	DataProcessorTemplate
}

func (p *JSONDataProcessor) Validate(data []interface{}) []interface{} {
	fmt.Println("[JSON] Validating JSON data...")
	// JSON特定验证逻辑
	return data
}

func (p *JSONDataProcessor) Transform(data []interface{}) []interface{} {
	fmt.Println("[JSON] Transforming JSON data...")
	// JSON特定转换逻辑
	result := make([]interface{}, len(data))
	for i, item := range data {
		result[i] = fmt.Sprintf("JSON_TRANSFORMED(%v)", item)
	}
	return result
}

// ========== 实际应用：算法骨架模板 ==========

// AlgorithmTemplate 算法模板
type AlgorithmTemplate interface {
	Step1()
	Step2()
	Step3()
	Hook() bool // 钩子方法
}

// AlgorithmExecutor 算法执行器
type AlgorithmExecutor struct {
	template AlgorithmTemplate
}

func NewAlgorithmExecutor(template AlgorithmTemplate) *AlgorithmExecutor {
	return &AlgorithmExecutor{template: template}
}

func (e *AlgorithmExecutor) Execute() {
	fmt.Println("[ALGORITHM] Starting execution...")
	e.template.Step1()

	// 钩子方法决定是否执行Step2
	if e.template.Hook() {
		e.template.Step2()
	}

	e.template.Step3()
	fmt.Println("[ALGORITHM] Execution completed")
}

// FastAlgorithm 快速算法实现
type FastAlgorithm struct{}

func (a *FastAlgorithm) Step1() {
	fmt.Println("[FAST_ALGO] Step 1: Quick initialization")
}

func (a *FastAlgorithm) Step2() {
	fmt.Println("[FAST_ALGO] Step 2: Fast processing")
}

func (a *FastAlgorithm) Step3() {
	fmt.Println("[FAST_ALGO] Step 3: Quick cleanup")
}

func (a *FastAlgorithm) Hook() bool {
	return false // 跳过Step2
}

// DetailedAlgorithm 详细算法实现
type DetailedAlgorithm struct{}

func (a *DetailedAlgorithm) Step1() {
	fmt.Println("[DETAILED_ALGO] Step 1: Thorough initialization")
}

func (a *DetailedAlgorithm) Step2() {
	fmt.Println("[DETAILED_ALGO] Step 2: Detailed processing")
}

func (a *DetailedAlgorithm) Step3() {
	fmt.Println("[DETAILED_ALGO] Step 3: Comprehensive cleanup")
}

func (a *DetailedAlgorithm) Hook() bool {
	return true // 执行Step2
}

// ========== 实际应用：文件处理模板 ==========

// FileProcessor 文件处理器接口
type FileProcessor interface {
	Open(filename string) error
	Read() ([]byte, error)
	Process(data []byte) ([]byte, error)
	Write(data []byte) error
	Close() error
}

// FileProcessorTemplate 文件处理器模板
type FileProcessorTemplate struct {
	filename string
}

// Process 模板方法
func (t *FileProcessorTemplate) Process(processor FileProcessor, filename string) error {
	// 1. 打开文件
	if err := processor.Open(filename); err != nil {
		return err
	}
	defer processor.Close()

	// 2. 读取数据
	data, err := processor.Read()
	if err != nil {
		return err
	}

	// 3. 处理数据
	processed, err := processor.Process(data)
	if err != nil {
		return err
	}

	// 4. 写入结果
	return processor.Write(processed)
}

// TextFileProcessor 文本文件处理器
type TextFileProcessor struct {
	FileProcessorTemplate
	content string
}

func (p *TextFileProcessor) Open(filename string) error {
	fmt.Printf("[TEXT_FILE] Opening: %s\n", filename)
	p.content = "sample text content"
	return nil
}

func (p *TextFileProcessor) Read() ([]byte, error) {
	fmt.Println("[TEXT_FILE] Reading content...")
	return []byte(p.content), nil
}

func (p *TextFileProcessor) Process(data []byte) ([]byte, error) {
	fmt.Println("[TEXT_FILE] Processing text...")
	// 文本处理：转大写
	result := make([]byte, len(data))
	for i, b := range data {
		if b >= 'a' && b <= 'z' {
			result[i] = b - 32
		} else {
			result[i] = b
		}
	}
	return result, nil
}

func (p *TextFileProcessor) Write(data []byte) error {
	fmt.Printf("[TEXT_FILE] Writing: %s\n", string(data))
	return nil
}

func (p *TextFileProcessor) Close() error {
	fmt.Println("[TEXT_FILE] Closing file")
	return nil
}

// BinaryFileProcessor 二进制文件处理器
type BinaryFileProcessor struct {
	FileProcessorTemplate
	data []byte
}

func (p *BinaryFileProcessor) Open(filename string) error {
	fmt.Printf("[BINARY_FILE] Opening: %s\n", filename)
	p.data = []byte{0x01, 0x02, 0x03, 0x04}
	return nil
}

func (p *BinaryFileProcessor) Read() ([]byte, error) {
	fmt.Println("[BINARY_FILE] Reading binary data...")
	return p.data, nil
}

func (p *BinaryFileProcessor) Process(data []byte) ([]byte, error) {
	fmt.Println("[BINARY_FILE] Processing binary...")
	// 二进制处理：每个字节+1
	result := make([]byte, len(data))
	for i, b := range data {
		result[i] = b + 1
	}
	return result, nil
}

func (p *BinaryFileProcessor) Write(data []byte) error {
	fmt.Printf("[BINARY_FILE] Writing: %v\n", data)
	return nil
}

func (p *BinaryFileProcessor) Close() error {
	fmt.Println("[BINARY_FILE] Closing file")
	return nil
}

// ========== 实际应用：API请求模板 ==========

// APIRequestHandler API请求处理器接口
type APIRequestHandler interface {
	ValidateRequest(request map[string]interface{}) error
	Authenticate(token string) error
	Authorize(userID string, resource string) bool
	ProcessRequest(request map[string]interface{}) (interface{}, error)
	FormatResponse(result interface{}) map[string]interface{}
	LogRequest(request map[string]interface{}, response map[string]interface{})
}

// APIRequestTemplate API请求模板
type APIRequestTemplate struct {
	handler APIRequestHandler
}

func NewAPIRequestTemplate(handler APIRequestHandler) *APIRequestTemplate {
	return &APIRequestTemplate{handler: handler}
}

// HandleRequest 模板方法
func (t *APIRequestTemplate) HandleRequest(request map[string]interface{}) map[string]interface{} {
	// 1. 验证请求
	if err := t.handler.ValidateRequest(request); err != nil {
		return t.errorResponse(400, err.Error())
	}

	// 2. 认证
	token, _ := request["token"].(string)
	if err := t.handler.Authenticate(token); err != nil {
		return t.errorResponse(401, "Authentication failed")
	}

	// 3. 授权
	userID, _ := request["user_id"].(string)
	resource, _ := request["resource"].(string)
	if !t.handler.Authorize(userID, resource) {
		return t.errorResponse(403, "Authorization failed")
	}

	// 4. 处理请求
	result, err := t.handler.ProcessRequest(request)
	if err != nil {
		return t.errorResponse(500, err.Error())
	}

	// 5. 格式化响应
	response := t.handler.FormatResponse(result)

	// 6. 记录日志
	t.handler.LogRequest(request, response)

	return response
}

func (t *APIRequestTemplate) errorResponse(code int, message string) map[string]interface{} {
	return map[string]interface{}{
		"code":    code,
		"message": message,
	}
}

// UserAPIHandler 用户API处理器
type UserAPIHandler struct{}

func (h *UserAPIHandler) ValidateRequest(request map[string]interface{}) error {
	fmt.Println("[USER_API] Validating request...")
	if _, ok := request["user_id"]; !ok {
		return fmt.Errorf("missing user_id")
	}
	return nil
}

func (h *UserAPIHandler) Authenticate(token string) error {
	fmt.Printf("[USER_API] Authenticating token: %s\n", token)
	if token == "" {
		return fmt.Errorf("invalid token")
	}
	return nil
}

func (h *UserAPIHandler) Authorize(userID string, resource string) bool {
	fmt.Printf("[USER_API] Authorizing user %s for %s\n", userID, resource)
	return true
}

func (h *UserAPIHandler) ProcessRequest(request map[string]interface{}) (interface{}, error) {
	fmt.Println("[USER_API] Processing user request...")
	return map[string]interface{}{
		"user_id":   request["user_id"],
		"name":      "Test User",
		"email":     "test@example.com",
		"created":   time.Now(),
	}, nil
}

func (h *UserAPIHandler) FormatResponse(result interface{}) map[string]interface{} {
	fmt.Println("[USER_API] Formatting response...")
	return map[string]interface{}{
		"code":    200,
		"message": "success",
		"data":    result,
	}
}

func (h *UserAPIHandler) LogRequest(request map[string]interface{}, response map[string]interface{}) {
	fmt.Printf("[USER_API] Logging request: %v -> %v\n", request["user_id"], response["code"])
}

// ========== 实际应用：数据库操作模板 ==========

// DatabaseOperation 数据库操作接口
type DatabaseOperation interface {
	Connect() error
	BeginTransaction() error
	Execute(query string) (interface{}, error)
	Commit() error
	Rollback() error
	Close() error
}

// DatabaseOperationTemplate 数据库操作模板
type DatabaseOperationTemplate struct {
	db DatabaseOperation
}

func NewDatabaseOperationTemplate(db DatabaseOperation) *DatabaseOperationTemplate {
	return &DatabaseOperationTemplate{db: db}
}

// ExecuteWithTransaction 带事务的执行模板
func (t *DatabaseOperationTemplate) ExecuteWithTransaction(query string) (interface{}, error) {
	// 1. 连接数据库
	if err := t.db.Connect(); err != nil {
		return nil, err
	}
	defer t.db.Close()

	// 2. 开启事务
	if err := t.db.BeginTransaction(); err != nil {
		return nil, err
	}

	// 3. 执行操作
	result, err := t.db.Execute(query)
	if err != nil {
		// 出错回滚
		t.db.Rollback()
		return nil, err
	}

	// 4. 提交事务
	if err := t.db.Commit(); err != nil {
		return nil, err
	}

	return result, nil
}

// MySQLDatabase MySQL数据库实现
type MySQLDatabase struct {
	connected bool
	inTx      bool
}

func (d *MySQLDatabase) Connect() error {
	fmt.Println("[MYSQL] Connecting...")
	d.connected = true
	return nil
}

func (d *MySQLDatabase) BeginTransaction() error {
	fmt.Println("[MYSQL] Beginning transaction...")
	d.inTx = true
	return nil
}

func (d *MySQLDatabase) Execute(query string) (interface{}, error) {
	fmt.Printf("[MYSQL] Executing: %s\n", query)
	return fmt.Sprintf("Result of: %s", query), nil
}

func (d *MySQLDatabase) Commit() error {
	fmt.Println("[MYSQL] Committing...")
	d.inTx = false
	return nil
}

func (d *MySQLDatabase) Rollback() error {
	fmt.Println("[MYSQL] Rolling back...")
	d.inTx = false
	return nil
}

func (d *MySQLDatabase) Close() error {
	fmt.Println("[MYSQL] Closing connection...")
	d.connected = false
	return nil
}

// ========== 实际应用：消息处理模板 ==========

// MessageHandler 消息处理器接口
type MessageHandler interface {
	Decode(raw []byte) (interface{}, error)
	Validate(msg interface{}) error
	Process(msg interface{}) (interface{}, error)
	Encode(result interface{}) ([]byte, error)
}

// MessageProcessor 消息处理器模板
type MessageProcessor struct {
	handler MessageHandler
}

func NewMessageProcessor(handler MessageHandler) *MessageProcessor {
	return &MessageProcessor{handler: handler}
}

// Process 模板方法
func (p *MessageProcessor) Process(raw []byte) ([]byte, error) {
	// 1. 解码
	msg, err := p.handler.Decode(raw)
	if err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	// 2. 验证
	if err := p.handler.Validate(msg); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// 3. 处理
	result, err := p.handler.Process(msg)
	if err != nil {
		return nil, fmt.Errorf("process error: %w", err)
	}

	// 4. 编码
	output, err := p.handler.Encode(result)
	if err != nil {
		return nil, fmt.Errorf("encode error: %w", err)
	}

	return output, nil
}

// JSONMessageHandler JSON消息处理器
type JSONMessageHandler struct{}

func (h *JSONMessageHandler) Decode(raw []byte) (interface{}, error) {
	fmt.Printf("[JSON_MSG] Decoding: %s\n", string(raw))
	return map[string]interface{}{"data": string(raw)}, nil
}

func (h *JSONMessageHandler) Validate(msg interface{}) error {
	fmt.Println("[JSON_MSG] Validating...")
	return nil
}

func (h *JSONMessageHandler) Process(msg interface{}) (interface{}, error) {
	fmt.Println("[JSON_MSG] Processing...")
	return map[string]interface{}{"result": "processed"}, nil
}

func (h *JSONMessageHandler) Encode(result interface{}) ([]byte, error) {
	fmt.Println("[JSON_MSG] Encoding...")
	return []byte(`{"result": "processed"}`), nil
}

// ========== 实际应用：报表生成模板 ==========

// ReportGenerator 报表生成器接口
type ReportGenerator interface {
	CollectData() ([]interface{}, error)
	ProcessData(data []interface{}) ([]interface{}, error)
	GenerateContent(processed []interface{}) (string, error)
	ApplyStyle(content string) (string, error)
	Output(result string) error
}

// ReportTemplate 报表模板
type ReportTemplate struct {
	generator ReportGenerator
}

func NewReportTemplate(generator ReportGenerator) *ReportTemplate {
	return &ReportTemplate{generator: generator}
}

// Generate 模板方法
func (t *ReportTemplate) Generate() error {
	// 1. 收集数据
	data, err := t.generator.CollectData()
	if err != nil {
		return err
	}

	// 2. 处理数据
	processed, err := t.generator.ProcessData(data)
	if err != nil {
		return err
	}

	// 3. 生成内容
	content, err := t.generator.GenerateContent(processed)
	if err != nil {
		return err
	}

	// 4. 应用样式
	styled, err := t.generator.ApplyStyle(content)
	if err != nil {
		return err
	}

	// 5. 输出
	return t.generator.Output(styled)
}

// PDFReportGenerator PDF报表生成器
type PDFReportGenerator struct{}

func (g *PDFReportGenerator) CollectData() ([]interface{}, error) {
	fmt.Println("[PDF_REPORT] Collecting data...")
	return []interface{}{"item1", "item2", "item3"}, nil
}

func (g *PDFReportGenerator) ProcessData(data []interface{}) ([]interface{}, error) {
	fmt.Println("[PDF_REPORT] Processing data...")
	return data, nil
}

func (g *PDFReportGenerator) GenerateContent(processed []interface{}) (string, error) {
	fmt.Println("[PDF_REPORT] Generating PDF content...")
	return "PDF Content", nil
}

func (g *PDFReportGenerator) ApplyStyle(content string) (string, error) {
	fmt.Println("[PDF_REPORT] Applying PDF style...")
	return fmt.Sprintf("[PDF_STYLED] %s", content), nil
}

func (g *PDFReportGenerator) Output(result string) error {
	fmt.Printf("[PDF_REPORT] Output: %s\n", result)
	return nil
}

// ExcelReportGenerator Excel报表生成器
type ExcelReportGenerator struct{}

func (g *ExcelReportGenerator) CollectData() ([]interface{}, error) {
	fmt.Println("[EXCEL_REPORT] Collecting data...")
	return []interface{}{"row1", "row2", "row3"}, nil
}

func (g *ExcelReportGenerator) ProcessData(data []interface{}) ([]interface{}, error) {
	fmt.Println("[EXCEL_REPORT] Processing data...")
	return data, nil
}

func (g *ExcelReportGenerator) GenerateContent(processed []interface{}) (string, error) {
	fmt.Println("[EXCEL_REPORT] Generating Excel content...")
	return "Excel Content", nil
}

func (g *ExcelReportGenerator) ApplyStyle(content string) (string, error) {
	fmt.Println("[EXCEL_REPORT] Applying Excel style...")
	return fmt.Sprintf("[EXCEL_STYLED] %s", content), nil
}

func (g *ExcelReportGenerator) Output(result string) error {
	fmt.Printf("[EXCEL_REPORT] Output: %s\n", result)
	return nil
}

// ========== 实际应用：任务执行模板 ==========

// TaskExecutor 任务执行器接口
type TaskExecutor interface {
	Initialize() error
	PreExecute() error
	DoExecute() (interface{}, error)
	PostExecute() error
	Cleanup()
}

// TaskTemplate 任务模板
type TaskTemplate struct {
	executor TaskExecutor
}

func NewTaskTemplate(executor TaskExecutor) *TaskTemplate {
	return &TaskTemplate{executor: executor}
}

// Execute 模板方法
func (t *TaskTemplate) Execute() (interface{}, error) {
	// 1. 初始化
	if err := t.executor.Initialize(); err != nil {
		t.executor.Cleanup()
		return nil, err
	}

	// 2. 预执行
	if err := t.executor.PreExecute(); err != nil {
		t.executor.Cleanup()
		return nil, err
	}

	// 3. 执行
	result, err := t.executor.DoExecute()
	if err != nil {
		t.executor.Cleanup()
		return nil, err
	}

	// 4. 后执行
	if err := t.executor.PostExecute(); err != nil {
		t.executor.Cleanup()
		return nil, err
	}

	// 5. 清理
	t.executor.Cleanup()

	return result, nil
}

// BatchTaskExecutor 批量任务执行器
type BatchTaskExecutor struct {
	items []string
}

func (e *BatchTaskExecutor) Initialize() error {
	fmt.Println("[BATCH_TASK] Initializing...")
	e.items = []string{"task1", "task2", "task3"}
	return nil
}

func (e *BatchTaskExecutor) PreExecute() error {
	fmt.Println("[BATCH_TASK] Pre-executing...")
	return nil
}

func (e *BatchTaskExecutor) DoExecute() (interface{}, error) {
	fmt.Println("[BATCH_TASK] Executing batch tasks...")
	results := make([]string, len(e.items))
	for i, item := range e.items {
		results[i] = fmt.Sprintf("Processed: %s", item)
	}
	return results, nil
}

func (e *BatchTaskExecutor) PostExecute() error {
	fmt.Println("[BATCH_TASK] Post-executing...")
	return nil
}

func (e *BatchTaskExecutor) Cleanup() {
	fmt.Println("[BATCH_TASK] Cleaning up...")
	e.items = nil
}
