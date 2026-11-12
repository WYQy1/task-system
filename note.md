API	用途	示例
c.Query()	获取 URL 查询参数	/hello?name=Tom
c.Param()	获取 URL 路径参数	/users/1010
JSON 标签需要使用 json:"字段名" 的格式
路由里叫什么名字，后面取参数时就必须用同一个名字，前后要对上
比如路由写 :id，获取时就写 c.Param("id")；路由改成 :userID，获取时就写 c.Param("userID")。
*action：匹配后面剩余的路径，可以包含多个 / 例如 /user/Tom/send/message 有name = Tom action = /send/message

前端发送->{          
    "name": "Tom",     ->ShouldBindJSON()->User{
    "age": 22                                 Name: "Tom",   ->返回值为一个nil 
           }                                  Age:  22, }

gte=1：大于或等于 1。
lte=150：小于或等于 150

客户端 JSON
    ↓
CreateUserRequest
    ↓
数据绑定与合法性校验
    ↓
姓名预处理
    ↓
转换为 User
    ↓
返回响应