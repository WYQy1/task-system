# 任务管理系统（Task System）

**Version: v1.2.1**

基于 Go + Gin + MySQL 的员工与任务管理后端，采用 Handler / Service / Repository / Model 四层结构。v1.2.1 完成 JWT 鉴权、context 全链路超时控制与 Swagger 接口文档。

## 技术栈

| 技术 | 用途 |
|---|---|
| Go 1.25 | 后端语言 |
| Gin | HTTP 框架 |
| MySQL 8 | 数据库 |
| database/sql + go-sql-driver/mysql | 数据访问 |
| golang-jwt/jwt v5 | JWT 签发与校验 |
| bcrypt | 密码哈希 |
| swaggo/swag + gin-swagger | 接口文档 |
| testify | 测试断言 |
| godotenv | 环境变量 |

## 已实现功能

**员工管理**：增删改查；删除时用事务同步删除其关联任务。

**任务管理**：按员工创建任务、任务列表、单条查询、更新、删除；与员工外键关联；列表按 ID 升序。

**认证**：注册（密码 bcrypt 哈希存储）、登录（签发 24 小时 JWT）、登出（`token_version` 撤销旧 token）。

**接口文档**：Swagger UI，可在页面 Authorize 后直接调试受保护接口。

## 项目架构

| 层 | 目录 | 职责 |
|---|---|---|
| Handler | `handler/` | 解析请求、调用 Service、把错误映射为 HTTP 状态码 |
| Service | `service/` | 业务规则校验与编排，不写 SQL |
| Repository | `repository/` | SQL 与事务，不做业务判断 |
| Model | `model/` | 数据结构定义 |
| Middleware | `middleware/` | JWT 鉴权、请求超时 |

请求链路：

```text
Router → Timeout → Handler → Service → Repository → MySQL
```

`ctx` 从 `c.Request.Context()` 一路传到 `QueryRowContext`，中途任何一层断链，超时都无法穿透到数据库。

## v1.2.1 的三个关键设计

### 1. JWT 鉴权与 token 撤销

- `users` 表新增 `password`、`token_version`
- 登录签发 claims：`user_id` / `exp`（24 小时）/ `ver`
- `AuthMiddleware` 七步校验：取 `Authorization` 头 → 拆 `Bearer` → 验签 → 查库取当前版本号 → 比对 → 注入 `userID` → 放行
- 登出执行 `token_version + 1`，旧 token 因版本号不一致立即 401
- 两个关卡分工：验签管"是否由本服务签发"，版本号管"是否仍然有效"

JWT 本身无状态、签发后无法收回，"每次请求回库比对版本号"是唯一可行的撤销手段，代价是每个受保护请求多一次查询。

### 2. context 全链路与超时

- 13 个 repository 方法、14 个 service 方法全部以 `ctx context.Context` 为首参
- 全局 `middleware.Timeout(5 * time.Second)`：`context.WithTimeout` + `defer cancel()` 把带截止时间的 ctx 注入请求
- 超时通过 `errors.Is(err, context.DeadlineExceeded)` 映射为 504 `请求超时`
- 已验证：`SELECT SLEEP(2)` 在 100ms 超时下 100.39ms 被掐断（`ctx_test.go`）

取消是**协作式**的：`context` 只负责广播，只有愿意监听 `ctx.Done()` 的代码（如 SQL 驱动）才会真正中止；不检查 ctx 的计算（如 bcrypt）会照常跑完。

### 3. 数据库连接池

```go
SetMaxOpenConns(25)
SetMaxIdleConns(10)
SetConnMaxLifetime(5 * time.Minute)
```

## 项目结构

```text
task-system/
├── handler/         # user / task / auth 处理器 + helpers（错误映射）
├── service/         # 业务逻辑与校验，含层内测试
├── repository/      # SQL 与事务
├── model/           # User / Task
├── middleware/      # auth.go（JWT）、timeout.go（超时）
├── docs/            # swag init 生成物，勿手改
├── main.go          # 依赖注入与路由注册、Swagger 全局注解
├── database.go      # 连接初始化与连接池
├── schema.sql       # 正式库 + 两个测试库的表结构
├── *_test.go        # 根目录：HTTP 全链路集成测试
└── README.md
```

根目录的 `auth_test.go`、`user_test.go`、`task_test.go`、`ctx_test.go`、`http_helpers_test.go` 均属 `package main`，必须与被测代码同目录，这是 Go 对同包测试的约束，不是文件乱放。

## 快速开始

1. 配置环境变量

```bash
cp .env.example .env
```

填写 `DB_DSN`、`TEST_DB_DSN`、`SERVICE_TEST_DB_DSN`、`JWT_SECRET`。

2. 初始化数据库

```bash
mysql -u root -p < schema.sql
```

3. 安装依赖并启动

```bash
go mod tidy
go run .
```

- 服务地址：http://localhost:8080
- 接口文档：http://localhost:8080/swagger/index.html

修改接口注解后需重新生成文档：

```bash
swag init
```

## 测试与静态检查

```bash
go test -count=1 ./...
go vet ./...
gofmt -l .
```

测试库相互隔离：`task_system_test` 用于 HTTP 集成测试，`task_system_service_test` 用于 Service 层测试。

## 已知限制

- 只做认证未做授权：任何登录用户都能修改任意员工（`/users` 系列未读取 `userID`）
- `AuthMiddleware` 查库失败一律返回 401，超时和数据库故障被误判为身份错误
- `users.name` 未加 UNIQUE 约束，同名账号可重复注册
- 登出为全设备失效，无法只踢掉单个设备

## 版本记录

**v1.2.1** —— JWT 鉴权（bcrypt + token_version 撤销）、context 全链路 + 5 秒超时中间件 + 504 映射、连接池配置、Swagger 接口文档

**v1.2.0** —— 三层彻底解耦，业务校验下沉 Service，统一错误映射，测试库隔离

**v1.1.0** —— 拆分 Handler / Service / Repository / Model，构造函数依赖注入

**v1.0.0** —— 员工与任务 CRUD、基础参数校验、自动化测试

## 后续计划（v1.3）

- 统一响应结构（code / message / data）
- 授权模型：基于 `userID` 限制只能操作本人数据
- 任务列表分页与状态筛选
- 数据库迁移工具，消除 `schema.sql` 中三库重复 DDL
- 接口日志与结构化日志
