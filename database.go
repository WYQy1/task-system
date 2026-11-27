package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var DB *sql.DB

// utcSessionParams 让驱动和 MySQL 会话都以 UTC 为准：
//   - parseTime=true  DATETIME 直接扫进 time.Time，而不是 []byte
//   - loc=UTC         驱动解析/序列化时间时用 UTC
//   - time_zone='+00:00'  服务端 NOW()/CURRENT_TIMESTAMP 也按 UTC 计算
//
// 三者必须同时生效，只改其中一个会让写入和读取差出一个时区偏移。
var utcSessionParams = []string{"parseTime=true", "loc=UTC", "time_zone=%27%2B00%3A00%27"}

func dsnHasParam(dsn, key string) bool {
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

// withUTCSession 补齐时区参数；已有同名参数时不覆盖，允许调用方显式指定。
func withUTCSession(dsn string) string {
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	for _, param := range utcSessionParams {
		key := strings.SplitN(param, "=", 2)[0]
		if dsnHasParam(dsn, key) {
			continue
		}
		dsn += sep + param
		sep = "&"
	}
	return dsn
}

func initDBFromEnv(envKey string) error {
	// .env 只服务于本地开发；没有这个文件时继续读进程环境变量
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("加载 .env 文件失败: %w", err)
	}

	dsn := strings.TrimSpace(os.Getenv(envKey))
	if dsn == "" {
		return fmt.Errorf("%s 环境变量未设置", envKey)
	}

	db, err := sql.Open("mysql", withUTCSession(dsn))
	if err != nil {
		return fmt.Errorf("创建数据库连接失败: %w", err)
	}

	// 连接池参数：最大打开 25、最大空闲 10、单条连接最长存活 5 分钟
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	DB = db
	return nil
}

// 初始化正式数据库
func InitDB() error {
	return initDBFromEnv("DB_DSN")
}

// 初始化测试数据库
func InitTestDB() error {
	return initDBFromEnv("TEST_DB_DSN")
}
