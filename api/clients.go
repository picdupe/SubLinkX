package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"sublink/models"
	"sublink/node"
	"sublink/services"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

const subscriptionTokenLength = 32

func hashSubscriptionToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func GetClient(c *gin.Context) {
	token := strings.ToLower(c.Query("token"))
	if len(token) != subscriptionTokenLength {
		c.String(http.StatusNotFound, "订阅不存在")
		return
	}
	tokenHash := hashSubscriptionToken(token)
	var matched *models.Subcription
	var matchedCount int32
	var subscriptions []models.Subcription
	if err := models.DB.Preload("Nodes").Preload("RelayRoutes").Where("access_token_hash = ?", tokenHash).Find(&subscriptions).Error; err != nil {
		c.String(http.StatusInternalServerError, "读取订阅失败")
		return
	}
	for i := range subscriptions {
		if subtle.ConstantTimeCompare([]byte(subscriptions[i].AccessTokenHash), []byte(tokenHash)) == 1 {
			matched = &subscriptions[i]
			matchedCount++
		}
	}
	if matched != nil && matchedCount == 1 {
		sub := *matched
		client := c.Query("client")
		if client == "" {
			ua := strings.ToLower(c.GetHeader("User-Agent"))
			switch {
			case strings.Contains(ua, "clash"):
				client = "clash"
			case strings.Contains(ua, "surge"):
				client = "surge"
			default:
				client = "v2ray"
			}
		}
		switch client {
		case "clash":
			writeClash(c, sub)
		case "surge":
			writeSurge(c, sub)
		case "v2ray":
			writeV2ray(c, sub)
		default:
			c.String(http.StatusBadRequest, "不支持的客户端类型")
		}
		return
	}
	c.String(http.StatusNotFound, "订阅不存在")
}

func fetchSubscription(rawURL string) (string, error) {
	body, err := utils.FetchPublicHTTP(rawURL, 5<<20)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func subscriptionLinks(sub models.Subcription) ([]string, error) {
	hasRelaySelection := len(sub.RelayRoutes) > 0
	routes := make([]models.RelayRoute, 0, len(sub.RelayRoutes))
	serverIDs := make([]uint, 0, len(sub.RelayRoutes))
	nodeIDs := make([]int, 0, len(sub.RelayRoutes))
	for _, route := range sub.RelayRoutes {
		if route.Enabled && route.PublishEnabled {
			routes = append(routes, route)
			serverIDs = append(serverIDs, route.RelayServerID)
			nodeIDs = append(nodeIDs, route.LandingNodeID)
		}
	}
	if len(routes) == 0 {
		if hasRelaySelection {
			return []string{}, nil
		}
		return subscriptionNodeLinks(sub.Nodes)
	}
	var servers []models.RelayServer
	if err := models.DB.Where("id IN ? AND enabled = ?", serverIDs, true).Find(&servers).Error; err != nil {
		return nil, err
	}
	serverByID := make(map[uint]models.RelayServer, len(servers))
	for _, server := range servers {
		serverByID[server.ID] = server
	}
	var landingNodes []models.Node
	if err := models.DB.Where("id IN ?", nodeIDs).Find(&landingNodes).Error; err != nil {
		return nil, err
	}
	nodeByID := make(map[int]models.Node, len(landingNodes))
	for _, landing := range landingNodes {
		nodeByID[landing.ID] = landing
	}
	var relayLinks []string
	for _, route := range routes {
		server, ok := serverByID[route.RelayServerID]
		if !ok {
			continue
		}
		landing, ok := nodeByID[route.LandingNodeID]
		if !ok {
			continue
		}
		link, err := services.BuildRelaySubscriptionLink(landing.Link, server.PublicHost, route.ListenPort, route.Name, route.Protocol)
		if err != nil {
			return nil, fmt.Errorf("派生中转线路 %q 失败: %w", route.Name, err)
		}
		relayLinks = append(relayLinks, link)
	}
	return relayLinks, nil
}

func subscriptionNodeLinks(nodes []models.Node) ([]string, error) {
	var urls []string
	for _, n := range nodes {
		if strings.Contains(n.Link, ",") {
			urls = append(urls, strings.Split(n.Link, ",")...)
			continue
		}
		if strings.HasPrefix(strings.ToLower(n.Link), "http://") || strings.HasPrefix(strings.ToLower(n.Link), "https://") {
			body, err := fetchSubscription(n.Link)
			if err != nil {
				return nil, err
			}
			urls = append(urls, strings.Split(node.Base64Decode(body), "\n")...)
			continue
		}
		urls = append(urls, n.Link)
	}
	return urls, nil
}

func writeV2ray(c *gin.Context, sub models.Subcription) {
	lines, err := subscriptionLinks(sub)
	if err != nil {
		c.String(http.StatusBadGateway, "获取订阅节点失败: %s", err.Error())
		return
	}
	setSubscriptionHeaders(c, sub.Name, "text/plain; charset=utf-8")
	c.String(http.StatusOK, node.Base64Encode(strings.Join(lines, "\n")+"\n"))
}

func writeClash(c *gin.Context, sub models.Subcription) {
	urls, err := subscriptionLinks(sub)
	if err != nil {
		c.String(http.StatusBadGateway, "获取上游订阅失败")
		return
	}
	var configs node.SqlConfig
	if err := json.Unmarshal([]byte(sub.Config), &configs); err != nil {
		c.String(http.StatusInternalServerError, "配置读取错误")
		return
	}
	output, err := node.EncodeClash(urls, configs)
	if err != nil {
		c.String(http.StatusBadRequest, "%s", err.Error())
		return
	}
	setSubscriptionHeaders(c, sub.Name, "text/plain; charset=utf-8")
	c.String(http.StatusOK, string(output))
}

func writeSurge(c *gin.Context, sub models.Subcription) {
	urls, err := subscriptionLinks(sub)
	if err != nil {
		c.String(http.StatusBadGateway, "获取上游订阅失败")
		return
	}
	var configs node.SqlConfig
	if err := json.Unmarshal([]byte(sub.Config), &configs); err != nil {
		c.String(http.StatusInternalServerError, "配置读取错误")
		return
	}
	output, err := node.EncodeSurge(urls, configs)
	if err != nil {
		c.String(http.StatusBadRequest, "%s", err.Error())
		return
	}
	setSubscriptionHeaders(c, sub.Name, "text/plain; charset=utf-8")
	if !strings.Contains(output, "#!MANAGED-CONFIG") {
		output = fmt.Sprintf("#!MANAGED-CONFIG %s interval=86400 strict=false\n%s", c.Request.Host+c.Request.URL.String(), output)
	}
	c.String(http.StatusOK, output)
}

func setSubscriptionHeaders(c *gin.Context, name, contentType string) {
	filename := url.PathEscape(name)
	c.Header("Content-Disposition", "inline; filename*=utf-8''"+filename)
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	c.Set("subname", name)
}
