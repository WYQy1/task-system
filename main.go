// @title           Task System API
// @version         1.2.1
// @description     员工-任务管理系统接口文档（JWT 鉴权）
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     格式：Bearer <token>
package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "task-system/docs"
	"task-system/handler"
	"task-system/middleware"
	"task-system/repository"
	"task-system/service"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(middleware.Timeout(5 * time.Second))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	userRepo := repository.NewUserRepository(DB)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	authService := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authService)

	taskRepo := repository.NewTaskRepository(DB)
	taskService := service.NewTaskService(userRepo, taskRepo)
	taskHandler := handler.NewTaskHandler(taskService)

	r.GET("/ping", handler.Ping)
	r.GET("/hello", handler.GetHello)
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)

	auth := r.Group("", middleware.AuthMiddleware(userRepo))
	{
		auth.POST("/auth/logout", authHandler.Logout)

		auth.POST("/users", userHandler.CreateUser)
		auth.GET("/users", userHandler.GetUsers)
		auth.GET("/users/:id", userHandler.GetUser)
		auth.PUT("/users/:id", userHandler.UpdateUser)
		auth.DELETE("/users/:id", userHandler.DeleteUser)

		auth.POST("/users/:id/tasks", taskHandler.CreateTask)
		auth.GET("/users/:id/tasks", taskHandler.GetTasks)
		auth.GET("/users/tasks/:id", taskHandler.GetTask)
		auth.PUT("/tasks/:id", taskHandler.UpdateTask)
		auth.DELETE("/tasks/:id", taskHandler.DeleteTask)
	}

	return r
}

func main() {
	if err := InitDB(); err != nil {
		log.Fatal("数据库连接失败：", err)
	}
	r := SetupRouter()
	if err := r.Run(":8080"); err != nil {
		log.Fatal("服务启动失败：", err)
	}
}
