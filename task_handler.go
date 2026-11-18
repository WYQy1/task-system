package main

import (
	"database/sql"
	"net/http" //这是 Go 标准库中的 HTTP 包
	"strconv"  //idStr->idint
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func parseTaskID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "任务ID格式错误"})
		return 0, false
	}
	return id, true
}

type CreateTaskRequest struct {
	Title       string `json:"title" binding:"required,max=100"`
	Description string `json:"description"`
}

func CreateTask(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "任务标题不能为空"})
		return
	}
	var userID int
	err := DB.QueryRow("SELECT id FROM users WHERE id = ?", id).Scan(&userID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"message": "员工不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库查询失败"})
		return
	}
	result, err := DB.Exec("INSERT INTO tasks (user_id,title,description) VALUES (?, ?, ?)", userID, req.Title, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "创建任务失败"})
		return
	}
	taskID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取任务id失败"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "任务创建成功", "id": taskID, "title": req.Title, "description": req.Description})
}

type Task struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"create_at"`
}
type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (Task, error) {
	var task Task
	err := scanner.Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.CreatedAt)
	if err != nil {
		return Task{}, err
	}
	return task, nil
}
func GetTasks(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}
	rows, err := DB.Query("SELECT id, user_id, title, description, status, created_at FROM tasks WHERE user_id = ? ORDER BY id ASC", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取任务失败"})
		return
	}
	defer rows.Close()
	tasks := make([]Task, 0) //没有值是默认返回0
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "读取任务数据失败"})
			return
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "遍历任务数据失败"})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func GetTask(c *gin.Context) {
	id, ok := parseTaskID(c)
	if !ok {
		return
	}
	task, err := scanTask(DB.QueryRow("SELECT id, user_id, title, description, status, created_at FROM tasks WHERE id = ?", id))
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"message": "任务不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询任务失败"})
		return
	}
	c.JSON(http.StatusOK, task)
}

type UpdateTaskRequest struct {
	Title       string `json:"title" binding:"required,max=100"`
	Description string `json:"description"`
	Status      string `json:"status" binding:"required"`
}

func isValidTaskStatus(status string) bool {
	switch status {
	case "pending", "in_progress", "completed":
		return true
	default:
		return false
	}
}
func UpdateTask(c *gin.Context) {
	id, ok := parseTaskID(c)
	if !ok {
		return
	}
	var uptask UpdateTaskRequest
	if err := c.ShouldBindJSON(&uptask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求数据格式错误"})
		return
	}
	uptask.Title = strings.TrimSpace(uptask.Title)
	if uptask.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "任务标题不能为空"})
		return
	}
	if !isValidTaskStatus(uptask.Status) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "任务状态不合法"})
		return
	}
	result, err := DB.Exec("UPDATE tasks SET title=?,description=?,status=? WHERE id=?", uptask.Title, uptask.Description, uptask.Status, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "更新任务失败"})
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取更新记录数失败"})
		return
	}
	if rowsAffected == 0 {
		err = DB.QueryRow("SELECT id FROM tasks WHERE id = ?", id).Scan(&id)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"message": "任务id不存在"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "查询任务失败"})
			return
		}
	}
	c.JSON(http.StatusOK, uptask)
}

func DeleteTask(c *gin.Context) {
	id, ok := parseTaskID(c)
	if !ok {
		return
	}
	result, err := DB.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "删除任务失败"})
		return
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取删除记录失败"})
		return
	}
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "任务不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功", "id": id})
}
