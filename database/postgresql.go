package database

import (
	"context"
	"fmt"
	"time"

	"go-web-template/config"

	"github.com/openzipkin/zipkin-go"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init 初始化数据库连接
func Init(cfg *config.DatabaseConfig, log *zap.Logger) error {
	return InitWithTracer(cfg, log, nil)
}

// InitWithTracer 初始化带追踪的数据库连接
func InitWithTracer(cfg *config.DatabaseConfig, log *zap.Logger, tracer *zipkin.Tracer) error {
	// 构建DSN
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode)

	// 配置GORM日志
	gormLogger := logger.New(
		&GormZapWriter{Logger: log},
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      false,
		},
	)

	// 打开数据库连接
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 获取底层的sql.DB以配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取底层数据库连接失败: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 如果有tracer，添加追踪插件
	if tracer != nil {
		if err := db.Use(&ZipkinPlugin{tracer: tracer}); err != nil {
			log.Warn("添加Zipkin追踪插件失败", zap.Error(err))
		} else {
			log.Info("Zipkin数据库追踪插件已启用")
		}
	}

	DB = db

	log.Info("数据库连接初始化成功",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.DBName),
		zap.Int("max_idle_conns", cfg.MaxIdleConns),
		zap.Int("max_open_conns", cfg.MaxOpenConns),
	)

	return nil
}

// Close 关闭数据库连接
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// GormZapWriter GORM的Zap日志写入器
type GormZapWriter struct {
	Logger *zap.Logger
}

func (g *GormZapWriter) Printf(format string, args ...interface{}) {
	g.Logger.Info(fmt.Sprintf(format, args...))
}
