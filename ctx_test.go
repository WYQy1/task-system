package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// 验证 ctx 超时真的能掐断一条正在执行的 SQL（而不是等它跑完）
func TestSlowQueryAbortedByContextDeadline(t *testing.T) {
	setupTestDB(t)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := DB.ExecContext(ctx, "SELECT SLEEP(2)")
	elapsed := time.Since(start)

	require.Error(t, err, "超时后 SQL 应返回错误")
	require.ErrorIs(t, err, context.DeadlineExceeded, "错误应能被 errors.Is 识别为超时")

	// 关键断言：必须在 SLEEP 的 2 秒之前就返回
	require.Less(t, elapsed, time.Second, "应在 1 秒内被掐断，而不是等 SLEEP(2) 跑完")
}
