# task-system

基于 Go + Gin + MySQL 的后端接口练习项目，采用 RESTful 风格提供员工（及后续任务）管理接口。

> 当前进度：员工模块 CRUD 已完成；任务模块待开发。

## 技术栈

- Go 1.25
- Gin v1.12（Web 框架）
- MySQL（驱动：go-sql-driver/mysql）

## 目录结构

| 文件 | 说明 |
| --- | --- |
| `main.go` | 路由注册、数据库初始化、启动 HTTP 服务 |
| `handler.go` | 请求处理函数与业务逻辑（参数校验、数据库读写、响应） |
| `database.go` | MySQL 连接初始化（全局 `DB`） |
| `go.mod` / `go.sum` | Go 依赖声明与版本校验 |
| `note.md` | 学习笔记（参数获取、绑定校验等） |

## 已完成接口

服务默认监听 `:8080`。

| 方法 | 路径 | 说明 | 成功状态码 |
| --- | --- | --- | --- |
| GET | `/ping` | 健康检查，返回 `pong` | 200 |
| GET | `/hello?name=` | 问候接口，`name` 缺省为 `Guest` | 200 |
| POST | `/users` | 新增员工，ID 由 MySQL 自增生成 | 201 |
| GET | `/users` | 查询员工列表 | 200 |
| GET | `/users/:id` | 按 ID 查询单条员工 | 200 |
| PUT | `/users/:id` | 更新指定 ID 的员工信息 | 200 |
| DELETE | `/users/:id` | 删除指定 ID 的员工 | 200 |

### 新增员工请求示例

`POST /users`

```json
{
  "name": "Tom",
  "age": 22,
  "gender": "male"
}
```

## 参数与校验规则

- `name`：必填，去除首尾空格后不能为空，最长 20 个字符。
- `age`：取值范围 1 ~ 150。
- `gender`：可空，读取时用 `sql.NullString` 兼容数据库 NULL。

## 状态码约定

| 状态码 | 使用场景 |
| --- | --- |
| 200 OK | 查询、更新、删除成功 |
| 201 Created | 新增资源成功 |
| 400 Bad Request | 请求体格式错误、字段校验不通过、ID 格式错误、姓名为空 |
| 404 Not Found | 按 ID 查询 / 更新 / 删除的员工不存在 |
| 500 Internal Server Error | 数据库执行、读取或遍历过程中出现异常 |

## 当前进度

- [x] MySQL 数据库连接初始化与连通性检查
- [x] 员工新增（POST `/users`）
- [x] 员工列表查询（GET `/users`）
- [x] 按 ID 查询员工（GET `/users/:id`）
- [x] 更新员工信息（PUT `/users/:id`）
- [x] 删除员工（DELETE `/users/:id`）
- [x] 请求参数绑定与合法性校验
- [x] 统一错误处理与状态码返回

## 待办（TODO）

- [ ] 任务（task）模块接口：任务的增删改查、与员工的关联
- [ ] 删除员工前校验是否存在关联任务（当前未做关联拦截）
- [ ] 员工列表分页、模糊搜索等查询能力
- [ ] 单元测试 / 接口测试
- [ ] 统一响应结构与错误信息规范

## 本地运行

### 1. 环境要求

- Go 1.25+
- MySQL
- Git

### 2. 创建数据库

在 MySQL 中执行项目根目录下的 `schema.sql`，创建所需数据库和数据表。

### 3. 配置数据库连接

项目通过环境变量 `DB_DSN` 读取数据库连接信息。

Linux / WSL 环境下执行：

```bash
read -rsp "请输入 DB_DSN: " DB_DSN
echo
export DB_DSN
```

DSN 格式：

```text
用户名:密码@tcp(数据库主机:3306)/task_system
```

请根据自己的 MySQL 环境填写连接信息，不要将真实账号密码提交到 GitHub。

### 4. 安装依赖并启动

```bash
go mod tidy
go run .
```

服务默认监听：

```text
http://localhost:8080
```

健康检查：

```text
GET /ping
```

### 参考建表语句

统一以 schema.sql 为准

## 进度记录

- 2026-11-12：完成员工模块 CRUD、参数校验、数据库连接配置改造，准备首次推送 GitHub。
