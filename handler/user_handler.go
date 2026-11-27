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
	authService *service.AuthService
}

func NewUserHandler(userService *service.UserService, authService *service.AuthService) *UserHandler {
	return &UserHandler{userService: userService, authService: authService}
}

// CreateUserRequest 管理员代建账号。密码必填：没有密码的员工既登不上来，
// 也会在 users 表里留下一条 password 为空的脏数据。
type CreateUserRequest struct {
	Name     string `json:"name"`
	Age      int    `json:"age"`
	Gender   string `json:"gender"`
	Password string `json:"password"`
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
//	@Summary		新增员工（管理员）
//	@Description	管理员代建账号，密码必填（≥6 位，入库前 bcrypt 哈希）
//	@Tags			用户
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateUserRequest	true	"员工信息"
//	@Success		201		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		401		{object}	map[string]interface{}
//	@Failure		403		{object}	map[string]interface{} "无权访问该员工的数据"
//	@Failure		409		{object}	map[string]interface{} "该姓名已被注册"
//	@Security		BearerAuth
//	@Router			/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	v, authed := currentViewer(c)
	if !authed {
		return
	}
	if !v.isAdmin() {
		respondForbidden(c)
		return
	}
	var request CreateUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求格式错误"})
		return
	}
	id, err := h.authService.Register(c.Request.Context(), request.Name, request.Password, request.Age, request.Gender)
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
//	@Failure	403	{object}	map[string]interface{} "无权访问该员工的数据"
//	@Security	BearerAuth
//	@Router		/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id, ok := parseID(c, "员工ID格式错误")
	if !ok {
		return
	}
	user, allowed := fetchUserForViewer(c, h.userService, id)
	if !allowed {
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
//	@Failure	403	{object}	map[string]interface{} "无权访问该员工的数据"
//	@Security	BearerAuth
//	@Router		/users [get]
func (h *UserHandler) GetUsers(c *gin.Context) {
	v, authed := currentViewer(c)
	if !authed {
		return
	}
	// 列表同样不能泄露别人的数据：管理员看全部，普通员工只能看到自己。
	if !v.isAdmin() {
		user, err := h.userService.GetByID(c.Request.Context(), v.ID)
		if err != nil {
			respondError(c, err, "员工不存在", "查询员工失败")
			return
		}
		c.JSON(http.StatusOK, []model.User{user})
		return
	}
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
//	@Failure	403	{object}	map[string]interface{} "无权访问该员工的数据"
//	@Security	BearerAuth
//	@Router		/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, ok := parseID(c, "员工ID格式错误")
	if !ok {
		return
	}
	if _, allowed := fetchUserForViewer(c, h.userService, id); !allowed {
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
//	@Failure	403	{object}	map[string]interface{} "无权访问该员工的数据"
//	@Security	BearerAuth
//	@Router		/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, ok := parseID(c, "员工ID格式错误")
	if !ok {
		return
	}
	if _, allowed := fetchUserForViewer(c, h.userService, id); !allowed {
		return
	}
	// 越权校验通过后再执行：删除会级联删掉该员工名下的任务，只允许本人或管理员。
	err := h.userService.Delete(c.Request.Context(), id)
	if err != nil {
		respondError(c, err, "删除的员工不存在", "删除员工失败")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功", "id": id})
}
