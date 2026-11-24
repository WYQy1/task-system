package service

import (
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"task-system/model"
	"task-system/repository"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// openTestDB 打开测试数据库并注册关闭清理。
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	if err := godotenv.Load("../.env"); err != nil && !os.IsNotExist(err) {
		t.Fatalf("加载环境配置失败: %v", err)
	}

	dsn := os.Getenv("SERVICE_TEST_DB_DSN")
	require.NotEmpty(t, dsn, "SERVICE_TEST_DB_DSN 未设置")

	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err, "打开测试数据库失败")

	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("连接测试数据库失败: %v", err)
	}

	t.Cleanup(func() {
		assert.NoError(t, db.Close(), "关闭测试数据库失败")
	})
	return db
}

// newTestUserService 返回测试数据库和 UserService。
func newTestUserService(t *testing.T) (*sql.DB, *UserService) {
	t.Helper()
	db := openTestDB(t)
	return db, NewUserService(repository.NewUserRepository(db))
}

// newTestTaskService 返回测试数据库和 TaskService。
func newTestTaskService(t *testing.T) (*sql.DB, *TaskService) {
	t.Helper()
	db := openTestDB(t)
	return db, NewTaskService(repository.NewUserRepository(db), repository.NewTaskRepository(db))
}

// cleanupTestUser 为指定用户注册清理（先删任务，再删用户）。
func cleanupTestUser(t *testing.T, db *sql.DB, userID int64) {
	t.Helper()
	t.Cleanup(func() {
		_, err := db.Exec("DELETE FROM tasks WHERE user_id = ?", userID)
		assert.NoError(t, err, "清理用户任务失败")
		_, err = db.Exec("DELETE FROM users WHERE id = ?", userID)
		assert.NoError(t, err, "清理测试用户失败")
	})
}

// createTestUser 插入一个测试用户并注册清理。
func createTestUser(t *testing.T, db *sql.DB) int {
	t.Helper()
	result, err := db.Exec("INSERT INTO users (name, age, gender, password) VALUES (?, ?, ?, ?)", "TestUser", 25, "male", "test-password")
	require.NoError(t, err, "插入测试用户失败")
	userID, err := result.LastInsertId()
	require.NoError(t, err, "获取用户ID失败")
	cleanupTestUser(t, db, userID)
	return int(userID)
}

// createTestTask 通过 TaskService 创建测试任务并返回 ID。
func createTestTask(t *testing.T, taskService *TaskService, task model.Task) int {
	t.Helper()
	taskID, err := taskService.Create(task)
	require.NoError(t, err, "创建测试任务失败")
	return int(taskID)
}

// getUserFromDB 直接从数据库读取用户。
func getUserFromDB(t *testing.T, db *sql.DB, userID int64) model.User {
	t.Helper()
	var user model.User
	require.NoError(t, db.QueryRow("SELECT id, name, age, gender FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.Name, &user.Age, &user.Gender), "查询用户失败")
	return user
}

// deleteUserForTest 直接删除用户（用于构造 not found 场景）。
func deleteUserForTest(t *testing.T, db *sql.DB, id int) {
	t.Helper()
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	require.NoError(t, err, "删除测试用户失败")
}

// deleteTaskForTest 直接删除任务（用于构造 not found 场景）。
func deleteTaskForTest(t *testing.T, db *sql.DB, id int) {
	t.Helper()
	_, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	require.NoError(t, err, "删除测试任务失败")
}

// checkUser 校验用户各字段是否符合预期。
func checkUser(t *testing.T, u model.User, wantID int, wantName string, wantAge int, wantGender string) {
	t.Helper()
	assert.Equal(t, wantID, u.ID, "用户ID不符")
	assert.Equal(t, wantName, u.Name, "用户名不符")
	assert.Equal(t, wantAge, u.Age, "年龄不符")
	assert.Equal(t, wantGender, u.Gender, "性别不符")
}

// checkTask 校验任务各字段是否符合预期。
func checkTask(t *testing.T, task model.Task, wantID, wantUserID int, wantTitle, wantDesc, wantStatus string) {
	t.Helper()
	assert.Equal(t, wantID, task.ID, "任务ID不符")
	assert.Equal(t, wantUserID, task.UserID, "所属用户不符")
	assert.Equal(t, wantTitle, task.Title, "任务标题不符")
	assert.Equal(t, wantDesc, task.Description, "任务描述不符")
	assert.Equal(t, wantStatus, task.Status, "任务状态不符")
}
