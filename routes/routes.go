package routes

import (
	"go-web-template/config"
	"go-web-template/middlewares"
	"go-web-template/routes/rest"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupRoutes 设置主路由
func SetupRoutes(cfg *config.Config) *gin.Engine {
	// 根据配置设置 Gin 模式
	if !cfg.App.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建 Gin 路由器（不使用默认中间件）
	r := gin.New()

	// 设置 multipart form 内存限制为 100MB
	r.MaxMultipartMemory = 100 << 20 // 100 MB

	// 添加自定义中间件
	r.Use(middlewares.CORS())              // CORS跨域处理（需要在其他中间件之前）
	r.Use(middlewares.GinLogger())         // 结构化日志
	r.Use(middlewares.GinRecovery())       // 异常恢复
	r.Use(middlewares.ErrorLogging())      // 错误响应日志（用于记录逻辑异常）
	r.Use(middlewares.TracingMiddleware()) // 链路追踪中间件

	// API 路由组
	api := r.Group("/api")

	// 公开路由组（无需认证）
	public := api.Group("")
	rest.ApplyPublic(public)

	// 私有路由组（需要 JWT 认证）
	private := api.Group("/private")
	private.Use(middlewares.JWTAuth()) // 启用JWT认证中间件
	rest.ApplyPrivate(private)

	// 打印路由统计信息
	rest.PrintStats()

	middlewares.Logger.Info("主路由设置完成",
		zap.Any("rest_stats", rest.GetRegistrarStats()),
	)

	return r
}
