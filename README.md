# 分布式任务系统（Distributed Task System）

**Version: v1.1.0**

基于 Go、Gin 和 MySQL 开发的任务管理后端项目，主要实现员工信息管理与个人任务管理。

项目围绕 RESTful API、数据库操作、业务逻辑分层、自动化测试及基础工程化实践展开。

v1.1 在 v1.0 基础 CRUD 功能之上，完成了第一轮分层架构重构，将 HTTP 请求处理、业务逻辑、数据库访问和数据模型进行职责分离。

当前版本已完成核心 CRUD 接口、基础参数校验、错误处理及自动化测试，后续将继续完善测试体系与工程化能力。

## 一、技术栈

| 技术 | 用途 |
|---|---|
| Go 1.25 | 后端开发语言 |
| Gin | HTTP Web 框架 |
| MySQL | 关系型数据库 |
| database/sql | 数据库操作 |
| go-sql-driver/mysql | MySQL 驱动 |
| godotenv | 环境变量配置管理 |
| Go testing | 自动化测试 |

## 二、已实现功能

### 1. 员工管理

- 创建员工
- 查询员工列表
- 根据 ID 查询员工
- 更新员工信息
- 删除员工
- 删除员工时通过数据库事务同步删除关联任务

### 2. 任务管理

- 创建员工任务
- 查询员工任务列表
- 查询单条任务
- 更新任务信息
- 删除任务
- 任务与员工建立外键关联关系
- 任务列表按照 ID 升序排列

### 3. 参数校验与错误处理

- 员工 ID、任务 ID 格式校验
- 必填参数校验
- 任务标题长度限制（最大 100 个字符）
- 任务标题空值及纯空格校验
- 任务状态合法性校验
- 数据库查询与操作错误处理
- 业务错误与数据库错误分层处理
- HTTP 状态码处理

任务状态包括：

- `pending`：待处理
- `in_progress`：进行中
- `completed`：已完成

## 三、项目架构

当前项目采用基础分层架构，将不同职责拆分到独立 Package 中。

| 层级 | 目录 | 主要职责 |
|---|---|---|
| Handler | `handler/` | 接收 HTTP 请求、参数解析、响应处理 |
| Service | `service/` | 业务逻辑、业务规则校验、业务错误处理 |
| Repository | `repository/` | 数据库查询、插入、更新、删除 |
| Model | `model/` | 定义员工与任务的数据模型 |

主要请求处理流程：

```text
Client
  ↓
Gin Router
  ↓
Handler
  ↓
Service
  ↓
Repository
  ↓
MySQL
```

各层通过结构体和构造函数进行依赖注入，减少业务逻辑与数据库操作之间的耦合。

## 四、项目结构

```text
task-system/
├── handler/
│   ├── user_handler.go
│   └── task_handler.go
│
├── model/
│   ├── user.go
│   └── task.go
│
├── repository/
│   ├── user_repository.go
│   └── task_repository.go
│
├── service/
│   ├── errors.go
│   ├── user_service.go
│   └── task_service.go
│
├── main.go                  # 程序入口、Gin 初始化及路由注册
├── database.go              # MySQL 数据库连接初始化
├── schema.sql               # 数据库表结构
│
├── main_test.go             # 员工模块相关测试
├── task_test.go             # 任务模块相关测试
├── test_helpers_test.go     # 测试辅助函数
│
├── .env                     # 本地环境配置（不提交 Git）
├── .env.example             # 环境变量配置示例
├── .gitignore
├── go.mod
├── go.sum
├── README.md
└── note.md
```

## 五、数据库设计

目前使用 MySQL，主要包含两张数据表。

### 1. users

员工信息表，保存员工基本信息。

主要字段：

- `id`：员工 ID
- `name`：员工姓名
- `age`：员工年龄
- `gender`：员工性别

### 2. tasks

任务信息表，保存员工任务。

主要字段：

- `id`：任务 ID
- `user_id`：所属员工 ID
- `title`：任务标题
- `description`：任务描述
- `status`：任务状态
- `created_at`：任务创建时间

通过 `tasks.user_id` 与 `users.id` 建立外键关联，实现员工与任务之间的一对多关系。

员工删除操作使用数据库事务，先删除其关联任务，再删除员工记录，以保证操作的一致性。

## 六、自动化测试

目前已经建立基础自动化测试体系。

测试范围包括：

- 员工 CRUD 接口
- 任务创建、查询、更新和删除
- 非法 ID 参数
- 员工不存在
- 任务不存在
- 任务标题参数校验
- 任务状态合法性校验
- 数据库操作结果验证

测试辅助函数统一放在 `test_helpers_test.go`，减少重复测试代码。

执行测试：

```bash
go test -count=1 ./...
```

执行静态检查：

```bash
go vet ./...
```

## 七、环境配置与运行

### 1. 配置环境变量

复制 `.env.example` 并创建本地 `.env` 文件，根据实际 MySQL 配置填写数据库连接信息。

```bash
cp .env.example .env
```

注意：`.env` 包含本地数据库连接配置，不应提交到 GitHub。

### 2. 初始化数据库

根据 `schema.sql` 创建数据库及相关数据表。

### 3. 安装依赖

```bash
go mod tidy
```

### 4. 启动项目

```bash
go run .
```

服务默认监听：

```text
http://localhost:8080
```

## 八、当前开发进度

### v1.0：基础 CRUD 阶段

已完成：

- [x] Go/Gin 基础项目搭建
- [x] MySQL 数据库连接
- [x] 员工 CRUD 接口
- [x] 任务 CRUD 接口
- [x] 基础参数校验与错误处理
- [x] 员工与任务关联关系
- [x] 基础自动化测试

### v1.1：分层架构重构阶段

已完成：

- [x] Handler、Service、Repository 分层
- [x] Model 数据模型独立
- [x] 构造函数与依赖注入
- [x] 数据库操作从 Handler 中分离
- [x] 业务校验集中到 Service
- [x] 数据库错误与业务错误分层处理
- [x] 员工删除事务处理
- [x] 测试代码适配新架构
- [x] 清理旧类型别名文件
- [x] Go 代码格式化
- [x] `go build ./...` 构建通过
- [x] `go test ./...` 全量测试通过

## 九、后续开发计划

后续将以现有项目为基础，逐步提升工程化能力：

- 完善 Handler 代码组织，减少重复逻辑
- 补充 Service、Repository 层测试
- 深入数据库事务、索引与查询优化
- 完善日志记录与错误处理
- 引入身份认证与权限控制
- 完善接口测试与边界测试
- 学习并实践 Redis、消息队列等组件
- 探索分布式任务调度与任务执行机制
- Docker 容器化部署
- 系统性能测试与优化

以上为后续计划，不代表相关功能已经实现。

## 十、版本记录

### v1.1.0

完成第一轮分层架构重构。

- 将项目拆分为 Handler、Service、Repository、Model 四层
- 使用构造函数进行依赖注入
- 将数据库访问逻辑从 Handler 中分离
- 将业务校验集中到 Service
- 完善业务错误与数据库错误处理
- 完善员工删除事务处理
- 调整测试代码以适配新架构
- 全量自动化测试通过

### v1.0.0

完成基础员工与任务管理功能。

- 实现员工及任务 CRUD 接口
- 完成基础参数校验与错误处理
- 建立自动化测试体系
- 抽取公共测试辅助函数
- 完成初步 Handler 重构

---

**项目当前阶段：v1.1.0 —— 第一轮分层架构重构完成，自动化测试通过。**