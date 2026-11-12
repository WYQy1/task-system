package main
import (
     "github.com/gin-gonic/gin"
     "log"
)

func main() {
    if err := InitDB(); err != nil {
        log.Fatal("数据库连接失败：", err)
    }
    r := gin.Default() //Default 创建了一个带默认中间件的 Gin 引擎 r就是router的简写 router是路由的意思
    r.GET("/ping",Ping)
    r.GET("/hello",GetHello)
    r.GET("/users/:id",GetUser)
    r.POST("/users",CreateUser) //新增员工，由 MySQL 自动生成 ID
    r.GET("/users",GetUsers)
    r.PUT("/users/:id",UpdateUser)  //修改指定 ID 的已有员工，不会自动新增记录
    r.DELETE("/users/:id", DeleteUser)
    r.Run(":8080") //这句代码负责启动 HTTP 服务
}