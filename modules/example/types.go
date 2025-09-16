package example

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// HealthStatus 健康状态枚举
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// HealthCheck 健康检查项
type HealthCheck struct {
	Name     string       `json:"name"`
	Status   HealthStatus `json:"status"`
	Message  string       `json:"message,omitempty"`
	Duration string       `json:"duration,omitempty"`
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    HealthStatus  `json:"status"`
	Version   string        `json:"version"`
	Timestamp time.Time     `json:"timestamp"`
	Checks    []HealthCheck `json:"checks"`
}

// PingResponse ping接口响应
type PingResponse struct {
	Message     string `json:"message"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	Timezone    string `json:"timezone"`
	CurrentTime string `json:"current_time"`
}

// APIResponse 通用API响应
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// HealthChecker 健康检查器
type HealthChecker struct {
	serviceName string
	version     string
	logger      *zap.Logger
	checks      []func() HealthCheck
	mu          sync.RWMutex
}

// 常量定义
const (
	StatusSuccess      = 200
	StatusError        = 500
	StatusUnauthorized = 401
	StatusForbidden    = 403
	StatusNotFound     = 404
	StatusBadRequest   = 400
)
