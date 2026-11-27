package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"task-system/middleware"
	"task-system/model"
)

// TestMain 先给一份足长的测试密钥：InitJWTSecret 要求 ≥32 字节，
// 而本地 .env 里的开发密钥可能很短，不覆盖的话一启动就报错。
func TestMain(m *testing.M) {
	if len(strings.TrimSpace(os.Getenv("JWT_SECRET"))) < 32 {
		os.Setenv("JWT_SECRET", "test-secret-0123456789abcdef0123456789")
	}
	os.Exit(m.Run())
}

// users.name 上有唯一索引，助手账号必须用互不相同的名字。
var testUserSeq atomic.Int64

func nextTestUserName() string {
	return fmt.Sprintf("TestUser%d", testUserSeq.Add(1))
}

func setupTestDB(t *testing.T) {
	t.Helper()
	require.NoError(t, InitTestDB(), "连接测试数据库失败")
	t.Cleanup(func() {
		assert.NoError(t, DB.Close(), "关闭测试数据库失败")
	})
}

func testToken(t *testing.T) string {
	t.Helper()
	id := createTestUser(t)
	token, err := middleware.GenerateToken(int(id), 0)
	require.NoError(t, err, "生成测试token失败")
	return token
}

func createTestUser(t *testing.T) int64 {
	t.Helper()
	result, err := DB.Exec("INSERT INTO users (name, age, gender, password) VALUES (?, ?, ?, ?)", nextTestUserName(), 23, "male", "test-password")
	require.NoError(t, err, "插入测试员工失败")
	userID, err := result.LastInsertId()
	require.NoError(t, err, "获取员工ID失败")
	t.Cleanup(func() {
		deleteTestUser(t, userID)
	})
	return userID
}

// createTestUserWithRole 插入指定角色的员工，用来造管理员账号。
func createTestUserWithRole(t *testing.T, role int) int64 {
	t.Helper()
	result, err := DB.Exec("INSERT INTO users (name, age, gender, password, role) VALUES (?, ?, ?, ?, ?)", nextTestUserName(), 23, "male", "test-password", role)
	require.NoError(t, err, "插入测试员工失败")
	userID, err := result.LastInsertId()
	require.NoError(t, err, "获取员工ID失败")
	t.Cleanup(func() {
		deleteTestUser(t, userID)
	})
	return userID
}

// createTestAdminWithToken 造一个管理员并返回它的 token：新增员工、跨员工查询都要用它。
func createTestAdminWithToken(t *testing.T) (int64, string) {
	t.Helper()
	id := createTestUserWithRole(t, model.RoleAdmin)
	token, err := middleware.GenerateToken(int(id), 0)
	require.NoError(t, err, "生成测试token失败")
	return id, token
}

func createTestUserWithToken(t *testing.T) (int64, string) {
	t.Helper()
	id := createTestUser(t)
	token, err := middleware.GenerateToken(int(id), 0)
	require.NoError(t, err, "生成测试token失败")
	return id, token
}

// testUserName 读回助手账号的真实姓名：名字由 nextTestUserName 生成，测试里不能写死。
func testUserName(t *testing.T, userID int64) string {
	t.Helper()
	var name string
	require.NoError(t, DB.QueryRow("SELECT name FROM users WHERE id = ?", userID).Scan(&name))
	return name
}

func deleteTestUser(t *testing.T, userID int64) {
	t.Helper()
	_, err := DB.Exec("DELETE FROM tasks WHERE user_id = ?", userID)
	assert.NoError(t, err, "删除测试任务失败")
	_, err = DB.Exec("DELETE FROM users WHERE id = ?", userID)
	assert.NoError(t, err, "删除测试员工失败")
}

// createTestTask 直接插入任务；任务随所属用户的清理一起删除。
func createTestTask(t *testing.T, userID int64, title, description string) int64 {
	t.Helper()
	result, err := DB.Exec("INSERT INTO tasks (user_id, title, description) VALUES (?, ?, ?)", userID, title, description)
	require.NoError(t, err, "插入测试任务失败")
	taskID, err := result.LastInsertId()
	require.NoError(t, err, "获取任务ID失败")
	return taskID
}

func performRequest(r http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func performRequestWithToken(r http.Handler, method, path, token string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func checkUser(t *testing.T, u model.User, wantID int64, wantName string, wantAge int, wantGender string) {
	t.Helper()
	assert.Equal(t, wantID, int64(u.ID), "用户ID不符")
	assert.Equal(t, wantName, u.Name, "用户名不符")
	assert.Equal(t, wantAge, u.Age, "年龄不符")
	assert.Equal(t, wantGender, u.Gender, "性别不符")
}

func checkTask(t *testing.T, task model.Task, wantID, wantUserID int64, wantTitle, wantDescription, wantStatus string) {
	t.Helper()
	assert.Equal(t, wantID, int64(task.ID), "任务ID不符")
	assert.Equal(t, wantUserID, int64(task.UserID), "所属用户不符")
	assert.Equal(t, wantTitle, task.Title, "任务标题不符")
	assert.Equal(t, wantDescription, task.Description, "任务描述不符")
	assert.Equal(t, wantStatus, task.Status, "任务状态不符")
}
