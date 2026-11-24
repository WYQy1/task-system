package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"task-system/service"
)

func parseID(c *gin.Context, wrongMsg string) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": wrongMsg})
		return 0, false
	}
	return id, true
}

func respondError(c *gin.Context, err error, notFoundMsg, fallback string) {
	switch {
	case errors.Is(err, service.ErrInvalidPassword):
		c.JSON(http.StatusBadRequest, gin.H{"message": "密码不能少于6位"})
	case errors.Is(err, service.ErrWrongPassword):
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户名或密码错误"})
	case errors.Is(err, service.ErrInvalidUserID):
		c.JSON(http.StatusBadRequest, gin.H{"message": "员工ID不合法"})
	case errors.Is(err, service.ErrInvalidUserName):
		c.JSON(http.StatusBadRequest, gin.H{"message": "员工姓名不能为空"})
	case errors.Is(err, service.ErrNameTooLong):
		c.JSON(http.StatusBadRequest, gin.H{"message": "员工姓名超过最大长度"})
	case errors.Is(err, service.ErrInvalidUserAge):
		c.JSON(http.StatusBadRequest, gin.H{"message": "员工年龄不符合要求"})
	case errors.Is(err, service.ErrInvalidUserGender):
		c.JSON(http.StatusBadRequest, gin.H{"message": "员工性别格式不符合要求"})
	case errors.Is(err, service.ErrInvalidTaskID):
		c.JSON(http.StatusBadRequest, gin.H{"message": "任务id无效"})
	case errors.Is(err, service.ErrInvalidTaskTitle):
		c.JSON(http.StatusBadRequest, gin.H{"message": "任务标题不合法"})
	case errors.Is(err, service.ErrInvalidTaskStatus):
		c.JSON(http.StatusBadRequest, gin.H{"message": "任务状态不合法"})
	case errors.Is(err, service.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": notFoundMsg})
	case errors.Is(err, service.ErrTaskNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": "任务不存在"})
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"message": fallback})
	}
}
