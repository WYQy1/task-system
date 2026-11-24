package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"task-system/middleware"
	"task-system/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Age      int    `json:"age"`
	Gender   string `json:"gender"`
}

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (a *AuthHandler) Register(c *gin.Context) {
	var request RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "注册信息失败"})
		return
	}
	id, err := a.authService.Register(request.Name, request.Password, request.Age, request.Gender)
	if err != nil {
		respondError(c, err, "", "注册信息失败")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "注册成功", "id": id})
}
func (a *AuthHandler) Login(c *gin.Context) {
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "登录信息失败"})
		return
	}
	user, err := a.authService.Login(request.Name, request.Password)
	if err != nil {
		respondError(c, err, "", "登录信息失败")
		return
	}
	ver, err := a.authService.GetTokenVersion(user.ID)
	if err != nil {
		respondError(c, err, "", "获取token版本失败")
		return
	}
	token, err := middleware.GenerateToken(user.ID, ver)
	if err != nil {
		respondError(c, err, "", "获取token失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "登录成功", "token": token, "user": user})
}

func (a *AuthHandler) Logout(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未通过认证"})
		return
	}
	userID, ok := userIDValue.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "认证信息异常"})
		return
	}

	if err := a.authService.Logout(int(userID)); err != nil {
		respondError(c, err, "", "退出登录失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已退出登录"})
}
