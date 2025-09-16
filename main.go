package main

import (
	"context"
	"fmt"
	"go-web-template/config"
	"go-web-template/database"
	"go-web-template/middlewares"
	"go-web-template/routes"
	_ "go-web-template/routes/rest" // 导入触发 init() 自动注册路由
	"go-web-template/utils"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/openzipkin/zipkin-go"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 初始化日志系统
	if err := middlewares.InitLogger(&cfg.Logger); err != nil {
		log.Fatal("初始化日志系统失败:", err)
	}

	// 初始化JWT配置
	middlewares.InitJWT(cfg.JWT.Secret, cfg.JWT.ExpireHours, cfg.JWT.Issuer)

	// 确保在程序退出时同步日志缓冲区
	defer middlewares.Sync()

	// 使用结构化日志记录启动信息
	zap.L().Info("应用程序启动",
		zap.String("version", cfg.App.Version),
		zap.String("environment", cfg.App.Environment),
		zap.Bool("debug", cfg.App.Debug),
	)

	// 初始化时区
	if err := utils.InitTimezone(cfg.App.Timezone); err != nil {
		zap.L().Warn("时区初始化警告", zap.Error(err))
	} else {
		zap.L().Info("时区设置成功", zap.String("timezone", cfg.App.Timezone))
	}

	// 初始化 Zipkin
	var zipkinTracer *zipkin.Tracer
	if cfg.Zipkin.Enabled {
		tracer, err := utils.InitZipkin(&cfg.Zipkin, zap.L())
		if err != nil {
			zap.L().Fatal("Zipkin初始化失败", zap.Error(err))
		}
		zipkinTracer = tracer
		// 设置全局 Zipkin tracer
		middlewares.SetZipkinTracer(tracer)
		zap.L().Info("Zipkin初始化成功",
			zap.String("service_name", cfg.Zipkin.ServiceName),
			zap.String("endpoint", cfg.Zipkin.Endpoint),
		)

		// 测试Zipkin连接状态
		healthChecker := utils.NewZipkinHealthChecker(cfg.Zipkin.Endpoint, zap.L())
		if healthChecker.CheckConnection() {
			zap.L().Info("Zipkin服务连接测试成功")
		} else {
			zap.L().Warn("Zipkin服务连接测试失败")
		}

		// 监控tracer状态
		utils.MonitorZipkinReporter(tracer, zap.L())
	}

	// 初始化数据库（带 Zipkin tracer 支持）
	if err := database.InitWithTracer(&cfg.Database, zap.L(), zipkinTracer); err != nil {
		zap.L().Fatal("数据库初始化失败", zap.Error(err))
	}

	// 设置路由
	r := routes.SetupRoutes(cfg)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    cfg.GetAddr(),
		Handler: r,
	}

	// 启动服务器
	go func() {
		zap.L().Info("HTTP 服务器启动",
			zap.String("address", cfg.GetAddr()),
			zap.String("version", cfg.App.Version),
			zap.String("environment", cfg.App.Environment),
			zap.Bool("debug", cfg.App.Debug),
		)

		fmt.Printf("服务器启动在: %s (版本: %s, 环境: %s)\n",
			cfg.GetAddr(), cfg.App.Version, cfg.App.Environment)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("启动服务器失败", zap.Error(err))
		}
	}()

	// 注册优雅关闭
	setupGracefulShutdown(srv, zipkinTracer)

	// 等待关闭信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zap.L().Info("正在关闭服务器...")
	gracefulShutdown(srv, zipkinTracer)
}

// setupGracefulShutdown 设置优雅关闭
func setupGracefulShutdown(srv *http.Server, tracer *zipkin.Tracer) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		zap.L().Info("接收到关闭信号，开始优雅关闭...")
		gracefulShutdown(srv, tracer)
		os.Exit(0)
	}()
}

// gracefulShutdown 执行优雅关闭
func gracefulShutdown(srv *http.Server, tracer *zipkin.Tracer) {
	// 设置关闭超时
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1. 关闭HTTP服务器
	zap.L().Info("正在关闭HTTP服务器...")
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Error("HTTP服务器关闭失败", zap.Error(err))
		// 强制关闭
		if err := srv.Close(); err != nil {
			zap.L().Error("强制关闭HTTP服务器失败", zap.Error(err))
		}
	} else {
		zap.L().Info("HTTP服务器关闭成功")
	}

	// 2. 关闭数据库连接
	if err := database.Close(); err != nil {
		zap.L().Error("关闭数据库连接失败", zap.Error(err))
	} else {
		zap.L().Info("数据库连接已关闭")
	}

	// 关闭 Zipkin tracer
	if tracer != nil {
		zap.L().Info("Zipkin tracer资源清理完成")
	}

	// 3. 同步日志缓冲区
	middlewares.Sync()

	zap.L().Info("应用程序已优雅关闭")
}
