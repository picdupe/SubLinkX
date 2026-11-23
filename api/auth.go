package api

import (
	"container/list"
	"log"
	"net/http"
	"sublink/middlewares"
	"sublink/models"
	"sublink/utils"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type loginAttempt struct {
	window  time.Time
	count   int
	orderEl *list.Element
}

var loginAttempts = struct {
	sync.Mutex
	entries map[string]loginAttempt
	order   *list.List
}{entries: make(map[string]loginAttempt), order: list.New()}

const (
	loginAttemptWindow  = 15 * time.Minute
	maxLoginAttemptKeys = 10000
)

func allowLogin(key string) bool {
	now := time.Now()
	loginAttempts.Lock()
	defer loginAttempts.Unlock()
	a, exists := loginAttempts.entries[key]
	if exists && now.Sub(a.window) >= loginAttemptWindow {
		a.count = 0
		a.window = now
	}
	if !exists {
		if len(loginAttempts.entries) >= maxLoginAttemptKeys {
			oldest := loginAttempts.order.Back()
			if oldest != nil {
				delete(loginAttempts.entries, oldest.Value.(string))
				loginAttempts.order.Remove(oldest)
			}
		}
		a = loginAttempt{window: now}
		a.orderEl = loginAttempts.order.PushFront(key)
	} else {
		loginAttempts.order.MoveToFront(a.orderEl)
	}
	if a.count >= 10 {
		loginAttempts.entries[key] = a
		return false
	}
	a.count++
	loginAttempts.entries[key] = a
	return true
}

func resetLoginAttempts(key string) {
	loginAttempts.Lock()
	if attempt, exists := loginAttempts.entries[key]; exists {
		loginAttempts.order.Remove(attempt.orderEl)
	}
	delete(loginAttempts.entries, key)
	loginAttempts.Unlock()
}

// 获取token
func GetToken(username string) (string, error) {
	// 过期时间天
	ExpireDays := models.ReadConfig().ExpireDays
	if ExpireDays <= 0 {
		ExpireDays = 14
	}
	user := models.User{Username: username}
	if err := user.Find(); err != nil {
		return "", err
	}
	c := &middlewares.JwtClaims{
		Username:     username,
		TokenVersion: user.TokenVersion,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * time.Duration(ExpireDays)).Unix(), // 设置过期时间
			IssuedAt:  time.Now().Unix(),                                                 // 签发时间
			Subject:   username,                                                          // 用户
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(middlewares.SecretBytes())
}

// 获取captcha图形验证码
func GetCaptcha(c *gin.Context) {
	id, bs4, _, err := utils.GetCaptcha()
	if err != nil {
		log.Println("获取验证码失败")
		c.JSON(400, gin.H{
			"msg": "获取验证码失败",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": "00000",
		"data": gin.H{
			"captchaKey":    id,
			"captchaBase64": bs4,
		},
		"msg": "获取验证码成功",
	})

}

// 用户登录
func UserLogin(c *gin.Context) {
	ipKey := c.ClientIP() + ":*"
	if !allowLogin(ipKey) {
		c.Header("Retry-After", "900")
		c.JSON(429, gin.H{"msg": "登录尝试过多，请稍后再试"})
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	username := c.PostForm("username")
	if len(username) > 128 {
		c.JSON(400, gin.H{"msg": "用户名长度超出限制"})
		return
	}
	key := c.ClientIP() + ":" + username
	if !allowLogin(key) {
		c.Header("Retry-After", "900")
		c.JSON(429, gin.H{"msg": "登录尝试过多，请稍后再试"})
		return
	}
	password := c.PostForm("password")
	captchaCode := c.PostForm("captchaCode")
	captchaKey := c.PostForm("captchaKey")
	// 验证验证码
	if !utils.VerifyCaptcha(captchaKey, captchaCode) {
		log.Println("验证码错误")
		c.JSON(400, gin.H{
			"msg": "验证码错误",
		})
		return
	}
	user := &models.User{Username: username, Password: password}
	err := user.Verify()
	if err != nil {
		log.Println("账号或者密码错误")
		c.JSON(400, gin.H{
			"msg": "账号或者密码错误",
		})
		return
	}
	resetLoginAttempts(key)
	resetLoginAttempts(c.ClientIP() + ":*")
	// 生成token
	token, err := GetToken(username)
	if err != nil {
		log.Println("获取token失败", err)
		c.JSON(400, gin.H{
			"msg": "获取token失败",
		})
		return
	}
	// 登录成功返回token
	c.JSON(200, gin.H{
		"code": "00000",
		"data": gin.H{
			"accessToken":  token,
			"tokenType":    "Bearer",
			"refreshToken": nil,
			"expires":      nil,
		},
		"msg": "登录成功",
	})
}
func UserOut(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		c.JSON(401, gin.H{"code": 401, "msg": "unauthorized"})
		return
	}
	result := models.DB.Model(&models.User{}).
		Where("username = ?", username.(string)).
		Update("token_version", gorm.Expr("token_version + 1"))
	if result.Error != nil {
		c.JSON(500, gin.H{"code": 500, "msg": "退出登录失败"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(401, gin.H{"code": 401, "msg": "unauthorized"})
		return
	}
	c.JSON(200, gin.H{"code": "00000", "msg": "退出成功"})
}
