package main
import (
    "io"
    "net/http"
    "net/http/httptest"
    "testing"
)
func setupTestDB(t *testing.T) {
	err := InitTestDB()
	if err != nil {
		t.Fatalf("连接测试数据库失败：%v", err)
	}
	t.Cleanup(func() {
		DB.Close()
	})
}
func createTestUser(t *testing.T) int64 {
    result, err := DB.Exec("INSERT INTO users (name, age, gender) VALUES (?, ?, ?)","TestUser", 23, "男")
    if err != nil {
        t.Fatalf("插入测试员工失败：%v", err)
    }
    userID, err := result.LastInsertId()
    if err != nil {
        t.Fatalf("获取员工ID失败：%v", err)
    }
    return userID
}
func deleteTestUser(t *testing.T, userID int64) {
    _, err := DB.Exec("DELETE FROM tasks WHERE user_id = ?", userID)
    if err != nil {
        t.Errorf("删除测试任务失败：%v", err)
    }
    _, err = DB.Exec("DELETE FROM users WHERE id = ?", userID)
    if err != nil {
        t.Errorf("删除测试员工失败：%v", err)
    }
}
func performRequest(r http.Handler,method string,path string,body io.Reader,) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}