package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"task-system/model"
	"task-system/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type CreateUserRequest struct {
	Name   string `json:"name"`
	Age    int    `json:"age"`
	Gender string `json:"gender"`
}

type UpdateUserRequest struct {
	Name   string `json:"name"`
	Age    int    `json:"age"`
	Gender string `json:"gender"`
}

// Ping 健康检查
//
//	@Summary	健康检查
//	@Tags		系统
//	@Produce	json
//	@Success	200	{object}	map[string]interface{}
//	@Router		/ping [get]
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

// GetHello 打招呼
//
//	@Summary	打招呼
//	@Tags		系统
//	@Produce	json
//	@Param		name	query	string	false	"名字，缺省 Guest"
//	@Success	200		{object}	map[string]interface{}
//	@Router		/hello [get]
func GetHello(c *gin.Context) {
	name := c.DefaultQuery("name", "Guest")
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("hello %s", name)})
}

// CreateUser 新增员工
//
//	@Summary		新增员工
//	@Description	创建一条员工记录
//	@Tags			用户
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateUserRequest	true	"员工信息"
//	@Success		201		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		401		{object}	map[string]interface{}
//	@Security		BearerAuth
//	@Router			/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var request CreateUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求格式错误"})
		return
	}
	user := model.User{Name: request.Name, Age: request.Age, Gender: request.Gender}
	id, err := h.userService.Create(c.Request.Context(), user)
	if err != nil {
		respondError(c, err, "", "员工添加失败")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "创建成功", "id": id})
}

// GetUser 查询单个员工
//
//	@Summary	查询单个员工
//	@Tags		用户
//	@Produce	json
//	@Param		id	path	int	true	"员工ID"
//	@Success	200	{object}	model.User
//	@Failure	404	{object}	map[string]interface{}
//	@Security	BearerAuth
//	@Router		/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id, ok := parseID(c, "员工ID格式错误")
	if !ok {
		return
	}
	user, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err, "员工不存在", "查询员工失败")
		return
	}
	c.JSON(http.StatusOK, user)
}

// GetUsers 员工列表
//
//	@Summary	员工列表
//	@Tags		用户
//	@Produce	json
//	@Success	200	{array}	model.User
//	@Security	BearerAuth
//	@Router		/users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.userService.GetAll(c.Request.Context())
	if err != nil {
		respondError(c, err, "", "查询员工失败")
		return
	}
	c.JSON(http.StatusOK, users)
}

// UpdateUser 更新员工
//
//	@Summary	更新员工
//	@Tags		用户
//	@Accept		json
//	@Produce	json
//	@Param		id		path	int					true	"员工ID"
//	@Param		request	body	UpdateUserRequest	true	"员工信息"
//	@Success	200		{object}	map[string]interface{}
//	@Failure	404		{object}	map[string]interface{}
//	@Security	BearerAuth
//	@Router		/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, ok := parseID(c, "员工ID格式错误")
	if !ok {
		return
	}
	var request UpdateUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求数据格式错误"})
		return
	}
	user := model.User{ID: id, Name: request.Name, Age: request.Age, Gender: request.Gender}
	_, err := h.userService.Update(c.Request.Context(), user)
	if err != nil {
		respondError(c, err, "员工id不存在", "员工信息更新失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "员工信息更新成功", "id": id})
}

// DeleteUser 删除员工
//
//	@Summary	删除员工（连同其任务）
//	@Tags		用户
//	@Produce	json
//	@Param		id	path	int	true	"员工ID"
//	@Success	200	{object}	map[string]interface{}
//	@Failure	404	{object}	map[string]interface{}
//	@Security	BearerAuth
//	@Router		/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, ok := parseID(c, "员工ID格式错误")
	if !ok {
		return
	}
	err := h.userService.Delete(c.Request.Context(), id)
	if err != nil {
		respondError(c, err, "删除的员工不存在", "删除员工失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功", "id": id})
}
