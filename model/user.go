package model

// 角色：0=普通员工，1=管理员。
const (
	RoleEmployee = 0
	RoleAdmin    = 1
)

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Age      int    `json:"age"`
	Gender   string `json:"gender"`
	Password string `json:"-"` // 序列化时忽略，永不输出到 JSON
	Role     int    `json:"role"`
}

// IsAdmin 用于授权判断：只有管理员能跨用户操作数据。
func (u User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
