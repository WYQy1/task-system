package main

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"task-system/model"
)

func TestPing(t *testing.T) {
	r := SetupRouter()
	w := performRequest(r, "GET", "/ping", nil)
	if w.Code != 200 {
		t.Errorf("状态码错误,期望 200,实际得到 %d", w.Code)
	}
	if w.Body.String() != `{"message":"pong"}` {
		t.Errorf("响应内容错误,期望pong,实际得到 %s", w.Body.String())
	}
}

func TestCreateUser(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()

	body := `{"name":"TestUser","age":23,"gender":"男"}`
	w := performRequest(r, "POST", "/users", strings.NewReader(body))
	if w.Code != 201 {
		t.Fatalf("状态码错误,期望 201,实际得到 %d", w.Code)
	}

	var response struct {
		Message string `json:"message"`
		ID      int64  `json:"id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("解析响应失败:%v", err)
	}
	if response.ID <= 0 {
		t.Fatalf("接口返回的员工ID无效：%d", response.ID)
	}
	t.Cleanup(func() {
		deleteTestUser(t, response.ID)
	})

	var user model.User
	if err := DB.QueryRow("SELECT id,name,age,gender FROM users WHERE id = ?", response.ID).
		Scan(&user.ID, &user.Name, &user.Age, &user.Gender); err != nil {
		t.Fatalf("查询信息失败:%v", err)
	}
	checkUser(t, user, response.ID, "TestUser", 23, "male")
}

func TestCreateUserInvalid(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()

	var before int
	if err := DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&before); err != nil {
		t.Fatalf("查询初始记录数量失败：%v", err)
	}

	body := `{"name":"","age":23,"gender":"男"}`
	w := performRequest(r, "POST", "/users", strings.NewReader(body))
	if w.Code != 400 {
		t.Fatalf("状态码错误,期望400,实际得到 %d", w.Code)
	}

	var after int
	if err := DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&after); err != nil {
		t.Fatalf("查询最终记录数量失败：%v", err)
	}
	if after != before {
		t.Fatalf("数据库记录数量异常，操作前=%d，操作后=%d", before, after)
	}
}

func TestGetUser(t *testing.T) {
	setupTestDB(t)
	id := createTestUser(t)
	r := SetupRouter()

	w := performRequest(r, "GET", "/users/"+strconv.FormatInt(id, 10), nil)
	if w.Code != 200 {
		t.Fatalf("状态码错误,期望200,实际得到 %d", w.Code)
	}

	var user model.User
	if err := json.Unmarshal(w.Body.Bytes(), &user); err != nil {
		t.Fatalf("解析响应失败:%v", err)
	}
	checkUser(t, user, id, "TestUser", 23, "male")
}

func TestGetUsers(t *testing.T) {
	setupTestDB(t)
	userID := createTestUser(t)
	r := SetupRouter()

	w := performRequest(r, "GET", "/users", nil)
	if w.Code != 200 {
		t.Fatalf("状态码错误,期望200,实际得到 %d", w.Code)
	}

	var users []model.User
	if err := json.Unmarshal(w.Body.Bytes(), &users); err != nil {
		t.Fatalf("解析响应失败:%v", err)
	}

	found := false
	for _, user := range users {
		if int64(user.ID) == userID {
			found = true
			checkUser(t, user, userID, "TestUser", 23, "male")
		}
	}
	if !found {
		t.Fatal("员工列表中没有找到刚创建的测试员工")
	}
}

func TestGetUserNotFound(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()
	w := performRequest(r, "GET", "/users/99999", nil)
	if w.Code != 404 {
		t.Fatalf("状态码错误,希望得到404,实际得到%v", w.Code)
	}
}

func TestGetUserInvalidID(t *testing.T) {
	r := SetupRouter()
	w := performRequest(r, "GET", "/users/abc", nil)
	if w.Code != 400 {
		t.Fatalf("状态码错误,希望得到400,实际得到%v", w.Code)
	}
}

func TestGetUsersEmpty(t *testing.T) {
	setupTestDB(t)
	if _, err := DB.Exec("DELETE FROM tasks"); err != nil {
		t.Fatalf("清理任务表失败: %v", err)
	}
	if _, err := DB.Exec("DELETE FROM users"); err != nil {
		t.Fatalf("清理员工表失败: %v", err)
	}
	r := SetupRouter()
	w := performRequest(r, "GET", "/users", nil)
	if w.Code != 200 {
		t.Fatalf("状态码错误，期望 200，实际得到 %d", w.Code)
	}

	var users []model.User
	if err := json.Unmarshal(w.Body.Bytes(), &users); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if users == nil {
		t.Fatal("期望返回空数组 []，实际返回 null")
	}
	if len(users) != 0 {
		t.Fatalf("期望员工列表为空，实际得到 %d 条记录", len(users))
	}
}

func TestUpdateUser(t *testing.T) {
	setupTestDB(t)
	id := createTestUser(t)
	r := SetupRouter()

	body := `{"name":"TestUser","age":25,"gender":"男"}`
	w := performRequest(r, "PUT", "/users/"+strconv.FormatInt(id, 10), strings.NewReader(body))
	if w.Code != 200 {
		t.Fatalf("状态码错误,希望得到200,实际得到%v", w.Code)
	}

	var user model.User
	if err := DB.QueryRow("SELECT id, name, age, gender FROM users WHERE id = ?", id).
		Scan(&user.ID, &user.Name, &user.Age, &user.Gender); err != nil {
		t.Fatalf("查询数据失败 %v", err)
	}
	checkUser(t, user, id, "TestUser", 25, "male")
}

func TestUpdateUserNotFound(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()

	body := `{"name":"TestUser","age":25,"gender":"男"}`
	w := performRequest(r, "PUT", "/users/99999", strings.NewReader(body))
	if w.Code != 404 {
		t.Fatalf("状态码错误,希望得到404,实际得到%v", w.Code)
	}
}

func TestUpdateUserInvalidID(t *testing.T) {
	r := SetupRouter()

	body := `{"name":"TestUser","age":25,"gender":"男"}`
	w := performRequest(r, "PUT", "/users/abc", strings.NewReader(body))
	if w.Code != 400 {
		t.Fatalf("状态码错误,希望得到400,实际得到%v", w.Code)
	}
}

func TestUpdateUserInvalid(t *testing.T) {
	setupTestDB(t)
	id := createTestUser(t)
	r := SetupRouter()

	body := `{"name":"","age":25,"gender":"男"}`
	w := performRequest(r, "PUT", "/users/"+strconv.FormatInt(id, 10), strings.NewReader(body))
	if w.Code != 400 {
		t.Fatalf("状态码错误,期望400,实际得到%d", w.Code)
	}

	var user model.User
	if err := DB.QueryRow("SELECT id, name, age, gender FROM users WHERE id = ?", id).
		Scan(&user.ID, &user.Name, &user.Age, &user.Gender); err != nil {
		t.Fatalf("查询数据失败：%v", err)
	}
	checkUser(t, user, id, "TestUser", 23, "male")
}

func TestDeleteUser(t *testing.T) {
	setupTestDB(t)
	id := createTestUser(t)
	r := SetupRouter()

	w := performRequest(r, "DELETE", "/users/"+strconv.FormatInt(id, 10), nil)
	if w.Code != 200 {
		t.Fatalf("状态码错误,期望200,实际得到%d", w.Code)
	}

	var gotID int
	err := DB.QueryRow("SELECT id FROM users WHERE id = ?", id).Scan(&gotID)
	if err == nil {
		t.Fatal("员工仍然存在，删除失败")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("查询数据库出现异常：%v", err)
	}
}

func TestDeleteUserNotFound(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()
	w := performRequest(r, "DELETE", "/users/99999", nil)
	if w.Code != 404 {
		t.Fatalf("状态码错误,期望404,实际得到%d", w.Code)
	}
}

func TestDeleteUserInvalidID(t *testing.T) {
	r := SetupRouter()
	w := performRequest(r, "DELETE", "/users/abc", nil)
	if w.Code != 400 {
		t.Fatalf("状态码错误,期望400,实际得到%d", w.Code)
	}
}
