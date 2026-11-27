package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"task-system/model"
	"task-system/service"
)

type TaskHandler struct {
	taskService *service.TaskService
	userService *service.UserService // 用来校验 URL 上的员工 id 存在且可访问
}

func NewTaskHandler(taskService *service.TaskService, userService *service.UserService) *TaskHandler {
	return &TaskHandler{taskService: taskService, userService: userService}
}

type CreateTaskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	DueAt       *time.Time `json:"due_at"` // 截止时间，RFC3339，例如 2026-12-01T18:00:00+08:00；不传表示不设截止
}

type UpdateTaskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	DueAt       *time.Time `json:"due_at"`
}

// CreateTask 创建任务
//
//	@Summary	给员工创建任务
//	@Tags		任务
//	@Accept		json
//	@Produce	json
//	@Param		id		path	int					true	"员工ID"
//	@Param		request	body	CreateTaskRequest	true	"任务信息"
//	@Success	201		{object}	map[string]interface{}
//	@Failure	404		{object}	map[string]interface{}
//	@Failure	403	{object}	map[string]interface{} "无权访问该员工的数据"
//	@Security	BearerAuth
//	@Router		/users/{id}/tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	userID, ok := parseID(c, "员工ID格式错误")
	if !ok {
		return
	}
	// 员工不存在 → 404；存在但不是自己的（且不是管理员）→ 403
	if _, allowed := fetchUserForViewer(c, h.userService, userID); !allowed {
		return
	}
	var request CreateTaskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求格式错误"})
		return
	}
	task := model.Task{UserID: userID, Title: request.Title, Description: request.Description, DueAt: request.DueAt}
	id, err := h.taskService.Create(c.Request.Context(), task)
	if err != nil {
		respondError(c, err, "员工不存在", "创建任务失败")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "创建成功", "id": id})
}

// GetTasks 员工的任务列表
//
//	@Summary	员工的全部任务
//	@Tags		任务
//	@Produce	json
//	@Param		id	path	int	true	"员工ID"
//	@Success	200	{array}	model.Task
//	@Failure	404	{object}	map[string]interface{}
//	@Failure	403	{object}	map[string]interface{} "无权访问该员工的数据"
//	@Security	BearerAuth
//	@Router		/users/{id}/tasks [get]
func (h *TaskHandler) GetTasks(c *gin.Context) {
	userID, ok := parseID(c, "员工ID格式错误")
	if !ok {
		return
	}
	// 员工不存在 → 404；存在但不是自己的（且不是管理员）→ 403
	if _, allowed := fetchUserForViewer(c, h.userService, userID); !allowed {
		return
	}
	tasks, err := h.taskService.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		respondError(c, err, "员工不存在", "查询任务列表失败")
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// GetTask 查询单个任务
//
//	@Summary	查询单个任务
//	@Tags		任务
//	@Produce	json
//	@Param		id	path	int	true	"任务ID"
//	@Success	200	{object}	model.Task
//	@Failure	404	{object}	map[string]interface{}
//	@Failure	403	{object}	map[string]interface{} "无权访问该员工的数据"
//	@Security	BearerAuth
//	@Router		/users/tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	id, ok := parseID(c, "任务ID格式错误")
	if !ok {
		return
	}
	task, allowed := fetchTaskForViewer(c, h.taskService, id)
	if !allowed {
		return
	}
	c.JSON(http.StatusOK, task)
}

// UpdateTask 更新任务
//
//	@Summary	更新任务
//	@Tags		任务
//	@Accept		json
//	@Produce	json
//	@Param		id		path	int					true	"任务ID"
//	@Param		request	body	UpdateTaskRequest	true	"任务信息"
//	@Success	200		{object}	map[string]interface{}
//	@Failure	404		{object}	map[string]interface{}
//	@Failure	403	{object}	map[string]interface{} "无权访问该员工的数据"
//	@Security	BearerAuth
//	@Router		/tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id, ok := parseID(c, "任务ID格式错误")
	if !ok {
		return
	}
	// 先确认这条任务属于当前登录者（或登录者是管理员），再允许改
	task, allowed := fetchTaskForViewer(c, h.taskService, id)
	if !allowed {
		return
	}
	var request UpdateTaskRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求数据格式错误"})
		return
	}
	task.Title = request.Title
	task.Description = request.Description
	task.Status = request.Status
	task.DueAt = request.DueAt
	_, err := h.taskService.Update(c.Request.Context(), task)
	if err != nil {
		respondError(c, err, "员工不存在", "更新任务失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "任务更新成功", "id": id})
}

// DeleteTask 删除任务
//
//	@Summary	删除任务
//	@Tags		任务
//	@Produce	json
//	@Param		id	path	int	true	"任务ID"
//	@Success	200	{object}	map[string]interface{}
//	@Failure	404	{object}	map[string]interface{}
//	@Failure	403	{object}	map[string]interface{} "无权访问该员工的数据"
//	@Security	BearerAuth
//	@Router		/tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, ok := parseID(c, "任务ID格式错误")
	if !ok {
		return
	}
	if _, allowed := fetchTaskForViewer(c, h.taskService, id); !allowed {
		return
	}
	_, err := h.taskService.Delete(c.Request.Context(), id)
	if err != nil {
		respondError(c, err, "员工不存在", "删除任务失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功", "id": id})
}
