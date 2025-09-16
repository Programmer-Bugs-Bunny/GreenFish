package rest

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 子路由注册函数签名
type Registrar func(*gin.RouterGroup)

// 两类注册器：公开/私有（是否需要 JWT）
var (
	publicRegistrars  []Registrar
	privateRegistrars []Registrar
)

// —— 添加注册器（给各业务子路由在 init() 调用）——

// RegisterPublic 注册公开路由（无需认证）
func RegisterPublic(fn Registrar) {
	publicRegistrars = append(publicRegistrars, fn)
	// 使用 zap.L() 全局logger，避免使用可能未初始化的 middlewares.Logger
	zap.L().Debug("公开路由注册器已添加",
		zap.Int("total_public", len(publicRegistrars)),
	)
}

// RegisterPrivate 注册私有路由（需要 JWT 认证）
func RegisterPrivate(fn Registrar) {
	privateRegistrars = append(privateRegistrars, fn)
	zap.L().Debug("私有路由注册器已添加",
		zap.Int("total_private", len(privateRegistrars)),
	)
}

// —— 批量应用（供主路由调用）——

// ApplyPublic 应用所有公开路由
func ApplyPublic(group *gin.RouterGroup) {
	zap.L().Info("开始应用公开路由",
		zap.Int("count", len(publicRegistrars)),
	)

	for i, fn := range publicRegistrars {
		fn(group)
		zap.L().Debug("公开路由已应用",
			zap.Int("index", i+1),
		)
	}

	zap.L().Info("公开路由应用完成",
		zap.Int("applied_count", len(publicRegistrars)),
	)
}

// ApplyPrivate 应用所有私有路由（带 JWT 中间件）
func ApplyPrivate(group *gin.RouterGroup) {
	zap.L().Info("开始应用私有路由",
		zap.Int("count", len(privateRegistrars)),
	)

	for i, fn := range privateRegistrars {
		fn(group)
		zap.L().Debug("私有路由已应用",
			zap.Int("index", i+1),
		)
	}

	zap.L().Info("私有路由应用完成",
		zap.Int("applied_count", len(privateRegistrars)),
	)
}

// GetRegistrarStats 获取注册器统计信息（调试用）
func GetRegistrarStats() map[string]int {
	return map[string]int{
		"public":  len(publicRegistrars),
		"private": len(privateRegistrars),
		"total":   len(publicRegistrars) + len(privateRegistrars),
	}
}

// PrintStats 打印注册器统计信息
func PrintStats() {
	stats := GetRegistrarStats()
	zap.L().Info("路由注册器统计",
		zap.Int("public_count", stats["public"]),
		zap.Int("private_count", stats["private"]),
		zap.Int("total_count", stats["total"]),
	)
}
