package handler

import (
	"net/http"

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
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID, ok := parseID(c, "员工ID格式错误")
	if !ok {
		return
	}
	var request CreateTaskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求格式错误"})
		return
	}
	task := model.Task{UserID: userID, Title: request.Title, Description: request.Description}
	id, err := h.taskService.Create(c.Request.Context(), task)
	if err != nil {
		respondError(c, err, "员工不存在", "创建任务失败")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "创建成功", "id": id})
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	userID, ok := parseID(c, "员工ID格式错误")
	if !ok {
		return
	}
	tasks, err := h.taskService.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		respondError(c, err, "员工不存在", "查询任务列表失败")
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) GetTask(c *gin.Context) {
	id, ok := parseID(c, "任务ID格式错误")
	if !ok {
		return
	}
	task, err := h.taskService.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err, "员工不存在", "查询任务失败")
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id, ok := parseID(c, "任务ID格式错误")
	if !ok {
		return
	}
	var request UpdateTaskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求数据格式错误"})
		return
	}
	task := model.Task{ID: id, Title: request.Title, Description: request.Description, Status: request.Status}
	_, err := h.taskService.Update(c.Request.Context(), task)
	if err != nil {
		respondError(c, err, "员工不存在", "更新任务失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "任务更新成功", "id": id})
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, ok := parseID(c, "任务ID格式错误")
	if !ok {
		return
	}
	_, err := h.taskService.Delete(c.Request.Context(), id)
	if err != nil {
		respondError(c, err, "员工不存在", "删除任务失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功", "id": id})
}
