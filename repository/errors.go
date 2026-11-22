package repository

import "errors"

var (
	ErrUserNotFound = errors.New("员工不存在")
	ErrTaskNotFound = errors.New("任务不存在")
)
