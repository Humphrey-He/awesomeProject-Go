# Seata 分布式事务实现（Go版）

## 项目简介

本项目实现了Seata分布式事务框架的两种核心模式：
- **AT模式**（Automatic Transaction）：自动补偿模式
- **TCC模式**（Try-Confirm-Cancel）：手动补偿模式

适用于微服务架构下的跨服务事务一致性保障。

---

## AT模式（自动补偿）

### 核心原理

AT模式通过**拦截SQL**生成**前后镜像**和**undo log**，在回滚时自动恢复数据。

#### 工作流程

```
┌─────────────┐      ┌─────────────┐      ┌─────────────┐
│ TM (发起者)  │      │ RM1 (订单)   │      │ RM2 (库存)   │
└──────┬──────┘      └──────┬──────┘      └──────┬──────┘
       │                    │                    │
       │ 1. Begin(xid)      │                    │
       ├───────────────────>│                    │
       │                    │                    │
       │ 2. 执行SQL + 生成undo_log               │
       │                    ├───────────────────>│
       │                    │                    │
       │ 3. 一阶段提交       │                    │
       │<───────────────────┤<───────────────────┤
       │                    │                    │
       │ 4. Commit(xid)     │                    │
       │ 或 Rollback(xid)   │                    │
       │                    │                    │
       │ 5. 删除undo_log    │                    │
       │    或执行回滚SQL    │                    │
       └────────────────────┴────────────────────┘
```

### 核心组件

#### 1. TM（Transaction Manager）- 事务管理器

```go
tm := NewTransactionManager()

// 开启全局事务
xid, err := tm.Begin(30 * time.Second)

// 将XID放入Context传递给下游
ctx := WithXID(context.Background(), xid)

// 提交或回滚
tm.Commit(xid)
tm.Rollback(xid)
```

#### 2. RM（Resource Manager）- 资源管理器

```go
rm := NewResourceManager(db, "order_db")

// 执行SQL并自动生成undo log
err := rm.ExecuteWithUndoLog(ctx, 
    "UPDATE stock SET quantity = quantity - ? WHERE product_id = ?", 
    10, 100)
```

### 订单-库存完整示例

#### 订单服务（TM发起者）

```go
type OrderService struct {
    tm          *TransactionManager
    stockClient *StockClient
    db          *sql.DB
}

func (s *OrderService) CreateOrder(ctx context.Context, productID int64, quantity int) error {
    // 1. 开启全局事务
    xid, err := s.tm.Begin(30 * time.Second)
    if err != nil {
        return err
    }
    
    // 2. 将XID传递给下游
    ctx = WithXID(ctx, xid)
    
    defer func() {
        if r := recover(); r != nil {
            s.tm.Rollback(xid)
            panic(r)
        }
    }()
    
    // 3. 执行本地订单业务
    rm := NewResourceManager(s.db, "order_db")
    err = rm.ExecuteWithUndoLog(ctx, 
        "INSERT INTO orders (product_id, quantity) VALUES (?, ?)", 
        productID, quantity)
    if err != nil {
        s.tm.Rollback(xid)
        return err
    }
    
    // 4. 调用库存服务扣减库存
    err = s.stockClient.DeductStock(ctx, productID, quantity)
    if err != nil {
        // 库存扣减失败，触发全局回滚
        s.tm.Rollback(xid)
        return err
    }
    
    // 5. 提交全局事务
    return s.tm.Commit(xid)
}
```

#### 库存服务（RM参与者）

```go
type StockService struct {
    db *sql.DB
    rm *ResourceManager
}

func (s *StockService) DeductStock(ctx context.Context, productID int64, quantity int) error {
    // 1. 从Context或HTTP Header解析XID
    xid := GetXIDFromContext(ctx)
    if xid == "" {
        return fmt.Errorf("XID not found")
    }
    
    // 2. 执行库存扣减（自动生成undo log）
    query := "UPDATE stock SET quantity = quantity - ? WHERE product_id = ? AND quantity >= ?"
    return s.rm.ExecuteWithUndoLog(ctx, query, quantity, productID, quantity)
}
```

### Undo Log表结构

```sql
CREATE TABLE IF NOT EXISTS undo_log (
    id BIGINT NOT NULL AUTO_INCREMENT COMMENT 'increment id',
    branch_id BIGINT NOT NULL COMMENT 'branch transaction id',
    xid VARCHAR(100) NOT NULL COMMENT 'global transaction id',
    context VARCHAR(128) NOT NULL COMMENT 'undo_log context',
    rollback_info LONGBLOB NOT NULL COMMENT 'rollback info',
    log_status INT(11) NOT NULL COMMENT '0:normal, 1:defense',
    log_created DATETIME NOT NULL COMMENT 'create datetime',
    log_modified DATETIME NOT NULL COMMENT 'modify datetime',
    PRIMARY KEY (id),
    UNIQUE KEY ux_undo_log (xid, branch_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
```

### 异常处理

当任何分支事务失败时，TM会触发全局回滚：

1. 通知所有RM执行回滚
2. RM读取undo log中的回滚SQL
3. 执行回滚SQL恢复数据
4. 删除undo log

---

## TCC模式（手动补偿）

### 核心原理

TCC通过业务层面的**三阶段提交**实现分布式事务：
- **Try**：尝试执行，预留资源（冻结）
- **Confirm**：确认执行，正式提交
- **Cancel**：取消执行，释放资源

### 工作流程

```
┌─────────────┐
│ TCC Resource│
└──────┬──────┘
       │
       │ Phase 1: Try (冻结资源)
       ├────────────────────────────┐
       │                            │
       │ balance: 1000              │
       │ frozen: 0                  │
       │                            │
       │ Try(100元)                 │
       │ balance: 1000              │
       │ frozen: 100  ◄─────────────┘
       │
       │ Phase 2a: Confirm (正式扣减)
       ├────────────────────────────┐
       │                            │
       │ balance: 900               │
       │ frozen: 0    ◄─────────────┘
       │
       │ 或 Phase 2b: Cancel (解冻回滚)
       ├────────────────────────────┐
       │                            │
       │ balance: 1000              │
       │ frozen: 0    ◄─────────────┘
       │
       └────────────────────────────
```

### TCC接口定义

```go
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
```

### 账户扣费TCC实现

```go
type AccountTCCService struct {
    db         *sql.DB
    resourceID string
}

// Prepare 冻结金额
func (s *AccountTCCService) Prepare(ctx context.Context, businessKey string, params map[string]interface{}) error {
    // 幂等性检查
    if s.isAlreadyPrepared(businessKey) {
        return nil
    }
    
    userID := params["user_id"].(int64)
    amount := params["amount"].(float64)
    
    tx, _ := s.db.BeginTx(ctx, nil)
    defer tx.Rollback()
    
    // 1. 检查余额
    var balance float64
    tx.QueryRow("SELECT balance FROM account WHERE user_id = ?", userID).Scan(&balance)
    
    if balance < amount {
        return fmt.Errorf("insufficient balance")
    }
    
    // 2. 冻结金额
    _, err := tx.Exec(`
        UPDATE account 
        SET frozen_amount = frozen_amount + ? 
        WHERE user_id = ?`, amount, userID)
    
    // 3. 记录TCC事务状态
    tx.Exec(`
        INSERT INTO tcc_transaction (xid, business_key, status, user_id, amount, create_time)
        VALUES (?, ?, ?, ?, ?, NOW())`,
        GetXIDFromContext(ctx), businessKey, TCCTrying, userID, amount)
    
    return tx.Commit()
}

// Commit 正式扣减
func (s *AccountTCCService) Commit(ctx context.Context, businessKey string) error {
    // 幂等性检查
    if s.isAlreadyCommitted(businessKey) {
        return nil
    }
    
    // 防止悬挂
    if !s.isAlreadyPrepared(businessKey) {
        return nil
    }
    
    tx, _ := s.db.BeginTx(ctx, nil)
    defer tx.Rollback()
    
    // 1. 查询事务信息
    var userID int64
    var amount float64
    tx.QueryRow(`
        SELECT user_id, amount FROM tcc_transaction 
        WHERE business_key = ?`, businessKey).Scan(&userID, &amount)
    
    // 2. 正式扣减并解冻
    tx.Exec(`
        UPDATE account 
        SET balance = balance - ?,
            frozen_amount = frozen_amount - ?
        WHERE user_id = ?`, amount, amount, userID)
    
    // 3. 更新状态
    tx.Exec(`
        UPDATE tcc_transaction 
        SET status = ?, update_time = NOW()
        WHERE business_key = ?`, TCCCommitted, businessKey)
    
    return tx.Commit()
}

// Cancel 解冻回滚
func (s *AccountTCCService) Cancel(ctx context.Context, businessKey string) error {
    // 幂等性检查
    if s.isAlreadyCancelled(businessKey) {
        return nil
    }
    
    // 空回滚检查
    if !s.isAlreadyPrepared(businessKey) {
        return nil
    }
    
    tx, _ := s.db.BeginTx(ctx, nil)
    defer tx.Rollback()
    
    // 1. 查询事务信息
    var userID int64
    var amount float64
    tx.QueryRow(`
        SELECT user_id, amount FROM tcc_transaction 
        WHERE business_key = ?`, businessKey).Scan(&userID, &amount)
    
    // 2. 解冻金额
    tx.Exec(`
        UPDATE account 
        SET frozen_amount = frozen_amount - ?
        WHERE user_id = ?`, amount, userID)
    
    // 3. 更新状态
    tx.Exec(`
        UPDATE tcc_transaction 
        SET status = ?, update_time = NOW()
        WHERE business_key = ?`, TCCCancelled, businessKey)
    
    return tx.Commit()
}
```

### TCC数据库表结构

```sql
-- 账户表（包含冻结金额字段）
CREATE TABLE IF NOT EXISTS account (
    user_id BIGINT PRIMARY KEY,
    balance DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '可用余额',
    frozen_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '冻结金额',
    create_time DATETIME NOT NULL,
    update_time DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- TCC事务记录表
CREATE TABLE IF NOT EXISTS tcc_transaction (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    xid VARCHAR(100) NOT NULL COMMENT '全局事务ID',
    business_key VARCHAR(100) NOT NULL COMMENT '业务主键（幂等性）',
    status INT NOT NULL COMMENT '状态：0-Trying, 1-Committed, 2-Cancelled',
    user_id BIGINT NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    create_time DATETIME NOT NULL,
    update_time DATETIME NOT NULL,
    UNIQUE KEY uk_business_key (business_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### TCC关键问题处理

#### 1. 幂等性

通过`business_key`和内存状态双重检查，防止重复执行：

```go
func (s *AccountTCCService) isAlreadyCommitted(businessKey string) bool {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    state, exists := s.tccStates[businessKey]
    return exists && state.Status == TCCCommitted
}
```

#### 2. 空回滚

Cancel在Prepare之前执行：

```go
func (s *AccountTCCService) Cancel(ctx context.Context, businessKey string) error {
    // 如果Prepare未执行，直接返回成功
    if !s.isAlreadyPrepared(businessKey) {
        return nil
    }
    // ... 执行回滚逻辑
}
```

#### 3. 悬挂

Prepare在Cancel之后执行：

```go
func (s *AccountTCCService) Prepare(ctx context.Context, businessKey string, params map[string]interface{}) error {
    // 检查是否已经Cancel
    if s.isAlreadyCancelled(businessKey) {
        return fmt.Errorf("transaction already cancelled")
    }
    // ... 执行冻结逻辑
}
```

### TCC注册和使用

```go
// 1. 创建TCC资源管理器
tccRM := NewTCCResourceManager()

// 2. 注册TCC资源
accountTCC := NewAccountTCCService(db, "account_service")
tccRM.RegisterResource(accountTCC)

// 3. 使用TCC执行事务
paymentService := NewOrderPaymentService(tm, accountTCC)
err := paymentService.PayOrder(ctx, orderID, userID, amount)
```

---

## AT vs TCC 对比

| 特性 | AT模式 | TCC模式 |
|------|--------|---------|
| **补偿方式** | 自动（undo log） | 手动（Try/Confirm/Cancel） |
| **业务侵入** | 低（只需代理数据源） | 高（需改造业务逻辑） |
| **性能** | 较低（生成镜像开销） | 高（无额外镜像） |
| **适用场景** | 简单CRUD | 高并发、精细资源控制 |
| **实现难度** | 简单 | 复杂（需处理幂等、悬挂） |
| **数据一致性** | 最终一致性 | 最终一致性 |
| **隔离性** | 读未提交 | 读已提交（冻结） |
| **并发控制** | 依赖数据库锁 | 业务层控制 |

### 选择建议

- **使用AT模式**：
  - 普通CRUD业务
  - 快速开发需求
  - 对性能要求不高
  
- **使用TCC模式**：
  - 高并发扣费场景
  - 库存扣减场景
  - 需要精细资源控制
  - 对性能要求高

---

## 运行测试

```bash
# 运行所有测试
go test -v ./seata_transaction/

# 运行基准测试
go test -bench=. ./seata_transaction/

# 运行集成测试（需要MySQL）
go test -v ./seata_transaction/ -run TestIntegration
```

### 测试输出示例

```
========== Test AT Mode: Success Commit ==========
[TM] Global transaction started: xid-1707123456-1
Branch registered: ID=1
[TM] Committing global transaction: xid-1707123456-1
[TM] Global transaction committed: xid-1707123456-1
✓ Test passed: Transaction committed successfully

========== Test TCC Mode: Prepare -> Commit ==========
[TCC Prepare] Freezing amount: user=1001, amount=100.50, key=test_pay_001
[TCC Prepare] Amount frozen successfully: test_pay_001
✓ Prepare phase completed
[TCC Commit] Confirming deduction: test_pay_001
[TCC Commit] Amount deducted successfully: test_pay_001
✓ Commit phase completed
✓ Test passed: TCC Prepare-Commit flow
```

---

## 实际应用场景

### 1. 电商订单场景（AT模式）

```
用户下单 → 创建订单 → 扣减库存 → 扣减积分 → 创建物流单
        ↓
    如果任何步骤失败，自动回滚所有操作
```

### 2. 金融转账场景（TCC模式）

```
转账操作 → Try: 冻结转出账户金额
         → Try: 冻结转入账户（防重入）
         → Confirm: 扣减转出账户
         → Confirm: 增加转入账户
         
如果失败 → Cancel: 解冻所有账户
```

### 3. 高并发秒杀场景（TCC模式）

```
秒杀下单 → Try: 冻结库存1件
         → Try: 冻结用户金额
         → Confirm: 扣减库存
         → Confirm: 扣减金额
         → Confirm: 创建订单
```

---

## 最佳实践

### 1. XID传递

- **同步调用**：通过Context传递
- **HTTP调用**：通过Header传递（`TX_XID`）
- **消息队列**：放入消息属性

### 2. 超时设置

```go
// 根据业务复杂度设置合理超时
xid, _ := tm.Begin(30 * time.Second)  // 简单业务
xid, _ := tm.Begin(60 * time.Second)  // 复杂业务
```

### 3. 异常处理

```go
defer func() {
    if r := recover(); r != nil {
        tm.Rollback(xid)
        panic(r)
    }
}()
```

### 4. TCC幂等性

- 使用唯一的`business_key`
- 双重检查（内存+数据库）
- 状态机管理

### 5. 性能优化

- AT模式：异步删除undo log
- TCC模式：减少数据库查询次数
- 使用连接池
- 批量提交

---

## 与Seata官方对比

| 功能 | 官方Seata（Java） | 本实现（Go） |
|------|-------------------|--------------|
| AT模式 | ✅ 完整支持 | ✅ 核心流程 |
| TCC模式 | ✅ 完整支持 | ✅ 完整实现 |
| Saga模式 | ✅ 支持 | ❌ 未实现 |
| XA模式 | ✅ 支持 | ❌ 未实现 |
| 全局锁 | ✅ 支持 | ⚠️ 简化版 |
| 协调器集群 | ✅ 支持 | ❌ 单机版 |
| 多语言客户端 | ✅ 支持 | ✅ Go实现 |

本实现主要用于**学习和理解**Seata的核心原理，生产环境建议使用官方版本。

---

## 参考资料

- [Seata官方文档](https://seata.io/zh-cn/docs/overview/what-is-seata.html)
- [分布式事务Seata原理](https://seata.io/zh-cn/docs/dev/mode/at-mode.html)
- [TCC模式设计](https://seata.io/zh-cn/docs/dev/mode/tcc-mode.html)

---

## 作者

本项目为学习实践项目，展示Seata分布式事务在Go语言中的实现方式。

