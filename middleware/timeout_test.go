package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newTimeoutRouter(d time.Duration) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Timeout(d))
	return r
}

// 验证中间件真的给请求 ctx 挂了 deadline
func TestTimeoutAddsDeadlineToRequestContext(t *testing.T) {
	r := newTimeoutRouter(100 * time.Millisecond)

	var hasDeadline bool
	var remaining time.Duration
	r.GET("/check", func(c *gin.Context) {
		deadline, ok := c.Request.Context().Deadline()
		hasDeadline = ok
		if ok {
			remaining = time.Until(deadline)
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/check", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.True(t, hasDeadline, "请求 ctx 应带 deadline")
	require.Positive(t, remaining, "deadline 应在未来")
	require.LessOrEqual(t, remaining, 100*time.Millisecond, "剩余时间不应超过设定值")
	require.Equal(t, http.StatusOK, w.Code)
}

// 验证超时到点后 ctx 真的被 cancel
func TestTimeoutCancelsRequestContext(t *testing.T) {
	r := newTimeoutRouter(30 * time.Millisecond)

	var ctxErr error
	r.GET("/wait", func(c *gin.Context) {
		<-c.Request.Context().Done()
		ctxErr = c.Request.Context().Err()
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/wait", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.ErrorIs(t, ctxErr, context.DeadlineExceeded)
	require.Equal(t, http.StatusOK, w.Code)
}
