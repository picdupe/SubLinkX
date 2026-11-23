package middlewares

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"sublink/models"

	"github.com/gin-gonic/gin"
)

func RelayAgentAuth(c *gin.Context) {
	header := c.GetHeader("Authorization")
	if !strings.HasPrefix(header, "Bearer ") || len(header) <= len("Bearer ") {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "agent credentials required"})
		return
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	sum := sha256.Sum256([]byte(token))
	hash := hex.EncodeToString(sum[:])
	var server models.RelayServer
	// Keep Agent authentication available for disabled servers so an already
	// queued empty configuration can be applied and remove active listeners.
	if err := models.DB.Where("agent_token_hash = ?", hash).First(&server).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "invalid agent credentials"})
		return
	}
	c.Set("relay_server_id", server.ID)
	c.Next()
}
