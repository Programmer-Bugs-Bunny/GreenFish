package utils

import (
	"fmt"
	"net/http"
	"time"

	"go-web-template/config"

	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/reporter"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"go.uber.org/zap"
)

// InitZipkin 初始化Zipkin追踪
func InitZipkin(cfg *config.ZipkinConfig, logger *zap.Logger) (*zipkin.Tracer, error) {
	// 创建HTTP reporter
	httpReporter := zipkinhttp.NewReporter(cfg.Endpoint)

	// 创建tracer
	tracer, err := zipkin.NewTracer(
		httpReporter,
		zipkin.WithLocalEndpoint(&model.Endpoint{
			ServiceName: cfg.ServiceName,
		}),
		zipkin.WithSampler(zipkin.NewModuloSampler(uint64(1.0/cfg.SampleRate))),
	)

	if err != nil {
		httpReporter.Close()
		return nil, fmt.Errorf("创建Zipkin tracer失败: %w", err)
	}

	logger.Info("Zipkin tracer创建成功",
		zap.String("service_name", cfg.ServiceName),
		zap.String("endpoint", cfg.Endpoint),
		zap.Float64("sample_rate", cfg.SampleRate),
	)

	return tracer, nil
}

// ZipkinHealthChecker Zipkin健康检查器
type ZipkinHealthChecker struct {
	endpoint string
	logger   *zap.Logger
}

// NewZipkinHealthChecker 创建Zipkin健康检查器
func NewZipkinHealthChecker(endpoint string, logger *zap.Logger) *ZipkinHealthChecker {
	return &ZipkinHealthChecker{
		endpoint: endpoint,
		logger:   logger,
	}
}

// CheckConnection 检查Zipkin连接状态
func (z *ZipkinHealthChecker) CheckConnection() bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 尝试访问Zipkin健康检查端点
	healthEndpoint := fmt.Sprintf("%s/../health", z.endpoint)
	resp, err := client.Get(healthEndpoint)
	if err != nil {
		z.logger.Debug("Zipkin健康检查失败", zap.Error(err))
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// MonitorZipkinReporter 监控Zipkin reporter状态
func MonitorZipkinReporter(tracer *zipkin.Tracer, logger *zap.Logger) {
	// 这是一个简单的监控示例
	// 在实际项目中，你可能需要更复杂的监控逻辑
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			logger.Debug("Zipkin tracer运行状态检查")
			// 这里可以添加更多的监控逻辑
		}
	}()
}

// CloseReporter 关闭reporter（优雅关闭时使用）
func CloseReporter(rep reporter.Reporter) error {
	if rep != nil {
		return rep.Close()
	}
	return nil
}
