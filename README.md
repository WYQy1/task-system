# 任务管理系统（Task System）

**v1.3.0** — Go + Gin + MySQL 的员工/任务管理后端，Handler / Service / Repository / Model 四层。v1.3.0 在 JWT 鉴权、context 全链路超时、Swagger 之上，补了按角色的授权模型与超期任务的背景扫描。

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

## 功能

- **员工**：增删改查；删除时事务级联删除其名下任务
- **任务**：按员工创建（可带截止时间 `due_at`）、列表、单条查询、更新、删除；外键关联
- **认证**：注册（bcrypt 哈希）、登录（24h JWT）、登出（`token_version` 撤销旧 token）
- **授权**：员工带 `role`（0 普通 / 1 管理员）。受保护接口统一从 `ctx` 取 `userID` 与 `role`，只能动自己的数据，跨员工访问 403；新增员工、查看全部员工仅管理员可用
- **超期提醒**：后台每分钟捞一批「未完成 + 已过截止时间 + 未提醒过」的任务，交给固定大小的 worker 池并发提醒；成功才写 `reminded_at`
- **文档**：Swagger UI，Authorize 后可直接调试受保护接口

## 项目架构

| 层 | 目录 | 职责 |
|---|---|---|
| Handler | `handler/` | 解析请求、调 Service、把错误映射为 HTTP 状态码 |
| Service | `service/` | 业务校验与编排，不写 SQL |
| Repository | `repository/` | SQL 与事务，不做业务判断 |
| Model | `model/` | 数据结构 |
| Middleware | `middleware/` | JWT 鉴权、请求超时 |
| Worker | `worker/` | 后台任务：超期扫描、并发分发、优雅退出（不在请求链路上） |

```text
Router → Timeout → Handler → Service → Repository → MySQL
```

`ctx` 从 `c.Request.Context()` 一路传到 `QueryRowContext`，任何一层断链，超时都穿不到数据库。

## v1.2.1 的三个关键设计

### 1. JWT 鉴权与 token 撤销

- `users` 表有 `password`、`token_version`
- 登录签发 `user_id` / `exp`（24 小时）/ `ver`
- `AuthMiddleware`：取 `Authorization` → 拆 `Bearer` → 验签 → 查库取版本号 → 比对 → 注入 `userID`、`role` → 放行
- 登出执行 `token_version + 1`，旧 token 版本号不符，立即 401
- 两个关卡分工：验签管「是否由本服务签发」，版本号管「是否仍然有效」

JWT 无状态、签发后收不回，回库比对版本号是唯一可行的撤销手段，代价是每个受保护请求多一次查询。

### 2. context 全链路与超时

- 13 个 repository 方法、14 个 service 方法全部以 `ctx context.Context` 为首参
- 全局 `middleware.Timeout(5 * time.Second)`：`context.WithTimeout` + `defer cancel()` 注入带截止时间的 ctx
- `errors.Is(err, context.DeadlineExceeded)` 映射为 504
- `SELECT SLEEP(2)` 在 100ms 超时下 100.39ms 被掐断（`ctx_test.go`）

取消是协作式的：`context` 只负责广播，只有监听 `ctx.Done()` 的代码（如 SQL 驱动）才会真正中止；不检查 ctx 的计算（如 bcrypt）照常跑完。

### 3. 连接池

```go
SetMaxOpenConns(25)
SetMaxIdleConns(10)
SetConnMaxLifetime(5 * time.Minute)
```

## v1.3.0 的两个新设计

### 1. 授权：认证答「你是谁」，授权答「你能动谁」

v1.2.1 只有认证。中间件验完 token 把 `userID` 放进 `ctx`，但业务 Handler 一个都没读——任何登录用户传别人的 id 就能改别人的数据。现在统一成一个判定点：

```go
type viewer struct{ ID, Role int }

func (v viewer) canAccess(ownerID int) bool { return v.ID == ownerID || v.isAdmin() }
```

- 中间件用一条 SQL 同时取回 `token_version` 和 `role`（`GetAuthInfo`），两个值都塞进 `ctx`，不多查一次库
- Handler 只做两件事：`currentViewer(c)` 取身份，`fetchUserForViewer` / `fetchTaskForViewer` 取数据并判归属
- **先判存在、再判归属**：不存在返回 404、存在但无权返回 403。否则「这个 id 存在但不是你的」和「压根不存在」会混成一个响应
- `TestGetUsersOnlySelf` 原本是假测试（清表后只建一个用户，再断言 `len == 1`），改成建两个用户：员工只看到自己，管理员看到全部

### 2. 后台 worker：并发放这里才有必要

批量插入用 goroutine 会被单条多值 `INSERT` 打败（并发抢同一把表锁）；看板按状态 fan-out 一条 `GROUP BY status` 就够。真正无法用一条 SQL 表达的是后台超期扫描：捞一批、逐个做外部动作、还要能取消。所以并发放这里：

```go
for i := 0; i < cfg.Workers; i++ {
    wg.Add(1)
    go func() { defer wg.Done(); for t := range jobs { s.handle(ctx, t) } }()
}
```

- 有界并发：`jobs` 缓冲 = `BatchSize`（50），worker 固定 4 个，扫描批量与并发度分开配
- 取消：`Run` 监听 `ctx.Done()`，收到就 `close(jobs)`，worker 排空队列再退出；`main` 里先 `srv.Shutdown`，再 `background.Wait()`
- 幂等靠数据库：`MarkReminded` 带 `AND reminded_at IS NULL`；提醒失败就不标，下一轮重扫——宁可重复，不可漏

取舍：单次提醒若比扫描周期还长，同一任务可能被两个 worker 同时提醒（数据库只保证 `reminded_at` 不被重复写）。彻底解决要进程内 in-flight 集合（单实例）或分布式锁（多实例）。

## 项目结构

```text
task-system/
├── handler/         # user / task / auth + helpers（错误映射）+ authz.go（授权判定）
├── service/         # 业务逻辑与校验
├── repository/      # SQL 与事务
├── model/           # User / Task
├── middleware/      # auth.go（JWT）、timeout.go（超时）
├── worker/          # overdue.go（超期扫描：ticker + worker pool + 优雅退出）
├── docs/            # swag init 生成物，勿手改
├── main.go          # 依赖注入、路由注册、优雅退出
├── database.go      # 连接初始化与连接池
├── schema.sql       # 正式库 + 两个测试库的表结构
├── *_test.go        # 根目录：HTTP 全链路集成测试
└── README.md
```

根目录的 `*_test.go` 都是 `package main`，必须与被测代码同目录——这是 Go 同包测试的约束，不是文件乱放。

## 快速开始

```bash
cp .env.example .env      # 填 DB_DSN / TEST_DB_DSN / SERVICE_TEST_DB_DSN / JWT_SECRET
mysql -u root -p < schema.sql
go mod tidy && go run .
```

- 服务：http://localhost:8080
- 文档：http://localhost:8080/swagger/index.html

从 v1.2.x 升级：执行 `schema.sql` 末尾「4. v1.2.x → v1.3.0 升级」的 `ALTER`（`ADD COLUMN` 不幂等，别重复跑），再把自己的账号提为管理员（注册接口固定写 `role = 0`，没开注册管理员的口子）：

```bash
mysql -u root -p task_system -e "UPDATE users SET role = 1 WHERE id = 1"
```

改完注解要重新生成文档：`swag init`。

## 测试与静态检查

```bash
go test -count=1 ./...
go vet ./...
gofmt -l .
```

83 个测试函数、18 组 `t.Run` 表驱动：根目录 42（HTTP 全链路）、`service/` 34、`worker/` 5、`middleware/` 2。

- 测试库隔离：`task_system_test` 给 HTTP 集成测试，`task_system_service_test` 给 Service 层
- `worker/` 用 fake store / fake notifier，不碰数据库，可单跑：`go test ./worker/ -count=1 -v`
- 并发部分用竞态检测器验过：`CGO_ENABLED=1 go test ./... -race -count=1`（需 gcc）

## 已知限制

- 超期提醒没有进程内 in-flight 去重：单次提醒超过扫描周期时可能重复提醒（数据库只保证 `reminded_at` 不被重复写）
- 提醒只打日志（`LogNotifier`），未接邮件 / IM
- 路由风格不统一：读单条是 `GET /users/tasks/:id`，改/删是 `PUT /tasks/:id`、`DELETE /tasks/:id`（历史遗留，改动要同步文档）
- `AuthMiddleware` 查库失败一律 401，超时与数据库故障被误判成身份错误
- `users.name` 无 UNIQUE 约束，同名可重复注册
- 登出为全设备失效，不能只踢单个设备

## 版本记录

**v1.3.0** — 基于角色的授权（`role` + `viewer.canAccess`，跨员工 403）、任务截止时间与超期后台扫描（ticker + 有界 worker 池 + ctx 取消 + 优雅退出）、`schema.sql` 补三库升级段与 `idx_tasks_overdue` 索引

**v1.2.1** — JWT 鉴权（bcrypt + token_version 撤销）、context 全链路 + 5 秒超时 + 504 映射、连接池、Swagger

**v1.2.0** — 三层彻底解耦，业务校验下沉 Service，统一错误映射，测试库隔离

**v1.1.0** — 拆分 Handler / Service / Repository / Model，构造函数依赖注入

**v1.0.0** — 员工与任务 CRUD、参数校验、自动化测试

## 后续计划（v1.4）

- 统一响应结构（code / message / data）
- 任务列表分页与状态筛选
- 超期提醒接真实通知渠道，并加 in-flight 去重
- Dockerfile + compose，一条命令拉起
- 数据库迁移工具，消除 `schema.sql` 的三库重复 DDL
- 结构化日志
