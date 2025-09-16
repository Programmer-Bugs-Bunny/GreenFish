package middlewares

import (
	"os"
	"strings"

	"go-web-template/config"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

// InitLogger 初始化日志系统
func InitLogger(cfg *config.LoggerConfig) error {
	// 配置编码器
	var encoderConfig zapcore.EncoderConfig
	if cfg.Format == "json" {
		encoderConfig = zap.NewProductionEncoderConfig()
	} else {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// 选择编码器
	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 配置日志级别
	level := zap.InfoLevel
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	}

	// 配置输出
	var writeSyncer zapcore.WriteSyncer
	if cfg.Output == "file" {
		lumberjackLogger := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxBackups,
			Compress:   cfg.Compress,
		}
		writeSyncer = zapcore.AddSync(lumberjackLogger)
	} else {
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// 创建core
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 创建logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	// 设置为全局logger，这样就可以使用 zap.L() 访问
	zap.ReplaceGlobals(Logger)

	return nil
}

// GinLogger 返回gin的日志中间件
func GinLogger() gin.HandlerFunc {
	return gin.LoggerWithWriter(os.Stdout)
}

// GinRecovery 返回gin的恢复中间件
func GinRecovery() gin.HandlerFunc {
	return gin.RecoveryWithWriter(os.Stdout)
}

// ErrorLogging 错误响应日志中间件
func ErrorLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 记录错误状态码的响应
		if c.Writer.Status() >= 400 {
			Logger.Warn("HTTP错误响应",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
				zap.Int("status_code", c.Writer.Status()),
				zap.String("user_agent", c.GetHeader("User-Agent")),
			)
		}
	}
}

// Sync 同步日志缓冲区
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}
