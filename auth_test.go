package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogoutRevokesToken(t *testing.T) {
	setupTestDB(t)
	_, token := createTestUserWithToken(t)
	r := SetupRouter()

	w := performRequestWithToken(r, "POST", "/auth/logout", token, nil)
	assert.Equal(t, 200, w.Code)

	w = performRequestWithToken(r, "GET", "/users", token, nil)
	assert.Equal(t, 401, w.Code)
	assert.Equal(t, `{"message":"token 已失效，请重新登录"}`, w.Body.String())
}

func TestLogoutWithoutToken(t *testing.T) {
	r := SetupRouter()
	w := performRequest(r, "POST", "/auth/logout", nil)
	assert.Equal(t, 401, w.Code)
}

// 注册和代建账号共用同一套唯一约束：同名第二次注册应当 409。
func TestRegisterDuplicateName(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()
	body := `{"name":"RegUser","password":"secret123","age":23,"gender":"男"}`

	w := performRequest(r, "POST", "/auth/register", strings.NewReader(body))
	require.Equal(t, 201, w.Code)
	var created struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	t.Cleanup(func() { deleteTestUser(t, created.ID) })

	w = performRequest(r, "POST", "/auth/register", strings.NewReader(body))
	assert.Equal(t, 409, w.Code)
	assert.Equal(t, `{"message":"该姓名已被注册"}`, strings.TrimSpace(w.Body.String()))
}
