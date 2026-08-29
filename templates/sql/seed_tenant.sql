-- Demo Tenant Database Creation & Seed
CREATE DATABASE IF NOT EXISTS `{{.DemoDBName}}` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `{{.DemoDBName}}`;

-- Demo Tutor Sessions Table
CREATE TABLE IF NOT EXISTS `tutor_sessions` (
    `id` CHAR(36) NOT NULL PRIMARY KEY,
    `student_name` VARCHAR(255) NOT NULL,
    `student_id` VARCHAR(100) NULL,
    `volunteer_id` CHAR(36) NULL,
    `ai_prep_notes` TEXT NULL,
    `scheduled_for` TIMESTAMP NULL,
    `status` VARCHAR(50) DEFAULT 'scheduled',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Demo Volunteers Table
CREATE TABLE IF NOT EXISTS `volunteers` (
    `id` CHAR(36) NOT NULL PRIMARY KEY,
    `name` VARCHAR(255) NOT NULL,
    `email` VARCHAR(255) NOT NULL,
    `global_user_id` CHAR(36) NULL,
    `availability` JSON NULL,
    `status` VARCHAR(50) DEFAULT 'active',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Insert Sample Volunteer
INSERT IGNORE INTO `volunteers` (`id`, `name`, `email`, `status`)
VALUES ('v-101-demo-uuid', 'Alex Rivera (Volunteer Tutor)', 'alex.tutor@example.com', 'active');

-- Register Demo Tenant in Core Database
USE `{{.CoreDBName}}`;
INSERT INTO `tenants` (`id`, `domain`, `trade_module`, `db_name`, `api_keys`)
VALUES (
    '{{.DemoTenantID}}',
    '{{.DemoDomain}}',
    'Tutor',
    '{{.DemoDBName}}',
    JSON_OBJECT('description', 'Demo Seed Tenant')
) ON DUPLICATE KEY UPDATE `trade_module` = 'Tutor', `db_name` = '{{.DemoDBName}}';
