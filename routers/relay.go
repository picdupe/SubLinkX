package routers

import (
	"sublink/api"
	"sublink/middlewares"

	"github.com/gin-gonic/gin"
)

func Relay(r *gin.Engine) {
	group := r.Group("/api/v1/admin/relay")
	group.Use(middlewares.RequireAdmin)
	{
		group.GET("/servers", api.RelayServersList)
		group.POST("/servers", api.RelayServerCreate)
		group.GET("/servers/:id", api.RelayServerGet)
		group.PUT("/servers/:id", api.RelayServerUpdate)
		group.DELETE("/servers/:id", api.RelayServerDelete)
		group.POST("/servers/:id/rotate-token", api.RelayServerRotateToken)
		group.GET("/servers/:id/events", api.RelayServerEvents)
		group.POST("/servers/:id/reconcile", api.RelayServerReconcile)
		group.POST("/servers/:id/rollback", api.RelayServerRollback)
		group.GET("/servers/:id/revisions", api.RelayServerRevisions)
		group.GET("/servers/:id/tasks", api.RelayServerTasks)
		group.GET("/routes", api.RelayRoutesList)
		group.POST("/routes", api.RelayRouteCreate)
		group.GET("/routes/:id", api.RelayRouteGet)
		group.PUT("/routes/:id", api.RelayRouteUpdate)
		group.DELETE("/routes/:id", api.RelayRouteDelete)
		group.POST("/routes/:id/:action", api.RelayRouteSetEnabled)
	}

	agent := r.Group("/api/agent/v1")
	agent.Use(middlewares.RelayAgentAuth)
	{
		agent.POST("/enroll", api.RelayAgentEnroll)
		agent.POST("/heartbeat", api.RelayAgentHeartbeat)
		agent.GET("/tasks/next", api.RelayAgentTaskNext)
		agent.POST("/tasks/:id/ack", api.RelayAgentTaskAck)
		agent.GET("/configs/:revision", api.RelayAgentConfigGet)
		agent.POST("/tasks/:id/result", api.RelayAgentTaskResult)
		agent.GET("/status", api.RelayAgentStatus)
	}
}
