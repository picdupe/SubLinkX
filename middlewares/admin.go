package middlewares

import (
	"net/http"
	"strings"

	"sublink/models"

	"github.com/gin-gonic/gin"
)

// RequireAdmin performs server-side authorization for privileged management APIs.
func RequireAdmin(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "unauthorized"})
		return
	}
	user := models.User{Username: username.(string)}
	if err := user.Find(); err != nil || !strings.EqualFold(user.Role, "admin") {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "msg": "admin role required"})
		return
	}
	c.Next()
}
