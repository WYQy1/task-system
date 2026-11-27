package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"task-system/middleware"
	"task-system/model"
	"task-system/service"
)

// viewer 是"当前登录者"，只由认证中间件的解析结果构造。
// 授权判断一律以它为准，绝不用 URL / body 里的 id 当作身份，
// 否则把 /users/1/tasks 改成 /users/2/tasks 就能读到别人的数据（水平越权）。
type viewer struct {
	ID   int
	Role int
}

func (v viewer) isAdmin() bool { return v.Role == model.RoleAdmin }

// canAccess 判断能否操作某个员工名下的数据：本人，或管理员。
func (v viewer) canAccess(ownerID int) bool { return v.isAdmin() || v.ID == ownerID }

// currentViewer 取出认证中间件写入的身份；取不到说明路由没挂中间件，按未认证处理。
func currentViewer(c *gin.Context) (viewer, bool) {
	rawID, exists := c.Get(middleware.CtxUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未认证"})
		return viewer{}, false
	}
	id, ok := rawID.(int64)
	if !ok || id <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "未认证"})
		return viewer{}, false
	}
	role := model.RoleEmployee
	if rawRole, exists := c.Get(middleware.CtxUserRole); exists {
		if r, ok := rawRole.(int); ok {
			role = r
		}
	}
	return viewer{ID: int(id), Role: role}, true
}

// respondForbidden 越权统一返回 403：请求方身份合法，但无权碰这条数据。
func respondForbidden(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{"message": "无权访问该员工的数据"})
}

// fetchTaskForViewer 取任务并校验归属：任务不存在 → 404；存在但不属于自己、又不是管理员 → 403。
// 任务的归属写在库里（tasks.user_id），所以只能先查出来再判断，不能信 URL 上的 id。
func fetchTaskForViewer(c *gin.Context, taskService *service.TaskService, taskID int) (model.Task, bool) {
	v, ok := currentViewer(c)
	if !ok {
		return model.Task{}, false
	}
	task, err := taskService.GetByID(c.Request.Context(), taskID)
	if err != nil {
		respondError(c, err, "任务不存在", "查询任务失败")
		return model.Task{}, false
	}
	if !v.canAccess(task.UserID) {
		respondForbidden(c)
		return model.Task{}, false
	}
	return task, true
}

// fetchUserForViewer 取员工并校验归属：不存在 → 404；存在但不是自己、又不是管理员 → 403。
// 先判存在再判权限，是为了让「查不存在的员工」仍然返回 404，而不是把 403 当成万能的兜底。
func fetchUserForViewer(c *gin.Context, userService *service.UserService, userID int) (model.User, bool) {
	v, ok := currentViewer(c)
	if !ok {
		return model.User{}, false
	}
	target, err := userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		respondError(c, err, "员工不存在", "查询员工失败")
		return model.User{}, false
	}
	if !v.canAccess(target.ID) {
		respondForbidden(c)
		return model.User{}, false
	}
	return target, true
}
