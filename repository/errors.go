package repository

import "errors"

var (
	ErrUserNotFound = errors.New("员工不存在")
	ErrTaskNotFound = errors.New("任务不存在")
	// ErrDuplicateName 由 users.name 的唯一索引触发（MySQL 1062），
	// 单独提出来是为了让上层能返回 409 而不是笼统的 500。
	ErrDuplicateName = errors.New("员工姓名已存在")
)
