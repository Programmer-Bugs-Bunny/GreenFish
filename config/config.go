package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// AppConfig 应用配置结构
type AppConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Version     string `yaml:"version"`
	Debug       bool   `yaml:"debug"`
	Timezone    string `yaml:"timezone"`
	Environment string `yaml:"environment"`
}

// LoggerConfig 日志配置结构
type LoggerConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"max_size"` // MB
	MaxAge     int    `yaml:"max_age"`  // days
	MaxBackups int    `yaml:"max_backups"`
	Compress   bool   `yaml:"compress"`
}

// DatabaseConfig 数据库配置结构
type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	DBName          string `yaml:"dbname"`
	SSLMode         string `yaml:"sslmode"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"` // minutes
}

// JWTConfig JWT配置结构
type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
	Issuer      string `yaml:"issuer"`
}

// ConsulConfig Consul配置结构
type ConsulConfig struct {
	Enabled     bool   `yaml:"enabled" json:"enabled"`           // 是否启用Consul
	Address     string `yaml:"address" json:"address"`           // Consul地址
	ServiceName string `yaml:"service_name" json:"service_name"` // 服务名称
}

// ZipkinConfig Zipkin配置结构
type ZipkinConfig struct {
	Enabled     bool    `yaml:"enabled"`
	ServiceName string  `yaml:"service_name"`
	Endpoint    string  `yaml:"endpoint"`
	SampleRate  float64 `yaml:"sample_rate"`
}

// Config 总配置结构
type Config struct {
	App      AppConfig      `yaml:"app"`
	Logger   LoggerConfig   `yaml:"logger"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Consul   ConsulConfig   `yaml:"consul"`
	Zipkin   ZipkinConfig   `yaml:"zipkin"`
}

// Load 加载配置文件
func Load() (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile("config/app.yaml")
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析 YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置默认值
	config.setDefaults()

	return &config, nil
}

// setDefaults 设置默认值
func (c *Config) setDefaults() {
	// App 默认值
	if c.App.Timezone == "" {
		c.App.Timezone = "UTC"
	}
	if c.App.Environment == "" {
		c.App.Environment = "development"
	}

	// Logger 默认值
	if c.Logger.Level == "" {
		c.Logger.Level = "info"
	}
	if c.Logger.Format == "" {
		c.Logger.Format = "console"
	}
	if c.Logger.Output == "" {
		c.Logger.Output = "stdout"
	}
	if c.Logger.MaxSize == 0 {
		c.Logger.MaxSize = 100
	}
	if c.Logger.MaxAge == 0 {
		c.Logger.MaxAge = 30
	}
	if c.Logger.MaxBackups == 0 {
		c.Logger.MaxBackups = 3
	}

	// Database 默认值
	if c.Database.SSLMode == "" {
		c.Database.SSLMode = "disable"
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = 10
	}
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = 100
	}
	if c.Database.ConnMaxLifetime == 0 {
		c.Database.ConnMaxLifetime = 60
	}

	// JWT 默认值
	if c.JWT.Secret == "" {
		c.JWT.Secret = "change-this-secret-key-in-production"
	}
	if c.JWT.ExpireHours == 0 {
		c.JWT.ExpireHours = 24
	}
	if c.JWT.Issuer == "" {
		c.JWT.Issuer = "go-web-template"
	}

	// Consul 默认值
	if c.Consul.ServiceName == "" {
		c.Consul.ServiceName = "go-web-template"
	}
	if c.Consul.Address == "" {
		c.Consul.Address = "http://localhost:8500"
	}

	// Zipkin 默认值
	if c.Zipkin.ServiceName == "" {
		c.Zipkin.ServiceName = "go-web-template"
	}
	if c.Zipkin.Endpoint == "" {
		c.Zipkin.Endpoint = "http://localhost:9411/api/v2/spans"
	}
	if c.Zipkin.SampleRate == 0 {
		c.Zipkin.SampleRate = 1.0
	}
}

// GetAddr 获取完整的监听地址
func (c *Config) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.App.Host, c.App.Port)
}

// GetDSN 获取数据库连接字符串
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.Username,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}
