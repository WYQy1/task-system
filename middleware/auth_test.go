package middleware

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"task-system/repository"
)

// 查库炸了不等于"token 无效"：返回 401 会让客户端反复重新登录却永远登不上。
// 这里连一个必然拒绝连接的地址，断言中间件返回 500 而不是 401。
func TestAuthMiddlewareDatabaseErrorIs500(t *testing.T) {
	t.Setenv("JWT_SECRET", strings.Repeat("k", 32))
	gin.SetMode(gin.TestMode)

	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:1)/task_system?timeout=300ms")
	require.NoError(t, err)
	defer db.Close()

	token, err := GenerateToken(1, 0)
	require.NoError(t, err)

	r := gin.New()
	r.Use(AuthMiddleware(repository.NewUserRepository(db)))
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, `{"message":"服务内部错误"}`, strings.TrimSpace(w.Body.String()))
}
