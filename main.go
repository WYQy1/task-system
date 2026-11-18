package main

import (
	"github.com/gin-gonic/gin"
	"log"
)

// 写成两个函数可以让路由初始化与服务启动分离
func SetupRouter() *gin.Engine {
	r := gin.Default() //Default 创建了一个带默认中间件的 Gin 引擎 r就是router的简写 router是路由的意思
	r.GET("/ping", Ping)
	r.GET("/hello", GetHello)
	r.GET("/users/:id", GetUser) //冒号 :是用来标记动态路由参数的
	r.POST("/users", CreateUser) //新增员工，由 MySQL 自动生成 ID
	r.GET("/users", GetUsers)
	r.PUT("/users/:id", UpdateUser) //修改指定 ID 的已有员工，不会自动新增记录
	r.DELETE("/users/:id", DeleteUser)
	r.POST("/users/:id/tasks", CreateTask)
	r.GET("/users/:id/tasks", GetTasks) // 员工 ID 	该员工的所有任务
	r.GET("/users/tasks/:id", GetTask)  //任务 ID  	单条任务
	r.PUT("/tasks/:id", UpdateTask)
	r.DELETE("/tasks/:id", DeleteTask)
	return r
}
func main() {
	if err := InitDB(); err != nil {
		log.Fatal("数据库连接失败：", err)
	}
	r := SetupRouter()
	r.Run(":8080") //这句代码负责启动 HTTP 服务
}
