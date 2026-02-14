package seata_transaction

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// ========== TCC模式核心接口 ==========

// TCCResource TCC资源接口
type TCCResource interface {
	// Prepare 一阶段：尝试执行，预留资源
	Prepare(ctx context.Context, businessKey string, params map[string]interface{}) error
	
	// Commit 二阶段提交：确认执行
	Commit(ctx context.Context, businessKey string) error
	
	// Cancel 二阶段取消：回滚操作
	Cancel(ctx context.Context, businessKey string) error
	
	// GetResourceID 获取资源ID
	GetResourceID() string
}

// ========== TCC事务状态管理 ==========

// TCCTransactionStatus TCC事务状态
type TCCTransactionStatus int

const (
	TCCTrying TCCTransactionStatus = iota
	TCCCommitted
	TCCCancelled
)

// TCCTransaction TCC事务记录
type TCCTransaction struct {
	XID          string
	BusinessKey  string
	Status       TCCTransactionStatus
	CreateTime   time.Time
	UpdateTime   time.Time
	RetryCount   int
	IsDeleted    bool
}

// ========== TCC资源管理器 ==========

// TCCResourceManager TCC资源管理器
type TCCResourceManager struct {
	resources map[string]TCCResource
	mu        sync.RWMutex
}

// NewTCCResourceManager 创建TCC资源管理器
func NewTCCResourceManager() *TCCResourceManager {
	return &TCCResourceManager{
		resources: make(map[string]TCCResource),
	}
}

// RegisterResource 注册TCC资源
func (trm *TCCResourceManager) RegisterResource(resource TCCResource) {
	trm.mu.Lock()
	defer trm.mu.Unlock()
	
	resourceID := resource.GetResourceID()
	trm.resources[resourceID] = resource
	
	fmt.Printf("[TCC RM] Resource registered: %s\n", resourceID)
}

// GetResource 获取TCC资源
func (trm *TCCResourceManager) GetResource(resourceID string) (TCCResource, error) {
	trm.mu.RLock()
	defer trm.mu.RUnlock()
	
	resource, exists := trm.resources[resourceID]
	if !exists {
		return nil, fmt.Errorf("TCC resource not found: %s", resourceID)
	}
	
	return resource, nil
}

// ========== 账户扣费TCC实现示例 ==========

// AccountTCCService 账户TCC服务（高并发扣费场景）
type AccountTCCService struct {
	db         *sql.DB
	resourceID string
	mu         sync.RWMutex
	// 内存中记录TCC状态，防止重复调用
	tccStates  map[string]*TCCState
}

// TCCState TCC状态记录
type TCCState struct {
	BusinessKey string
	Status      TCCTransactionStatus
	PrepareTime time.Time
	CommitTime  time.Time
	CancelTime  time.Time
}

// NewAccountTCCService 创建账户TCC服务
func NewAccountTCCService(db *sql.DB, resourceID string) *AccountTCCService {
	return &AccountTCCService{
		db:         db,
		resourceID: resourceID,
		tccStates:  make(map[string]*TCCState),
	}
}

// Prepare 一阶段：冻结金额
func (s *AccountTCCService) Prepare(ctx context.Context, businessKey string, params map[string]interface{}) error {
	// 幂等性检查
	if s.isAlreadyPrepared(businessKey) {
		fmt.Printf("[TCC] Prepare already executed: %s\n", businessKey)
		return nil
	}
	
	userID := params["user_id"].(int64)
	amount := params["amount"].(float64)
	
	fmt.Printf("[TCC Prepare] Freezing amount: user=%d, amount=%.2f, key=%s\n", 
		userID, amount, businessKey)
	
	// 检查数据库连接
	if s.db == nil {
		// 无数据库连接时，只记录状态（用于测试）
		s.mu.Lock()
		s.tccStates[businessKey] = &TCCState{
			BusinessKey: businessKey,
			Status:      TCCTrying,
			PrepareTime: time.Now(),
		}
		s.mu.Unlock()
		fmt.Printf("[TCC Prepare] Amount frozen successfully (no DB): %s\n", businessKey)
		return nil
	}
	
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// 1. 检查余额是否充足
	var balance float64
	err = tx.QueryRow("SELECT balance FROM account WHERE user_id = ? FOR UPDATE", userID).Scan(&balance)
	if err != nil {
		return fmt.Errorf("query balance failed: %w", err)
	}
	
	if balance < amount {
		return fmt.Errorf("insufficient balance: have=%.2f, need=%.2f", balance, amount)
	}
	
	// 2. 冻结金额（不真正扣减，只是标记冻结）
	_, err = tx.Exec(`
		UPDATE account 
		SET frozen_amount = frozen_amount + ? 
		WHERE user_id = ? AND balance >= ?`,
		amount, userID, amount)
	
	if err != nil {
		return fmt.Errorf("freeze amount failed: %w", err)
	}
	
	// 3. 记录TCC事务状态
	_, err = tx.Exec(`
		INSERT INTO tcc_transaction (xid, business_key, status, user_id, amount, create_time, update_time)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())`,
		GetXIDFromContext(ctx), businessKey, TCCTrying, userID, amount)
	
	if err != nil {
		return fmt.Errorf("save tcc transaction failed: %w", err)
	}
	
	if err := tx.Commit(); err != nil {
		return err
	}
	
	// 4. 记录内存状态（用于幂等性）
	s.mu.Lock()
	s.tccStates[businessKey] = &TCCState{
		BusinessKey: businessKey,
		Status:      TCCTrying,
		PrepareTime: time.Now(),
	}
	s.mu.Unlock()
	
	fmt.Printf("[TCC Prepare] Amount frozen successfully: %s\n", businessKey)
	return nil
}

// Commit 二阶段提交：正式扣减金额
func (s *AccountTCCService) Commit(ctx context.Context, businessKey string) error {
	// 幂等性检查
	if s.isAlreadyCommitted(businessKey) {
		fmt.Printf("[TCC] Commit already executed: %s\n", businessKey)
		return nil
	}
	
	// 防止悬挂：如果Prepare未执行，直接返回
	if !s.isAlreadyPrepared(businessKey) {
		fmt.Printf("[TCC] Prepare not executed, skip commit: %s\n", businessKey)
		return nil
	}
	
	fmt.Printf("[TCC Commit] Confirming deduction: %s\n", businessKey)
	
	// 检查数据库连接
	if s.db == nil {
		// 无数据库连接时，只更新状态（用于测试）
		s.mu.Lock()
		if state, ok := s.tccStates[businessKey]; ok {
			state.Status = TCCCommitted
			state.CommitTime = time.Now()
		}
		s.mu.Unlock()
		fmt.Printf("[TCC Commit] Amount deducted successfully (no DB): %s\n", businessKey)
		return nil
	}
	
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// 1. 查询TCC事务信息
	var userID int64
	var amount float64
	err = tx.QueryRow(`
		SELECT user_id, amount FROM tcc_transaction 
		WHERE business_key = ? AND status = ?`,
		businessKey, TCCTrying).Scan(&userID, &amount)
	
	if err == sql.ErrNoRows {
		// 事务不存在或已处理，幂等返回
		return nil
	}
	if err != nil {
		return err
	}
	
	// 2. 正式扣减金额并解冻
	_, err = tx.Exec(`
		UPDATE account 
		SET balance = balance - ?,
			frozen_amount = frozen_amount - ?
		WHERE user_id = ?`,
		amount, amount, userID)
	
	if err != nil {
		return fmt.Errorf("deduct amount failed: %w", err)
	}
	
	// 3. 更新TCC事务状态为已提交
	_, err = tx.Exec(`
		UPDATE tcc_transaction 
		SET status = ?, update_time = NOW()
		WHERE business_key = ?`,
		TCCCommitted, businessKey)
	
	if err != nil {
		return err
	}
	
	if err := tx.Commit(); err != nil {
		return err
	}
	
	// 4. 更新内存状态
	s.mu.Lock()
	if state, ok := s.tccStates[businessKey]; ok {
		state.Status = TCCCommitted
		state.CommitTime = time.Now()
	}
	s.mu.Unlock()
	
	fmt.Printf("[TCC Commit] Amount deducted successfully: %s\n", businessKey)
	return nil
}

// Cancel 二阶段取消：解冻金额
func (s *AccountTCCService) Cancel(ctx context.Context, businessKey string) error {
	// 幂等性检查
	if s.isAlreadyCancelled(businessKey) {
		fmt.Printf("[TCC] Cancel already executed: %s\n", businessKey)
		return nil
	}
	
	// 防止悬挂：如果Prepare未执行，直接返回
	if !s.isAlreadyPrepared(businessKey) {
		fmt.Printf("[TCC] Prepare not executed, skip cancel: %s\n", businessKey)
		return nil
	}
	
	fmt.Printf("[TCC Cancel] Rolling back: %s\n", businessKey)
	
	// 检查数据库连接
	if s.db == nil {
		// 无数据库连接时，只更新状态（用于测试）
		s.mu.Lock()
		if state, ok := s.tccStates[businessKey]; ok {
			state.Status = TCCCancelled
			state.CancelTime = time.Now()
		}
		s.mu.Unlock()
		fmt.Printf("[TCC Cancel] Amount unfrozen successfully (no DB): %s\n", businessKey)
		return nil
	}
	
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// 1. 查询TCC事务信息
	var userID int64
	var amount float64
	err = tx.QueryRow(`
		SELECT user_id, amount FROM tcc_transaction 
		WHERE business_key = ? AND status = ?`,
		businessKey, TCCTrying).Scan(&userID, &amount)
	
	if err == sql.ErrNoRows {
		// 事务不存在或已处理
		return nil
	}
	if err != nil {
		return err
	}
	
	// 2. 解冻金额（恢复可用余额）
	_, err = tx.Exec(`
		UPDATE account 
		SET frozen_amount = frozen_amount - ?
		WHERE user_id = ?`,
		amount, userID)
	
	if err != nil {
		return fmt.Errorf("unfreeze amount failed: %w", err)
	}
	
	// 3. 更新TCC事务状态为已取消
	_, err = tx.Exec(`
		UPDATE tcc_transaction 
		SET status = ?, update_time = NOW()
		WHERE business_key = ?`,
		TCCCancelled, businessKey)
	
	if err != nil {
		return err
	}
	
	if err := tx.Commit(); err != nil {
		return err
	}
	
	// 4. 更新内存状态
	s.mu.Lock()
	if state, ok := s.tccStates[businessKey]; ok {
		state.Status = TCCCancelled
		state.CancelTime = time.Now()
	}
	s.mu.Unlock()
	
	fmt.Printf("[TCC Cancel] Amount unfrozen successfully: %s\n", businessKey)
	return nil
}

// GetResourceID 获取资源ID
func (s *AccountTCCService) GetResourceID() string {
	return s.resourceID
}

// ========== 幂等性检查（防止重复调用）==========

func (s *AccountTCCService) isAlreadyPrepared(businessKey string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	state, exists := s.tccStates[businessKey]
	return exists && state.Status >= TCCTrying
}

func (s *AccountTCCService) isAlreadyCommitted(businessKey string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	state, exists := s.tccStates[businessKey]
	return exists && state.Status == TCCCommitted
}

func (s *AccountTCCService) isAlreadyCancelled(businessKey string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	state, exists := s.tccStates[businessKey]
	return exists && state.Status == TCCCancelled
}

// ========== TCC数据库表结构 ==========

// TCCTableDDL TCC相关表结构
const TCCTableDDL = `
-- 账户表（包含冻结金额字段）
CREATE TABLE IF NOT EXISTS account (
    user_id BIGINT PRIMARY KEY,
    balance DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '可用余额',
    frozen_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '冻结金额',
    create_time DATETIME NOT NULL,
    update_time DATETIME NOT NULL,
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='账户表';

-- TCC事务记录表
CREATE TABLE IF NOT EXISTS tcc_transaction (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    xid VARCHAR(100) NOT NULL COMMENT '全局事务ID',
    business_key VARCHAR(100) NOT NULL COMMENT '业务主键（幂等性）',
    status INT NOT NULL COMMENT '状态：0-Trying, 1-Committed, 2-Cancelled',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    amount DECIMAL(15,2) NOT NULL COMMENT '金额',
    create_time DATETIME NOT NULL COMMENT '创建时间',
    update_time DATETIME NOT NULL COMMENT '更新时间',
    retry_count INT NOT NULL DEFAULT 0 COMMENT '重试次数',
    UNIQUE KEY uk_business_key (business_key),
    INDEX idx_xid (xid),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='TCC事务记录表';
`

// ========== TCC完整业务示例：订单扣费 ==========

// OrderPaymentService 订单支付服务（使用TCC）
type OrderPaymentService struct {
	tm              *TransactionManager
	accountTCC      *AccountTCCService
	tccRM           *TCCResourceManager
}

// NewOrderPaymentService 创建订单支付服务
func NewOrderPaymentService(tm *TransactionManager, accountTCC *AccountTCCService) *OrderPaymentService {
	tccRM := NewTCCResourceManager()
	tccRM.RegisterResource(accountTCC)
	
	return &OrderPaymentService{
		tm:         tm,
		accountTCC: accountTCC,
		tccRM:      tccRM,
	}
}

// PayOrder 支付订单（TCC全局事务）
func (s *OrderPaymentService) PayOrder(ctx context.Context, orderID, userID int64, amount float64) error {
	// 1. 开启全局事务
	xid, err := s.tm.Begin(30 * time.Second)
	if err != nil {
		return err
	}
	
	ctx = WithXID(ctx, xid)
	businessKey := fmt.Sprintf("order_pay_%d_%d", orderID, time.Now().UnixNano())
	
	fmt.Printf("[OrderPayment] Starting TCC transaction: XID=%s\n", xid)
	
	defer func() {
		if r := recover(); r != nil {
			s.tm.Rollback(xid)
			panic(r)
		}
	}()
	
	// 2. 调用TCC的Prepare阶段（冻结金额）
	params := map[string]interface{}{
		"user_id": userID,
		"amount":  amount,
	}
	
	err = s.accountTCC.Prepare(ctx, businessKey, params)
	if err != nil {
		// Prepare失败，回滚全局事务
		fmt.Printf("[OrderPayment] Prepare failed, rolling back: %v\n", err)
		s.executeTCCCancel(ctx, businessKey)
		s.tm.Rollback(xid)
		return fmt.Errorf("payment failed: %w", err)
	}
	
	// 3. 执行其他业务逻辑...
	// 例如：创建支付记录、通知订单服务等
	
	// 4. 提交全局事务（触发TCC的Commit阶段）
	err = s.executeTCCCommit(ctx, businessKey)
	if err != nil {
		// Commit失败，尝试回滚
		fmt.Printf("[OrderPayment] Commit failed: %v\n", err)
		s.executeTCCCancel(ctx, businessKey)
		s.tm.Rollback(xid)
		return err
	}
	
	s.tm.Commit(xid)
	
	fmt.Printf("[OrderPayment] Payment completed successfully: order=%d, amount=%.2f\n", 
		orderID, amount)
	return nil
}

func (s *OrderPaymentService) executeTCCCommit(ctx context.Context, businessKey string) error {
	return s.accountTCC.Commit(ctx, businessKey)
}

func (s *OrderPaymentService) executeTCCCancel(ctx context.Context, businessKey string) error {
	return s.accountTCC.Cancel(ctx, businessKey)
}

// ========== TCC vs AT 对比 ==========

/*
AT模式特点：
- 自动补偿：通过undo log自动回滚
- 对业务侵入小：只需要代理数据源
- 适合简单CRUD场景
- 性能较低：需要生成前后镜像

TCC模式特点：
- 手动补偿：需要实现Try/Confirm/Cancel三个方法
- 对业务侵入大：需要改造业务逻辑
- 适合高并发、复杂业务场景
- 性能较高：没有额外的镜像开销
- 需要处理幂等性和悬挂问题

选择建议：
- 普通CRUD业务：使用AT模式
- 高并发扣费/库存场景：使用TCC模式
- 需要精细资源控制：使用TCC模式
- 快速开发：使用AT模式
*/

// ========== TCC异常场景处理 ==========

// 1. 空回滚（Cancel在Prepare之前执行）
// 解决：在Cancel中检查Prepare是否执行过

// 2. 悬挂（Prepare在Cancel之后执行）
// 解决：在Prepare中检查是否已经Cancel过

// 3. 幂等性（Commit/Cancel被重复调用）
// 解决：记录状态，重复调用直接返回成功

// TCCAntiHang TCC防悬挂检查
func (s *AccountTCCService) checkAntiHang(businessKey string) error {
	// 查询是否已经Cancel
	var status int
	err := s.db.QueryRow(`
		SELECT status FROM tcc_transaction 
		WHERE business_key = ?`, businessKey).Scan(&status)
	
	if err == sql.ErrNoRows {
		return nil // 未执行过，可以继续
	}
	
	if err != nil {
		return err
	}
	
	if status == int(TCCCancelled) {
		return fmt.Errorf("transaction already cancelled (hang prevention)")
	}
	
	return nil
}

