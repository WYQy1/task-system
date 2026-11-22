package service

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateUserID(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		wantErr error
	}{
		{"正常ID", 1, nil},
		{"ID为0", 0, ErrInvalidUserID},
		{"负数ID", -1, ErrInvalidUserID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateUserID(tt.id); !errors.Is(err, tt.wantErr) {
				t.Fatalf("期望错误 %v，实际得到 %v", tt.wantErr, err)
			}
		})
	}
}

func TestNormalizeAndValidateUserName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
		wantErr  error
	}{
		{"正常姓名", "张三", "张三", nil},
		{"首尾空格", " 张三 ", "张三", nil},
		{"空字符串", "", "", ErrInvalidUserName},
		{"超长姓名", strings.Repeat("测", 21), "", ErrNameTooLong},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeAndValidateUserName(tt.input)
			if got != tt.wantName {
				t.Errorf("期望姓名 %q，实际得到 %q", tt.wantName, got)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("期望错误 %v，实际得到 %v", tt.wantErr, err)
			}
		})
	}
}

func TestValidateUserAge(t *testing.T) {
	tests := []struct {
		name    string
		age     int
		wantErr error
	}{
		{"最小合法年龄", 1, nil},
		{"正常年龄", 23, nil},
		{"最大合法年龄", 150, nil},
		{"年龄为0", 0, ErrInvalidUserAge},
		{"负数年龄", -1, ErrInvalidUserAge},
		{"超过150岁", 151, ErrInvalidUserAge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateUserAge(tt.age); !errors.Is(err, tt.wantErr) {
				t.Errorf("期望错误 %v，实际得到 %v", tt.wantErr, err)
			}
		})
	}
}

func TestNormalizeAndValidateUserGender(t *testing.T) {
	tests := []struct {
		name       string
		gender     string
		wantGender string
		wantErr    error
	}{
		{"中文男性", "男", "male", nil},
		{"英文男性", "male", "male", nil},
		{"英文大写男性", "MALE", "male", nil},
		{"中文女性", "女", "female", nil},
		{"英文女性", "female", "female", nil},
		{"英文大写女性", "FEMALE", "female", nil},
		{"首尾空格", " male ", "male", nil},
		{"空字符串", "", "", nil},
		{"非法性别", "unknown", "", ErrInvalidUserGender},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeAndValidateUserGender(tt.gender)
			if got != tt.wantGender {
				t.Errorf("期望性别 %q，实际得到 %q", tt.wantGender, got)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("期望错误 %v，实际得到 %v", tt.wantErr, err)
			}
		})
	}
}

func TestValidateTaskID(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		wantErr error
	}{
		{"正常ID", 1, nil},
		{"较大ID", 100, nil},
		{"ID为0", 0, ErrInvalidTaskID},
		{"负数ID", -1, ErrInvalidTaskID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateTaskID(tt.id); !errors.Is(err, tt.wantErr) {
				t.Errorf("validateTaskID(%d) 应返回 %v，实际返回 %v", tt.id, tt.wantErr, err)
			}
		})
	}
}

func TestNormalizeAndValidateTaskTitle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{"正常标题", "完成任务", "完成任务", nil},
		{"首尾空格", " 完成任务 ", "完成任务", nil},
		{"空字符串", "", "", ErrInvalidTaskTitle},
		{"100个中文字符", strings.Repeat("任", 100), strings.Repeat("任", 100), nil},
		{"101个中文字符", strings.Repeat("任", 101), "", ErrInvalidTaskTitle},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeAndValidateTaskTitle(tt.input)
			if got != tt.want {
				t.Errorf("normalizeAndValidateTaskTitle(%q) 标题：期望 %q，实际 %q", tt.input, tt.want, got)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("normalizeAndValidateTaskTitle(%q) 错误：期望 %v，实际 %v", tt.input, tt.wantErr, err)
			}
		})
	}
}

func TestIsValidTaskStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{"pending", "pending", true},
		{"in_progress", "in_progress", true},
		{"completed", "completed", true},
		{"非法状态", "invalid", false},
		{"空字符串", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidTaskStatus(tt.status); got != tt.want {
				t.Errorf("isValidTaskStatus(%q) 期望 %v，实际 %v", tt.status, tt.want, got)
			}
		})
	}
}
