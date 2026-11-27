// Package dsn 统一 DSN 的会话参数，避免每个入口各写一遍。
package dsn

import "strings"

// utcSessionParams 让驱动和 MySQL 会话都以 UTC 为准：
//   - parseTime=true     DATETIME 直接扫进 time.Time，而不是 []byte
//   - loc=UTC            驱动解析/序列化时间时用 UTC
//   - time_zone='+00:00' 服务端 NOW()/CURRENT_TIMESTAMP 也按 UTC 计算
//
// 三者必须同时生效：只改其中一个，写入和读取之间就会差出一个时区偏移。
var utcSessionParams = []string{"parseTime=true", "loc=UTC", "time_zone=%27%2B00%3A00%27"}

func hasParam(dsn, key string) bool {
	i := strings.Index(dsn, "?")
	if i < 0 {
		return false
	}
	for _, kv := range strings.Split(dsn[i+1:], "&") {
		if strings.SplitN(kv, "=", 2)[0] == key {
			return true
		}
	}
	return false
}

// WithUTCSession 补齐时区参数；同名参数已存在时不覆盖，允许调用方显式指定。
func WithUTCSession(dsn string) string {
	dsn = strings.TrimSpace(dsn)
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	for _, param := range utcSessionParams {
		key := strings.SplitN(param, "=", 2)[0]
		if hasParam(dsn, key) {
			continue
		}
		dsn += sep + param
		sep = "&"
	}
	return dsn
}
