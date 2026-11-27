package middleware

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"task-system/repository"
)

// 认证通过后写进 gin.Context 的键。handler 层只从这里取当前登录者，
// 不再相信请求里传来的 id，避免"改个 URL 就变成别人"的水平越权。
const (
	CtxUserID   = "userID"
	CtxUserRole = "userRole"
)

func GenerateToken(userID int, tokenVersion int) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	//`time.Now()` → 当前时间  `Add(24 * time.Hour)` → 往后加 24 小时 `.Unix()` → 转成 Unix 时间戳（一个整数，表示秒数）
	claims := jwt.MapClaims{"user_id": userID, "exp": time.Now().Add(24 * time.Hour).Unix(), "ver": tokenVersion} // 24小时后过期
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)                                                    //`HS256` 是对称签名算法（同一个密钥签名和校验
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
func ParseToken(tokenString string) (int64, int, error) {
	secret := os.Getenv("JWT_SECRET")
	token, err := jwt.Parse(tokenString,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()})) //显式限定了允许的签名算法
	if err != nil {
		return 0, 0, err
	}
	claims, ok := token.Claims.(jwt.MapClaims) //`x.(T)` 的含义：断言 "x 其实是 T 类型"
	if !ok || !token.Valid {
		return 0, 0, errors.New("token无效")
	}
	// 注意：JSON 里的数字会被解析成 float64
	userIDRaw, ok := claims["user_id"].(float64)
	if !ok {
		return 0, 0, errors.New("token 缺少 user_id")
	}
	verRaw, ok := claims["ver"].(float64)
	if !ok {
		return 0, 0, errors.New("token 缺少 ver")
	}
	return int64(userIDRaw), int(verRaw), nil
}
func AuthMiddleware(userRepo *repository.UserRepository) gin.HandlerFunc { //中间件的类型，就是一个 `func(*gin.Context)` 的别名
	return func(c *gin.Context) {
		// 1. 取 Authorization 头
		authHeader := c.GetHeader("Authorization") //**`Authorization` 头本身**：HTTP 标准（RFC 7235）规定的**通用认证头**
		// 2. 按空格拆开，格式必须是 "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		// 3. 格式不对 → 401 并中止
		if len(parts) != 2 || parts[0] != "Bearer" { //Bearer"叫 "认证方案"（scheme），表示 "这次用哪种认证方式"
			c.JSON(http.StatusUnauthorized, gin.H{"message": "token 格式错误"})
			c.Abort() // 中止，不再往下走
			return
		}
		// 4. 校验 token 签名，取出 userID 和签发时的版本号
		userID, ver, err := ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "token 无效"})
			c.Abort()
			return
		}
		// 5. 查库里当前的版本号和角色（顺带校验用户是否还存在）
		currentVer, role, err := userRepo.GetAuthInfo(c.Request.Context(), int(userID))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "token 无效"})
			c.Abort()
			return
		}
		// 6. 版本不一致 → 说明已被登出/踢下线
		if currentVer != ver {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "token 已失效，请重新登录"})
			c.Abort()
			return
		}
		c.Set(CtxUserID, userID)
		c.Set(CtxUserRole, role)
		// 7. 校验通过，放行
		c.Next()
	}
}
