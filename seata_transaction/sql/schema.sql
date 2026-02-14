-- ========================================
-- Seata 分布式事务数据库表结构
-- ========================================

-- ========== AT 模式表结构 ==========

-- 1. Undo Log 表（AT模式必需）
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

-- 2. 订单表（业务表示例）
CREATE TABLE IF NOT EXISTS orders (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    product_id BIGINT NOT NULL COMMENT '商品ID',
    quantity INT NOT NULL COMMENT '购买数量',
    status VARCHAR(20) NOT NULL COMMENT '订单状态：PENDING/SUCCESS/FAILED',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_product_id (product_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单表';

-- 3. 库存表（业务表示例）
CREATE TABLE IF NOT EXISTS stock (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    product_id BIGINT NOT NULL UNIQUE COMMENT '商品ID',
    quantity INT NOT NULL COMMENT '库存数量',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_product_id (product_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存表';

-- 插入测试数据
INSERT INTO stock (product_id, quantity) VALUES 
(100, 1000),
(200, 500),
(300, 2000)
ON DUPLICATE KEY UPDATE quantity=VALUES(quantity);


-- ========== TCC 模式表结构 ==========

-- 1. 账户表（包含冻结金额字段）
CREATE TABLE IF NOT EXISTS account (
    user_id BIGINT PRIMARY KEY COMMENT '用户ID',
    balance DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '可用余额',
    frozen_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00 COMMENT '冻结金额',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    CHECK (balance >= 0),
    CHECK (frozen_amount >= 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='账户表';

-- 2. TCC事务记录表（用于幂等性和状态管理）
CREATE TABLE IF NOT EXISTS tcc_transaction (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    xid VARCHAR(100) NOT NULL COMMENT '全局事务ID',
    business_key VARCHAR(100) NOT NULL COMMENT '业务主键（幂等性）',
    status INT NOT NULL COMMENT '状态：0-Trying, 1-Committed, 2-Cancelled',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    amount DECIMAL(15,2) NOT NULL COMMENT '金额',
    create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    retry_count INT NOT NULL DEFAULT 0 COMMENT '重试次数',
    UNIQUE KEY uk_business_key (business_key),
    INDEX idx_xid (xid),
    INDEX idx_status (status),
    INDEX idx_create_time (create_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='TCC事务记录表';

-- 插入测试账户数据
INSERT INTO account (user_id, balance) VALUES 
(10001, 10000.00),
(10002, 5000.00),
(10003, 20000.00)
ON DUPLICATE KEY UPDATE balance=VALUES(balance);


-- ========== 全局事务记录表（可选）==========

-- 全局事务表（用于事务协调器TC）
CREATE TABLE IF NOT EXISTS global_transaction (
    xid VARCHAR(100) PRIMARY KEY COMMENT '全局事务ID',
    transaction_id BIGINT NOT NULL COMMENT '事务ID',
    status INT NOT NULL COMMENT '状态：0-Begin,1-Committing,2-Committed,3-Rollbacking,4-Rollbacked',
    application_id VARCHAR(64) NOT NULL COMMENT '应用ID',
    transaction_service_group VARCHAR(64) NOT NULL COMMENT '事务服务组',
    transaction_name VARCHAR(128) COMMENT '事务名称',
    timeout INT NOT NULL COMMENT '超时时间（毫秒）',
    begin_time BIGINT NOT NULL COMMENT '开始时间',
    application_data VARCHAR(2000) COMMENT '应用数据',
    gmt_create DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    gmt_modified DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_gmt_modified (gmt_modified),
    INDEX idx_transaction_id (transaction_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='全局事务表';

-- 分支事务表（用于事务协调器TC）
CREATE TABLE IF NOT EXISTS branch_transaction (
    branch_id BIGINT PRIMARY KEY COMMENT '分支ID',
    xid VARCHAR(100) NOT NULL COMMENT '全局事务ID',
    transaction_id BIGINT NOT NULL COMMENT '事务ID',
    resource_group_id VARCHAR(32) COMMENT '资源组ID',
    resource_id VARCHAR(256) NOT NULL COMMENT '资源ID',
    branch_type VARCHAR(8) NOT NULL COMMENT '分支类型：AT/TCC/SAGA/XA',
    status INT NOT NULL COMMENT '状态：1-Registered,2-PhaseOneDone,3-PhaseTwoCommitted,4-PhaseTwoRollbacked',
    client_id VARCHAR(64) NOT NULL COMMENT '客户端ID',
    application_data VARCHAR(2000) COMMENT '应用数据',
    gmt_create DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    gmt_modified DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_xid (xid),
    INDEX idx_gmt_modified (gmt_modified)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='分支事务表';

-- ========== 使用示例 ==========

-- AT模式示例查询
-- 查看undo_log
-- SELECT * FROM undo_log WHERE xid = 'your-xid';

-- TCC模式示例查询
-- 查看账户余额和冻结金额
-- SELECT user_id, balance, frozen_amount, (balance + frozen_amount) as total 
-- FROM account WHERE user_id = 10001;

-- 查看TCC事务状态
-- SELECT * FROM tcc_transaction WHERE xid = 'your-xid';

