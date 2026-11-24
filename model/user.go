package model

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Age      int    `json:"age"`
	Gender   string `json:"gender"`
	Password string `json:"-"` // 序列化时忽略，永不输出到 JSON
}
