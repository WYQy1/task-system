
package main

import (
	"log"
	"task-system/handler"
	"task-system/repository"
	"task-system/service"
	

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	userRepo := repository.NewUserRepository(DB)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	taskRepo := repository.NewTaskRepository(DB)
	taskService := service.NewTaskService(taskRepo)
	taskHandler := handler.NewTaskHandler(taskService)

	r.GET("/ping", handler.Ping)
	r.GET("/hello", handler.GetHello)

	r.POST("/users", userHandler.CreateUser)
	r.GET("/users", userHandler.GetUsers)
	r.GET("/users/:id", userHandler.GetUser)
	r.PUT("/users/:id", userHandler.UpdateUser)
	r.DELETE("/users/:id", userHandler.DeleteUser)

	r.POST("/users/:id/tasks", taskHandler.CreateTask)
	r.GET("/users/:id/tasks", taskHandler.GetTasks)
	r.GET("/users/tasks/:id", taskHandler.GetTask)
	r.PUT("/tasks/:id", taskHandler.UpdateTask)
	r.DELETE("/tasks/:id", taskHandler.DeleteTask)

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
