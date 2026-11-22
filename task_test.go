package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"task-system/model"
)

func TestCreateTask(t *testing.T) {
	setupTestDB(t)
	userID := createTestUser(t)
	r := SetupRouter()

	body := `{"title":"go任务测试","description":"自动化任务测试"}`
	w := performRequest(r, "POST", "/users/"+strconv.FormatInt(userID, 10)+"/tasks", strings.NewReader(body))
	if w.Code != 201 {
		t.Fatalf("状态码错误,期望201,实际得到%d", w.Code)
	}

	var response struct {
		Message string `json:"message"`
		ID      int64  `json:"id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("解析响应失败：%v", err)
	}
	if response.ID <= 0 {
		t.Fatalf("接口返回的任务ID无效：%d", response.ID)
	}

	var task model.Task
	if err := DB.QueryRow("SELECT id, user_id, title, description, status FROM tasks WHERE id = ?", response.ID).
		Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status); err != nil {
		t.Fatalf("查询数据失败：%v", err)
	}
	checkTask(t, task, response.ID, userID, "go任务测试", "自动化任务测试", "pending")
}

func TestCreateTaskUserNotFound(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()

	body := `{"title":"go任务测试","description":"自动化任务测试"}`
	w := performRequest(r, "POST", "/users/99999/tasks", strings.NewReader(body))
	if w.Code != 404 {
		t.Fatalf("状态码错误,期望404,实际得到%d", w.Code)
	}
	if w.Body.String() != `{"message":"员工不存在"}` {
		t.Errorf("响应内容错误,期望员工不存在,实际得到 %s", w.Body.String())
	}
}

func TestCreateTaskUserInvalidID(t *testing.T) {
	r := SetupRouter()
	w := performRequest(r, "POST", "/users/abc/tasks", nil)
	if w.Code != 400 {
		t.Fatalf("状态码错误,期望400,实际得到%d", w.Code)
	}
	if w.Body.String() != `{"message":"员工ID格式错误"}` {
		t.Errorf("响应内容错误，实际得到：%s", w.Body.String())
	}
}

func TestCreateTaskInvalidTitle(t *testing.T) {
	setupTestDB(t)
	userID := createTestUser(t)
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
			w := performRequest(r, "POST", "/users/"+strconv.FormatInt(userID, 10)+"/tasks", strings.NewReader(body))
			if w.Code != 400 {
				t.Fatalf("状态码错误,期望400,实际得到%d", w.Code)
			}

			var count int
			if err := DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE user_id = ?", userID).Scan(&count); err != nil {
				t.Fatalf("查询任务数量失败：%v", err)
			}
			if count != 0 {
				t.Errorf("非法标题不应该创建任务,实际创建了 %d 条", count)
			}
		})
	}
}

func TestGetTasks(t *testing.T) {
	setupTestDB(t)
	userID := createTestUser(t)
	createTestTask(t, userID, "Go任务测试1", "第一个测试任务")
	createTestTask(t, userID, "Go任务测试2", "第二个测试任务")
	r := SetupRouter()

	w := performRequest(r, "GET", "/users/"+strconv.FormatInt(userID, 10)+"/tasks", nil)
	if w.Code != 200 {
		t.Fatalf("状态码错误,期望200,实际得到%d", w.Code)
	}

	var tasks []model.Task
	if err := json.Unmarshal(w.Body.Bytes(), &tasks); err != nil {
		t.Fatalf("解析任务列表失败：%v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("任务数量错误，期望 2，实际得到 %d", len(tasks))
	}

	found1, found2 := false, false
	for _, task := range tasks {
		if task.Title == "Go任务测试1" {
			found1 = true
		}
		if task.Title == "Go任务测试2" {
			found2 = true
		}
		if int64(task.UserID) != userID {
			t.Errorf("任务归属错误，期望 UserID=%d，实际得到 %d", userID, task.UserID)
		}
	}
	if !found1 || !found2 {
		t.Fatalf("任务列表缺少预期任务")
	}
}

func TestGetTasksEmpty(t *testing.T) {
	setupTestDB(t)
	userID := createTestUser(t)
	r := SetupRouter()

	w := performRequest(r, "GET", "/users/"+strconv.FormatInt(userID, 10)+"/tasks", nil)
	if w.Code != 200 {
		t.Fatalf("状态码错误,期望200,实际得到%d", w.Code)
	}

	tasks := make([]model.Task, 0)
	if err := json.Unmarshal(w.Body.Bytes(), &tasks); err != nil {
		t.Fatalf("解析任务列表失败：%v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("预期返回空列表，实际返回 %d 条任务", len(tasks))
	}
}

func TestGetTasksUserNotFound(t *testing.T) {
	setupTestDB(t)
	userID := createTestUser(t)
	if _, err := DB.Exec("DELETE FROM users WHERE id = ?", userID); err != nil {
		t.Fatalf("删除测试员工失败：%v", err)
	}
	r := SetupRouter()

	w := performRequest(r, "GET", "/users/"+strconv.FormatInt(userID, 10)+"/tasks", nil)
	if w.Code != 404 {
		t.Fatalf("状态码错误，期望404，实际得到%d，响应：%s", w.Code, w.Body.String())
	}
	if strings.TrimSpace(w.Body.String()) != `{"message":"员工不存在"}` {
		t.Errorf("响应内容错误，期望员工不存在，实际得到：%s", w.Body.String())
	}
}

func TestGetTasksInvalidID(t *testing.T) {
	r := SetupRouter()
	w := performRequest(r, "GET", "/users/abc/tasks", nil)
	if w.Code != 400 {
		t.Fatalf("状态码错误,期望400,实际得到%d", w.Code)
	}
}

func TestGetTasksUserIsolation(t *testing.T) {
	setupTestDB(t)
	userAID := createTestUser(t)
	userBID := createTestUser(t)
	createTestTask(t, userAID, "Go任务测试1", "第一个测试任务")
	createTestTask(t, userAID, "Go任务测试2", "第二个测试任务")
	createTestTask(t, userBID, "Go任务测试3", "第三个测试任务")
	r := SetupRouter()

	w := performRequest(r, "GET", "/users/"+strconv.FormatInt(userAID, 10)+"/tasks", nil)
	if w.Code != 200 {
		t.Fatalf("状态码错误,期望200,实际得到%d", w.Code)
	}

	tasks := make([]model.Task, 0)
	if err := json.Unmarshal(w.Body.Bytes(), &tasks); err != nil {
		t.Fatalf("解析任务列表失败：%v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("预期2个列表,实际返回 %d 条任务", len(tasks))
	}
	for _, task := range tasks {
		if int64(task.UserID) != userAID {
			t.Errorf("任务归属错误，期望 UserID=%d，实际得到 %d", userAID, task.UserID)
		}
	}
}

func TestGetTask(t *testing.T) {
	setupTestDB(t)
	userID := createTestUser(t)
	taskID := createTestTask(t, userID, "Go单条任务测试", "测试单条任务查询")
	r := SetupRouter()

	w := performRequest(r, "GET", "/users/tasks/"+strconv.FormatInt(taskID, 10), nil)
	if w.Code != 200 {
		t.Fatalf("状态码错误，期望 200，实际得到 %d", w.Code)
	}

	var task model.Task
	if err := json.Unmarshal(w.Body.Bytes(), &task); err != nil {
		t.Fatalf("解析任务数据失败：%v", err)
	}
	checkTask(t, task, taskID, userID, "Go单条任务测试", "测试单条任务查询", "pending")
}

func TestGetTaskNotFound(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()
	w := performRequest(r, "GET", "/users/tasks/9999", nil)
	if w.Code != 404 {
		t.Fatalf("状态码错误,期望404,实际得到 %d", w.Code)
	}
}

func TestGetTaskInvalidID(t *testing.T) {
	r := SetupRouter()
	w := performRequest(r, "GET", "/users/tasks/abc", nil)
	if w.Code != 400 {
		t.Fatalf("状态码错误,期望400,实际得到 %d", w.Code)
	}
}

func TestDeleteTask(t *testing.T) {
	setupTestDB(t)
	userID := createTestUser(t)
	taskID := createTestTask(t, userID, "Go单条任务测试", "测试单条任务查询")
	r := SetupRouter()

	w := performRequest(r, "DELETE", "/tasks/"+strconv.FormatInt(taskID, 10), nil)
	if w.Code != 200 {
		t.Fatalf("状态码错误,期望200,实际得到%d", w.Code)
	}

	var id int
	err := DB.QueryRow("SELECT id FROM tasks WHERE id = ?", taskID).Scan(&id)
	if err == nil {
		t.Fatal("任务仍然存在，删除失败")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("查询数据库出现异常%v", err)
	}
}

func TestDeleteTaskNotFound(t *testing.T) {
	setupTestDB(t)
	r := SetupRouter()
	w := performRequest(r, "DELETE", "/tasks/99999", nil)
	if w.Code != 404 {
		t.Fatalf("状态码错误,期望404,实际得到%d", w.Code)
	}
}

func TestDeleteTaskInvalidID(t *testing.T) {
	r := SetupRouter()
	w := performRequest(r, "DELETE", "/tasks/abc", nil)
	if w.Code != 400 {
		t.Fatalf("状态码错误，期望400，实际得到%d", w.Code)
	}
}

func TestUpdateTask(t *testing.T) {
	setupTestDB(t)
	userID := createTestUser(t)
	taskID := createTestTask(t, userID, "测试任务", "原始描述")
	r := SetupRouter()

	body := `{"title":"更新后的任务","description":"更新后的描述","status":"completed"}`
	w := performRequest(r, "PUT", "/tasks/"+strconv.FormatInt(taskID, 10), strings.NewReader(body))
	if w.Code != 200 {
		t.Fatalf("状态码错误，期望200，实际得到%d，响应：%s", w.Code, w.Body.String())
	}

	var task model.Task
	if err := DB.QueryRow("SELECT id, user_id, title, description, status FROM tasks WHERE id = ?", taskID).
		Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status); err != nil {
		t.Fatalf("查询更新后的任务失败：%v", err)
	}
	checkTask(t, task, taskID, userID, "更新后的任务", "更新后的描述", "completed")
}

func TestUpdateTaskNotFound(t *testing.T) {
	setupTestDB(t)
	body := `{"title":"更新后的任务","description":"更新后的描述","status":"completed"}`
	r := SetupRouter()
	w := performRequest(r, "PUT", "/tasks/999999", strings.NewReader(body))
	if w.Code != 404 {
		t.Fatalf("状态码错误,期望404,实际得到%d,响应:%s", w.Code, w.Body.String())
	}
}

func TestUpdateTaskInvalidID(t *testing.T) {
	r := SetupRouter()
	body := `{"title":"测试任务","description":"测试描述","status":"pending"}`
	w := performRequest(r, "PUT", "/tasks/abc", strings.NewReader(body))
	if w.Code != 400 {
		t.Fatalf("状态码错误，期望400，实际得到%d", w.Code)
	}
}

func TestUpdateTaskInvalidInput(t *testing.T) {
	setupTestDB(t)
	userID := createTestUser(t)
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
			w := performRequest(r, "PUT", "/tasks/"+strconv.FormatInt(taskID, 10), strings.NewReader(body))
			if w.Code != 400 {
				t.Fatalf("状态码错误，期望400，实际得到%d，响应：%s", w.Code, w.Body.String())
			}

			var task model.Task
			if err := DB.QueryRow("SELECT id, user_id, title, description, status FROM tasks WHERE id = ?", taskID).
				Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status); err != nil {
				t.Fatalf("查询任务失败：%v", err)
			}
			checkTask(t, task, taskID, userID, "原始任务", "原始描述", "pending")
		})
	}
}
