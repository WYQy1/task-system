package service

import (
	"database/sql"
	"os"
	"testing"

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
	if dsn == "" {
		t.Fatal("SERVICE_TEST_DB_DSN 未设置")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("连接测试数据库失败: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("关闭测试数据库失败: %v", err)
		}
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
		if _, err := db.Exec("DELETE FROM tasks WHERE user_id = ?", userID); err != nil {
			t.Errorf("清理用户任务失败: %v", err)
		}
		if _, err := db.Exec("DELETE FROM users WHERE id = ?", userID); err != nil {
			t.Errorf("清理测试用户失败: %v", err)
		}
	})
}

// createTestUser 插入一个测试用户并注册清理。
func createTestUser(t *testing.T, db *sql.DB) int {
	t.Helper()
	result, err := db.Exec("INSERT INTO users (name, age, gender) VALUES (?, ?, ?)", "TestUser", 25, "male")
	if err != nil {
		t.Fatalf("插入测试用户失败: %v", err)
	}
	userID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("获取用户ID失败: %v", err)
	}
	cleanupTestUser(t, db, userID)
	return int(userID)
}

// createTestTask 通过 TaskService 创建测试任务并返回 ID。
func createTestTask(t *testing.T, taskService *TaskService, task model.Task) int {
	t.Helper()
	taskID, err := taskService.Create(task)
	if err != nil {
		t.Fatalf("创建测试任务失败: %v", err)
	}
	return int(taskID)
}

// getUserFromDB 直接从数据库读取用户。
func getUserFromDB(t *testing.T, db *sql.DB, userID int64) model.User {
	t.Helper()
	var user model.User
	if err := db.QueryRow("SELECT id, name, age, gender FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.Name, &user.Age, &user.Gender); err != nil {
		t.Fatalf("查询用户失败: %v", err)
	}
	return user
}

// deleteUserForTest 直接删除用户（用于构造 not found 场景）。
func deleteUserForTest(t *testing.T, db *sql.DB, id int) {
	t.Helper()
	if _, err := db.Exec("DELETE FROM users WHERE id = ?", id); err != nil {
		t.Fatalf("删除测试用户失败: %v", err)
	}
}

// deleteTaskForTest 直接删除任务（用于构造 not found 场景）。
func deleteTaskForTest(t *testing.T, db *sql.DB, id int) {
	t.Helper()
	if _, err := db.Exec("DELETE FROM tasks WHERE id = ?", id); err != nil {
		t.Fatalf("删除测试任务失败: %v", err)
	}
}

// checkUser 校验用户各字段是否符合预期。
func checkUser(t *testing.T, u model.User, wantID int, wantName string, wantAge int, wantGender string) {
	t.Helper()
	if u.ID != wantID || u.Name != wantName || u.Age != wantAge || u.Gender != wantGender {
		t.Errorf("用户不符合预期: %+v", u)
	}
}

// checkTask 校验任务各字段是否符合预期。
func checkTask(t *testing.T, task model.Task, wantID, wantUserID int, wantTitle, wantDesc, wantStatus string) {
	t.Helper()
	if task.ID != wantID || task.UserID != wantUserID || task.Title != wantTitle ||
		task.Description != wantDesc || task.Status != wantStatus {
		t.Errorf("任务不符合预期: %+v", task)
	}
}
