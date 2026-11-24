package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
