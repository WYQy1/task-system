# 分布式任务系统（Distributed Task System）

**Version: v1.2.0**

基于 Go、Gin 和 MySQL 开发的任务管理后端项目，主要实现员工信息管理与个人任务管理。

项目围绕 RESTful API、数据库操作、业务逻辑分层、自动化测试及基础工程化实践展开。

v1.1 在 v1.0 基础 CRUD 功能之上，完成了第一轮分层架构重构，将 HTTP 请求处理、业务逻辑、数据库访问和数据模型进行职责分离。

v1.2 在 v1.1 分层架构之上，完成三层彻底解耦与测试体系完善：业务校验全部下沉到 Service 层，错误响应统一收敛，Service 层单元/集成测试与全局接口集成测试相互配合，测试库相互隔离。

当前版本已完成核心 CRUD 接口、参数校验、统一错误处理及覆盖各层的自动化测试，后续将进入工程化增强阶段，并逐步向分布式任务系统演进。

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
- 员工姓名长度限制（最大 20 个字符）
- 员工年龄范围校验（1 - 150）
- 员工性别归一化（男/女/大小写英文 → male/female）
- 业务错误与数据库错误分层定义（哨兵错误）
- Handler 统一将业务错误映射为 HTTP 状态码
- 数据库查询与操作错误处理

任务状态包括：

- `pending`：待处理
- `in_progress`：进行中
- `completed`：已完成

## 三、项目架构

当前项目采用分层架构，将不同职责拆分到独立 Package 中。

| 层级 | 目录 | 主要职责 |
|---|---|---|
| Handler | `handler/` | 接收 HTTP 请求、参数解析、错误到 HTTP 状态码的映射 |
| Service | `service/` | 业务逻辑、业务规则校验、业务错误定义 |
| Repository | `repository/` | 数据库查询、插入、更新、删除、事务 |
| Model | `model/` | 定义员工与任务的数据模型 |

### 三层彻底解耦原则

- Handler 只做 HTTP 相关的事：解析请求、调用 Service、映射响应，不写任何业务规则与 SQL
- Service 只做业务逻辑：参数校验、状态判断、编排 Repository，不接触 SQL 细节
- Repository 只做数据库访问：封装 SQL 与事务，不包含业务判断
- 各层通过构造函数进行依赖注入，不依赖全局状态（Service 组合所需 Repository，跨实体校验通过注入另一个 Repository 实现）
- 错误分层传递：业务错误在 Service 层定义、数据库错误在 Repository 层定义，Handler 统一映射为 HTTP 状态码

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

## 四、项目结构

```text
task-system/
├── handler/                      # Handler 层
│   ├── user_handler.go           # 员工接口处理器
│   ├── task_handler.go           # 任务接口处理器
│   └── helpers.go                # 公共函数：路径参数解析、业务错误→HTTP 响应映射
│
├── service/                      # Service 层
│   ├── user_service.go           # 员工业务逻辑与校验
│   ├── task_service.go           # 任务业务逻辑与校验
│   ├── errors.go                 # 业务错误定义（哨兵错误）
│   ├── validation_test.go        # Service 层测试：校验函数单元测试（表驱动）
│   ├── user_service_test.go      # Service 层测试：员工业务逻辑测试
│   ├── task_service_test.go      # Service 层测试：任务业务逻辑测试
│   └── test_helpers_test.go      # Service 层测试辅助函数（连库、造数、清理、断言）
│
├── repository/                   # Repository 层
│   ├── user_repository.go        # 员工表数据访问
│   ├── task_repository.go        # 任务表数据访问
│   └── errors.go                 # 数据库错误定义
│
├── model/                        # Model 层
│   ├── user.go                   # 员工数据模型
│   └── task.go                   # 任务数据模型
│
├── main.go                       # 程序入口、依赖注入、路由注册
├── database.go                   # 数据库连接初始化
├── schema.sql                    # 数据库与表结构（含正式库与两个测试库）
│
├── user_test.go                  # 全局测试：员工接口集成测试（HTTP 全链路）
├── task_test.go                  # 全局测试：任务接口集成测试（HTTP 全链路）
├── test_helpers_test.go          # 全局测试辅助函数（连库、造数、发请求、断言）
│
├── .env                          # 本地环境配置（不提交 Git）
├── .env.example                  # 环境变量配置示例
├── .gitignore
├── go.mod
├── go.sum
├── README.md
└── note.md
```

说明：根目录的三个测试文件属于 `package main`，通过真实 HTTP 请求走完整链路（路由 → Handler → Service → Repository → MySQL），属于全局接口集成测试；`service/` 下的测试文件直接调用 Service 方法，属于对应层的测试。

## 五、数据库设计

目前使用 MySQL，包含正式库与两个测试库，共三套数据库，表结构完全一致，统一由 `schema.sql` 维护。

### 1. users

员工信息表，保存员工基本信息。

主要字段：

- `id`：员工 ID
- `name`：员工姓名
- `age`：员工年龄
- `gender`：员工性别（统一存储 male/female）

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

当前测试体系分为两层，各使用独立测试库，并行执行时互不干扰。

### 1. 全局测试（接口级集成测试，package main）

位于项目根目录，通过真实 HTTP 请求走「路由 → Handler → Service → Repository → MySQL」完整链路，连接 `task_system_test` 库。

- `user_test.go`：员工模块全部接口的集成测试（CRUD、非法参数、不存在、空列表等场景）
- `task_test.go`：任务模块全部接口的集成测试（CRUD、员工不存在、标题校验、状态校验、员工数据隔离等场景）
- `test_helpers_test.go`：全局测试辅助函数（setupTestDB、createTestUser、createTestTask、performRequest、checkUser、checkTask）

### 2. Service 层测试（package service）

位于 `service/` 目录，直接调用 Service 方法验证业务逻辑，连接 `task_system_service_test` 库。

- `validation_test.go`：校验函数单元测试（ID 校验、姓名/标题归一化、年龄/性别/状态校验，表驱动，不连数据库）
- `user_service_test.go`：员工业务逻辑测试（创建归一化、非法输入、不存在、无变化更新、删除等边界）
- `task_service_test.go`：任务业务逻辑测试（创建、外键不存在、非法输入、更新、删除等边界）
- `test_helpers_test.go`：Service 层测试辅助函数（openTestDB、造数、清理、checkUser、checkTask）

Handler 层逻辑较薄（仅参数解析与错误映射），未单独建立测试文件，由全局接口集成测试完整覆盖。

执行测试：

```bash
go test -count=1 ./...
```

执行静态检查：

```bash
go vet ./...
```

执行代码格式化检查：

```bash
gofmt -l .
```

## 七、下载方式

### 1. Git 克隆项目

```bash
git clone git@github.com:WYQy1/task-system.git
```

HTTPS 方式：

```bash
git clone https://github.com/WYQy1/task-system.git
```

### 2. 查看与切换版本

项目每个大版本都有对应的 tag：

```bash
# 查看所有版本标签
git tag

# 切换到指定版本，例如 v1.2.0
git checkout v1.2.0
```

克隆后需要按照「环境配置与运行」章节创建本地 `.env` 并初始化数据库，才能启动项目。

## 八、环境配置与运行

### 1. 配置环境变量

复制 `.env.example` 并创建本地 `.env` 文件，根据实际 MySQL 配置填写连接信息。

```bash
cp .env.example .env
```

`.env` 需要配置三个 DSN：

| 变量 | 对应数据库 | 用途 |
|---|---|---|
| `DB_DSN` | `task_system` | 正式运行 |
| `TEST_DB_DSN` | `task_system_test` | 全局接口集成测试 |
| `SERVICE_TEST_DB_DSN` | `task_system_service_test` | Service 层测试 |

注意：`.env` 包含本地数据库连接配置，不应提交到 GitHub。

### 2. 初始化数据库

执行 `schema.sql`，一次创建正式库和两个测试库及对应数据表：

```bash
mysql -u root -p < schema.sql
```

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

## 九、当前开发进度

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

### v1.2：三层解耦深化与测试体系完善阶段

已完成：

- [x] 业务校验从 Handler 彻底下沉到 Service
- [x] 错误响应格式统一，Handler 错误映射收敛到 `helpers.go`
- [x] 员工性别统一归一化为 male/female
- [x] Repository 层错误独立定义，数据库错误与业务错误彻底分层
- [x] TaskService 注入 UserRepository，跨实体校验员工存在
- [x] Service 层校验函数单元测试（表驱动）
- [x] Service 层员工/任务业务逻辑测试
- [x] 全局接口集成测试重构（helper 复用、表驱动合并重复用例）
- [x] 测试库隔离：全局测试与 Service 测试使用独立测试库
- [x] `schema.sql` 覆盖正式库与两个测试库
- [x] `gofmt` / `go vet` / `go build` / `go test` 全部通过

## 十、后续开发计划

### v1.3：工程化增强（预期）

- 统一 API 响应结构（code/message/data）
- 用户登录与身份认证（JWT）
- 请求日志中间件与结构化日志
- 任务列表分页与状态筛选
- 数据库迁移工具（消除 schema.sql 中三库重复 DDL）
- 补充并发、超长等边界测试

### 长远方向：向分布式任务系统演进

- 引入 Redis 缓存与分布式锁
- 引入消息队列，实现任务投递与异步执行
- 分布式任务调度（多实例、定时任务）
- 任务重试、幂等与状态机
- Docker 容器化部署
- Kubernetes 编排（云原生方向）
- 服务拆分与微服务化

以上为后续计划，不代表相关功能已经实现。

## 十一、版本记录

### v1.2.0

完成三层彻底解耦与测试体系完善。

- 业务校验从 Handler 彻底下沉到 Service
- 统一错误响应格式，Handler 错误映射收敛
- 性别字段统一归一化为 male/female
- 补全 Service 层单元测试与业务逻辑测试
- 重构全局接口集成测试（helper 复用、表驱动）
- 测试库隔离（task_system_test / task_system_service_test）
- TaskService 注入 UserRepository 校验员工存在
- schema.sql 覆盖正式库与两个测试库
- gofmt / go vet / go build / go test 全量通过

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

**项目当前阶段：v1.2.0 —— 三层彻底解耦与测试体系完善完成，全量自动化测试通过。**
