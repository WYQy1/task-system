package main
import (
    "net/http"  //这是 Go 标准库中的 HTTP 包
    "github.com/gin-gonic/gin"
	"fmt"
	"strings"
	"database/sql"
	"strconv"  //idStr->idint 

)
//Context 里装着本次请求的所有信息（路径、查询参数、请求体、响应写入器等），Gin 会自动传进来
func Ping(c *gin.Context){
    //c.JSON(状态码, gin.H{...})
	c.JSON(http.StatusOK,gin.H{"message":"pong"})
}

func GetHello(c *gin.Context){
	name := c.DefaultQuery("name", "Guest")  //DefaultQuery() 可以直接设置默认值
	// name:=c.Query("name")
	// if name==""{
	// 	name="Guest"     //添加默认值处理
	// }  下边这四行代码等价于上边一行
	message:=fmt.Sprintf("hello %s",name) 
	c.JSON(http.StatusOK,gin.H{"message":message})
}

//定义结构体时，可以直接写 JSON Tag，不需要导入任何包
type CreateUserRequest struct{  //请求结构体负责限制客户端能提交什么
	Name string `json:"name" binding:"required,max=20"`//binding:"required" 要求该字段必须提供有效的非空值->用作合法性校验
	Age  int    `json:"age" binding:"gte=1,lte=150"` 
    Gender string `json:"gender"`
}
type User struct {  //业务数据结构体负责描述服务端实际管理什么 
    ID     int    `json:"id"`
    Name   string `json:"name"`
    Age    int    `json:"age"`
    Gender string `json:"gender"`
}   
func CreateUser(c *gin.Context){
	var request CreateUserRequest
	err:=c.ShouldBindJSON(&request) //将请求体 JSON 绑定到结构体
	if err!=nil{ //err != nil 表示出错了
       c.JSON(http.StatusBadRequest,gin.H{"message":"请求格式错误"})
	   return   //return 就是告诉程序：错误已经处理完了，当前函数到此结束，不要再往下执行
	}
	request.Name=strings.TrimSpace(request.Name) //去除首位空格
	if request.Name==""{  // //为什么加这一步 因为ShouldBindJSON中requierd可以拦截""但是无法处理" " 所以要二次检验
		c.JSON(http.StatusBadRequest, gin.H{
        "message": "员工姓名不能为空",
    })
         return
	}
	user := User{  //将前端的数据导入到数据库的预处理
    Name: request.Name,
    Age:  request.Age,
	Gender: request.Gender,
    }  
   //Go 里很多函数返回两个值：结果, 错误。        
   result, err := DB.Exec("INSERT INTO users (name, age, gender) VALUES (?, ?, ?)",user.Name,user.Age,user.Gender) //Exec执行 INSERT、UPDATE、DELETE 等操作
   if err!=nil{
	   c.JSON(http.StatusInternalServerError,gin.H{"message":"员工添加失败"})
	   return
   }
   id, err := result.LastInsertId() //取刚刚插入那一行的自增主键 ID
   if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"message": "获取员工ID失败"})
    return
   }
   c.JSON(http.StatusCreated, gin.H{"message": "员工添加成功","id":id,"name":user.Name,"age":user.Age,"gender":user.Gender})
}

func GetUser(c *gin.Context){
	idStr:=c.Param("id")
	id,err:=strconv.Atoi(idStr)
	if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{ "message": "员工ID格式错误"})
    return
    }
	var user User  //QueryRow 根据 ID 查询一条员工记录。 ?：SQL 参数占位符，实际查询时由 id 替换。
	err=DB.QueryRow("SELECT id, name, age, gender FROM users WHERE id = ?",id).Scan(&user.ID, &user.Name, &user.Age, &user.Gender)
	if err==sql.ErrNoRows{ //是一个特定错误值 查询结果一条都没有"
		c.JSON(http.StatusNotFound,gin.H{"message":"员工不存在"})
		return
	}
	if err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{"message":"查询员工失败"})
		return
	}
	c.JSON(http.StatusOK,gin.H{"message":"查询员工信息成功","id":user.ID,"name":user.Name,"age":user.Age, "gender":user.Gender})
	
}
func GetUsers(c *gin.Context) {
    rows, err := DB.Query("SELECT id, name, age, gender FROM users") //执行 SELECT 查询，返回查询结果集合 rows
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "查询员工失败"})
        return
    }
    defer rows.Close()
    var users []User  //声明一个用于存放多个员工对象的切片->users = append(users, User{ID: 1,Name: "Tom", Age:  23,})

    for rows.Next() {  //rows.Next() 每调用一次，移动到下一条记录。
        var user User
        var gender sql.NullString  //表示 "可能为 NULL 的字符串
        err := rows.Scan(&user.ID,&user.Name,&user.Age, &gender) //把当前记录的字段值写入 Go 变量因为 Scan 需要变量地址，才能将查询结果写进去
		if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"message": "读取员工数据失败"})
            return
        }
        if gender.Valid {user.Gender = gender.String}
        users = append(users, user)
    }
	if err := rows.Err(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "遍历员工数据失败"})
        return
    }

    c.JSON(http.StatusOK, users)
}

func UpdateUser(c *gin.Context){
    idStr:=c.Param("id")
    id,err:=strconv.Atoi(idStr)
    if err!=nil{
        c.JSON(http.StatusBadRequest,gin.H{"message":"员工ID格式错误"})
        return
    }
    var request CreateUserRequest
    if err:=c.ShouldBindJSON(&request);err!=nil{
         c.JSON(http.StatusBadRequest, gin.H{"message": "请求数据格式错误"})
         return
    }
    request.Name = strings.TrimSpace(request.Name)
    if request.Name==""{
        c.JSON(http.StatusBadRequest, gin.H{"message": "员工姓名不能为空"})
        return
    }
    //_：暂时丢弃 SQL 执行结果。  这里使用 =，因为前面已经声明过 err
    result,err:= DB.Exec("UPDATE users SET name = ?, age = ?, gender = ? WHERE id = ?",request.Name,request.Age,request.Gender,id)
    if err!=nil{
        c.JSON(http.StatusInternalServerError,gin.H{"message":"员工信息更新失败"})
        return
    }
    rowsAffected, err := result.RowsAffected() //RowsAffected() 获取受影响的记录数。
    if err!=nil{
        c.JSON(http.StatusInternalServerError,gin.H{"message":"获取更新记录数失败"})
        return
    }
    if rowsAffected==0{
        err=DB.QueryRow("SELECT id FROM users WHERE id = ?",id).Scan(&id)
        if err==sql.ErrNoRows{
             c.JSON(http.StatusNotFound,gin.H{"message":"员工id不存在"})
             return
        }
        if err!=nil{
            c.JSON(http.StatusInternalServerError,gin.H{"message":"查询员工失败"})
             return
        }
    }   //注意最后没有 return 是正确的。因为员工存在但数据没变化时，仍然应该返回更新成功。
    c.JSON(http.StatusOK, gin.H{"message":"员工信息更新成功","id":id,"name":request.Name,"age":request.Age,"gender":request.Gender})
}
//目前暂未实现删除有关联任务的员工
func DeleteUser(c *gin.Context){
    idStr:=c.Param("id")
    id,err:=strconv.Atoi(idStr)
    if err!=nil{
        c.JSON(http.StatusBadRequest,gin.H{"message":"员工ID格式错误"})
        return
    }
    result,err:=DB.Exec("DELETE FROM users WHERE id = ?",id) //DELETE 是删除操作，不需要读取查询结果，所以不需要 Scan()
    if err!=nil{
        c.JSON(http.StatusInternalServerError,gin.H{"message":"删除员工失败"})
        return
    }
    rowsAffected, err := result.RowsAffected() 
    if err!=nil{
        c.JSON(http.StatusInternalServerError,gin.H{"message":"获取删除记录失败"})
        return
    }
    if rowsAffected==0{
        c.JSON(http.StatusNotFound,gin.H{"message":"删除的员工不存在"})
        return
    }
    c.JSON(http.StatusOK,gin.H{"message":"删除成功","id":id})

}