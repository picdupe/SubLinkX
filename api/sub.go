// api/subcription.go

package api

import (
	// 导入 json 包，用于解析 config 字符串

	"errors"
	"log"
	"strconv"
	"strings"
	"sublink/models" // 导入 models 包
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

func SubTotal(c *gin.Context) {
	var Sub models.Subcription
	subs, err := Sub.List()
	count := len(subs)
	if err != nil {
		c.JSON(500, gin.H{
			"msg": "取得订阅总数失败",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": "00000",
		"data": count,
		"msg":  "取得订阅总数",
	})
}

// 获取订阅列表
func SubGet(c *gin.Context) {
	var Sub models.Subcription
	Subs, err := Sub.List()
	if err != nil {
		c.JSON(500, gin.H{
			"msg": "node list error",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": "00000",
		"data": Subs,
		"msg":  "node get",
	})
}

// 添加订阅
func SubAdd(c *gin.Context) {
	name := c.PostForm("name")
	configs := c.PostForm("config") // 这里的 configString 是前端传来的 JSON 字符串
	nodes := c.PostForm("nodes")
	relayRoutes, routeErr := selectedRelayRoutes(c.PostForm("relay_routes"))
	if routeErr != nil {
		c.JSON(400, gin.H{"msg": routeErr.Error()})
		return
	}

	if name == "" || (nodes == "" && len(relayRoutes) == 0) {
		c.JSON(400, gin.H{
			"msg": "订阅名称或节点不能为空",
		})
		return
	}

	// 1. 根据 nodesString 字符串，构建 models.Node 数组
	var NodesData []models.Node

	for _, nodeName := range strings.Split(nodes, ",") {
		if strings.TrimSpace(nodeName) == "" {
			continue
		}
		FirstNode := models.Node{
			Name: nodeName,
		}

		// 查出node的数据
		result := models.DB.Model(models.Node{}).Where("name = ?", FirstNode.Name).First(&FirstNode)
		if result.Error != nil {
			log.Println(result.Error)
			c.JSON(400, gin.H{
				"msg": result.Error,
			})
			return
		}
		// 插入nodes
		NodesData = append(NodesData, FirstNode)
	}
	if err := validateRelayRouteCoverage(NodesData, relayRoutes); err != nil {
		c.JSON(400, gin.H{"msg": err.Error()})
		return
	}
	sub := models.Subcription{
		Name:        name,
		Config:      configs,   // 这里直接赋值字符串
		NodeOrder:   nodes,     // 这里直接赋值字符串
		Nodes:       NodesData, // 这里直接赋值 nodes 数组
		RelayRoutes: relayRoutes,
	}
	initialToken, err := utils.RandString(subscriptionTokenLength)
	if err != nil {
		c.JSON(500, gin.H{"msg": "生成订阅令牌失败"})
		return
	}
	initialToken = strings.ToLower(initialToken)
	sub.AccessTokenHash = hashSubscriptionToken(initialToken)
	err = sub.Add()
	if err != nil {
		c.JSON(400, gin.H{
			"msg": "添加订阅失败: " + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "添加订阅成功",
	})
}

func SubRotateToken(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"msg": "无效的订阅 ID"})
		return
	}
	token, err := utils.RandString(subscriptionTokenLength)
	if err != nil {
		c.JSON(500, gin.H{"msg": "生成订阅令牌失败"})
		return
	}
	token = strings.ToLower(token)
	result := models.DB.Model(&models.Subcription{}).Where("id = ?", id).Update("access_token_hash", hashSubscriptionToken(token))
	if result.Error != nil {
		c.JSON(500, gin.H{"msg": "更新订阅令牌失败"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(404, gin.H{"msg": "订阅不存在"})
		return
	}
	c.JSON(200, gin.H{"code": "00000", "data": gin.H{"token": token}, "msg": "订阅令牌已更新"})
}

// 更新订阅
func SubUpdate(c *gin.Context) {
	NewName := c.PostForm("name")
	OldName := c.PostForm("oldname")
	configs := c.PostForm("config") // 这里的 configString 是前端传来的 JSON 字符串
	nodes := c.PostForm("nodes")
	relayRoutes, routeErr := selectedRelayRoutes(c.PostForm("relay_routes"))
	if routeErr != nil {
		c.JSON(400, gin.H{"msg": routeErr.Error()})
		return
	}

	if NewName == "" || (nodes == "" && len(relayRoutes) == 0) {
		c.JSON(400, gin.H{
			"msg": "订阅名称或节点不能为空",
		})
		return
	}

	// 1. 根据 nodesString 字符串，构建 models.Node 数组
	var NodesData []models.Node

	for _, nodeName := range strings.Split(nodes, ",") {
		if strings.TrimSpace(nodeName) == "" {
			continue
		}
		FirstNode := models.Node{
			Name: nodeName,
		}

		// 查出node的数据
		result := models.DB.Model(models.Node{}).Where("name = ?", FirstNode.Name).First(&FirstNode)
		if result.Error != nil {
			log.Println(result.Error)
			c.JSON(400, gin.H{
				"msg": result.Error,
			})
			return
		}
		// 插入nodes
		NodesData = append(NodesData, FirstNode)
	}
	if err := validateRelayRouteCoverage(NodesData, relayRoutes); err != nil {
		c.JSON(400, gin.H{"msg": err.Error()})
		return
	}
	OldSub := models.Subcription{
		Name: OldName,
	}
	NewSub := models.Subcription{
		Name:        NewName,
		Config:      configs,   // 这里直接赋值字符串
		NodeOrder:   nodes,     // 这里直接赋值字符串
		Nodes:       NodesData, // 这里直接赋值 nodes 数组
		RelayRoutes: relayRoutes,
	}

	err := OldSub.Update(&NewSub)
	if err != nil {
		c.JSON(400, gin.H{
			"msg": "更新订阅失败: " + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "更新订阅成功",
	})
}

func validateRelayRouteCoverage(nodes []models.Node, routes []models.RelayRoute) error {
	if len(routes) == 0 || len(nodes) == 0 {
		return nil
	}
	covered := make(map[int]struct{}, len(routes))
	for _, route := range routes {
		covered[route.LandingNodeID] = struct{}{}
	}
	for _, n := range nodes {
		if _, ok := covered[n.ID]; !ok {
			return errors.New("已选择中转线路时，订阅中的每个原始节点都必须选择对应的中转线路")
		}
	}
	for _, route := range routes {
		found := false
		for _, n := range nodes {
			if n.ID == route.LandingNodeID {
				found = true
				break
			}
		}
		if !found {
			return errors.New("所选中转线路对应的原始节点必须先加入订阅")
		}
	}
	return nil
}

func selectedRelayRoutes(raw string) ([]models.RelayRoute, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var ids []uint
	seen := make(map[uint]struct{})
	for _, value := range strings.Split(raw, ",") {
		parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 32)
		if err != nil || parsed == 0 {
			return nil, errors.New("无效的中转线路 ID")
		}
		id := uint(parsed)
		if _, exists := seen[id]; !exists {
			ids = append(ids, id)
			seen[id] = struct{}{}
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}
	var routes []models.RelayRoute
	if err := models.DB.Where("id IN ? AND enabled = ? AND publish_enabled = ?", ids, true, true).Find(&routes).Error; err != nil {
		return nil, err
	}
	if len(routes) != len(ids) {
		return nil, errors.New("所选中转线路不存在、未启用或未允许订阅发布")
	}
	serverIDs := make([]uint, 0, len(routes))
	for _, route := range routes {
		serverIDs = append(serverIDs, route.RelayServerID)
	}
	var servers []models.RelayServer
	if err := models.DB.Where("id IN ? AND enabled = ?", serverIDs, true).Find(&servers).Error; err != nil {
		return nil, err
	}
	enabledServers := make(map[uint]struct{}, len(servers))
	for _, server := range servers {
		enabledServers[server.ID] = struct{}{}
	}
	for _, route := range routes {
		if _, ok := enabledServers[route.RelayServerID]; !ok {
			return nil, errors.New("所选线路的中转服务器未启用")
		}
	}
	return routes, nil
}

// 删除订阅 (无需修改)
func SubDel(c *gin.Context) {
	var sub models.Subcription
	id := c.Query("id")
	if id == "" {
		c.JSON(400, gin.H{
			"msg": "id 不能为空",
		})
		return
	}
	x, err := strconv.Atoi(id) // 增加错误检查
	if err != nil {
		c.JSON(400, gin.H{
			"msg": "无效的 ID: " + err.Error(),
		})
		return
	}
	sub.ID = x
	err = sub.Find()
	if err != nil {
		c.JSON(400, gin.H{
			"msg": "查找订阅失败: " + err.Error(),
		})
		return
	}
	err = sub.Del()
	if err != nil {
		c.JSON(400, gin.H{
			"msg": "删除订阅失败: " + err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "删除订阅成功",
	})
}
