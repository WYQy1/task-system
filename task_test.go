package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestCreateTask(t *testing.T) {
	setupTestDB(t)
	userID := createTestUser(t)
	defer deleteTestUser(t, userID)
	r := SetupRouter()
	body := `{"title":"go任务测试","description":"自动化任务测试"}`
	w := performRequest(r, "POST", "/users/"+strconv.FormatInt(userID, 10)+"/tasks", strings.NewReader(body))
	if w.Code != 201 {
		t.Fatalf("状态码错误,期望201,实际得到%d", w.Code)
	}
	var task Task
	err := DB.QueryRow("SELECT id, user_id, title,description,status FROM tasks WHERE user_id = ?", userID).Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status)
	if err != nil {
		t.Fatalf("查询数据失败：%v", err)
	}
	if task.UserID != int(userID) {
		t.Errorf("员工ID错误,期望 %d,实际得到 %d", userID, task.UserID)
	}
	if task.Title != "go任务测试" {
		t.Errorf("任务标题错误，实际得到 %s", task.Title)
	}
	if task.Description != "自动化任务测试" {
		t.Errorf("任务描述错误，实际得到 %s", task.Description)
	}
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
	defer deleteTestUser(t, userID)
	body := `{"title":"","description":"自动化任务测试"}`
	r := SetupRouter()
	w := performRequest(r, "POST", "/users/"+strconv.FormatInt(userID, 10)+"/tasks", strings.NewReader(body))
	if w.Code != 400 {
		t.Fatalf("状态码错误,期望400,实际得到%d", w.Code)
	}
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE user_id = ?", userID).Scan(&count)
	if err != nil {
		t.Fatalf("查询任务数量失败：%v", err)
	}
	if count != 0 {
		t.Errorf("标题为空时不应该创建任务，实际创建了 %d 条", count)
	}
}

func TestCreateTaskTitleTooLong(t *testing.T) {
	setupTestDB(t)
	userID := createTestUser(t)
	defer deleteTestUser(t, userID)
	title := strings.Repeat("a", 101)
	body := fmt.Sprintf(`{"title":"%s","description":"自动化任务测试"}`, title)
	r := SetupRouter()
	w := performRequest(r, "POST", "/users/"+strconv.FormatInt(userID, 10)+"/tasks", strings.NewReader(body))
	if w.Code != 400 {
		t.Fatalf("状态码错误,期望400,实际得到%d", w.Code)
	}
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE user_id = ?", userID).Scan(&count)
	if err != nil {
		t.Fatalf("查询任务数量失败：%v", err)
	}
	if count != 0 {
		t.Errorf("标题为空时不应该创建任务，实际创建了 %d 条", count)
	}
}
func TestCreateTaskTitleEmpty(t *testing.T){
	setupTestDB(t)
	userID := createTestUser(t)
	defer deleteTestUser(t, userID)
	body := `{"title":"","description":"自动化任务测试"}`
	r := SetupRouter()
	w := performRequest(r, "POST", "/users/"+strconv.FormatInt(userID, 10)+"/tasks", strings.NewReader(body))
	if w.Code != 400 {
		t.Fatalf("状态码错误,期望400,实际得到%d", w.Code)
	}
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE user_id = ?", userID).Scan(&count)
	if err != nil {
		t.Fatalf("查询任务数量失败：%v", err)
	}
	if count != 0 {
		t.Errorf("标题为空时不应该创建任务，实际创建了 %d 条", count)
	}

}
//纯空白
func TestCreateTaskTitleWhitespace(t *testing.T) {
    setupTestDB(t)
    userID := createTestUser(t)
    defer deleteTestUser(t, userID)

    body := `{"title":"   ","description":"自动化任务测试"}`
    r := SetupRouter()
    w := performRequest(r, "POST", "/users/"+strconv.FormatInt(userID, 10)+"/tasks", strings.NewReader(body))

    if w.Code != 400 {
        t.Fatalf("状态码错误，期望400，实际得到%d，响应：%s", w.Code, w.Body.String())
    }

    var count int
    err := DB.QueryRow("SELECT COUNT(*) FROM tasks WHERE user_id = ?", userID).Scan(&count)
    if err != nil {
        t.Fatalf("查询任务数量失败：%v", err)
    }
    if count != 0 {
        t.Errorf("纯空格标题不应该创建任务，实际创建了%d条", count)
    }
}
func TestGetTasks(t *testing.T) {
	setupTestDB(t)
	userID := createTestUser(t)
	defer deleteTestUser(t, userID)
	_, err := DB.Exec("INSERT INTO tasks (user_id, title, description) VALUES (?, ?, ?)", userID, "Go任务测试1", "第一个测试任务")
	if err != nil {
		t.Fatalf("插入第一个任务失败：%v", err)
	}
	_, err = DB.Exec("INSERT INTO tasks (user_id, title, description) VALUES (?, ?, ?)", userID, "Go任务测试2", "第二个测试任务")
	if err != nil {
		t.Fatalf("插入第二个任务失败：%v", err)
	}
	r := SetupRouter()
	w := performRequest(r, "GET", "/users/"+strconv.FormatInt(userID, 10)+"/tasks", nil)
	if w.Code != 200 {
		t.Fatalf("状态码错误,期望200,实际得到%d", w.Code)
	}
	var tasks []Task
	err = json.Unmarshal(w.Body.Bytes(), &tasks)
	if err != nil {
		t.Fatalf("解析任务列表失败：%v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("任务数量错误，期望 2，实际得到 %d", len(tasks))
	}
	if tasks[0].Title != "Go任务测试1" {
		t.Errorf("第一条任务标题错误，实际得到：%s", tasks[0].Title)
	}
	if tasks[1].Title != "Go任务测试2" {
		t.Errorf("第二条任务标题错误，实际得到：%s", tasks[1].Title)
	}

	for _, task := range tasks {
		if task.UserID != int(userID) {
			t.Errorf("任务归属错误，期望 UserID=%d，实际得到 %d", userID, task.UserID)
		}
	}
}

func TestGetTasksEmpty(t *testing.T) {
	setupTestDB(t)
	userID := createTestUser(t)
	defer deleteTestUser(t, userID)
	r := SetupRouter()
	w := performRequest(r, "GET", "/users/"+strconv.FormatInt(userID, 10)+"/tasks", nil)
	if w.Code != 200 {
		t.Fatalf("状态码错误,期望200,实际得到%d", w.Code)
	}
	var tasks = make([]Task, 0)
	err := json.Unmarshal(w.Body.Bytes(), &tasks)
	if err != nil {
		t.Fatalf("解析任务列表失败：%v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("预期返回空列表，实际返回 %d 条任务", len(tasks))
	}
}

func TestGetTasksInvalidID(t *testing.T) {
	r := SetupRouter()
	w := performRequest(r, "GET", "/users/abc/tasks", nil)
	if w.Code != 400 {
		t.Fatalf("状态码错误,期望400,实际得到%d", w.Code)
	}
}

// 任务归属隔离
func TestGetTasksUserIsolation(t *testing.T) {
	setupTestDB(t)
	userAID := createTestUser(t)
	userBID := createTestUser(t)
	defer deleteTestUser(t, userAID)
	defer deleteTestUser(t, userBID)
	_, err := DB.Exec("INSERT INTO tasks (user_id, title, description) VALUES (?, ?, ?)", userAID, "Go任务测试1", "第一个测试任务")
	if err != nil {
		t.Fatalf("插入A员工第一个任务失败：%v", err)
	}
	_, err = DB.Exec("INSERT INTO tasks (user_id, title, description) VALUES (?, ?, ?)", userAID, "Go任务测试2", "第二个测试任务")
	if err != nil {
		t.Fatalf("插入A员工第二个任务失败：%v", err)
	}
	_, err = DB.Exec("INSERT INTO tasks (user_id, title, description) VALUES (?, ?, ?)", userBID, "Go任务测试3", "第三个测试任务")
	if err != nil {
		t.Fatalf("插入B员工测试任务失败：%v", err)
	}
	r := SetupRouter()
	w := performRequest(r, "GET", "/users/"+strconv.FormatInt(userAID, 10)+"/tasks", nil)
	if w.Code != 200 {
		t.Fatalf("状态码错误,期望200,实际得到%d", w.Code)
	}
	var tasks = make([]Task, 0)
	err = json.Unmarshal(w.Body.Bytes(), &tasks)
	if err != nil {
		t.Fatalf("解析任务列表失败：%v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("预期2个列表,实际返回 %d 条任务", len(tasks))
	}
	for _, task := range tasks {
		if task.UserID != int(userAID) {
			t.Errorf("任务归属错误，期望 UserID=%d，实际得到 %d", userAID, task.UserID)
		}
	}
}

func TestGetTask(t *testing.T) {
	setupTestDB(t)
	userID := createTestUser(t)
	defer deleteTestUser(t, userID)
	result, err := DB.Exec("INSERT INTO tasks (user_id, title, description) VALUES (?, ?, ?)", userID, "Go单条任务测试", "测试单条任务查询")
	if err != nil {
		t.Fatalf("插入测试任务失败：%v", err)
	}
	taskID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("获取任务ID失败：%v", err)
	}

	r := SetupRouter()
	w := performRequest(r, "GET", "/users/tasks/"+strconv.FormatInt(taskID, 10), nil)
	if w.Code != 200 {
		t.Fatalf("状态码错误，期望 200，实际得到 %d", w.Code)
	}
	var task Task
	err = json.Unmarshal(w.Body.Bytes(), &task)
	if err != nil {
		t.Fatalf("解析任务数据失败：%v", err)
	}
	if task.ID != int(taskID) {
		t.Errorf("任务ID错误，期望 %d，实际得到 %d", taskID, task.ID)
	}
	if task.UserID != int(userID) {
		t.Errorf("员工ID错误，期望 %d，实际得到 %d", userID, task.UserID)
	}
	if task.Title != "Go单条任务测试" {
		t.Errorf("任务标题错误，实际得到：%s", task.Title)
	}
	if task.Description != "测试单条任务查询" {
		t.Errorf("任务描述错误，实际得到：%s", task.Description)
	}
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
	defer deleteTestUser(t, userID)
	result, err := DB.Exec("INSERT INTO tasks (user_id, title, description) VALUES (?, ?, ?)", userID, "Go单条任务测试", "测试单条任务查询")
	if err != nil {
		t.Fatalf("插入测试任务失败 %v", err)
	}
	taskID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("获取任务id失败 %v", err)
	}
	r := SetupRouter()
	w := performRequest(r, "DELETE", "/tasks/"+strconv.FormatInt(taskID, 10), nil)
	if w.Code != 200 {
		t.Fatalf("状态码错误,期望200,实际得到%d", w.Code)
	}
	var id int
	err = DB.QueryRow("SELECT id FROM tasks WHERE id = ?", taskID).Scan(&id)
	if err != sql.ErrNoRows {
		t.Fatalf("删除任务失败%v", err)
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
func TestUpdateTaskInvalidStatus(t *testing.T) {
	setupTestDB(t)
	userID := createTestUser(t)
	defer deleteTestUser(t, userID)
	result, err := DB.Exec("INSERT INTO tasks (user_id, title, description, status) VALUES (?, ?, ?, ?)", userID, "测试任务", "测试描述", "pending")
	if err != nil {
		t.Fatalf("创建测试任务失败：%v", err)
	}
	taskID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("获取任务ID失败：%v", err)
	}
	body := `{"title": "测试任务","description": "测试描述", "status": "invalid" }`
	r := SetupRouter()
	w := performRequest(r, "PUT", "/tasks/"+strconv.FormatInt(taskID, 10), strings.NewReader(body))
	if w.Code != 400 {
		t.Fatalf("状态码错误，期望400，实际得到%d，响应：%s", w.Code, w.Body.String())
	}
	var status string
	err = DB.QueryRow("SELECT status FROM tasks WHERE id = ?", taskID).Scan(&status)
	if err != nil {
		t.Fatalf("查询任务状态失败：%v", err)
	}
	if status != "pending" {
		t.Fatalf("任务状态不应被修改，实际得到：%s", status)
	}
}
func TestUpdateTask(t *testing.T) {
	setupTestDB(t)
	userID := createTestUser(t)
	defer deleteTestUser(t, userID)
	result, err := DB.Exec("INSERT INTO tasks (user_id, title, description, status) VALUES (?, ?, ?, ?)", userID, "测试任务", "原始描述", "pending")
	if err != nil {
		t.Fatalf("创建测试任务失败：%v", err)
	}
	taskID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("获取任务ID失败：%v", err)
	}
	body := `{"title": "更新后的任务","description": "更新后的描述","status": "completed"}`
	r := SetupRouter()
	w := performRequest(r, "PUT", "/tasks/"+strconv.FormatInt(taskID, 10), strings.NewReader(body))
	if w.Code != 200 {
		t.Fatalf("状态码错误，期望200，实际得到%d，响应：%s", w.Code, w.Body.String())
	}
	var title, description, status string
	err = DB.QueryRow("SELECT title, description, status FROM tasks WHERE id = ?", taskID).Scan(&title, &description, &status)
	if err != nil {
		t.Fatalf("查询更新后的任务失败：%v", err)
	}
	if title != "更新后的任务" || description != "更新后的描述" || status != "completed" {
		t.Fatalf("任务更新结果错误:title=%s, description=%s, status=%s", title, description, status)
	}
}

func TestUpdateTaskNotFound(t *testing.T) {
	setupTestDB(t)
	body := `{"title": "更新后的任务","description": "更新后的描述","status": "completed"}`
	r := SetupRouter()
	w := performRequest(r, "PUT", "/tasks/999999", strings.NewReader(body))
	if w.Code != 404 {
		t.Fatalf("状态码错误,期望404,实际得到%d,响应:%s", w.Code, w.Body.String())
	}
}
func TestUpdateTaskInvalidTitle(t *testing.T) {
	setupTestDB(t)
	userID := createTestUser(t)
	defer deleteTestUser(t, userID)
	result, err := DB.Exec("INSERT INTO tasks (user_id, title, description, status) VALUES (?, ?, ?, ?)", userID, "原始任务", "原始描述", "pending")
	if err != nil {
		t.Fatalf("创建测试任务失败：%v", err)
	}
	taskID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("获取任务ID失败：%v", err)
	}
	body := `{"title": "", "description": "更新后的描述", "status": "completed"}`
	r := SetupRouter()
	w := performRequest(r, "PUT", "/tasks/"+strconv.FormatInt(taskID, 10), strings.NewReader(body))
	if w.Code != 400 {
		t.Fatalf("状态码错误，期望400，实际得到%d，响应：%s", w.Code, w.Body.String())
	}
}
func TestUpdateTaskTitleTooLong(t *testing.T) {
	setupTestDB(t)
	userID := createTestUser(t)
	defer deleteTestUser(t, userID)
	result, err := DB.Exec("INSERT INTO tasks (user_id, title, description, status) VALUES (?, ?, ?, ?)", userID, "原始任务", "原始描述", "pending")
	if err != nil {
		t.Fatalf("创建测试任务失败：%v", err)
	}
	taskID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("获取任务ID失败：%v", err)
	}
	body := `{"title": "` + strings.Repeat("a", 101) + `", "description": "更新后的描述", "status": "completed"}`
	r := SetupRouter()
	w := performRequest(r, "PUT", "/tasks/"+strconv.FormatInt(taskID, 10), strings.NewReader(body))
	if w.Code != 400 {
		t.Fatalf("状态码错误，期望400，实际得到%d，响应：%s", w.Code, w.Body.String())
	}
}
func TestUpdateTaskTitleEmpty(t *testing.T){
   setupTestDB(t)
   userID := createTestUser(t)
   defer deleteTestUser(t, userID)
   result, err := DB.Exec("INSERT INTO tasks (user_id, title, description, status) VALUES (?, ?, ?, ?)", userID, "原始任务", "原始描述", "pending")
	if err != nil {
		t.Fatalf("创建测试任务失败：%v", err)
	}
	taskID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("获取任务ID失败：%v", err)
	}
	body := `{"title":"", "description": "更新后的描述", "status": "completed"}`
	r := SetupRouter()
	w := performRequest(r, "PUT", "/tasks/"+strconv.FormatInt(taskID, 10), strings.NewReader(body))
	if w.Code != 400 {
		t.Fatalf("状态码错误，期望400，实际得到%d，响应：%s", w.Code, w.Body.String())
	}
	var title string
    err = DB.QueryRow("SELECT title FROM tasks WHERE id = ?", taskID).Scan(&title)
    if err != nil {
      t.Fatalf("查询任务失败：%v", err)
    }
    if title != "原始任务" {
      t.Errorf("标题不应该被修改，实际得到：%s", title)
    }
}
