-- OIP Backend Database Schema
-- 创建时间: 2025-12-23

-- ============================================
-- Table: accounts
-- 说明: 账号表
-- 注意: ID使用分布式ID生成器，不使用AUTO_INCREMENT
-- ============================================
CREATE TABLE IF NOT EXISTS accounts (
    id BIGINT PRIMARY KEY COMMENT '账号ID（分布式ID）',
    name VARCHAR(255) NOT NULL COMMENT '账号名称',
    email VARCHAR(255) NOT NULL UNIQUE COMMENT '邮箱地址',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='账号表';

-- ============================================
-- Table: orders
-- 说明: 订单表
-- ============================================
CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(64) PRIMARY KEY COMMENT '订单ID (UUID)',
    account_id BIGINT NOT NULL COMMENT '账号ID',
    merchant_order_no VARCHAR(255) NOT NULL COMMENT '商户订单号',
    shipment JSON NOT NULL COMMENT '货件信息（包含发件地址、收件地址、包裹详情）',
    status VARCHAR(50) NOT NULL COMMENT '订单状态: DIAGNOSING/DIAGNOSED/FAILED',
    diagnose_result JSON COMMENT '诊断结果（包含诊断项列表）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    INDEX idx_account_id (account_id),
    INDEX idx_merchant_order_no (merchant_order_no),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    UNIQUE KEY uk_account_merchant (account_id, merchant_order_no) COMMENT '防止重复订单'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';
