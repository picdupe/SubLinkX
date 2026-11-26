package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"sublink/models"
	"sublink/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type relayServerRequest struct {
	Name       string `json:"name"`
	PublicHost string `json:"public_host"`
	Enabled    *bool  `json:"enabled"`
}

type relayRouteRequest struct {
	Name           string `json:"name"`
	RelayServerID  uint   `json:"relay_server_id"`
	LandingNodeID  int    `json:"landing_node_id"`
	Protocol       string `json:"protocol"`
	ListenAddress  string `json:"listen_address"`
	ListenPort     uint16 `json:"listen_port"`
	Enabled        bool   `json:"enabled"`
	PublishEnabled bool   `json:"publish_enabled"`
	Priority       int    `json:"priority"`
	Description    string `json:"description"`
}

var errRelayRouteAssigned = errors.New("route is still assigned to subscriptions; remove it from subscriptions before deleting")

func RelayServersList(c *gin.Context) {
	var servers []models.RelayServer
	if err := models.DB.Order("id desc").Find(&servers).Error; err != nil {
		relayError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "00000", "data": servers, "msg": "ok"})
}

func RelayServerGet(c *gin.Context) {
	var server models.RelayServer
	if err := models.DB.First(&server, "id = ?", c.Param("id")).Error; err != nil {
		relayError(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "00000", "data": server, "msg": "ok"})
}

// RelayServerCreate returns the generated enrollment token once. Only its SHA-256 hash is stored.
func RelayServerCreate(c *gin.Context) {
	var req relayServerRequest
	if err := c.ShouldBindJSON(&req); err != nil || !validRelayServer(req) {
		c.JSON(http.StatusBadRequest, gin.H{"code": "00001", "msg": "invalid server name or public_host"})
		return
	}
	credential, err := relaySecret()
	if err != nil {
		relayError(c, http.StatusInternalServerError, err)
		return
	}
	agentID, err := relaySecret()
	if err != nil {
		relayError(c, http.StatusInternalServerError, err)
		return
	}
	server := models.RelayServer{
		Name: strings.TrimSpace(req.Name), PublicHost: strings.TrimSpace(req.PublicHost),
		AgentID: "relay-" + agentID[:24], AgentTokenHash: hashRelayToken(credential),
		AgentTokenCreatedAt: timePtr(time.Now().UTC()), HAProxyStatus: "unknown", Enabled: true,
	}
	if req.Enabled != nil {
		server.Enabled = *req.Enabled
	}
	if err := models.DB.Create(&server).Error; err != nil {
		relayError(c, http.StatusConflict, err)
		return
	}
	recordRelayAdminEvent(c, server.ID, "server_created", "")
	c.JSON(http.StatusCreated, gin.H{"code": "00000", "data": gin.H{"server": server, "agent_token": credential}, "msg": "created; token is shown once"})
}

func RelayServerUpdate(c *gin.Context) {
	var req relayServerRequest
	if err := c.ShouldBindJSON(&req); err != nil || !validRelayServer(req) {
		c.JSON(http.StatusBadRequest, gin.H{"code": "00001", "msg": "invalid server name or public_host"})
		return
	}
	var server models.RelayServer
	var revision models.RelayConfigRevision
	var task models.RelayConfigTask
	var queued bool
	err := models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&server, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		wasEnabled := server.Enabled
		enabled := wasEnabled
		if req.Enabled != nil {
			enabled = *req.Enabled
		}
		fields := map[string]any{
			"name":        strings.TrimSpace(req.Name),
			"public_host": strings.TrimSpace(req.PublicHost),
		}
		if wasEnabled && !enabled {
			// Keep the Agent authorized until it applies the empty config that
			// removes all public listeners from HAProxy.
			if err := tx.Model(&server).Updates(fields).Error; err != nil {
				return err
			}
			var err error
			revision, task, err = queueRelayRevision(tx, server.ID, "", "server disabled; clear relay listeners", relayActorID(c))
			if err != nil {
				return err
			}
			if err := tx.Model(&server).Update("enabled", false).Error; err != nil {
				return err
			}
			queued = true
		} else {
			fields["enabled"] = enabled
			if err := tx.Model(&server).Updates(fields).Error; err != nil {
				return err
			}
			if !wasEnabled && enabled {
				config, err := services.RenderRelayConfig(tx, server.ID)
				if err != nil {
					return err
				}
				revision, task, err = queueRelayRevision(tx, server.ID, config, "server enabled; restore relay listeners", relayActorID(c))
				if err != nil {
					return err
				}
				queued = true
			}
		}
		return tx.First(&server, server.ID).Error
	})
	if err != nil {
		relayError(c, http.StatusConflict, err)
		return
	}
	recordRelayAdminEvent(c, server.ID, "server_updated", "")
	if queued {
		recordRelayAdminEvent(c, server.ID, "config_queued", fmt.Sprintf("revision=%d task_id=%s", revision.Revision, task.TaskID))
	}
	c.JSON(http.StatusOK, gin.H{"code": "00000", "data": server, "msg": "updated"})
}

func RelayServerDelete(c *gin.Context) {
	id, ok := relayID(c)
	if !ok {
		return
	}
	var routes, tasks, revisions int64
	if err := models.DB.Model(&models.RelayRoute{}).Where("relay_server_id = ?", id).Count(&routes).Error; err != nil {
		relayError(c, http.StatusInternalServerError, err)
		return
	}
	if err := models.DB.Model(&models.RelayConfigTask{}).Where("relay_server_id = ?", id).Count(&tasks).Error; err != nil {
		relayError(c, http.StatusInternalServerError, err)
		return
	}
	if err := models.DB.Model(&models.RelayConfigRevision{}).Where("relay_server_id = ?", id).Count(&revisions).Error; err != nil {
		relayError(c, http.StatusInternalServerError, err)
		return
	}
	if routes > 0 || tasks > 0 || revisions > 0 {
		c.JSON(http.StatusConflict, gin.H{"code": "00001", "msg": "server has routes or configuration history; disable it instead"})
		return
	}
	if err := models.DB.Delete(&models.RelayServer{}, id).Error; err != nil {
		relayError(c, http.StatusInternalServerError, err)
		return
	}
	recordRelayAdminEvent(c, id, "server_deleted", "")
	c.JSON(http.StatusOK, gin.H{"code": "00000", "msg": "deleted"})
}

func RelayServerRotateToken(c *gin.Context) {
	var server models.RelayServer
	if err := models.DB.First(&server, "id = ?", c.Param("id")).Error; err != nil {
		relayError(c, http.StatusNotFound, err)
		return
	}
	if !server.Enabled && server.AppliedRevision < server.DesiredRevision {
		c.JSON(http.StatusConflict, gin.H{"code": "00001", "msg": "cannot rotate the Agent token until the disabled configuration has been applied"})
		return
	}
	token, err := relaySecret()
	if err != nil {
		relayError(c, http.StatusInternalServerError, err)
		return
	}
	server.AgentTokenHash = hashRelayToken(token)
	server.AgentTokenCreatedAt = timePtr(time.Now().UTC())
	if err := models.DB.Save(&server).Error; err != nil {
		relayError(c, http.StatusInternalServerError, err)
		return
	}
	recordRelayAdminEvent(c, server.ID, "token_rotated", "")
	c.JSON(http.StatusOK, gin.H{"code": "00000", "data": gin.H{"agent_token": token}, "msg": "token rotated; token is shown once"})
}

func RelayServerEvents(c *gin.Context) {
	id, ok := relayID(c)
	if !ok {
		return
	}
	var events []models.RelayAgentEvent
	if err := models.DB.Where("relay_server_id = ?", id).Order("id desc").Limit(100).Find(&events).Error; err != nil {
		relayError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "00000", "data": events, "msg": "ok"})
}

func RelayServerReconcile(c *gin.Context) {
	id, ok := relayID(c)
	if !ok {
		return
	}
	var req struct {
		ChangeSummary string `json:"change_summary"`
	}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": "00001", "msg": "invalid request"})
			return
		}
	}
	actorID := relayActorID(c)
	var revision models.RelayConfigRevision
	var task models.RelayConfigTask
	var err error
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		var server models.RelayServer
		if err := tx.First(&server, id).Error; err != nil {
			return err
		}
		var config string
		if server.Enabled {
			var err error
			config, err = services.RenderRelayConfig(tx, id)
			if err != nil {
				return err
			}
		}
		revision, task, err = queueRelayRevision(tx, id, config, strings.TrimSpace(req.ChangeSummary), actorID)
		return err
	})
	if err != nil {
		relayError(c, http.StatusBadRequest, err)
		return
	}
	recordRelayAdminEvent(c, id, "config_reconciled", fmt.Sprintf("revision=%d task_id=%s", revision.Revision, task.TaskID))
	c.JSON(http.StatusCreated, gin.H{"code": "00000", "data": gin.H{"revision": revision, "task": task}, "msg": "configuration queued"})
}

func RelayServerRollback(c *gin.Context) {
	id, ok := relayID(c)
	if !ok {
		return
	}
	var req struct {
		Revision uint64 `json:"revision"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Revision == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "00001", "msg": "revision is required"})
		return
	}
	var revision models.RelayConfigRevision
	var task models.RelayConfigTask
	actorID := relayActorID(c)
	err := models.DB.Transaction(func(tx *gorm.DB) error {
		var source models.RelayConfigRevision
		if err := tx.Where("relay_server_id = ? AND revision = ?", id, req.Revision).First(&source).Error; err != nil {
			return err
		}
		var err error
		revision, task, err = queueRelayRevision(tx, id, source.ConfigContent, fmt.Sprintf("rollback to revision %d", source.Revision), actorID)
		return err
	})
	if err != nil {
		relayError(c, http.StatusBadRequest, err)
		return
	}
	recordRelayAdminEvent(c, id, "config_rollback_queued", fmt.Sprintf("source_revision=%d revision=%d task_id=%s", req.Revision, revision.Revision, task.TaskID))
	c.JSON(http.StatusCreated, gin.H{"code": "00000", "data": gin.H{"revision": revision, "task": task}, "msg": "rollback configuration queued"})
}

func RelayServerRevisions(c *gin.Context) {
	id, ok := relayID(c)
	if !ok {
		return
	}
	var revisions []models.RelayConfigRevision
	if err := models.DB.Where("relay_server_id = ?", id).Order("revision desc").Limit(100).Find(&revisions).Error; err != nil {
		relayError(c, http.StatusInternalServerError, err)
		return
	}
	for i := range revisions {
		revisions[i].ConfigContent = ""
	}
	c.JSON(http.StatusOK, gin.H{"code": "00000", "data": revisions, "msg": "ok"})
}

func RelayServerTasks(c *gin.Context) {
	id, ok := relayID(c)
	if !ok {
		return
	}
	var tasks []models.RelayConfigTask
	if err := models.DB.Where("relay_server_id = ?", id).Order("id desc").Limit(100).Find(&tasks).Error; err != nil {
		relayError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "00000", "data": tasks, "msg": "ok"})
}

func RelayRoutesList(c *gin.Context) {
	var routes []models.RelayRoute
	query := models.DB.Order("priority desc, id asc")
	if id := c.Query("relay_server_id"); id != "" {
		query = query.Where("relay_server_id = ?", id)
	}
	if err := query.Find(&routes).Error; err != nil {
		relayError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "00000", "data": routes, "msg": "ok"})
}

func RelayRouteGet(c *gin.Context) {
	var route models.RelayRoute
	if err := models.DB.First(&route, "id = ?", c.Param("id")).Error; err != nil {
		relayError(c, http.StatusNotFound, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "00000", "data": route, "msg": "ok"})
}

func RelayRouteCreate(c *gin.Context) {
	var req relayRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "00001", "msg": "invalid request"})
		return
	}
	route, err := relayRouteFromRequest(req)
	if err != nil {
		relayError(c, http.StatusBadRequest, err)
		return
	}
	var revision models.RelayConfigRevision
	var task models.RelayConfigTask
	actorID := relayActorID(c)
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		if err := ensureRelayLandingNode(tx, route.LandingNodeID); err != nil {
			return err
		}
		if err := tx.Create(&route).Error; err != nil {
			return err
		}
		config, err := services.RenderRelayConfig(tx, route.RelayServerID)
		if err != nil {
			return err
		}
		revision, task, err = queueRelayRevision(tx, route.RelayServerID, config, "route created", actorID)
		return err
	})
	if err != nil {
		relayError(c, http.StatusConflict, err)
		return
	}
	recordRelayAdminEvent(c, route.RelayServerID, "route_created", fmt.Sprintf("route_id=%d", route.ID))
	recordRelayAdminEvent(c, route.RelayServerID, "config_queued", fmt.Sprintf("revision=%d task_id=%s", revision.Revision, task.TaskID))
	c.JSON(http.StatusCreated, gin.H{"code": "00000", "data": route, "msg": "created"})
}

func RelayRouteUpdate(c *gin.Context) {
	var req relayRouteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "00001", "msg": "invalid request"})
		return
	}
	route, err := relayRouteFromRequest(req)
	if err != nil {
		relayError(c, http.StatusBadRequest, err)
		return
	}
	var existing models.RelayRoute
	if err := models.DB.First(&existing, "id = ?", c.Param("id")).Error; err != nil {
		relayError(c, http.StatusNotFound, err)
		return
	}
	if route.RelayServerID != existing.RelayServerID {
		c.JSON(http.StatusBadRequest, gin.H{"code": "00001", "msg": "moving a route to another server requires delete and recreate"})
		return
	}
	route.ID, route.CreatedAt = existing.ID, existing.CreatedAt
	var revision models.RelayConfigRevision
	var task models.RelayConfigTask
	actorID := relayActorID(c)
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		if err := ensureRelayLandingNode(tx, route.LandingNodeID); err != nil {
			return err
		}
		if err := tx.Model(&existing).Select("Name", "RelayServerID", "LandingNodeID", "Protocol", "ListenAddress", "ListenPort", "Enabled", "PublishEnabled", "Priority", "Description").Updates(&route).Error; err != nil {
			return err
		}
		config, err := services.RenderRelayConfig(tx, route.RelayServerID)
		if err != nil {
			return err
		}
		revision, task, err = queueRelayRevision(tx, route.RelayServerID, config, "route updated", actorID)
		return err
	})
	if err != nil {
		relayError(c, http.StatusConflict, err)
		return
	}
	recordRelayAdminEvent(c, route.RelayServerID, "route_updated", fmt.Sprintf("route_id=%d", route.ID))
	recordRelayAdminEvent(c, route.RelayServerID, "config_queued", fmt.Sprintf("revision=%d task_id=%s", revision.Revision, task.TaskID))
	c.JSON(http.StatusOK, gin.H{"code": "00000", "data": route, "msg": "updated"})
}

func RelayRouteDelete(c *gin.Context) {
	id, ok := relayID(c)
	if !ok {
		return
	}
	var route models.RelayRoute
	if err := models.DB.First(&route, id).Error; err != nil {
		relayError(c, http.StatusNotFound, err)
		return
	}
	var revision models.RelayConfigRevision
	var task models.RelayConfigTask
	actorID := relayActorID(c)
	err := models.DB.Transaction(func(tx *gorm.DB) error {
		association := tx.Model(&route).Association("Subscriptions")
		if association.Error != nil {
			return association.Error
		}
		if association.Count() > 0 {
			// Keep the subscription's relay selection intact. Clearing this
			// relation would make subscription generation fall back to the raw
			// landing node, exposing its direct address.
			return errRelayRouteAssigned
		}
		if err := tx.Unscoped().Delete(&route).Error; err != nil {
			return err
		}
		config, err := services.RenderRelayConfig(tx, route.RelayServerID)
		if err != nil {
			return err
		}
		revision, task, err = queueRelayRevision(tx, route.RelayServerID, config, "route deleted", actorID)
		return err
	})
	if err != nil {
		if errors.Is(err, errRelayRouteAssigned) {
			relayError(c, http.StatusConflict, err)
			return
		}
		relayError(c, http.StatusInternalServerError, err)
		return
	}
	recordRelayAdminEvent(c, route.RelayServerID, "route_deleted", fmt.Sprintf("route_id=%d", route.ID))
	recordRelayAdminEvent(c, route.RelayServerID, "config_queued", fmt.Sprintf("revision=%d task_id=%s", revision.Revision, task.TaskID))
	c.JSON(http.StatusOK, gin.H{"code": "00000", "msg": "deleted"})
}

func RelayRouteSetEnabled(c *gin.Context) {
	id, ok := relayID(c)
	if !ok {
		return
	}
	action := c.Param("action")
	if action != "enable" && action != "disable" {
		c.JSON(http.StatusNotFound, gin.H{"code": "00001", "msg": "unknown action"})
		return
	}
	enabled := action == "enable"
	var route models.RelayRoute
	if err := models.DB.First(&route, id).Error; err != nil {
		relayError(c, http.StatusNotFound, err)
		return
	}
	actorID := relayActorID(c)
	var revision models.RelayConfigRevision
	var task models.RelayConfigTask
	err := models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&route).Update("enabled", enabled).Error; err != nil {
			return err
		}
		config, err := services.RenderRelayConfig(tx, route.RelayServerID)
		if err != nil {
			return err
		}
		revision, task, err = queueRelayRevision(tx, route.RelayServerID, config, "route "+action+"d", actorID)
		return err
	})
	if err != nil {
		relayError(c, http.StatusConflict, err)
		return
	}
	recordRelayAdminEvent(c, route.RelayServerID, "route_"+action+"d", fmt.Sprintf("route_id=%d", id))
	recordRelayAdminEvent(c, route.RelayServerID, "config_queued", fmt.Sprintf("revision=%d task_id=%s", revision.Revision, task.TaskID))
	c.JSON(http.StatusOK, gin.H{"code": "00000", "msg": "updated"})
}

func relayRouteFromRequest(req relayRouteRequest) (models.RelayRoute, error) {
	req.Name, req.Protocol = strings.TrimSpace(req.Name), strings.ToLower(strings.TrimSpace(req.Protocol))
	if req.Protocol == "" {
		req.Protocol = "tcp"
	}
	if req.Name == "" || len(req.Name) > 100 || len(req.Description) > 1000 || req.RelayServerID == 0 || req.LandingNodeID <= 0 || (req.Protocol != "tcp" && req.Protocol != "udp") || req.ListenPort == 0 {
		return models.RelayRoute{}, errors.New("name, server, landing node, TCP/UDP protocol and port 1-65535 are required")
	}
	if req.ListenAddress == "" {
		req.ListenAddress = "0.0.0.0"
	}
	if net.ParseIP(req.ListenAddress) == nil {
		return models.RelayRoute{}, errors.New("listen_address must be an IP address")
	}
	var server models.RelayServer
	if err := models.DB.First(&server, req.RelayServerID).Error; err != nil {
		return models.RelayRoute{}, errors.New("relay server not found")
	}
	if req.Protocol == "udp" && !relayAgentSupportsUDP(server.AgentVersion) {
		return models.RelayRoute{}, errors.New("UDP relay requires Agent 0.2.0 or newer; update the Agent first")
	}
	var node models.Node
	if err := models.DB.First(&node, "id = ?", req.LandingNodeID).Error; err != nil {
		return models.RelayRoute{}, errors.New("landing node not found")
	}
	if err := services.ValidateRelayNodeLinkForProtocol(node.Link, req.Protocol); err != nil {
		return models.RelayRoute{}, fmt.Errorf("landing node cannot be relayed: %w", err)
	}
	return models.RelayRoute{
		Name: req.Name, RelayServerID: req.RelayServerID, LandingNodeID: req.LandingNodeID,
		Protocol: req.Protocol, ListenAddress: req.ListenAddress, ListenPort: req.ListenPort,
		Enabled: req.Enabled, PublishEnabled: req.PublishEnabled, Priority: req.Priority, Description: req.Description,
	}, nil
}

func relayAgentSupportsUDP(version string) bool {
	parts := strings.SplitN(strings.TrimPrefix(strings.TrimSpace(version), "v"), ".", 3)
	if len(parts) != 3 {
		return false
	}
	major, errMajor := strconv.Atoi(parts[0])
	minor, errMinor := strconv.Atoi(parts[1])
	patch, errPatch := strconv.Atoi(strings.SplitN(parts[2], "-", 2)[0])
	if errMajor != nil || errMinor != nil || errPatch != nil {
		return false
	}
	return major > 0 || minor > 2 || (minor == 2 && patch >= 0)
}

func ensureRelayLandingNode(tx *gorm.DB, nodeID int) error {
	var node models.Node
	return tx.Select("id").First(&node, "id = ?", nodeID).Error
}

func validRelayServer(req relayServerRequest) bool {
	name, host := strings.TrimSpace(req.Name), strings.TrimSpace(req.PublicHost)
	return name != "" && len(name) <= 100 && host != "" && len(host) <= 255 &&
		validPublicHost(host)
}

func validPublicHost(host string) bool {
	if net.ParseIP(host) != nil {
		return true
	}
	if len(host) > 253 || strings.HasPrefix(host, ".") || strings.HasSuffix(host, ".") || strings.Contains(host, ":") {
		return false
	}
	for _, label := range strings.Split(host, ".") {
		if !relayHostnameLabel.MatchString(label) {
			return false
		}
	}
	return true
}

var relayHostnameLabel = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?$`)

func relayID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": "00001", "msg": "invalid id"})
		return 0, false
	}
	return uint(id), true
}

func relaySecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func hashRelayToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func timePtr(value time.Time) *time.Time { return &value }

func relayActorID(c *gin.Context) uint {
	username, _ := c.Get("username")
	var user models.User
	if err := models.DB.Select("id").Where("username = ?", fmt.Sprint(username)).First(&user).Error; err != nil || user.ID <= 0 {
		return 0
	}
	return uint(user.ID)
}

func queueRelayRevision(tx *gorm.DB, serverID uint, config, summary string, actorID uint) (models.RelayConfigRevision, models.RelayConfigTask, error) {
	var server models.RelayServer
	if err := tx.First(&server, serverID).Error; err != nil {
		return models.RelayConfigRevision{}, models.RelayConfigTask{}, err
	}
	if !server.Enabled && strings.TrimSpace(config) != "" {
		return models.RelayConfigRevision{}, models.RelayConfigTask{}, errors.New("relay server is disabled")
	}
	if strings.Contains(config, "# SUBLINKX-UDP-ROUTE\t") && !relayAgentSupportsUDP(server.AgentVersion) {
		return models.RelayConfigRevision{}, models.RelayConfigTask{}, errors.New("UDP relay requires Agent 0.2.0 or newer; update the Agent first")
	}
	revisionNumber := server.DesiredRevision + 1
	sum := sha256.Sum256([]byte(config))
	revision := models.RelayConfigRevision{
		RelayServerID: serverID, Revision: revisionNumber, ConfigSHA256: hex.EncodeToString(sum[:]),
		ConfigContent: config, ChangeSummary: summary, CreatedBy: actorID,
	}
	if err := tx.Create(&revision).Error; err != nil {
		return models.RelayConfigRevision{}, models.RelayConfigTask{}, err
	}
	// Coalesce only tasks that no Agent has leased yet. Leased work remains immutable.
	if err := tx.Model(&models.RelayConfigTask{}).Where("relay_server_id = ? AND status = ?", serverID, "pending").Updates(map[string]any{"status": "expired", "finished_at": time.Now().UTC(), "error_code": "SUPERSEDED", "error_message": "superseded by a newer desired revision"}).Error; err != nil {
		return models.RelayConfigRevision{}, models.RelayConfigTask{}, err
	}
	taskID, err := relaySecret()
	if err != nil {
		return models.RelayConfigRevision{}, models.RelayConfigTask{}, err
	}
	task := models.RelayConfigTask{
		TaskID:        fmt.Sprintf("%s-%s-%s-%s-%s", taskID[:8], taskID[8:12], taskID[12:16], taskID[16:20], taskID[20:32]),
		RelayServerID: serverID, RevisionID: revision.ID, TaskType: "apply_config", Status: "pending",
	}
	if err := tx.Create(&task).Error; err != nil {
		return models.RelayConfigRevision{}, models.RelayConfigTask{}, err
	}
	if err := tx.Model(&server).Update("desired_revision", revisionNumber).Error; err != nil {
		return models.RelayConfigRevision{}, models.RelayConfigTask{}, err
	}
	return revision, task, nil
}

func recordRelayAdminEvent(c *gin.Context, serverID uint, eventType, details string) {
	if serverID == 0 {
		return
	}
	username, _ := c.Get("username")
	actor := "agent"
	if username != nil {
		actor = fmt.Sprint(username)
	}
	details = strings.TrimSpace(details + " actor=" + actor)
	if err := models.DB.Create(&models.RelayAgentEvent{RelayServerID: serverID, EventType: eventType, Details: details}).Error; err != nil {
		log.Printf("relay audit event write failed (%s): %v", eventType, err)
	}
}

func relayError(c *gin.Context, status int, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		status = http.StatusNotFound
	}
	c.JSON(status, gin.H{"code": fmt.Sprintf("%05d", status), "msg": err.Error()})
}
