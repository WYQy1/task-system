package main

import (
    "database/sql"
    "errors"
    "os"

    _ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDB() error {
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

