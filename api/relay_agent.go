package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"sublink/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type relayHeartbeatRequest struct {
	AgentID         string `json:"agent_id"`
	AgentVersion    string `json:"agent_version"`
	OS              string `json:"os"`
	Architecture    string `json:"architecture"`
	ServiceManager  string `json:"service_manager"`
	HAProxyVersion  string `json:"haproxy_version"`
	HAProxyStatus   string `json:"haproxy_status"`
	AppliedRevision uint64 `json:"applied_revision"`
}

type relayTaskResultRequest struct {
	Attempt        uint   `json:"attempt"`
	Success        bool   `json:"success"`
	Revision       uint64 `json:"revision"`
	SHA256         string `json:"sha256"`
	Validation     string `json:"validation"`
	Reload         string `json:"reload"`
	Rollback       string `json:"rollback"`
	HAProxyStatus  string `json:"haproxy_status"`
	ErrorCode      string `json:"error_code"`
	ErrorMessage   string `json:"error_message"`
	DurationMillis int64  `json:"duration_ms"`
}

func RelayAgentHeartbeat(c *gin.Context) {
	var req relayHeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.AgentID) > 64 || len(req.AgentVersion) > 32 || len(req.OS) > 255 || len(req.Architecture) > 32 || len(req.ServiceManager) > 20 || len(req.HAProxyVersion) > 64 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "00001", "msg": "invalid heartbeat"})
		return
	}
	if req.HAProxyStatus != "running" && req.HAProxyStatus != "stopped" && req.HAProxyStatus != "unknown" && req.HAProxyStatus != "degraded" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "00001", "msg": "invalid haproxy_status"})
		return
	}
	serverID := relayAgentServerID(c)
	var identity models.RelayServer
	if err := models.DB.Select("agent_id", "desired_revision").First(&identity, serverID).Error; err != nil || identity.AgentID != req.AgentID {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "agent identity does not match credential"})
		return
	}
	result := models.DB.Model(&models.RelayServer{}).Where("id = ?", serverID).Updates(map[string]any{
		"agent_version": req.AgentVersion, "os_info": req.OS, "architecture": req.Architecture,
		"service_manager": req.ServiceManager, "ha_proxy_version": req.HAProxyVersion,
		"ha_proxy_status": req.HAProxyStatus, "heartbeat_at": time.Now().UTC(),
		"applied_revision": gorm.Expr("CASE WHEN applied_revision <= ? AND ? <= desired_revision THEN ? ELSE applied_revision END", req.AppliedRevision, req.AppliedRevision, req.AppliedRevision),
	})
	if result.Error != nil {
		relayError(c, http.StatusInternalServerError, result.Error)
		return
	}
	var current models.RelayServer
	if err := models.DB.First(&current, serverID).Error; err != nil {
		relayError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "00000", "data": gin.H{"desired_revision": current.DesiredRevision, "applied_revision": current.AppliedRevision}, "msg": "ok"})
}

func RelayAgentTaskNext(c *gin.Context) {
	waitSeconds, _ := strconv.Atoi(c.DefaultQuery("wait", "0"))
	if waitSeconds < 0 {
		waitSeconds = 0
	}
	if waitSeconds > 25 {
		waitSeconds = 25
	}
	deadline := time.NewTimer(time.Duration(waitSeconds) * time.Second)
	defer deadline.Stop()
	for {
		task, err := claimRelayTask(relayAgentServerID(c))
		if err != nil {
			relayError(c, http.StatusInternalServerError, err)
			return
		}
		if task != nil {
			var revision models.RelayConfigRevision
			if err := models.DB.First(&revision, task.RevisionID).Error; err != nil {
				relayError(c, http.StatusInternalServerError, err)
				return
			}
			c.JSON(http.StatusOK, gin.H{"code": "00000", "data": gin.H{
				"task_id": task.TaskID, "type": task.TaskType, "revision": revision.Revision,
				"attempt": task.Attempt, "config_url": fmt.Sprintf("/api/agent/v1/configs/%d", revision.Revision),
				"sha256": revision.ConfigSHA256, "created_at": task.CreatedAt,
			}, "msg": "task"})
			return
		}
		select {
		case <-c.Request.Context().Done():
			return
		case <-deadline.C:
			c.JSON(http.StatusOK, gin.H{"code": "00000", "data": nil, "msg": "no task"})
			return
		case <-time.After(400 * time.Millisecond):
		}
	}
}

func RelayAgentTaskAck(c *gin.Context) {
	var req struct {
		Attempt uint `json:"attempt"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Attempt == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "00001", "msg": "attempt is required"})
		return
	}
	result := models.DB.Model(&models.RelayConfigTask{}).
		Where("task_id = ? AND relay_server_id = ? AND status = ? AND attempt = ? AND lease_expires_at > ?", c.Param("id"), relayAgentServerID(c), "leased", req.Attempt, time.Now().UTC()).
		Updates(map[string]any{"status": "running", "started_at": time.Now().UTC(), "lease_expires_at": time.Now().UTC().Add(5 * time.Minute)})
	if result.Error != nil {
		relayError(c, http.StatusInternalServerError, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusConflict, gin.H{"code": "00001", "msg": "task lease is no longer current"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": "00000",
		"data": gin.H{"task_id": c.Param("id"), "attempt": req.Attempt},
		"msg":  "acknowledged",
	})
}

func RelayAgentConfigGet(c *gin.Context) {
	revisionNumber, err := strconv.ParseUint(c.Param("revision"), 10, 64)
	if err != nil || revisionNumber == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "00001", "msg": "invalid revision"})
		return
	}
	var revision models.RelayConfigRevision
	if err := models.DB.Where("relay_server_id = ? AND revision = ?", relayAgentServerID(c), revisionNumber).First(&revision).Error; err != nil {
		relayError(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "00000", "data": gin.H{"revision": revision.Revision, "sha256": revision.ConfigSHA256, "config": revision.ConfigContent}, "msg": "ok"})
}

func RelayAgentTaskResult(c *gin.Context) {
	var req relayTaskResultRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Attempt == 0 || req.DurationMillis < 0 || len(req.ErrorCode) > 64 || len(req.ErrorMessage) > 2000 || len(req.Validation) > 32 || len(req.Reload) > 32 || len(req.Rollback) > 32 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "00001", "msg": "invalid task result"})
		return
	}
	serverID := relayAgentServerID(c)
	var task models.RelayConfigTask
	if err := models.DB.Where("task_id = ? AND relay_server_id = ?", c.Param("id"), serverID).First(&task).Error; err != nil {
		relayError(c, http.StatusNotFound, err)
		return
	}
	if task.Attempt != req.Attempt || task.Status != "running" || task.LeaseExpiresAt == nil || !task.LeaseExpiresAt.After(time.Now().UTC()) {
		c.JSON(http.StatusConflict, gin.H{"code": "00001", "msg": "task lease is no longer current"})
		return
	}
	var revision models.RelayConfigRevision
	if err := models.DB.First(&revision, task.RevisionID).Error; err != nil {
		relayError(c, http.StatusInternalServerError, err)
		return
	}
	if req.Revision != revision.Revision || (req.Success && !strings.EqualFold(req.SHA256, revision.ConfigSHA256)) || (req.SHA256 != "" && !strings.EqualFold(req.SHA256, revision.ConfigSHA256)) {
		c.JSON(http.StatusBadRequest, gin.H{"code": "00001", "msg": "result revision or hash does not match task"})
		return
	}
	status := "failed"
	if req.Success {
		status = "success"
	}
	if req.HAProxyStatus != "" && req.HAProxyStatus != "running" && req.HAProxyStatus != "stopped" && req.HAProxyStatus != "unknown" && req.HAProxyStatus != "degraded" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "00001", "msg": "invalid haproxy_status"})
		return
	}
	now := time.Now().UTC()
	stale := false
	err := models.DB.Transaction(func(tx *gorm.DB) error {
		update := tx.Model(&models.RelayConfigTask{}).Where("id = ? AND attempt = ? AND status = ? AND lease_expires_at > ?", task.ID, req.Attempt, "running", now).Updates(map[string]any{
			"status": status, "finished_at": now, "error_code": req.ErrorCode, "error_message": sanitizeRelayError(req.ErrorMessage),
		})
		if update.Error != nil {
			return update.Error
		}
		if update.RowsAffected == 0 {
			stale = true
			return nil
		}
		fields := map[string]any{}
		if req.Success {
			fields["applied_revision"] = revision.Revision
		}
		if req.HAProxyStatus != "" {
			fields["ha_proxy_status"] = req.HAProxyStatus
		}
		if len(fields) > 0 {
			return tx.Model(&models.RelayServer{}).Where("id = ? AND desired_revision = ?", serverID, revision.Revision).Updates(fields).Error
		}
		return nil
	})
	if err != nil {
		relayError(c, http.StatusInternalServerError, err)
		return
	}
	if stale {
		c.JSON(http.StatusConflict, gin.H{"code": "00001", "msg": "task lease is no longer current"})
		return
	}
	eventType := "task_failed"
	if req.Success {
		eventType = "task_succeeded"
	}
	recordRelayAdminEvent(c, serverID, eventType, fmt.Sprintf("task_id=%s revision=%d", task.TaskID, revision.Revision))
	c.JSON(http.StatusOK, gin.H{"code": "00000", "data": nil, "msg": "result recorded"})
}

func RelayAgentStatus(c *gin.Context) {
	var server models.RelayServer
	if err := models.DB.First(&server, relayAgentServerID(c)).Error; err != nil {
		relayError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "00000", "data": gin.H{
		"agent_id": server.AgentID, "desired_revision": server.DesiredRevision,
		"applied_revision": server.AppliedRevision, "enabled": server.Enabled,
	}, "msg": "ok"})
}

func RelayAgentEnroll(c *gin.Context) {
	var server models.RelayServer
	if err := models.DB.First(&server, relayAgentServerID(c)).Error; err != nil {
		relayError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "00000", "data": gin.H{"agent_id": server.AgentID, "server_id": server.ID}, "msg": "enrolled"})
}

func claimRelayTask(serverID uint) (*models.RelayConfigTask, error) {
	var claimed models.RelayConfigTask
	now := time.Now().UTC()
	err := models.DB.Transaction(func(tx *gorm.DB) error {
		var candidate models.RelayConfigTask
		result := tx.Where("relay_server_id = ? AND (status = ? OR (status IN ? AND lease_expires_at IS NOT NULL AND lease_expires_at < ?))",
			serverID, "pending", []string{"leased", "running"}, now).Order("id asc").Limit(1).Find(&candidate)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		leaseExpiry := now.Add(60 * time.Second)
		updated := tx.Model(&models.RelayConfigTask{}).
			Where("id = ? AND ((status = ?) OR (status IN ? AND lease_expires_at IS NOT NULL AND lease_expires_at < ?))",
				candidate.ID, "pending", []string{"leased", "running"}, now).
			Updates(map[string]any{"status": "leased", "attempt": gorm.Expr("attempt + 1"), "lease_expires_at": leaseExpiry})
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected == 0 {
			return nil
		}
		return tx.First(&claimed, candidate.ID).Error
	})
	if err != nil {
		return nil, err
	}
	if claimed.ID == 0 {
		return nil, nil
	}
	return &claimed, nil
}

func relayAgentServerID(c *gin.Context) uint {
	id, _ := c.Get("relay_server_id")
	serverID, _ := id.(uint)
	return serverID
}

var relaySecretPattern = regexp.MustCompile(`(?i)(token|authorization|password|secret|private[_-]?key)\s*[:=]\s*[^\s,;]+`)

func sanitizeRelayError(message string) string {
	return relaySecretPattern.ReplaceAllString(message, "$1=[redacted]")
}
