package main

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"task-system/model"
)

func TestPing(t *testing.T) {
	r := SetupRouter()
	w := performRequest(r, "GET", "/ping", nil)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, `{"message":"pong"}`, w.Body.String())
}

func TestCreateUser(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()
	_, token := createTestAdminWithToken(t) // 新增员工是管理员权限
	body := `{"name":"TestUser","age":23,"gender":"男"}`
	w := performRequestWithToken(r, "POST", "/users", token, strings.NewReader(body))
	assert.Equal(t, 201, w.Code)
	var response struct {
		Message string `json:"message"`
		ID      int64  `json:"id"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	require.Positive(t, response.ID)
	t.Cleanup(func() {
		deleteTestUser(t, response.ID)
	})

	var user model.User
	require.NoError(t, DB.QueryRow("SELECT id,name,age,gender FROM users WHERE id = ?", response.ID).
		Scan(&user.ID, &user.Name, &user.Age, &user.Gender))
	checkUser(t, user, response.ID, "TestUser", 23, "male")
}

func TestCreateUserInvalid(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()
	_, token := createTestAdminWithToken(t)
	var before int
	require.NoError(t, DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&before))

	body := `{"name":"","age":23,"gender":"男"}`
	w := performRequestWithToken(r, "POST", "/users", token, strings.NewReader(body))
	assert.Equal(t, 400, w.Code)
	var after int
	require.NoError(t, DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&after))
	assert.Equal(t, before, after)
}

func TestGetUser(t *testing.T) {
	setupTestDB(t)
	id, token := createTestUserWithToken(t)
	r := SetupRouter()
	w := performRequestWithToken(r, "GET", "/users/"+strconv.FormatInt(id, 10), token, nil)
	assert.Equal(t, 200, w.Code)
	var user model.User
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &user))
	checkUser(t, user, id, "TestUser", 23, "male")
}

func TestGetUsers(t *testing.T) {
	setupTestDB(t)
	userID := createTestUser(t)
	_, token := createTestAdminWithToken(t) // 看全部员工需要管理员
	r := SetupRouter()
	w := performRequestWithToken(r, "GET", "/users", token, nil)
	assert.Equal(t, 200, w.Code)
	var users []model.User
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &users))
	found := false
	for _, user := range users {
		if int64(user.ID) == userID {
			found = true
			checkUser(t, user, userID, "TestUser", 23, "male")
		}
	}
	assert.True(t, found, "员工列表中没有找到刚创建的测试员工")
}

func TestGetUserNotFound(t *testing.T) {
	setupTestDB(t)
	token := testToken(t)
	r := SetupRouter()
	w := performRequestWithToken(r, "GET", "/users/99999", token, nil)
	assert.Equal(t, 404, w.Code)
}

func TestGetUserInvalidID(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()
	token := testToken(t)
	w := performRequestWithToken(r, "GET", "/users/abc", token, nil)
	assert.Equal(t, 400, w.Code)
}

func TestGetUsersOnlySelf(t *testing.T) {
	setupTestDB(t)
	_, err := DB.Exec("DELETE FROM tasks")
	require.NoError(t, err)
	_, err = DB.Exec("DELETE FROM users")
	require.NoError(t, err)

	// 两个员工都在库里，普通员工只能看到自己那一条
	userA, tokenA := createTestUserWithToken(t)
	userB := createTestUser(t)
	r := SetupRouter()

	w := performRequestWithToken(r, "GET", "/users", tokenA, nil)
	assert.Equal(t, 200, w.Code)
	var users []model.User
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &users))
	require.Len(t, users, 1, "普通员工不该看到其他员工")
	assert.Equal(t, userA, int64(users[0].ID))

	// 管理员能看到全部
	_, adminToken := createTestAdminWithToken(t)
	w = performRequestWithToken(r, "GET", "/users", adminToken, nil)
	assert.Equal(t, 200, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &users))
	ids := make(map[int64]bool, len(users))
	for _, u := range users {
		ids[int64(u.ID)] = true
	}
	assert.True(t, ids[userA], "管理员列表缺少A")
	assert.True(t, ids[userB], "管理员列表缺少B")
}

func TestCreateUserForbiddenForEmployee(t *testing.T) {
	setupTestDB(t)
	_, token := createTestUserWithToken(t) // 普通员工
	r := SetupRouter()

	var before int
	require.NoError(t, DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&before))
	body := `{"name":"越权新增","age":23,"gender":"男"}`
	w := performRequestWithToken(r, "POST", "/users", token, strings.NewReader(body))
	assert.Equal(t, 403, w.Code)
	assert.Equal(t, `{"message":"无权访问该员工的数据"}`, strings.TrimSpace(w.Body.String()))

	var after int
	require.NoError(t, DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&after))
	assert.Equal(t, before, after, "越权请求不该写库")
}

func TestUserCrossAccessForbidden(t *testing.T) {
	setupTestDB(t)
	_, tokenA := createTestUserWithToken(t)
	userB := createTestUser(t)
	r := SetupRouter()

	body := `{"name":"Hacked","age":30,"gender":"男"}`
	w := performRequestWithToken(r, "PUT", "/users/"+strconv.FormatInt(userB, 10), tokenA, strings.NewReader(body))
	assert.Equal(t, 403, w.Code)

	w = performRequestWithToken(r, "GET", "/users/"+strconv.FormatInt(userB, 10), tokenA, nil)
	assert.Equal(t, 403, w.Code)

	w = performRequestWithToken(r, "DELETE", "/users/"+strconv.FormatInt(userB, 10), tokenA, nil)
	assert.Equal(t, 403, w.Code)

	// B 没有被改、也没有被删
	var name string
	require.NoError(t, DB.QueryRow("SELECT name FROM users WHERE id = ?", userB).Scan(&name))
	assert.Equal(t, "TestUser", name)

	// 管理员可以跨员工查
	_, adminToken := createTestAdminWithToken(t)
	w = performRequestWithToken(r, "GET", "/users/"+strconv.FormatInt(userB, 10), adminToken, nil)
	assert.Equal(t, 200, w.Code)
}

func TestUpdateUser(t *testing.T) {
	setupTestDB(t)
	id, token := createTestUserWithToken(t)
	r := SetupRouter()
	body := `{"name":"TestUser","age":25,"gender":"男"}`
	w := performRequestWithToken(r, "PUT", "/users/"+strconv.FormatInt(id, 10), token, strings.NewReader(body))
	assert.Equal(t, 200, w.Code)
	var user model.User
	require.NoError(t, DB.QueryRow("SELECT id, name, age, gender FROM users WHERE id = ?", id).
		Scan(&user.ID, &user.Name, &user.Age, &user.Gender))
	checkUser(t, user, id, "TestUser", 25, "male")
}

func TestUpdateUserNotFound(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()
	token := testToken(t)
	body := `{"name":"TestUser","age":25,"gender":"男"}`
	w := performRequestWithToken(r, "PUT", "/users/99999", token, strings.NewReader(body))
	assert.Equal(t, 404, w.Code)
}

func TestUpdateUserInvalidID(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()
	token := testToken(t)
	body := `{"name":"TestUser","age":25,"gender":"男"}`
	w := performRequestWithToken(r, "PUT", "/users/abc", token, strings.NewReader(body))
	assert.Equal(t, 400, w.Code)
}

func TestUpdateUserInvalid(t *testing.T) {
	setupTestDB(t)
	id, token := createTestUserWithToken(t)
	r := SetupRouter()
	body := `{"name":"","age":25,"gender":"男"}`
	w := performRequestWithToken(r, "PUT", "/users/"+strconv.FormatInt(id, 10), token, strings.NewReader(body))
	assert.Equal(t, 400, w.Code)

	var user model.User
	require.NoError(t, DB.QueryRow("SELECT id, name, age, gender FROM users WHERE id = ?", id).
		Scan(&user.ID, &user.Name, &user.Age, &user.Gender))
	checkUser(t, user, id, "TestUser", 23, "male")
}

func TestDeleteUser(t *testing.T) {
	setupTestDB(t)
	id, token := createTestUserWithToken(t)
	r := SetupRouter()
	w := performRequestWithToken(r, "DELETE", "/users/"+strconv.FormatInt(id, 10), token, nil)
	assert.Equal(t, 200, w.Code)
	var gotID int
	err := DB.QueryRow("SELECT id FROM users WHERE id = ?", id).Scan(&gotID)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestDeleteUserNotFound(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()
	token := testToken(t)
	w := performRequestWithToken(r, "DELETE", "/users/99999", token, nil)
	assert.Equal(t, 404, w.Code)
}

func TestDeleteUserInvalidID(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()
	token := testToken(t)
	w := performRequestWithToken(r, "DELETE", "/users/abc", token, nil)
	assert.Equal(t, 400, w.Code)
}
func TestTokenVersionRevoked(t *testing.T) {
	setupTestDB(t)
	id, token := createTestUserWithToken(t)
	// 模拟"登出 / 踢下线"
	_, err := DB.Exec("UPDATE users SET token_version = token_version + 1 WHERE id = ?", id)
	require.NoError(t, err)
	r := SetupRouter()
	w := performRequestWithToken(r, "GET", "/users", token, nil)
	assert.Equal(t, 401, w.Code)
	assert.Equal(t, `{"message":"token 已失效，请重新登录"}`, w.Body.String())
}
