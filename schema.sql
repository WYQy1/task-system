-- ============================================================
-- 初始化脚本：正式库 + 两个测试库（三库表结构一致）
-- 使用方法：mysql -u root -p < schema.sql
-- ============================================================

-- 1. 正式库
CREATE DATABASE IF NOT EXISTS task_system CHARACTER SET utf8mb4;

USE task_system;

CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(20) NOT NULL,
    age INT,
    gender VARCHAR(10),
    password VARCHAR(255) NOT NULL,
    token_version INT NOT NULL DEFAULT 0,
    -- 0=普通员工，1=管理员；注册接口固定写入 0，管理员由 SQL 手工提升
    role TINYINT NOT NULL DEFAULT 0

);

CREATE TABLE IF NOT EXISTS tasks (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- 截止时间：为空表示不设截止；超期扫描只关心未完成且有截止时间且未提醒过/已到期的任务
    due_at DATETIME NULL,
    -- 最近一次超期提醒时间：作为提醒去重（幂等）的标记
    reminded_at DATETIME NULL,
    -- 超期扫描的主查询条件（reminded_at, due_at），fresh install 也要建，不能只写进升级注释
    KEY idx_tasks_overdue (reminded_at, due_at),
        FOREIGN KEY (user_id) REFERENCES users(id)
);

-- 2. 测试库：全局接口集成测试使用（对应 TEST_DB_DSN）
CREATE DATABASE IF NOT EXISTS task_system_test CHARACTER SET utf8mb4;

USE task_system_test;

CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(20) NOT NULL,
    age INT,
    gender VARCHAR(10),
    password VARCHAR(255) NOT NULL,
    token_version INT NOT NULL DEFAULT 0,
    -- 0=普通员工，1=管理员；注册接口固定写入 0，管理员由 SQL 手工提升
    role TINYINT NOT NULL DEFAULT 0

);

CREATE TABLE IF NOT EXISTS tasks (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- 截止时间：为空表示不设截止；超期扫描只关心未完成且有截止时间且未提醒过/已到期的任务
    due_at DATETIME NULL,
    -- 最近一次超期提醒时间：作为提醒去重（幂等）的标记
    reminded_at DATETIME NULL,
    -- 超期扫描的主查询条件（reminded_at, due_at），fresh install 也要建，不能只写进升级注释
    KEY idx_tasks_overdue (reminded_at, due_at),
        FOREIGN KEY (user_id) REFERENCES users(id)
);

-- 3. 测试库：Service 层测试使用（对应 SERVICE_TEST_DB_DSN）
CREATE DATABASE IF NOT EXISTS task_system_service_test CHARACTER SET utf8mb4;

USE task_system_service_test;

CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(20) NOT NULL,
    age INT,
    gender VARCHAR(10),
    password VARCHAR(255) NOT NULL,
    token_version INT NOT NULL DEFAULT 0,
    -- 0=普通员工，1=管理员；注册接口固定写入 0，管理员由 SQL 手工提升
    role TINYINT NOT NULL DEFAULT 0

);

CREATE TABLE IF NOT EXISTS tasks (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- 截止时间：为空表示不设截止；超期扫描只关心未完成且有截止时间且未提醒过/已到期的任务
    due_at DATETIME NULL,
    -- 最近一次超期提醒时间：作为提醒去重（幂等）的标记
    reminded_at DATETIME NULL,
    -- 超期扫描的主查询条件（reminded_at, due_at），fresh install 也要建，不能只写进升级注释
    KEY idx_tasks_overdue (reminded_at, due_at),
        FOREIGN KEY (user_id) REFERENCES users(id)
);

-- ============================================================
-- 4. v1.2.x → v1.3.0 升级（已存在的库执行这一段即可，新建库跳过）
-- ============================================================
-- ALTER TABLE users  ADD COLUMN role TINYINT NOT NULL DEFAULT 0;
-- ALTER TABLE tasks  ADD COLUMN due_at DATETIME NULL, ADD COLUMN reminded_at DATETIME NULL;
-- ALTER TABLE tasks  ADD INDEX idx_tasks_overdue (reminded_at, due_at);
-- UPDATE users SET role = 1 WHERE id = 1;  -- 把要当管理员的那条账号提权
