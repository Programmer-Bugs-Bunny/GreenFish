package example

import (
	"net/http"

	"go-web-template/middlewares"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Ping ping接口 - 简单的存活检查
func Ping(c *gin.Context) {
	middlewares.Logger.Debug("健康检查请求")

	c.JSON(http.StatusOK, gin.H{
		"message":      "pong",
		"version":      "1.0.0",
		"environment":  "development",
		"timezone":     "UTC",
		"current_time": getCurrentTimeString(),
	})
}

// Health health接口 - 完整的健康检查
func Health(c *gin.Context) {
	middlewares.Logger.Debug("完整健康检查请求",
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.GetHeader("User-Agent")),
	)

	// 返回简单的健康状态
	response := HealthResponse{
		Status:    HealthStatusHealthy,
		Version:   "1.0.0",
		Timestamp: getCurrentTime(),
		Checks: []HealthCheck{
			{
				Name:    "database",
				Status:  HealthStatusHealthy,
				Message: "数据库连接正常",
			},
		},
	}

	middlewares.Logger.Debug("健康检查通过",
		zap.String("status", string(response.Status)),
	)

	c.JSON(http.StatusOK, response)
}
