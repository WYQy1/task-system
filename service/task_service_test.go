package service

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"task-system/model"
)

func TestTaskServiceCreate(t *testing.T) {
	db, taskService := newTestTaskService(t)
	userID := createTestUser(t, db)
	taskID := createTestTask(t, taskService, model.Task{UserID: userID, Title: "  Go任务测试  ", Description: "测试任务描述"})
	task, err := taskService.GetByID(context.Background(), taskID)
	require.NoError(t, err)
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
			_, err := taskService.Create(context.Background(), tt.task)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestTaskServiceCreateUserNotFound(t *testing.T) {
	db, taskService := newTestTaskService(t)
	userID := createTestUser(t, db)
	deleteUserForTest(t, db, userID)
	_, err := taskService.Create(context.Background(), model.Task{UserID: userID, Title: "Go任务测试", Description: "测试任务描述"})
	require.ErrorIs(t, err, ErrUserNotFound)
}

func TestTaskServiceGetByID(t *testing.T) {
	db, taskService := newTestTaskService(t)
	userID := createTestUser(t, db)
	taskID := createTestTask(t, taskService, model.Task{UserID: userID, Title: "  Go任务测试  ", Description: "测试任务描述"})
	task, err := taskService.GetByID(context.Background(), taskID)
	require.NoError(t, err)
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
			_, err := taskService.GetByID(context.Background(), tt.id)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestTaskServiceGetByIDNotFound(t *testing.T) {
	db, taskService := newTestTaskService(t)
	userID := createTestUser(t, db)
	taskID := createTestTask(t, taskService, model.Task{UserID: userID, Title: "测试任务", Description: "测试任务描述"})
	deleteTaskForTest(t, db, taskID)
	_, err := taskService.GetByID(context.Background(), taskID)
	require.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskServiceGetByUserID(t *testing.T) {
	db, taskService := newTestTaskService(t)
	userID := createTestUser(t, db)
	titles := []string{"任务一", "任务二"}
	for _, title := range titles {
		createTestTask(t, taskService, model.Task{UserID: userID, Title: title})
	}
	tasks, err := taskService.GetByUserID(context.Background(), userID)
	require.NoError(t, err)
	assert.Len(t, tasks, len(titles))
	for i, wantTitle := range titles {
		assert.Equal(t, userID, tasks[i].UserID, "任务 %d：UserID 不符", i)
		assert.Equal(t, wantTitle, tasks[i].Title, "任务 %d：标题不符", i)
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
			_, err := taskService.GetByUserID(context.Background(), tt.userID)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestTaskServiceUpdate(t *testing.T) {
	db, taskService := newTestTaskService(t)
	userID := createTestUser(t, db)
	taskID := createTestTask(t, taskService, model.Task{UserID: userID, Title: "原始标题", Description: "原始描述"})
	rows, err := taskService.Update(context.Background(), model.Task{ID: taskID, UserID: userID, Title: " 修改后的标题 ", Description: "修改后的描述", Status: "completed"})
	require.NoError(t, err)
	require.Positive(t, rows)
	task, err := taskService.GetByID(context.Background(), taskID)
	require.NoError(t, err)
	checkTask(t, task, taskID, userID, "修改后的标题", "修改后的描述", "completed")
}

func TestTaskServiceUpdateNoChanges(t *testing.T) {
	db, taskService := newTestTaskService(t)
	userID := createTestUser(t, db)
	taskID := createTestTask(t, taskService, model.Task{UserID: userID, Title: "测试任务", Description: "测试任务描述"})
	_, err := taskService.Update(context.Background(), model.Task{ID: taskID, Title: "测试任务", Description: "测试任务描述", Status: "pending"})
	require.NoError(t, err)
	task, err := taskService.GetByID(context.Background(), taskID)
	require.NoError(t, err)
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
			_, err := taskService.Update(context.Background(), tt.task)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestTaskServiceUpdateNotFound(t *testing.T) {
	db, taskService := newTestTaskService(t)
	userID := createTestUser(t, db)
	taskID := createTestTask(t, taskService, model.Task{UserID: userID, Title: "测试任务", Description: "测试任务描述"})
	deleteTaskForTest(t, db, taskID)
	_, err := taskService.Update(context.Background(), model.Task{ID: taskID, Title: "修改后的标题", Description: "修改后的描述", Status: "completed"})
	require.ErrorIs(t, err, ErrTaskNotFound)
}

func TestTaskServiceDelete(t *testing.T) {
	db, taskService := newTestTaskService(t)
	userID := createTestUser(t, db)
	taskID := createTestTask(t, taskService, model.Task{UserID: userID, Title: "测试任务", Description: "测试任务描述"})
	rowsAffected, err := taskService.Delete(context.Background(), taskID)
	require.NoError(t, err)
	require.Positive(t, rowsAffected)
	_, err = taskService.GetByID(context.Background(), taskID)
	require.ErrorIs(t, err, ErrTaskNotFound)
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
			_, err := taskService.Delete(context.Background(), tt.id)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestTaskServiceDeleteNotFound(t *testing.T) {
	db, taskService := newTestTaskService(t)
	userID := createTestUser(t, db)
	taskID := createTestTask(t, taskService, model.Task{UserID: userID, Title: "测试任务", Description: "测试任务描述"})
	deleteTaskForTest(t, db, taskID)
	_, err := taskService.Delete(context.Background(), taskID)
	require.ErrorIs(t, err, ErrTaskNotFound)
}
