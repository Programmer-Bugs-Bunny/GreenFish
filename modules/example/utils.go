package example

import (
	"time"

	"go.uber.org/zap"
)

// newHealthChecker 创建新的健康检查器 (私有函数)
func newHealthChecker(serviceName, version string, logger *zap.Logger) *HealthChecker {
	return &HealthChecker{
		serviceName: serviceName,
		version:     version,
		logger:      logger,
		checks:      make([]func() HealthCheck, 0),
	}
}

// executeCheck 执行单个检查并捕获panic (私有方法)
func (hc *HealthChecker) executeCheck(checkFunc func() HealthCheck) HealthCheck {
	defer func() {
		if r := recover(); r != nil {
			hc.logger.Error("健康检查执行时发生panic",
				zap.Any("panic", r),
			)
		}
	}()

	startTime := time.Now()
	check := checkFunc()
	check.Duration = time.Since(startTime).String()

	return check
}

// databaseHealthCheck 数据库健康检查示例 (私有函数)
func databaseHealthCheck() HealthCheck {
	// 这里应该实际检查数据库连接
	// 为了模板简单，这里直接返回健康状态
	return HealthCheck{
		Name:    "database",
		Status:  HealthStatusHealthy,
		Message: "数据库连接正常",
	}
}

// externalServiceHealthCheck 外部服务健康检查示例 (私有函数)
func externalServiceHealthCheck(serviceName, endpoint string) func() HealthCheck {
	return func() HealthCheck {
		// 这里应该实际检查外部服务
		// 为了模板简单，这里直接返回健康状态
		return HealthCheck{
			Name:    serviceName,
			Status:  HealthStatusHealthy,
			Message: "外部服务连接正常",
		}
	}
}

// AddCheck 添加健康检查项 (HealthChecker方法)
func (hc *HealthChecker) AddCheck(checkFunc func() HealthCheck) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checks = append(hc.checks, checkFunc)
}

// Check 执行健康检查 (HealthChecker方法)
func (hc *HealthChecker) Check() HealthResponse {
	startTime := time.Now()

	hc.mu.RLock()
	defer hc.mu.RUnlock()

	response := HealthResponse{
		Status:    HealthStatusHealthy,
		Version:   hc.version,
		Timestamp: time.Now(),
		Checks:    make([]HealthCheck, 0, len(hc.checks)),
	}

	// 执行所有检查
	for _, checkFunc := range hc.checks {
		check := hc.executeCheck(checkFunc)
		response.Checks = append(response.Checks, check)

		// 根据检查结果更新整体状态
		if check.Status == HealthStatusUnhealthy {
			response.Status = HealthStatusUnhealthy
		} else if check.Status == HealthStatusDegraded && response.Status == HealthStatusHealthy {
			response.Status = HealthStatusDegraded
		}
	}

	duration := time.Since(startTime)
	hc.logger.Debug("健康检查执行完成",
		zap.String("status", string(response.Status)),
		zap.Duration("duration", duration),
		zap.Int("checks_count", len(response.Checks)),
	)

	return response
}

// getCurrentTime 获取当前时间 (私有函数)
func getCurrentTime() time.Time {
	return time.Now()
}

// getCurrentTimeString 获取当前时间字符串 (私有函数)
func getCurrentTimeString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
