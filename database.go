package main

import (
	"database/sql"
	"errors"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDB() error {
	// 加载项目目录下的 .env 文件
	if err := godotenv.Load(); err != nil {
		return err
	}

	dsn := os.Getenv("DB_DSN")

	if dsn == "" {
		return errors.New("DB_DSN 环境变量未设置")
	}

	var err error

	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	err = DB.Ping()
	if err != nil {
		return err
	}

	return nil
}