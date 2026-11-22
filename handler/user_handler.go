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

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

func GetHello(c *gin.Context) {
	name := c.DefaultQuery("name", "Guest")
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("hello %s", name)})
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var request CreateUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求格式错误"})
		return
	}
	user := model.User{Name: request.Name, Age: request.Age, Gender: request.Gender}
	id, err := h.userService.Create(user)
	if err != nil {
		respondError(c, err, "", "员工添加失败")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "创建成功", "id": id})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id, ok := parseID(c, "员工ID格式错误")
	if !ok {
		return
	}
	user, err := h.userService.GetByID(id)
	if err != nil {
		respondError(c, err, "员工不存在", "查询员工失败")
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.userService.GetAll()
	if err != nil {
		respondError(c, err, "", "查询员工失败")
		return
	}
	c.JSON(http.StatusOK, users)
}

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
	_, err := h.userService.Update(user)
	if err != nil {
		respondError(c, err, "员工id不存在", "员工信息更新失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "员工信息更新成功", "id": id})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, ok := parseID(c, "员工ID格式错误")
	if !ok {
		return
	}
	err := h.userService.Delete(id)
	if err != nil {
		respondError(c, err, "删除的员工不存在", "删除员工失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功", "id": id})
}
