package api

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sublink/models"
	"sublink/node"
	"sublink/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func DocodeNodeName(nd *models.Node) (models.Node, error) { // 解码节点名称
	if nd.Name == "" {
		u, err := url.Parse(nd.Link)
		if err != nil {
			log.Println(err)
			return *nd, err
		}
		switch {
		case u.Scheme == "ss":
			ss, err := node.DecodeSSURL(nd.Link)
			if err != nil {
				log.Println(err)
				return *nd, err
			}
			nd.Name = ss.Name
		case u.Scheme == "ssr":
			ssr, err := node.DecodeSSRURL(nd.Link)
			if err != nil {
				log.Println(err)
				return *nd, err
			}
			nd.Name = ssr.Qurey.Remarks
		case u.Scheme == "trojan":
			trojan, err := node.DecodeTrojanURL(nd.Link)
			if err != nil {
				log.Println(err)
				return *nd, err
			}
			nd.Name = trojan.Name
		case u.Scheme == "vmess":
			vmess, err := node.DecodeVMESSURL(nd.Link)
			if err != nil {
				log.Println(err)
				return *nd, err
			}
			nd.Name = vmess.Ps

		case u.Scheme == "vless":
			vless, err := node.DecodeVLESSURL(nd.Link)
			if err != nil {
				log.Println(err)
				return *nd, err
			}
			nd.Name = vless.Name
		case u.Scheme == "hy" || u.Scheme == "hysteria":
			hy, err := node.DecodeHYURL(nd.Link)
			if err != nil {
				log.Println(err)
				return *nd, err
			}
			nd.Name = hy.Name
		case u.Scheme == "hy2" || u.Scheme == "hysteria2":
			hy2, err := node.DecodeHY2URL(nd.Link)
			if err != nil {
				log.Println(err)
				return *nd, err
			}
			nd.Name = hy2.Name
		case u.Scheme == "tuic":
			tuic, err := node.DecodeTuicURL(nd.Link)
			if err != nil {
				log.Println(err)
				return *nd, err
			}
			nd.Name = tuic.Name
		}
	}
	return *nd, nil
}
func NodeUpdadte(c *gin.Context) {
	// var node models.Node
	NewName := c.PostForm("name")
	Newlink := c.PostForm("link")
	id := c.PostForm("id")
	group := c.PostForm("group")        // 分组
	groups := strings.Split(group, ",") // 分组列表
	index, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(400, gin.H{
			"msg": "id 不能为空或者格式不正确",
		})
		return

	}
	if NewName == "" || Newlink == "" {
		c.JSON(400, gin.H{
			"msg": "节点名称 or 备注不能为空",
		})
		return
	}
	OldNode := &models.Node{
		ID: index,
	}
	var currentNode models.Node
	if err := models.DB.First(&currentNode, index).Error; err != nil {
		c.JSON(404, gin.H{"msg": "节点不存在"})
		return
	}
	linkChanged := currentNode.Link != Newlink
	if linkChanged {
		var referencedRoutes []models.RelayRoute
		if err := models.DB.Where("landing_node_id = ?", index).Find(&referencedRoutes).Error; err != nil {
			c.JSON(500, gin.H{"msg": "检查中转线路引用失败"})
			return
		}
		validatedProtocols := make(map[string]struct{}, len(referencedRoutes))
		for _, route := range referencedRoutes {
			if _, checked := validatedProtocols[route.Protocol]; checked {
				continue
			}
			if err := services.ValidateRelayNodeLinkForProtocol(Newlink, route.Protocol); err != nil {
				c.JSON(400, gin.H{"msg": fmt.Sprintf("节点链接不兼容中转线路 %q: %v", route.Name, err)})
				return
			}
			validatedProtocols[route.Protocol] = struct{}{}
		}
	}
	var gns []models.GroupNode
	if groups != nil || len(groups) > 0 {
		for _, g := range groups {
			TempGn := models.GroupNode{
				Name: strings.TrimSpace(g), // 去除分组名称两端空格
			}
			gns = append(gns, TempGn) // 生成分组列表
		}

	}
	err = OldNode.UpdateGroup(gns) // 更新分组
	if err != nil {
		c.JSON(400, gin.H{
			"msg": fmt.Sprintf("更新失败: %s", err.Error()),
		})
		return
	}
	type queuedRelayConfig struct {
		serverID uint
		revision uint64
		taskID   string
	}
	var queued []queuedRelayConfig
	actorID := relayActorID(c)
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		var current models.Node
		if err := tx.First(&current, index).Error; err != nil {
			return err
		}
		if err := tx.Model(&current).Updates(map[string]any{"name": NewName, "link": Newlink}).Error; err != nil {
			return err
		}
		if !linkChanged {
			return nil
		}
		var serverIDs []uint
		if err := tx.Model(&models.RelayRoute{}).Where("landing_node_id = ? AND enabled = ?", index, true).
			Distinct().Pluck("relay_server_id", &serverIDs).Error; err != nil {
			return err
		}
		for _, serverID := range serverIDs {
			var server models.RelayServer
			if err := tx.First(&server, serverID).Error; err != nil {
				return err
			}
			if !server.Enabled {
				continue
			}
			config, err := services.RenderRelayConfig(tx, serverID)
			if err != nil {
				return err
			}
			revision, task, err := queueRelayRevision(tx, serverID, config, fmt.Sprintf("landing node %d updated", index), actorID)
			if err != nil {
				return err
			}
			queued = append(queued, queuedRelayConfig{serverID: serverID, revision: revision.Revision, taskID: task.TaskID})
		}
		return nil
	})
	if err != nil {
		c.JSON(400, gin.H{
			"msg": fmt.Sprintf("更新失败: %s", err.Error()),
		})
		return
	}
	for _, item := range queued {
		recordRelayAdminEvent(c, item.serverID, "landing_node_updated", fmt.Sprintf("node_id=%d", index))
		recordRelayAdminEvent(c, item.serverID, "config_queued", fmt.Sprintf("revision=%d task_id=%s", item.revision, item.taskID))
	}

	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "更新成功",
	})
}

// 获取节点列表
func NodeGet(c *gin.Context) {
	var ns []models.Node
	ns, err := models.GetNodeList()
	if err != nil {
		c.JSON(500, gin.H{
			"msg": "node list error",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": "00000",
		"data": ns,
		"msg":  "node get",
	})
}

// 获取分组列表
func GroupNodeGet(c *gin.Context) {
	var Gns []models.GroupNode
	Gns, err := models.GetGroupNodeList()
	var data []string
	for _, g := range Gns {
		data = append(data, g.Name)
	}
	if err != nil {
		c.JSON(400, gin.H{
			"msg": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"code": "00000",
		"data": data,
		"msg":  "GroupNode get",
	})
}

// 设置关联分组
func GroupNodeSet(c *gin.Context) {
	// var n models.Node
	var gns []models.GroupNode
	var FirstGroup models.GroupNode
	name := c.PostForm("name")
	group := c.PostForm("group")

	// 将group分割成多个分组
	groups := strings.Split(group, ",")
	if len(groups) == 0 {
		c.JSON(400, gin.H{
			"msg": "分组不能为空",
		})
		return
	}
	log.Println("分组列表:", groups, "数组长度", len(groups))

	// 循环生成或绑定分组
	for _, g := range groups {
		// 如果group为空，跳过
		if strings.TrimSpace(g) == "" {
			log.Println("分组名为空，跳过")
			continue
		}
		log.Println("分组名:", g)
		FirstGroup.Name = g
		err := FirstGroup.Add()
		if err != nil {
			log.Println("添加分组失败:", err)
			c.JSON(400, gin.H{
				"msg": err.Error(),
			})
			return
		}
		// 查找分组并将数据FirstGroup填充 并且插入给gns
		result := models.DB.Model(models.GroupNode{}).Where("name = ?", g).First(&FirstGroup)
		log.Println("FirstGroup", FirstGroup)
		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			log.Println(result.Error)
			c.JSON(400, gin.H{
				"msg": result.Error,
			})
			return
		}
		gns = append(gns, FirstGroup)
	}

	n := models.Node{Name: name}
	err := n.UpdateGroup(gns)
	if err != nil {
		c.JSON(400, gin.H{
			"msg": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "更新关联分组成功",
	})
}

// 添加节点
func NodeAdd(c *gin.Context) {
	var n models.Node
	link := c.PostForm("link")
	name := c.PostForm("name")
	group := c.PostForm("group")
	n = models.Node{
		Name: name,
		Link: link,
	}
	parsedLink, parseErr := url.ParseRequestURI(link)
	validScheme := false
	if parsedLink != nil {
		switch strings.ToLower(parsedLink.Scheme) {
		case "ss", "ssr", "trojan", "vmess", "vless", "hy", "hysteria", "hy2", "hysteria2", "tuic", "http", "https":
			validScheme = true
		}
	}
	if parseErr != nil || parsedLink == nil || !validScheme {
		c.JSON(400, gin.H{
			"msg": "link格式不正确或协议不受支持",
		})
		return
	}
	// 解码节点名称
	n, err := DocodeNodeName(&n)
	if err != nil {
		log.Println("解码节点名称错误:", err)
		c.JSON(400, gin.H{
			"msg": "解码节点名称错误",
		})
		return
	}

	// 添加节点
	err = n.Add()
	if err != nil {
		log.Println("添加节点失败:", err)
		c.JSON(400, gin.H{
			"msg": err.Error(),
		})
		return
	}

	// 关联分组开始
	if strings.TrimSpace(group) != "" { // 去除空格后判断分组是否为空
		groups := strings.Split(group, ",") // 允许多个分组用逗号分隔
		if groups != nil || len(groups) > 0 {
			for _, g := range groups {
				gn := &models.GroupNode{Name: g}
				err = gn.Add()
				if err != nil {
					// 分组不存在
					log.Println(err)
					c.JSON(400, gin.H{
						"msg": err,
					})
					return
				}
				// 分组存在，关联节点
				if err := gn.Ass(&n); err != nil {
					log.Println("关联失败:", err)
					c.JSON(400, gin.H{
						"msg": err,
					})
					return
				}

			}
		}
	}
	//关联分组结束

	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "添加成功",
	})
}

// 删除节点
func NodeDel(c *gin.Context) {
	var n models.Node
	id := c.Query("id")
	if id == "" {
		c.JSON(400, gin.H{
			"msg": "id 不能为空",
		})
		return
	}
	x, _ := strconv.Atoi(id)
	n.ID = x
	var relayReferences int64
	if err := models.DB.Model(&models.RelayRoute{}).Where("landing_node_id = ?", n.ID).Count(&relayReferences).Error; err != nil {
		c.JSON(500, gin.H{"msg": "检查中转线路引用失败"})
		return
	}
	if relayReferences > 0 {
		c.JSON(409, gin.H{"msg": "节点仍被中转线路引用，请先删除或更换线路目标"})
		return
	}
	err := n.Del()
	if err != nil {
		c.JSON(400, gin.H{
			"msg": "删除失败",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "删除成功",
	})
}

// 删除分组
func NodesGroup(c *gin.Context) {
	var gn models.GroupNode
	id := c.Query("id")
	if id == "" {
		c.JSON(400, gin.H{
			"msg": "id 不能为空",
		})
		return
	}
	x, _ := strconv.Atoi(id)
	gn.ID = x
	err := gn.Del()
	if err != nil {
		c.JSON(400, gin.H{
			"msg": "删除失败",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": "00000",
		"msg":  "删除成功",
	})
}

// 节点统计
func NodesTotal(c *gin.Context) {
	var nodes []models.Node
	nodes, err := models.GetNodeList()
	count := len(nodes)
	if err != nil {
		c.JSON(500, gin.H{
			"msg": "获取不到节点统计",
		})
		return
	}
	c.JSON(200, gin.H{
		"code": "00000",
		"data": count,
		"msg":  "取得节点统计",
	})
}
