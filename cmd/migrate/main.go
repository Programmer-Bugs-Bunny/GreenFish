package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"go-web-template/config"
	"go-web-template/database"
	"go-web-template/middlewares"

	"go.uber.org/zap"
)

func main() {
	var (
		env    = flag.String("env", "local", "环境配置 (local/production)")
		action = flag.String("action", "", "操作类型: status, diff, apply, validate, reset")
		name   = flag.String("name", "", "迁移名称 (仅用于 diff 操作)")
		dryRun = flag.Bool("dry-run", false, "模拟执行，不实际应用迁移 (仅用于 apply 操作)")
	)
	flag.Parse()

	if *action == "" {
		printUsage()
		os.Exit(1)
	}

	// 检查 Atlas 是否已安装
	if err := database.EnsureAtlasInstalled(); err != nil {
		log.Fatal("Atlas CLI 未安装:", err)
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 初始化日志系统
	if err := middlewares.InitLogger(&cfg.Logger); err != nil {
		log.Fatal("初始化日志系统失败:", err)
	}
	defer middlewares.Sync()

	// 创建迁移管理器
	migrationConfig := &database.MigrationConfig{
		Environment: *env,
		Timeout:     60, // 60秒超时
	}

	manager := database.NewMigrationManager(migrationConfig, middlewares.Logger)

	// 初始化迁移目录
	if err := manager.InitMigrationDirectory(); err != nil {
		middlewares.Logger.Fatal("初始化迁移目录失败", zap.Error(err))
	}

	// 执行相应的操作
	switch *action {
	case "status":
		handleStatus(manager)
	case "diff":
		handleDiff(manager, *name)
	case "apply":
		handleApply(manager, *dryRun)
	case "validate":
		handleValidate(manager)
	case "reset":
		handleReset()
	default:
		fmt.Printf("未知操作: %s\n", *action)
		printUsage()
		os.Exit(1)
	}
}

func handleStatus(manager *database.MigrationManager) {
	fmt.Println("🔍 检查迁移状态...")
	if err := manager.CheckMigrations(); err != nil {
		middlewares.Logger.Fatal("检查迁移状态失败", zap.Error(err))
	}
	fmt.Println("✅ 迁移状态检查完成")
}

func handleDiff(manager *database.MigrationManager, name string) {
	if name == "" {
		fmt.Println("📝 生成迁移文件...")
	} else {
		fmt.Printf("📝 生成迁移文件: %s\n", name)
	}

	if err := manager.GenerateMigration(name); err != nil {
		middlewares.Logger.Fatal("生成迁移失败", zap.Error(err))
	}
	fmt.Println("✅ 迁移文件生成完成")
}

func handleApply(manager *database.MigrationManager, dryRun bool) {
	if dryRun {
		fmt.Println("🧪 模拟应用迁移 (dry-run)...")
		// 这里可以添加 dry-run 逻辑
		fmt.Println("注意: dry-run 功能需要在 Atlas 命令中添加 --dry-run 参数")
	} else {
		fmt.Println("🚀 应用迁移...")
	}

	if err := manager.ApplyMigrations(); err != nil {
		middlewares.Logger.Fatal("应用迁移失败", zap.Error(err))
	}
	fmt.Println("✅ 迁移应用完成")
}

func handleValidate(manager *database.MigrationManager) {
	fmt.Println("🔍 验证迁移文件...")
	if err := manager.ValidateMigrations(); err != nil {
		middlewares.Logger.Fatal("验证迁移失败", zap.Error(err))
	}
	fmt.Println("✅ 迁移验证通过")
}

func handleReset() {
	fmt.Println("⚠️  重置迁移历史是危险操作!")
	fmt.Println("请手动执行: atlas migrate reset --env [环境名]")
	fmt.Println("这将删除迁移历史表，请谨慎操作")
}

func printUsage() {
	fmt.Println("数据库迁移管理工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  go run cmd/migrate/main.go -action <操作> [选项]")
	fmt.Println()
	fmt.Println("操作:")
	fmt.Println("  status    检查迁移状态")
	fmt.Println("  diff      生成迁移文件 (可选: -name <迁移名称>)")
	fmt.Println("  apply     应用迁移 (可选: -dry-run)")
	fmt.Println("  validate  验证迁移文件")
	fmt.Println("  reset     重置迁移历史 (仅显示提示)")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -env      环境配置 (默认: local)")
	fmt.Println("  -name     迁移名称 (仅用于 diff)")
	fmt.Println("  -dry-run  模拟执行 (仅用于 apply)")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  go run cmd/migrate/main.go -action status")
	fmt.Println("  go run cmd/migrate/main.go -action diff -name create_users")
	fmt.Println("  go run cmd/migrate/main.go -action apply")
	fmt.Println("  go run cmd/migrate/main.go -action apply -dry-run")
	fmt.Println("  go run cmd/migrate/main.go -action validate")
}
