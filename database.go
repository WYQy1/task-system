package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var DB *sql.DB

func initDBFromEnv(envKey string) error {
	// 加载项目目录下的 .env 文件
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("加载 .env 文件失败: %w", err)
	}

	// 获取数据库连接配置
	dsn := strings.TrimSpace(os.Getenv(envKey))
	if dsn == "" {
		return fmt.Errorf("%s 环境变量未设置", envKey)
	}

	// 创建数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("创建数据库连接失败: %w", err)
	}

	// 检查数据库连接是否正常
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
