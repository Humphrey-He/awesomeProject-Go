package seata_transaction

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// ========== AT模式核心组件 ==========

// XID 全局事务ID
type XID string

// GlobalTransaction 全局事务
type GlobalTransaction struct {
	XID        XID
	Status     TransactionStatus
	BeginTime  time.Time
	Timeout    time.Duration
	BranchList []*BranchTransaction
}

// BranchTransaction 分支事务
type BranchTransaction struct {
	BranchID      int64
	XID           XID
	ResourceID    string
	LockKey       string
	Status        BranchStatus
	ApplicationID string
}

// TransactionStatus 事务状态
type TransactionStatus int

const (
	Begin TransactionStatus = iota
	Committing
	Committed
	Rollbacking
	Rollbacked
	Failed
)

// BranchStatus 分支状态
type BranchStatus int

const (
	Registered BranchStatus = iota
	PhaseOneDone
	PhaseTwoCommitted
	PhaseTwoRollbacked
)

// ========== TM（Transaction Manager）事务管理器 ==========

// TransactionManager 事务管理器
type TransactionManager struct {
	transactions map[XID]*GlobalTransaction
}

// NewTransactionManager 创建事务管理器
func NewTransactionManager() *TransactionManager {
	return &TransactionManager{
		transactions: make(map[XID]*GlobalTransaction),
	}
}

// Begin 开启全局事务
func (tm *TransactionManager) Begin(timeout time.Duration) (XID, error) {
	xid := generateXID()
	
	gt := &GlobalTransaction{
		XID:        xid,
		Status:     Begin,
		BeginTime:  time.Now(),
		Timeout:    timeout,
		BranchList: make([]*BranchTransaction, 0),
	}
	
	tm.transactions[xid] = gt
	
	fmt.Printf("[TM] Global transaction started: %s\n", xid)
	return xid, nil
}

// Commit 提交全局事务
func (tm *TransactionManager) Commit(xid XID) error {
	gt, exists := tm.transactions[xid]
	if !exists {
		return fmt.Errorf("global transaction not found: %s", xid)
	}
	
	gt.Status = Committing
	fmt.Printf("[TM] Committing global transaction: %s\n", xid)
	
	// 通知所有分支提交（异步）
	for _, branch := range gt.BranchList {
		if err := tm.commitBranch(branch); err != nil {
			fmt.Printf("[TM] Branch commit failed: %v\n", err)
			// AT模式的二阶段提交失败可以重试
		}
	}
	
	gt.Status = Committed
	fmt.Printf("[TM] Global transaction committed: %s\n", xid)
	
	// 清理事务
	delete(tm.transactions, xid)
	return nil
}

// Rollback 回滚全局事务
func (tm *TransactionManager) Rollback(xid XID) error {
	gt, exists := tm.transactions[xid]
	if !exists {
		return fmt.Errorf("global transaction not found: %s", xid)
	}
	
	gt.Status = Rollbacking
	fmt.Printf("[TM] Rolling back global transaction: %s\n", xid)
	
	// 通知所有分支回滚
	for _, branch := range gt.BranchList {
		if err := tm.rollbackBranch(branch); err != nil {
			fmt.Printf("[TM] Branch rollback failed: %v\n", err)
			return err
		}
	}
	
	gt.Status = Rollbacked
	fmt.Printf("[TM] Global transaction rolled back: %s\n", xid)
	
	delete(tm.transactions, xid)
	return nil
}

// RegisterBranch 注册分支事务
func (tm *TransactionManager) RegisterBranch(xid XID, resourceID, lockKey string) (*BranchTransaction, error) {
	gt, exists := tm.transactions[xid]
	if !exists {
		return nil, fmt.Errorf("global transaction not found: %s", xid)
	}
	
	branch := &BranchTransaction{
		BranchID:   generateBranchID(),
		XID:        xid,
		ResourceID: resourceID,
		LockKey:    lockKey,
		Status:     Registered,
	}
	
	gt.BranchList = append(gt.BranchList, branch)
	fmt.Printf("[TM] Branch registered: XID=%s, BranchID=%d\n", xid, branch.BranchID)
	
	return branch, nil
}

func (tm *TransactionManager) commitBranch(branch *BranchTransaction) error {
	// 二阶段提交：删除undo log
	fmt.Printf("[TM] Committing branch: %d\n", branch.BranchID)
	branch.Status = PhaseTwoCommitted
	return nil
}

func (tm *TransactionManager) rollbackBranch(branch *BranchTransaction) error {
	// 二阶段回滚：使用undo log恢复数据
	fmt.Printf("[TM] Rolling back branch: %d\n", branch.BranchID)
	branch.Status = PhaseTwoRollbacked
	return nil
}

// ========== RM（Resource Manager）资源管理器 ==========

// ResourceManager 资源管理器
type ResourceManager struct {
	db         *sql.DB
	resourceID string
}

// NewResourceManager 创建资源管理器
func NewResourceManager(db *sql.DB, resourceID string) *ResourceManager {
	return &ResourceManager{
		db:         db,
		resourceID: resourceID,
	}
}

// ExecuteWithUndoLog 执行SQL并生成undo log
func (rm *ResourceManager) ExecuteWithUndoLog(ctx context.Context, query string, args ...interface{}) error {
	xid := GetXIDFromContext(ctx)
	if xid == "" {
		return fmt.Errorf("XID not found in context")
	}
	
	tx, err := rm.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// 1. 查询前镜像（before image）
	beforeImage, err := rm.queryBeforeImage(tx, query, args...)
	if err != nil {
		return err
	}
	
	// 2. 执行业务SQL
	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	
	// 3. 查询后镜像（after image）
	afterImage, err := rm.queryAfterImage(tx, query, args...)
	if err != nil {
		return err
	}
	
	// 4. 生成并保存undo log
	undoLog := &UndoLog{
		BranchID:    generateBranchID(),
		XID:         string(xid),
		Context:     "{}",
		Rollback:    rm.generateRollbackSQL(beforeImage, afterImage),
		LogStatus:   0,
		LogCreated:  time.Now(),
		LogModified: time.Now(),
	}
	
	if err := rm.saveUndoLog(tx, undoLog); err != nil {
		return err
	}
	
	// 5. 提交本地事务（一阶段提交）
	if err := tx.Commit(); err != nil {
		return err
	}
	
	fmt.Printf("[RM] Phase one committed: XID=%s, Resource=%s\n", xid, rm.resourceID)
	return nil
}

// queryBeforeImage 查询前镜像
func (rm *ResourceManager) queryBeforeImage(tx *sql.Tx, query string, args ...interface{}) (map[string]interface{}, error) {
	// 简化实现：解析SQL获取WHERE条件，查询当前数据
	// 实际应该使用SQL解析器
	fmt.Println("[RM] Querying before image...")
	return map[string]interface{}{
		"id":    1,
		"stock": 100,
	}, nil
}

// queryAfterImage 查询后镜像
func (rm *ResourceManager) queryAfterImage(tx *sql.Tx, query string, args ...interface{}) (map[string]interface{}, error) {
	fmt.Println("[RM] Querying after image...")
	return map[string]interface{}{
		"id":    1,
		"stock": 90,
	}, nil
}

// generateRollbackSQL 生成回滚SQL
func (rm *ResourceManager) generateRollbackSQL(before, after map[string]interface{}) string {
	// 根据前后镜像生成回滚SQL
	// 例如：UPDATE stock SET stock=100 WHERE id=1
	return fmt.Sprintf("UPDATE stock SET stock=%v WHERE id=%v", before["stock"], before["id"])
}

// saveUndoLog 保存undo log
func (rm *ResourceManager) saveUndoLog(tx *sql.Tx, undoLog *UndoLog) error {
	rollbackJSON, _ := json.Marshal(undoLog.Rollback)
	
	query := `
		INSERT INTO undo_log (branch_id, xid, context, rollback_info, log_status, log_created, log_modified)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := tx.Exec(query,
		undoLog.BranchID,
		undoLog.XID,
		undoLog.Context,
		rollbackJSON,
		undoLog.LogStatus,
		undoLog.LogCreated,
		undoLog.LogModified,
	)
	
	if err == nil {
		fmt.Printf("[RM] Undo log saved: BranchID=%d\n", undoLog.BranchID)
	}
	
	return err
}

// CommitBranch 提交分支（删除undo log）
func (rm *ResourceManager) CommitBranch(branchID int64) error {
	query := "DELETE FROM undo_log WHERE branch_id = ?"
	_, err := rm.db.Exec(query, branchID)
	
	if err == nil {
		fmt.Printf("[RM] Undo log deleted: BranchID=%d\n", branchID)
	}
	
	return err
}

// RollbackBranch 回滚分支（执行undo log）
func (rm *ResourceManager) RollbackBranch(branchID int64) error {
	// 1. 查询undo log
	undoLog, err := rm.queryUndoLog(branchID)
	if err != nil {
		return err
	}
	
	// 2. 执行回滚SQL
	fmt.Printf("[RM] Executing rollback SQL: %s\n", undoLog.Rollback)
	_, err = rm.db.Exec(undoLog.Rollback)
	if err != nil {
		return err
	}
	
	// 3. 删除undo log
	return rm.CommitBranch(branchID)
}

func (rm *ResourceManager) queryUndoLog(branchID int64) (*UndoLog, error) {
	query := "SELECT branch_id, xid, rollback_info FROM undo_log WHERE branch_id = ?"
	
	var undoLog UndoLog
	var rollbackJSON string
	
	err := rm.db.QueryRow(query, branchID).Scan(
		&undoLog.BranchID,
		&undoLog.XID,
		&rollbackJSON,
	)
	
	if err != nil {
		return nil, err
	}
	
	// 解析回滚信息
	undoLog.Rollback = rollbackJSON
	
	return &undoLog, nil
}

// ========== UndoLog 数据结构 ==========

// UndoLog undo日志
type UndoLog struct {
	ID          int64
	BranchID    int64
	XID         string
	Context     string
	Rollback    string
	LogStatus   int
	LogCreated  time.Time
	LogModified time.Time
}

// UndoLogTableDDL undo_log表结构
const UndoLogTableDDL = `
CREATE TABLE IF NOT EXISTS undo_log (
    id BIGINT NOT NULL AUTO_INCREMENT COMMENT 'increment id',
    branch_id BIGINT NOT NULL COMMENT 'branch transaction id',
    xid VARCHAR(100) NOT NULL COMMENT 'global transaction id',
    context VARCHAR(128) NOT NULL COMMENT 'undo_log context,such as serialization',
    rollback_info LONGBLOB NOT NULL COMMENT 'rollback info',
    log_status INT(11) NOT NULL COMMENT '0:normal status,1:defense status',
    log_created DATETIME NOT NULL COMMENT 'create datetime',
    log_modified DATETIME NOT NULL COMMENT 'modify datetime',
    PRIMARY KEY (id),
    UNIQUE KEY ux_undo_log (xid, branch_id),
    KEY idx_log_created (log_created)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='AT transaction mode undo table';
`

// ========== Context传递XID ==========

type contextKey string

const xidKey contextKey = "seata_xid"

// WithXID 将XID放入context
func WithXID(ctx context.Context, xid XID) context.Context {
	return context.WithValue(ctx, xidKey, xid)
}

// GetXIDFromContext 从context获取XID
func GetXIDFromContext(ctx context.Context) XID {
	xid, ok := ctx.Value(xidKey).(XID)
	if !ok {
		return ""
	}
	return xid
}

// ========== 辅助函数 ==========

var (
	xidCounter    int64
	branchCounter int64
)

func generateXID() XID {
	xidCounter++
	return XID(fmt.Sprintf("xid-%d-%d", time.Now().Unix(), xidCounter))
}

func generateBranchID() int64 {
	branchCounter++
	return branchCounter
}

// ========== 完整示例：订单-库存 ==========

// OrderService 订单服务（TM发起者）
type OrderService struct {
	tm          *TransactionManager
	stockClient *StockClient
	db          *sql.DB
}

// CreateOrder 创建订单（全局事务发起）
func (s *OrderService) CreateOrder(ctx context.Context, productID int64, quantity int) error {
	// 1. 开启全局事务
	xid, err := s.tm.Begin(30 * time.Second)
	if err != nil {
		return err
	}
	
	// 将XID放入context
	ctx = WithXID(ctx, xid)
	
	defer func() {
		if r := recover(); r != nil {
			// 发生panic，回滚事务
			s.tm.Rollback(xid)
			panic(r)
		}
	}()
	
	// 2. 执行本地订单业务
	err = s.executeLocalOrderBusiness(ctx, productID, quantity)
	if err != nil {
		s.tm.Rollback(xid)
		return fmt.Errorf("create order failed: %w", err)
	}
	
	// 3. 调用库存服务扣减库存（传递XID）
	err = s.stockClient.DeductStock(ctx, productID, quantity)
	if err != nil {
		// 库存扣减失败，触发全局回滚
		fmt.Printf("[OrderService] Stock deduction failed, rolling back: %v\n", err)
		s.tm.Rollback(xid)
		return fmt.Errorf("deduct stock failed: %w", err)
	}
	
	// 4. 提交全局事务
	err = s.tm.Commit(xid)
	if err != nil {
		return fmt.Errorf("commit transaction failed: %w", err)
	}
	
	fmt.Println("[OrderService] Order created successfully")
	return nil
}

func (s *OrderService) executeLocalOrderBusiness(ctx context.Context, productID int64, quantity int) error {
	rm := NewResourceManager(s.db, "order_db")
	
	query := "INSERT INTO orders (product_id, quantity, status) VALUES (?, ?, 'PENDING')"
	return rm.ExecuteWithUndoLog(ctx, query, productID, quantity)
}

// StockClient 库存服务客户端
type StockClient struct {
	baseURL string
}

func (c *StockClient) DeductStock(ctx context.Context, productID int64, quantity int) error {
	xid := GetXIDFromContext(ctx)
	
	// 实际应该通过HTTP Header传递XID
	fmt.Printf("[StockClient] Calling stock service with XID: %s\n", xid)
	
	// 模拟RPC调用
	// req.Header.Set("TX_XID", string(xid))
	
	return nil
}

// StockService 库存服务（RM参与者）
type StockService struct {
	db *sql.DB
	rm *ResourceManager
}

// DeductStock 扣减库存（分支事务）
func (s *StockService) DeductStock(ctx context.Context, productID int64, quantity int) error {
	// 从context或HTTP Header解析XID
	xid := GetXIDFromContext(ctx)
	if xid == "" {
		return fmt.Errorf("XID not found")
	}
	
	fmt.Printf("[StockService] Processing stock deduction: XID=%s\n", xid)
	
	// 执行库存扣减（自动生成undo log）
	query := "UPDATE stock SET quantity = quantity - ? WHERE product_id = ? AND quantity >= ?"
	err := s.rm.ExecuteWithUndoLog(ctx, query, quantity, productID, quantity)
	
	if err != nil {
		return fmt.Errorf("insufficient stock or update failed: %w", err)
	}
	
	fmt.Println("[StockService] Stock deducted successfully")
	return nil
}

// ========== HTTP中间件：解析XID ==========

// SeataMiddleware Seata中间件
func SeataMiddleware(next func(ctx context.Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		// 从HTTP Header解析XID
		// xid := r.Header.Get("TX_XID")
		// if xid != "" {
		//     ctx = WithXID(ctx, XID(xid))
		// }
		
		return next(ctx)
	}
}

