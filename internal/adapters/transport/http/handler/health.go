package handler

import (
	"net/http"

	runtimeport "fkteams/internal/ports/runtime"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
func HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		OK(c, gin.H{"status": "ok"})
	}
}

// ReadinessHandler 检查服务处理模型请求所需的核心 runtime。
func (rt *Runtime) ReadinessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := rt.InitializationError(); err != nil {
			c.JSON(http.StatusServiceUnavailable, Response{
				Code:      1,
				ErrorCode: "unavailable",
				Message:   "runtime initialization failed",
				Data:      gin.H{"status": "not_ready"},
			})
			return
		}
		if rt == nil || rt.Runtime == nil {
			c.JSON(http.StatusServiceUnavailable, Response{
				Code:      1,
				ErrorCode: "unavailable",
				Message:   "runtime is not configured",
				Data:      gin.H{"status": "not_ready"},
			})
			return
		}
		data := gin.H{"status": "ready"}
		if inspector, ok := rt.Runtime.(runtimeport.RuntimeInspector); ok {
			health := inspector.CheckHealth(c.Request.Context())
			data["runtime"] = health
			if !health.Ready {
				c.JSON(http.StatusServiceUnavailable, Response{
					Code:      1,
					ErrorCode: "unavailable",
					Message:   "runtime is not ready",
					Data:      data,
				})
				return
			}
		}
		OK(c, data)
	}
}
