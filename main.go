// @title           Task System API
// @version         1.3.0
// @description     员工-任务管理系统接口文档（JWT 鉴权 + 按角色授权 + 超期提醒）
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     格式：Bearer <token>
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "task-system/docs"
	"task-system/handler"
	"task-system/middleware"
	"task-system/repository"
	"task-system/service"
	"task-system/worker"
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
	taskHandler := handler.NewTaskHandler(taskService, userService)

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

	// 收到 Ctrl+C / SIGTERM 时取消这个 ctx：HTTP 服务和后台扫描器都盯着它退出
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 后台超期扫描：和 HTTP 服务同生命周期，但不依赖它
	scanner := worker.NewScanner(repository.NewTaskRepository(DB), worker.LogNotifier{}, worker.Config{
		Interval:  time.Minute,
		BatchSize: 50,
		Workers:   4,
	})
	var background sync.WaitGroup
	background.Add(1)
	go func() {
		defer background.Done()
		scanner.Run(ctx)
	}()

	srv := &http.Server{Addr: ":8080", Handler: SetupRouter()}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("服务启动失败：", err)
		}
	}()
	log.Println("服务已启动：http://localhost:8080")

	// 优雅退出：停止接收新请求 → 等在途请求结束 → 等后台扫描器收尾
	<-ctx.Done()
	stop()
	log.Println("收到退出信号，开始优雅关闭…")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP 服务关闭超时：%v", err)
	}
	background.Wait()
	log.Println("服务已退出")
}
