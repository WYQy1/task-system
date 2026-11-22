package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"task-system/model"
)

func setupTestDB(t *testing.T) {
	if err := InitTestDB(); err != nil {
		t.Fatalf("连接测试数据库失败：%v", err)
	}
	t.Cleanup(func() {
		DB.Close()
	})
}

func createTestUser(t *testing.T) int64 {
	result, err := DB.Exec("INSERT INTO users (name, age, gender) VALUES (?, ?, ?)", "TestUser", 23, "male")
	if err != nil {
		t.Fatalf("插入测试员工失败：%v", err)
	}
	userID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("获取员工ID失败：%v", err)
	}
	t.Cleanup(func() {
		deleteTestUser(t, userID)
	})
	return userID
}

func deleteTestUser(t *testing.T, userID int64) {
	if _, err := DB.Exec("DELETE FROM tasks WHERE user_id = ?", userID); err != nil {
		t.Errorf("删除测试任务失败：%v", err)
	}
	if _, err := DB.Exec("DELETE FROM users WHERE id = ?", userID); err != nil {
		t.Errorf("删除测试员工失败：%v", err)
	}
}

// createTestTask 直接插入任务；任务随所属用户的清理一起删除。
func createTestTask(t *testing.T, userID int64, title, description string) int64 {
	result, err := DB.Exec("INSERT INTO tasks (user_id, title, description) VALUES (?, ?, ?)", userID, title, description)
	if err != nil {
		t.Fatalf("插入测试任务失败：%v", err)
	}
	taskID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("获取任务ID失败：%v", err)
	}
	return taskID
}

func performRequest(r http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func checkUser(t *testing.T, u model.User, wantID int64, wantName string, wantAge int, wantGender string) {
	t.Helper()
	if int64(u.ID) != wantID || u.Name != wantName || u.Age != wantAge || u.Gender != wantGender {
		t.Errorf("用户不符合预期: %+v", u)
	}
}

func checkTask(t *testing.T, task model.Task, wantID, wantUserID int64, wantTitle, wantDescription, wantStatus string) {
	t.Helper()
	if int64(task.ID) != wantID || int64(task.UserID) != wantUserID || task.Title != wantTitle ||
		task.Description != wantDescription || task.Status != wantStatus {
		t.Errorf("任务不符合预期: %+v", task)
	}
}
