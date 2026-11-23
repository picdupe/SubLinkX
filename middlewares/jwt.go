package middlewares

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sublink/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// SecretBytes reads the configured key at use time, after ConfigInit has repaired
// or generated configuration defaults.
func SecretBytes() []byte {
	return []byte(models.ReadConfig().JwtSecret)
}

// JwtClaims jwt声明
type JwtClaims struct {
	Username     string `json:"username"`
	TokenVersion int64  `json:"token_version"`
	jwt.StandardClaims
}

// AuthorToken 验证token中间件
func AuthorToken(c *gin.Context) {
	// 定义白名单
	path := c.Request.URL.Path
	method := c.Request.Method
	// 如果是首页直接跳过
	if c.Request.URL.Path == "/" {
		c.Next()
		return
	}
	// 如果是白名单直接跳过
	if strings.HasPrefix(path, "/static/") || strings.HasPrefix(path, "/api/agent/v1/") || path == "/c/" ||
		(path == "/api/v1/auth/login" && method == http.MethodPost) ||
		(path == "/api/v1/auth/captcha" && method == http.MethodGet) ||
		(path == "/api/v1/version" && method == http.MethodGet) {
		c.Next()
		return
	}
	token := c.Request.Header.Get("Authorization")
	if token == "" {
		c.JSON(400, gin.H{"msg": "请求未携带token"})
		c.Abort()
		return
	}
	const prefix = "Bearer "
	if len(token) <= len(prefix) || token[:len(prefix)] != prefix {
		c.JSON(400, gin.H{"msg": "token格式错误"})
		c.Abort()
		return
	}
	mc, err := ParseToken(token[len(prefix):])
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": 401,
			"msg":  err.Error(),
		})
		c.Abort()
		return
	}
	user := models.User{Username: mc.Username}
	if err := user.Find(); err != nil || subtle.ConstantTimeCompare(
		[]byte(strconv.FormatInt(user.TokenVersion, 10)),
		[]byte(strconv.FormatInt(mc.TokenVersion, 10)),
	) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "token has been revoked"})
		c.Abort()
		return
	}
	c.Set("username", mc.Username)
	c.Next()
}

// ParseToken 解析JWT
func ParseToken(tokenString string) (*JwtClaims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		secret := SecretBytes()
		if len(secret) == 0 {
			return nil, errors.New("JWT secret is not configured")
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid { // 校验token
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
