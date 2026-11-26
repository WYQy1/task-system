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

// Register 注册
//
//	@Summary		注册新用户
//	@Description	密码以 bcrypt 哈希存储
//	@Tags			认证
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RegisterRequest	true	"注册信息"
//	@Success		201		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]interface{}
//	@Router			/auth/register [post]
func (a *AuthHandler) Register(c *gin.Context) {
	var request RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "注册信息失败"})
		return
	}
	id, err := a.authService.Register(c.Request.Context(), request.Name, request.Password, request.Age, request.Gender)
	if err != nil {
		respondError(c, err, "", "注册信息失败")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "注册成功", "id": id})
}

// Login 登录
//
//	@Summary		登录并获取 token
//	@Description	成功返回 JWT，之后请求头带 Authorization: Bearer <token>
//	@Tags			认证
//	@Accept			json
//	@Produce		json
//	@Param			request	body		LoginRequest	true	"用户名和密码"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		401		{object}	map[string]interface{}
//	@Router			/auth/login [post]
func (a *AuthHandler) Login(c *gin.Context) {
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "登录信息失败"})
		return
	}
	user, err := a.authService.Login(c.Request.Context(), request.Name, request.Password)
	if err != nil {
		respondError(c, err, "", "登录信息失败")
		return
	}
	ver, err := a.authService.GetTokenVersion(c.Request.Context(), user.ID)
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

// Logout 退出登录
//
//	@Summary		退出登录
//	@Description	token_version +1，该用户所有旧 token 立即失效
//	@Tags			认证
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Failure		401	{object}	map[string]interface{}
//	@Security			BearerAuth
//	@Router			/auth/logout [post]
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

	if err := a.authService.Logout(c.Request.Context(), int(userID)); err != nil {
		respondError(c, err, "", "退出登录失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已退出登录"})
}
