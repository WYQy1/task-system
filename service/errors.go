package service

import "errors"

var (
	ErrInvalidUserName   = errors.New("员工姓名不能为空")
	ErrInvalidTaskTitle  = errors.New("任务标题不合法")
	ErrInvalidTaskStatus = errors.New("任务状态不合法")
	ErrUserNotFound      = errors.New("员工不存在")
	ErrTaskNotFound      = errors.New("任务不存在")
	ErrNameTooLong       = errors.New("名字超过最大长度")
	ErrInvalidUserAge    = errors.New("年龄不合要求")
	ErrInvalidUserGender = errors.New("性别不符合要求")
	ErrInvalidUserID     = errors.New("员工ID不合法")
	ErrInvalidTaskID     = errors.New("任务ID不合法")
	ErrInvalidPassword   = errors.New("密码不合法")
	ErrWrongPassword     = errors.New("密码错误")
	ErrNameExists        = errors.New("该姓名已被注册")
)
