package rest

import (
	"go-web-template/modules/example"

	"github.com/gin-gonic/gin"
)

func init() {
	// 注册公开路由（无需认证）
	RegisterPublic(registerExamplePublicRoutes)

	// 注册私有路由（需要JWT认证）
	RegisterPrivate(registerExamplePrivateRoutes)
}

// registerExamplePublicRoutes 注册示例公开路由
func registerExamplePublicRoutes(r *gin.RouterGroup) {
	// 健康检查相关路由
	r.GET("/ping", example.Ping)     // ping接口 - 简单的存活检查
	r.GET("/health", example.Health) // health接口 - 完整的健康检查
}

// registerExamplePrivateRoutes 注册示例私有路由
func registerExamplePrivateRoutes(r *gin.RouterGroup) {
}
