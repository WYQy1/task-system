package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"task-system/model"
)

func TestCreateTask(t *testing.T) {
	setupTestDB(t)
	userID, token := createTestUserWithToken(t)
	r := SetupRouter()

	body := `{"title":"go任务测试","description":"自动化任务测试"}`
	w := performRequestWithToken(r, "POST", "/users/"+strconv.FormatInt(userID, 10)+"/tasks", token, strings.NewReader(body))
	assert.Equal(t, 201, w.Code)

	var response struct {
		Message string `json:"message"`
		ID      int64  `json:"id"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	require.Positive(t, response.ID)

	var task model.Task
	require.NoError(t, DB.QueryRow("SELECT id, user_id, title, description, status FROM tasks WHERE id = ?", response.ID).
		Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status))
	checkTask(t, task, response.ID, userID, "go任务测试", "自动化任务测试", "pending")
}

func TestCreateTaskUserNotFound(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()
	token := testToken(t)
	body := `{"title":"go任务测试","description":"自动化任务测试"}`
	w := performRequestWithToken(r, "POST", "/users/99999/tasks", token, strings.NewReader(body))
	assert.Equal(t, 404, w.Code)
	assert.Equal(t, `{"message":"员工不存在"}`, w.Body.String())
}

func TestCreateTaskUserInvalidID(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()
	token := testToken(t)
	w := performRequestWithToken(r, "POST", "/users/abc/tasks", token, nil)
	assert.Equal(t, 400, w.Code)
	assert.Equal(t, `{"message":"员工ID格式错误"}`, w.Body.String())
}

func TestCreateTaskInvalidTitle(t *testing.T) {
	setupTestDB(t)
	userID, token := createTestUserWithToken(t)
	r := SetupRouter()
	tests := []struct {
		name  string
		title string
	}{
		{"标题为空", ""},
		{"标题全空格", "   "},
		{"标题超过100字", strings.Repeat("a", 101)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := fmt.Sprintf(`{"title":%q,"description":"自动化任务测试"}`, tt.title)
			w := performRequestWithToken(r, "POST", "/users/"+strconv.FormatInt(userID, 10)+"/tasks", token, strings.NewReader(body))
			assert.Equal(t, 400, w.Code)
			var count int
			require.NoError(t, DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE user_id = ?", userID).Scan(&count))
			assert.Zero(t, count, "非法标题不应该创建任务")
		})
	}
}

func TestGetTasks(t *testing.T) {
	setupTestDB(t)
	userID, token := createTestUserWithToken(t)
	createTestTask(t, userID, "Go任务测试1", "第一个测试任务")
	createTestTask(t, userID, "Go任务测试2", "第二个测试任务")
	r := SetupRouter()

	w := performRequestWithToken(r, "GET", "/users/"+strconv.FormatInt(userID, 10)+"/tasks", token, nil)
	assert.Equal(t, 200, w.Code)

	var tasks []model.Task
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &tasks))
	assert.Len(t, tasks, 2)

	found1, found2 := false, false
	for _, task := range tasks {
		if task.Title == "Go任务测试1" {
			found1 = true
		}
		if task.Title == "Go任务测试2" {
			found2 = true
		}
		assert.Equal(t, userID, int64(task.UserID), "任务归属错误")
	}
	assert.True(t, found1, "任务列表缺少 Go任务测试1")
	assert.True(t, found2, "任务列表缺少 Go任务测试2")
}

func TestGetTasksEmpty(t *testing.T) {
	setupTestDB(t)
	userID, token := createTestUserWithToken(t)
	r := SetupRouter()
	w := performRequestWithToken(r, "GET", "/users/"+strconv.FormatInt(userID, 10)+"/tasks", token, nil)
	assert.Equal(t, 200, w.Code)
	tasks := make([]model.Task, 0)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &tasks))
	assert.Len(t, tasks, 0)
}

func TestGetTasksUserNotFound(t *testing.T) {
	setupTestDB(t)
	userID, token := createTestUserWithToken(t)
	_, err := DB.Exec("DELETE FROM users WHERE id = ?", userID)
	require.NoError(t, err)
	r := SetupRouter()

	w := performRequestWithToken(r, "GET", "/users/"+strconv.FormatInt(userID, 10)+"/tasks", token, nil)
	assert.Equal(t, 401, w.Code)
	assert.Equal(t, `{"message":"token 无效"}`, strings.TrimSpace(w.Body.String()))
}

func TestGetTasksInvalidID(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()
	token := testToken(t)
	w := performRequestWithToken(r, "GET", "/users/abc/tasks", token, nil)
	assert.Equal(t, 400, w.Code)
}

func TestGetTasksUserIsolation(t *testing.T) {
	setupTestDB(t)
	userAID, token := createTestUserWithToken(t)
	userBID := createTestUser(t)
	createTestTask(t, userAID, "Go任务测试1", "第一个测试任务")
	createTestTask(t, userAID, "Go任务测试2", "第二个测试任务")
	createTestTask(t, userBID, "Go任务测试3", "第三个测试任务")
	r := SetupRouter()

	w := performRequestWithToken(r, "GET", "/users/"+strconv.FormatInt(userAID, 10)+"/tasks", token, nil)
	assert.Equal(t, 200, w.Code)

	tasks := make([]model.Task, 0)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &tasks))
	assert.Len(t, tasks, 2)
	for _, task := range tasks {
		assert.Equal(t, userAID, int64(task.UserID), "任务归属错误")
	}
}

func TestGetTask(t *testing.T) {
	setupTestDB(t)
	userID, token := createTestUserWithToken(t)
	taskID := createTestTask(t, userID, "Go单条任务测试", "测试单条任务查询")
	r := SetupRouter()

	w := performRequestWithToken(r, "GET", "/users/tasks/"+strconv.FormatInt(taskID, 10), token, nil)
	assert.Equal(t, 200, w.Code)

	var task model.Task
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &task))
	checkTask(t, task, taskID, userID, "Go单条任务测试", "测试单条任务查询", "pending")
}

func TestGetTaskNotFound(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()
	token := testToken(t)
	w := performRequestWithToken(r, "GET", "/users/tasks/9999", token, nil)
	assert.Equal(t, 404, w.Code)
}

func TestGetTaskInvalidID(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()
	token := testToken(t)
	w := performRequestWithToken(r, "GET", "/users/tasks/abc", token, nil)
	assert.Equal(t, 400, w.Code)
}

func TestDeleteTask(t *testing.T) {
	setupTestDB(t)
	userID, token := createTestUserWithToken(t)
	taskID := createTestTask(t, userID, "Go单条任务测试", "测试单条任务查询")
	r := SetupRouter()

	w := performRequestWithToken(r, "DELETE", "/tasks/"+strconv.FormatInt(taskID, 10), token, nil)
	assert.Equal(t, 200, w.Code)

	var id int
	err := DB.QueryRow("SELECT id FROM tasks WHERE id = ?", taskID).Scan(&id)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestDeleteTaskNotFound(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()
	token := testToken(t)
	w := performRequestWithToken(r, "DELETE", "/tasks/99999", token, nil)
	assert.Equal(t, 404, w.Code)
}

func TestDeleteTaskInvalidID(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()
	token := testToken(t)
	w := performRequestWithToken(r, "DELETE", "/tasks/abc", token, nil)
	assert.Equal(t, 400, w.Code)
}

func TestUpdateTask(t *testing.T) {
	setupTestDB(t)
	userID, token := createTestUserWithToken(t)
	taskID := createTestTask(t, userID, "测试任务", "原始描述")
	r := SetupRouter()

	body := `{"title":"更新后的任务","description":"更新后的描述","status":"completed"}`
	w := performRequestWithToken(r, "PUT", "/tasks/"+strconv.FormatInt(taskID, 10), token, strings.NewReader(body))
	assert.Equal(t, 200, w.Code)

	var task model.Task
	require.NoError(t, DB.QueryRow("SELECT id, user_id, title, description, status FROM tasks WHERE id = ?", taskID).
		Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status))
	checkTask(t, task, taskID, userID, "更新后的任务", "更新后的描述", "completed")
}

func TestUpdateTaskNotFound(t *testing.T) {
	setupTestDB(t)
	body := `{"title":"更新后的任务","description":"更新后的描述","status":"completed"}`
	r := SetupRouter()
	token := testToken(t)
	w := performRequestWithToken(r, "PUT", "/tasks/999999", token, strings.NewReader(body))
	assert.Equal(t, 404, w.Code)
}

func TestUpdateTaskInvalidID(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()
	body := `{"title":"测试任务","description":"测试描述","status":"pending"}`
	token := testToken(t)
	w := performRequestWithToken(r, "PUT", "/tasks/abc", token, strings.NewReader(body))
	assert.Equal(t, 400, w.Code)
}

func TestUpdateTaskInvalidInput(t *testing.T) {
	setupTestDB(t)
	userID, token := createTestUserWithToken(t)
	taskID := createTestTask(t, userID, "原始任务", "原始描述")
	r := SetupRouter()

	tests := []struct {
		name   string
		title  string
		status string
	}{
		{"标题为空", "", "completed"},
		{"标题超过100字", strings.Repeat("a", 101), "completed"},
		{"非法状态", "原始任务", "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := fmt.Sprintf(`{"title":%q,"description":"更新后的描述","status":%q}`, tt.title, tt.status)
			w := performRequestWithToken(r, "PUT", "/tasks/"+strconv.FormatInt(taskID, 10), token, strings.NewReader(body))
			assert.Equal(t, 400, w.Code)
			var task model.Task
			require.NoError(t, DB.QueryRow("SELECT id, user_id, title, description, status FROM tasks WHERE id = ?", taskID).
				Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status))
			checkTask(t, task, taskID, userID, "原始任务", "原始描述", "pending")
		})
	}
}

// TestTaskCrossUserForbidden 越权路径：A 的 token 碰 B 的员工 id / 任务 id，必须 403 且不改库。
func TestTaskCrossUserForbidden(t *testing.T) {
	setupTestDB(t)
	_, tokenA := createTestUserWithToken(t)
	userB, tokenB := createTestUserWithToken(t)
	taskB := createTestTask(t, userB, "B的任务", "不该被别人动")
	r := SetupRouter()

	// 在别人的名下建任务
	w := performRequestWithToken(r, "POST", "/users/"+strconv.FormatInt(userB, 10)+"/tasks", tokenA, strings.NewReader(`{"title":"越权创建","description":"x"}`))
	assert.Equal(t, 403, w.Code)

	// 列别人的任务
	w = performRequestWithToken(r, "GET", "/users/"+strconv.FormatInt(userB, 10)+"/tasks", tokenA, nil)
	assert.Equal(t, 403, w.Code)

	// 读 / 改 / 删别人的任务（注意：读用 /users/tasks/{id}，改和删用 /tasks/{id}，是仓库原有的路由风格）
	readPath := "/users/tasks/" + strconv.FormatInt(taskB, 10)
	mutatePath := "/tasks/" + strconv.FormatInt(taskB, 10)
	w = performRequestWithToken(r, "GET", readPath, tokenA, nil)
	assert.Equal(t, 403, w.Code)
	w = performRequestWithToken(r, "PUT", mutatePath, tokenA, strings.NewReader(`{"title":"越权修改","description":"x","status":"completed"}`))
	assert.Equal(t, 403, w.Code)
	w = performRequestWithToken(r, "DELETE", mutatePath, tokenA, nil)
	assert.Equal(t, 403, w.Code)

	// 越权之后 B 的任务原样还在
	var title, status string
	require.NoError(t, DB.QueryRow("SELECT title, status FROM tasks WHERE id = ?", taskB).Scan(&title, &status))
	assert.Equal(t, "B的任务", title)
	assert.Equal(t, "pending", status)

	// 主人自己读得到
	w = performRequestWithToken(r, "GET", readPath, tokenB, nil)
	assert.Equal(t, 200, w.Code)
}

// TestCreateTaskWithDueAt 截止时间按 RFC3339 进出，NULL 表示不设截止时间。
func TestCreateTaskWithDueAt(t *testing.T) {
	setupTestDB(t)
	userID, token := createTestUserWithToken(t)
	r := SetupRouter()

	body := `{"title":"带截止时间的任务","description":"x","due_at":"2026-01-01T10:00:00+08:00"}`
	w := performRequestWithToken(r, "POST", "/users/"+strconv.FormatInt(userID, 10)+"/tasks", token, strings.NewReader(body))
	require.Equal(t, 201, w.Code)

	var response struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

	tasksW := performRequestWithToken(r, "GET", "/users/tasks/"+strconv.FormatInt(response.ID, 10), token, nil)
	require.Equal(t, 200, tasksW.Code)
	var task model.Task
	require.NoError(t, json.Unmarshal(tasksW.Body.Bytes(), &task))
	require.NotNil(t, task.DueAt, "due_at 没有写进库")
	assert.True(t, task.DueAt.Equal(time.Date(2026, 1, 1, 2, 0, 0, 0, time.UTC)))
}
