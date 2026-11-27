# 任务管理系统（Task System）

**Version: v1.3.0**

基于 Go + Gin + MySQL 的员工与任务管理后端，采用 Handler / Service / Repository / Model 四层结构。v1.3.0 在 JWT 鉴权、context 全链路超时、Swagger 文档之上，补了**基于角色的授权模型**和**超期任务的后台扫描提醒**。

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

**任务管理**：按员工创建任务（可带截止时间 `due_at`）、任务列表、单条查询、更新、删除；与员工外键关联；列表按 ID 升序。

**认证**：注册（密码 bcrypt 哈希存储）、登录（签发 24 小时 JWT）、登出（`token_version` 撤销旧 token）。

**授权**：员工带 `role`（0 普通员工 / 1 管理员）。受保护接口统一从 `ctx` 取 `userID` 与 `role`，只能动自己的数据，跨员工访问返回 403；新增员工、查看全部员工仅管理员可用，普通员工的列表接口只返回自己那一条。

**超期提醒**：后台扫描器每分钟捞一批「未完成 + 已过截止时间 + 没提醒过」的任务，交给固定大小的 worker 池并发提醒；提醒成功才写 `reminded_at`，失败的下一轮自动重试。

**接口文档**：Swagger UI，可在页面 Authorize 后直接调试受保护接口。

## 项目架构

| 层 | 目录 | 职责 |
|---|---|---|
| Handler | `handler/` | 解析请求、调用 Service、把错误映射为 HTTP 状态码 |
| Service | `service/` | 业务规则校验与编排，不写 SQL |
| Repository | `repository/` | SQL 与事务，不做业务判断 |
| Model | `model/` | 数据结构定义 |
| Middleware | `middleware/` | JWT 鉴权、请求超时 |
| Worker | `worker/` | 后台任务：超期扫描、并发分发、优雅退出（不在请求链路上） |

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

## v1.3.0 的两个新设计

### 1. 授权：认证回答「你是谁」，授权回答「你能动谁」

v1.2.1 只有认证。中间件验完 token 就把 `userID` 放进 `ctx`，但业务 Handler 一个都没读——任何登录用户传别人的 id 就能改别人的员工和任务，README 里原先也把这条写成了已知限制。

现在统一成一个判定点：

```go
type viewer struct{ ID, Role int }

func (v viewer) canAccess(ownerID int) bool { return v.ID == ownerID || v.isAdmin() }
```

- 中间件校验 token 时，用一条 SQL 同时取回 `token_version` 和 `role`（`GetAuthInfo`），两个值都塞进 `ctx`，不额外多查一次库。
- Handler 侧只做两件事：`currentViewer(c)` 取身份，`fetchUserForViewer` / `fetchTaskForViewer` 取数据并判归属。
- **先判存在、再判归属**：查不到的 id 返回 404 而不是 403，否则「这个 id 存在但不是你的」和「这个 id 压根不存在」会被混成同一个响应；同时也保住了原有的 404 用例。
- 原来有条测试写着却什么都没测到：`TestGetUsersOnlySelf` 清表后只建一个用户，再断言 `len(users) == 1`——库里只有一个用户，它当然只能是 1。现在改成建两个用户：普通员工只看到自己，管理员看到全部。

### 2. 后台 worker：并发放在这里才有必要

给这个项目加并发，最容易加错地方。试过两条路，都被否掉了：

- 批量创建任务用 goroutine 并发插入——标准答案是单条 `INSERT ... VALUES (...),(...)`，并发反而抢同一把表锁。
- 看板按状态 fan-out 查询——一条 `GROUP BY status` 就够了。

真正无法用单条 SQL 表达的是**后台超期扫描**：找出一批到期任务，逐个做外部动作（现在打日志，将来发通知），还要能被取消、能优雅退出。所以并发放在这里：

```go
for i := 0; i < cfg.Workers; i++ {
    wg.Add(1)
    go func() { defer wg.Done(); for t := range jobs { s.handle(ctx, t) } }()
}
```

- 有界并发：`jobs` 缓冲等于 `BatchSize`（50），worker 固定 4 个，扫描批量与并发度分开配。
- 取消：`Run` 监听 `ctx.Done()`，收到就 `close(jobs)`，worker 把队列里的活干完再退出；`main` 里先 `srv.Shutdown`，再等 `background.Wait()`。
- 幂等靠数据库：`MarkReminded` 带 `AND reminded_at IS NULL`，重复标记不会覆盖时间；提醒失败就不标，下一轮还会捞到它——**宁可重复提醒，不可漏掉**。

取舍写在明面上：单次提醒如果比扫描周期还长，同一任务可能被两个 worker 同时提醒，数据库只能保证 `reminded_at` 不被重复写。要彻底杜绝得加进程内 in-flight 集合（单实例）或分布式锁（多实例）。

## 项目结构

```text
task-system/
├── handler/         # user / task / auth 处理器 + helpers（错误映射）+ authz.go（授权判定）
├── service/         # 业务逻辑与校验，含层内测试
├── repository/      # SQL 与事务
├── model/           # User / Task
├── middleware/      # auth.go（JWT）、timeout.go（超时）
├── worker/          # overdue.go（超期扫描：ticker + worker pool + 优雅退出）
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

已经按 v1.2.x 建过库的，执行 `schema.sql` 末尾「4. v1.2.x → v1.3.0 升级」那段 `ALTER`；`ADD COLUMN` 不幂等，别重复跑。

把某个账号提升为管理员（注册接口固定写 `role = 0`，没有开放「注册成管理员」的口子）：

```bash
mysql -u root -p task_system -e "UPDATE users SET role = 1 WHERE id = 1"
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

共 83 个测试函数、18 组 `t.Run` 表驱动：根目录 42 个（HTTP 全链路）、`service/` 34 个、`worker/` 5 个、`middleware/` 2 个。

测试库相互隔离：`task_system_test` 用于 HTTP 集成测试，`task_system_service_test` 用于 Service 层测试。

`worker/` 的测试用 fake store / fake notifier，不需要数据库：

```bash
go test ./worker/ -count=1 -v
```

另外建议在自己机器上跑一次 `go test -race ./...`：本机 gcc 是 32 位，`-race` 编不过（`cc1.exe: unimplemented: 64-bit mode not compiled in`），所以没能用竞态检测器再验一遍。

## 已知限制

- 超期提醒没有进程内 in-flight 去重：单次提醒耗时超过扫描周期时，同一条任务可能被重复提醒（数据库只保证 `reminded_at` 不被重复写）
- 提醒目前只打日志（`LogNotifier`），没接邮件 / IM
- 路由风格不统一：读单个任务是 `GET /users/tasks/:id`，改和删是 `PUT /tasks/:id`、`DELETE /tasks/:id`（历史遗留，改动要同步前端与文档）
- `AuthMiddleware` 查库失败一律返回 401，超时和数据库故障被误判为身份错误
- `users.name` 未加 UNIQUE 约束，同名账号可重复注册
- 登出为全设备失效，无法只踢掉单个设备

## 版本记录

**v1.3.0** —— 基于角色的授权模型（`role` + `viewer.canAccess`，跨员工访问 403）、任务截止时间与超期后台扫描（ticker + 有界 worker 池 + ctx 取消 + 优雅退出）、`schema.sql` 补三库升级段与 `idx_tasks_overdue` 索引

**v1.2.1** —— JWT 鉴权（bcrypt + token_version 撤销）、context 全链路 + 5 秒超时中间件 + 504 映射、连接池配置、Swagger 接口文档

**v1.2.0** —— 三层彻底解耦，业务校验下沉 Service，统一错误映射，测试库隔离

**v1.1.0** —— 拆分 Handler / Service / Repository / Model，构造函数依赖注入

**v1.0.0** —— 员工与任务 CRUD、基础参数校验、自动化测试

## 后续计划（v1.4）

- 统一响应结构（code / message / data）
- 任务列表分页与状态筛选
- 超期提醒接真实通知渠道，并加 in-flight 去重
- Dockerfile + compose，一条命令拉起 MySQL 与服务
- 数据库迁移工具，消除 `schema.sql` 中三库重复 DDL
- 接口日志与结构化日志
