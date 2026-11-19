package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"task-system/model"
	"task-system/service"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

type CreateTaskRequest struct {
	Title       string `json:"title" binding:"required,max=100"`
	Description string `json:"description"`
}

type UpdateTaskRequest struct {
	Title       string `json:"title" binding:"required,max=100"`
	Description string `json:"description"`
	Status      string `json:"status" binding:"required"`
}

func parseTaskID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "任务ID格式错误"})
		return 0, false
	}
	return id, true
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "员工ID格式错误"})
		return
	}

	var request CreateTaskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求格式错误"})
		return
	}

	task := model.Task{UserID: userID, Title: strings.TrimSpace(request.Title), Description: request.Description}
	id, err := h.taskService.Create(task)
	if errors.Is(err, service.ErrInvalidTaskTitle) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "任务标题不合法"})
		return
	}
	if errors.Is(err, service.ErrUserNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": "员工不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "任务创建失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "创建成功", "id": id})
}
func (h *TaskHandler) GetTasks(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil || userID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "员工ID格式错误"})
		return
	}

	tasks, err := h.taskService.GetByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询任务失败"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	id, ok := parseTaskID(c)
	if !ok {
		return
	}

	task, err := h.taskService.GetByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"message": "任务不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "查询任务失败"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id, ok := parseTaskID(c)
	if !ok {
		return
	}

	var request UpdateTaskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求数据格式错误"})
		return
	}

	task := model.Task{ID: id, Title: strings.TrimSpace(request.Title), Description: request.Description, Status: request.Status}
	_, err := h.taskService.Update(task)
	if errors.Is(err, service.ErrInvalidTaskTitle) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "任务标题不能为空"})
		return
	}
	if errors.Is(err, service.ErrInvalidTaskStatus) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "任务状态不合法"})
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"message": "任务不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "任务更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "任务更新成功", "id": id, "data": request})
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, ok := parseTaskID(c)
	if !ok {
		return
	}
	rowsAffected, err := h.taskService.Delete(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "删除任务失败"})
		return
	}
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "任务不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功", "id": id})
}
