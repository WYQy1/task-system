package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"task-system/repository"
)

// 认证通过后写进 gin.Context 的键。handler 只从这里取当前登录者，
// 不再相信请求里传来的 id，避免改个 URL 就变成别人。
const (
	CtxUserID   = "userID"
	CtxUserRole = "userRole"
)

// HS256 的密钥长度直接决定被离线爆破的难度，32 字节（256 bit）是与 SHA-256 匹配的下限。
const minJWTSecretLen = 32

var jwtSecret []byte

// InitJWTSecret 在启动时校验密钥：缺失或过短直接失败，而不是带着空密钥上线。
func InitJWTSecret() error {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if len(secret) < minJWTSecretLen {
		return fmt.Errorf("JWT_SECRET 未设置或长度不足 %d 字节（当前 %d 字节），可用 `openssl rand -hex 32` 生成", minJWTSecretLen, len(secret))
	}
	jwtSecret = []byte(secret)
	return nil
}

func signingKey() ([]byte, error) {
	if len(jwtSecret) == 0 {
		if err := InitJWTSecret(); err != nil {
			return nil, err
		}
	}
	return jwtSecret, nil
}

func GenerateToken(userID int, tokenVersion int) (string, error) {
	key, err := signingKey()
	if err != nil {
		return "", err
	}
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"ver":     tokenVersion,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(key)
}

func ParseToken(tokenString string) (int64, int, error) {
	key, err := signingKey()
	if err != nil {
		return 0, 0, err
	}
	token, err := jwt.Parse(tokenString,
		func(t *jwt.Token) (interface{}, error) { return key, nil },
		// 显式限定算法，防止 alg 混淆（把 HS256 换成 none 或 RS256 绕过签名）
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return 0, 0, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, 0, errors.New("token无效")
	}
	// JSON 里的数字统一解析成 float64
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

// AuthMiddleware 验签后回库确认用户仍存在、版本号未变（登出即失效），并把角色放进 ctx。
func AuthMiddleware(userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.SplitN(c.GetHeader("Authorization"), " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "token 格式错误"})
			return
		}

		userID, ver, err := ParseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "token 无效"})
			return
		}

		currentVer, role, err := userRepo.GetAuthInfo(c.Request.Context(), int(userID))
		switch {
		case err == nil:
			// 继续走下面的版本号比对
		case errors.Is(err, repository.ErrUserNotFound):
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "token 无效"})
			return
		case errors.Is(err, context.DeadlineExceeded):
			// 查库超时是服务端的问题，不能伪装成"token 无效"，否则客户端会去重新登录却永远登录不上
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{"message": "请求超时"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "服务内部错误"})
			return
		}

		if currentVer != ver {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "token 已失效，请重新登录"})
			return
		}

		c.Set(CtxUserID, userID)
		c.Set(CtxUserRole, role)
		c.Next()
	}
}
