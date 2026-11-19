package main

import (
	"database/sql"
	"encoding/json"
	"net/http/httptest" //模拟 HTTP 请求和接收响应。
	"strconv"
	"strings"
	"task-system/model"
	"testing" //编写和执行自动化测试
)

func TestPing(t *testing.T) {
	r := SetupRouter()
	w := performRequest(r, "GET", "/ping", nil)
	if w.Code != 200 { //w.Code：获取实际 HTTP 响应状态码
		t.Errorf("状态码错误,期望 200,实际得到 %d", w.Code) //t.Errorf()：测试失败时报告错误，但不会立即终止测试
	}
	if w.Body.String() != `{"message":"pong"}` {
		t.Errorf("响应内容错误,期望pong,实际得到 %s", w.Body.String())
	}
}
func TestCreateUser(t *testing.T) {
	var response struct {
		Message string `json:"message"`
		ID      int64  `json:"id"`
	}
	setupTestDB(t)
	r := SetupRouter()
	body := `{"name":"TestUser","age":23,"gender":"男"}`
	w := performRequest(r, "POST", "/users", strings.NewReader(body))
	if w.Code != 201 {
		t.Errorf("状态码错误,期望 201,实际得到 %d", w.Code)
	}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("解析响应失败:%v", err)
	}
	var user model.User
	err = DB.QueryRow("SELECT id,name,age,gender FROM users WHERE id = ?", response.ID).Scan(&user.ID, &user.Name, &user.Age, &user.Gender)
	if err != nil {
		t.Fatalf("查询信息失败%v", err)
	}
	if user.ID != int(response.ID) {
		t.Errorf("员工ID不一致，数据库ID=%d，接口返回ID=%d", user.ID, response.ID)
	}
	// 清理本次测试创建的员工
	_, err = DB.Exec("DELETE FROM users WHERE id = ?", user.ID)
	if err != nil {
		t.Fatalf("清理测试数据失败：%v", err)
	}
}

func TestCreateUserInvalid(t *testing.T) {
	setupTestDB(t)
	var before int
	err := DB.QueryRow("SELECT COUNT(*) FROM users WHERE name = ?", "").Scan(&before)
	if err != nil {
		t.Fatalf("查询初始记录数量失败：%v", err)
	}
	r := SetupRouter()
	body := `{"name":"","age":23,"gender":"男"}`
	req := httptest.NewRequest("POST", "/users", strings.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatalf("状态码错误,期望400,实际得到 %d", w.Code)
	}
	var count int
	//这里的 COUNT(*) 无论有没有匹配记录，都会返回一行，因此可以使用 QueryRow()
	err = DB.QueryRow("SELECT COUNT(*) FROM users WHERE name = ?", "").Scan(&count)
	if err != nil {
		t.Fatalf("查询最终记录数量失败：%v", err)
	}
	if count != before {
		t.Fatalf("数据库记录数量异常，操作前=%d，操作后=%d", before, count)
	}
}

func TestGetUser(t *testing.T) {
	setupTestDB(t)
	id := createTestUser(t)
	defer deleteTestUser(t, id)
	r := SetupRouter()
	w := performRequest(r, "GET", "/users/"+strconv.FormatInt(id, 10), nil)
	if w.Code != 200 {
		t.Fatalf("状态码错误,期望200,实际得到 %d", w.Code)
	}
	var user model.User
	err := json.Unmarshal(w.Body.Bytes(), &user)
	if err != nil {
		t.Fatalf("解析响应失败:%v", err)
	}
	if int64(user.ID) != id {
		t.Errorf("员工ID不一致，数据库ID=%d，接口返回ID=%d", id, user.ID)
	}
	if user.Name != "TestUser" {
		t.Errorf("员工姓名不一致，实际得到 %s", user.Name)
	}
	if user.Age != 23 {
		t.Errorf("员工年龄不一致，实际得到 %d", user.Age)
	}
	if user.Gender != "男" {
		t.Errorf("员工性别不一致，实际得到 %s", user.Gender)
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

func TestUpdateUser(t *testing.T) {
	setupTestDB(t)
	id := createTestUser(t)
	defer deleteTestUser(t, id)
	r := SetupRouter()
	body := `{"name":"TestUser","age":25,"gender":"男"}`
	w := performRequest(r, "PUT", "/users/"+strconv.FormatInt(id, 10), strings.NewReader(body))
	if w.Code != 200 {
		t.Fatalf("状态码错误,希望得到200,实际得到%v", w.Code)
	}
	var user model.User
	err := DB.QueryRow("SELECT name,age,gender From users WHERE id=?", id).Scan(&user.Name, &user.Age, &user.Gender)
	if err != nil {
		t.Fatalf("查询数据失败 %v", err)
	}
	if user.Name != "TestUser" {
		t.Errorf("员工姓名不一致，实际得到 %s", user.Name)
	}
	if user.Age != 25 {
		t.Errorf("员工年龄不一致，实际得到 %d", user.Age)
	}
	if user.Gender != "男" {
		t.Errorf("员工性别不一致，实际得到 %s", user.Gender)
	}
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
	defer deleteTestUser(t, id)
	r := SetupRouter()
	body := `{"name":"","age":25,"gender":"男"}`
	w := performRequest(r, "PUT", "/users/"+strconv.FormatInt(id, 10), strings.NewReader(body))
	if w.Code != 400 {
		t.Fatalf("状态码错误,期望400,实际得到%d", w.Code)
	}
	var user model.User
	err := DB.QueryRow("SELECT name, age, gender FROM users WHERE id = ?", id).Scan(&user.Name, &user.Age, &user.Gender)
	if err != nil {
		t.Fatalf("查询数据失败：%v", err)
	}
	if user.Name != "TestUser" {
		t.Errorf("员工姓名发生异常变化，实际得到：%s", user.Name)
	}
	if user.Age != 23 {
		t.Errorf("员工年龄发生异常变化，实际得到：%d", user.Age)
	}
	if user.Gender != "男" {
		t.Errorf("员工性别发生异常变化，实际得到：%s", user.Gender)
	}
}

func TestDeleteUser(t *testing.T) {
	setupTestDB(t)
	id := createTestUser(t)
	r := SetupRouter()
	w := performRequest(r, "DELETE", "/users/"+strconv.FormatInt(id, 10), nil)
	if w.Code != 200 {
		t.Fatalf("状态码错误,期望200,实际得到%d", w.Code)
	}
	var user model.User
	err := DB.QueryRow("SELECT id FROM users WHERE id = ?", id).Scan(&user.ID)
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
