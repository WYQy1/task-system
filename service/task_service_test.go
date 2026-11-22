package service

import (
	"errors"
	"strings"
	"testing"

	"task-system/model"
)

func TestTaskServiceCreate(t *testing.T) {
	db, taskService := newTestTaskService(t)
	userID := createTestUser(t, db)
	taskID := createTestTask(t, taskService, model.Task{UserID: userID, Title: "  Go任务测试  ", Description: "测试任务描述"})

	task, err := taskService.GetByID(taskID)
	if err != nil {
		t.Fatalf("查询任务失败: %v", err)
	}
	checkTask(t, task, taskID, userID, "Go任务测试", "测试任务描述", "pending")
}

func TestTaskServiceCreateInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		task    model.Task
		wantErr error
	}{
		{"UserID为0", model.Task{UserID: 0, Title: "完成任务"}, ErrInvalidUserID},
		{"UserID为负数", model.Task{UserID: -1, Title: "完成任务"}, ErrInvalidUserID},
		{"标题为空", model.Task{UserID: 1, Title: ""}, ErrInvalidTaskTitle},
		{"标题全空格", model.Task{UserID: 1, Title: "   "}, ErrInvalidTaskTitle},
		{"标题超过100字", model.Task{UserID: 1, Title: strings.Repeat("任", 101)}, ErrInvalidTaskTitle},
	}

	taskService := NewTaskService(nil, nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := taskService.Create(tt.task); !errors.Is(err, tt.wantErr) {
				t.Errorf("期望错误 %v，实际得到 %v", tt.wantErr, err)
			}
		})
	}
}

func TestTaskServiceCreateUserNotFound(t *testing.T) {
	db, taskService := newTestTaskService(t)
	userID := createTestUser(t, db)
	deleteUserForTest(t, db, userID)

	if _, err := taskService.Create(model.Task{UserID: userID, Title: "Go任务测试", Description: "测试任务描述"}); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("期望 ErrUserNotFound，实际得到 %v", err)
	}
}

func TestTaskServiceGetByID(t *testing.T) {
	db, taskService := newTestTaskService(t)
	userID := createTestUser(t, db)
	taskID := createTestTask(t, taskService, model.Task{UserID: userID, Title: "  Go任务测试  ", Description: "测试任务描述"})

	task, err := taskService.GetByID(taskID)
	if err != nil {
		t.Fatalf("查询任务失败: %v", err)
	}
	checkTask(t, task, taskID, userID, "Go任务测试", "测试任务描述", "pending")
}

func TestTaskServiceGetByIDInvalidID(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		wantErr error
	}{
		{"ID为0", 0, ErrInvalidTaskID},
		{"负数ID", -1, ErrInvalidTaskID},
	}

	taskService := NewTaskService(nil, nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := taskService.GetByID(tt.id); !errors.Is(err, tt.wantErr) {
				t.Errorf("期望错误 %v，实际得到 %v", tt.wantErr, err)
			}
		})
	}
}

func TestTaskServiceGetByIDNotFound(t *testing.T) {
	db, taskService := newTestTaskService(t)
	userID := createTestUser(t, db)
	taskID := createTestTask(t, taskService, model.Task{UserID: userID, Title: "测试任务", Description: "测试任务描述"})
	deleteTaskForTest(t, db, taskID)

	if _, err := taskService.GetByID(taskID); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("期望 ErrTaskNotFound，实际得到 %v", err)
	}
}

func TestTaskServiceGetByUserID(t *testing.T) {
	db, taskService := newTestTaskService(t)
	userID := createTestUser(t, db)
	titles := []string{"任务一", "任务二"}

	for _, title := range titles {
		createTestTask(t, taskService, model.Task{UserID: userID, Title: title})
	}

	tasks, err := taskService.GetByUserID(userID)
	if err != nil {
		t.Fatalf("查询任务列表失败: %v", err)
	}
	if len(tasks) != len(titles) {
		t.Fatalf("期望 %d 个任务，实际 %d", len(titles), len(tasks))
	}

	for i, wantTitle := range titles {
		if tasks[i].UserID != userID {
			t.Errorf("任务 %d：期望 UserID %d，实际 %d", i, userID, tasks[i].UserID)
		}
		if tasks[i].Title != wantTitle {
			t.Errorf("任务 %d：期望标题 %q，实际 %q", i, wantTitle, tasks[i].Title)
		}
	}
}

func TestTaskServiceGetByUserIDInvalidID(t *testing.T) {
	tests := []struct {
		name    string
		userID  int
		wantErr error
	}{
		{"UserID为0", 0, ErrInvalidUserID},
		{"UserID为负数", -1, ErrInvalidUserID},
	}

	taskService := NewTaskService(nil, nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := taskService.GetByUserID(tt.userID); !errors.Is(err, tt.wantErr) {
				t.Errorf("期望错误 %v，实际得到 %v", tt.wantErr, err)
			}
		})
	}
}

func TestTaskServiceUpdate(t *testing.T) {
	db, taskService := newTestTaskService(t)
	userID := createTestUser(t, db)
	taskID := createTestTask(t, taskService, model.Task{UserID: userID, Title: "原始标题", Description: "原始描述"})

	rows, err := taskService.Update(model.Task{ID: taskID, UserID: userID, Title: " 修改后的标题 ", Description: "修改后的描述", Status: "completed"})
	if err != nil {
		t.Fatalf("更新任务失败: %v", err)
	}
	if rows <= 0 {
		t.Fatalf("期望影响行数大于 0，实际 %d", rows)
	}

	task, err := taskService.GetByID(taskID)
	if err != nil {
		t.Fatalf("查询任务失败: %v", err)
	}
	checkTask(t, task, taskID, userID, "修改后的标题", "修改后的描述", "completed")
}

func TestTaskServiceUpdateNoChanges(t *testing.T) {
	db, taskService := newTestTaskService(t)
	userID := createTestUser(t, db)
	taskID := createTestTask(t, taskService, model.Task{UserID: userID, Title: "测试任务", Description: "测试任务描述"})

	if _, err := taskService.Update(model.Task{ID: taskID, Title: "测试任务", Description: "测试任务描述", Status: "pending"}); err != nil {
		t.Fatalf("无变化更新不应报错: %v", err)
	}

	task, err := taskService.GetByID(taskID)
	if err != nil {
		t.Fatalf("任务应依然存在: %v", err)
	}
	checkTask(t, task, taskID, userID, "测试任务", "测试任务描述", "pending")
}

func TestTaskServiceUpdateInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		task    model.Task
		wantErr error
	}{
		{"ID为0", model.Task{ID: 0, Title: "测试任务", Status: "pending"}, ErrInvalidTaskID},
		{"ID为负数", model.Task{ID: -1, Title: "测试任务", Status: "pending"}, ErrInvalidTaskID},
		{"标题为空", model.Task{ID: 1, Title: "", Status: "pending"}, ErrInvalidTaskTitle},
		{"标题全空格", model.Task{ID: 1, Title: "   ", Status: "pending"}, ErrInvalidTaskTitle},
		{"标题超过100字", model.Task{ID: 1, Title: strings.Repeat("任", 101), Status: "pending"}, ErrInvalidTaskTitle},
		{"非法状态", model.Task{ID: 1, Title: "测试任务", Status: "invalid"}, ErrInvalidTaskStatus},
	}

	taskService := NewTaskService(nil, nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := taskService.Update(tt.task); !errors.Is(err, tt.wantErr) {
				t.Errorf("期望错误 %v，实际得到 %v", tt.wantErr, err)
			}
		})
	}
}

func TestTaskServiceUpdateNotFound(t *testing.T) {
	db, taskService := newTestTaskService(t)
	userID := createTestUser(t, db)
	taskID := createTestTask(t, taskService, model.Task{UserID: userID, Title: "测试任务", Description: "测试任务描述"})
	deleteTaskForTest(t, db, taskID)

	if _, err := taskService.Update(model.Task{ID: taskID, Title: "修改后的标题", Description: "修改后的描述", Status: "completed"}); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("期望 ErrTaskNotFound，实际得到 %v", err)
	}
}

func TestTaskServiceDelete(t *testing.T) {
	db, taskService := newTestTaskService(t)
	userID := createTestUser(t, db)
	taskID := createTestTask(t, taskService, model.Task{UserID: userID, Title: "测试任务", Description: "测试任务描述"})

	rowsAffected, err := taskService.Delete(taskID)
	if err != nil {
		t.Fatalf("删除任务失败: %v", err)
	}
	if rowsAffected <= 0 {
		t.Fatalf("期望影响行数大于 0，实际 %d", rowsAffected)
	}

	if _, err := taskService.GetByID(taskID); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("期望 ErrTaskNotFound，实际得到 %v", err)
	}
}

func TestTaskServiceDeleteInvalidID(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		wantErr error
	}{
		{"ID为0", 0, ErrInvalidTaskID},
		{"负数ID", -1, ErrInvalidTaskID},
	}

	taskService := NewTaskService(nil, nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := taskService.Delete(tt.id); !errors.Is(err, tt.wantErr) {
				t.Errorf("期望错误 %v，实际得到 %v", tt.wantErr, err)
			}
		})
	}
}

func TestTaskServiceDeleteNotFound(t *testing.T) {
	db, taskService := newTestTaskService(t)
	userID := createTestUser(t, db)
	taskID := createTestTask(t, taskService, model.Task{UserID: userID, Title: "测试任务", Description: "测试任务描述"})
	deleteTaskForTest(t, db, taskID)

	if _, err := taskService.Delete(taskID); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("期望 ErrTaskNotFound，实际得到 %v", err)
	}
}
