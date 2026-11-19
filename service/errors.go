package service

import "errors"

var (
	ErrInvalidUserName   = errors.New("员工姓名不能为空")
	ErrInvalidTaskTitle  = errors.New("任务标题不合法")
	ErrInvalidTaskStatus = errors.New("任务状态不合法")
	ErrUserNotFound      = errors.New("员工不存在")
)
