package database

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"go.uber.org/zap"
)

// MigrationConfig 迁移配置
type MigrationConfig struct {
	Environment string // local, production
	Timeout     int    // 超时时间（秒）
}

// MigrationManager 迁移管理器
type MigrationManager struct {
	config *MigrationConfig
	logger *zap.Logger
}

// NewMigrationManager 创建迁移管理器
func NewMigrationManager(config *MigrationConfig, logger *zap.Logger) *MigrationManager {
	return &MigrationManager{
		config: config,
		logger: logger,
	}
}

// CheckMigrations 检查是否有未应用的迁移
func (m *MigrationManager) CheckMigrations() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.config.Timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "atlas", "migrate", "status", "--env", m.config.Environment)
	output, err := cmd.CombinedOutput()

	if err != nil {
		m.logger.Error("检查迁移状态失败",
			zap.Error(err),
			zap.String("output", string(output)))
		return fmt.Errorf("检查迁移状态失败: %w", err)
	}

	m.logger.Info("迁移状态检查完成", zap.String("output", string(output)))
	return nil
}

// GenerateMigration 生成迁移文件
func (m *MigrationManager) GenerateMigration(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.config.Timeout)*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if name != "" {
		cmd = exec.CommandContext(ctx, "atlas", "migrate", "diff", name, "--env", m.config.Environment)
	} else {
		cmd = exec.CommandContext(ctx, "atlas", "migrate", "diff", "--env", m.config.Environment)
	}

	output, err := cmd.CombinedOutput()

	if err != nil {
		m.logger.Error("生成迁移失败",
			zap.Error(err),
			zap.String("output", string(output)))
		return fmt.Errorf("生成迁移失败: %w", err)
	}

	m.logger.Info("迁移文件生成成功",
		zap.String("name", name),
		zap.String("output", string(output)))
	return nil
}

// ApplyMigrations 应用迁移
func (m *MigrationManager) ApplyMigrations() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.config.Timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "atlas", "migrate", "apply", "--env", m.config.Environment)
	output, err := cmd.CombinedOutput()

	if err != nil {
		m.logger.Error("应用迁移失败",
			zap.Error(err),
			zap.String("output", string(output)))
		return fmt.Errorf("应用迁移失败: %w", err)
	}

	m.logger.Info("迁移应用成功", zap.String("output", string(output)))
	return nil
}

// ValidateMigrations 验证迁移
func (m *MigrationManager) ValidateMigrations() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.config.Timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "atlas", "migrate", "validate", "--env", m.config.Environment)
	output, err := cmd.CombinedOutput()

	if err != nil {
		m.logger.Error("验证迁移失败",
			zap.Error(err),
			zap.String("output", string(output)))
		return fmt.Errorf("验证迁移失败: %w", err)
	}

	m.logger.Info("迁移验证成功", zap.String("output", string(output)))
	return nil
}

// EnsureAtlasInstalled 确保Atlas已安装
func EnsureAtlasInstalled() error {
	cmd := exec.Command("atlas", "version")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Atlas CLI 未安装或不在 PATH 中，请安装: %w", err)
	}
	return nil
}

// InitMigrationDirectory 初始化迁移目录
func (m *MigrationManager) InitMigrationDirectory() error {
	migrationDir := "migrations"

	// 检查目录是否存在
	if _, err := os.Stat(migrationDir); os.IsNotExist(err) {
		if err := os.MkdirAll(migrationDir, 0755); err != nil {
			m.logger.Error("创建迁移目录失败", zap.Error(err))
			return fmt.Errorf("创建迁移目录失败: %w", err)
		}
		m.logger.Info("迁移目录创建成功", zap.String("dir", migrationDir))
	}

	return nil
}
