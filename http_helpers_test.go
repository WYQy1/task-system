package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"task-system/middleware"
	"task-system/model"
)

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
	token, err := middleware.GenerateToken(int(id), 0) // 中间件只验签名，不查库，随便给个 id 即可
	require.NoError(t, err, "生成测试token失败")
	return token
}

func createTestUser(t *testing.T) int64 {
	t.Helper()
	result, err := DB.Exec("INSERT INTO users (name, age, gender, password) VALUES (?, ?, ?, ?)", "TestUser", 23, "male", "test-password")
	require.NoError(t, err, "插入测试员工失败")
	userID, err := result.LastInsertId()
	require.NoError(t, err, "获取员工ID失败")
	t.Cleanup(func() {
		deleteTestUser(t, userID)
	})
	return userID
}

func createTestUserWithToken(t *testing.T) (int64, string) {
	t.Helper()
	id := createTestUser(t)
	token, err := middleware.GenerateToken(int(id), 0)
	require.NoError(t, err, "生成测试token失败")
	return id, token
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
