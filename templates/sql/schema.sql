-- Core Database Initialization for Agency Platform
CREATE DATABASE IF NOT EXISTS `{{.CoreDBName}}` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `{{.CoreDBName}}`;

-- Tenants Registry
CREATE TABLE IF NOT EXISTS `tenants` (
    `id` CHAR(36) NOT NULL PRIMARY KEY,
    `domain` VARCHAR(255) NOT NULL,
    `trade_module` VARCHAR(100) NOT NULL,
    `db_name` VARCHAR(100) NOT NULL,
    `api_keys` JSON NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY `idx_domain` (`domain`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Users Table (Centralized IdP)
CREATE TABLE IF NOT EXISTS `users` (
    `id` CHAR(36) NOT NULL PRIMARY KEY,
    `email` VARCHAR(255) NOT NULL,
    `password_hash` VARCHAR(255) NOT NULL,
    `tenant_id` CHAR(36) NULL,
    `role` VARCHAR(50) DEFAULT 'admin',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY `idx_email` (`email`),
    KEY `idx_tenant` (`tenant_id`),
    CONSTRAINT `fk_users_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Asynchronous AI Job Queue
CREATE TABLE IF NOT EXISTS `ai_job_queue` (
    `id` CHAR(36) NOT NULL PRIMARY KEY,
    `tenant_id` CHAR(36) NOT NULL,
    `user_id` CHAR(36) NULL,
    `status` ENUM('pending', 'dispatched', 'processing', 'completed', 'failed') DEFAULT 'pending',
    `payload` JSON NOT NULL,
    `response` JSON NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY `idx_status_created` (`status`, `created_at`),
    KEY `idx_tenant_jobs` (`tenant_id`),
    CONSTRAINT `fk_jobs_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
