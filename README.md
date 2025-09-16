# Go Web 基础框架模板

一个基于 **Gin + GORM(Postgres) + Zap + JWT** 的 Go Web 项目基础框架，支持模块化开发，内置日志、配置、数据库、路由、健康检查等常用功能。  
本仓库已设置为 **Template Repository**，可直接用来创建新项目。

---

## ✨ 特性

- **模块化结构**：支持 `modules` 目录按业务拆分，示例模块 `example` 已提供参考。
- **常用中间件**：CORS、JWT 鉴权、Zap 日志、全局 Recovery。
- **数据库支持**：GORM + PostgreSQL，包含基础配置与连接池设置。
- **数据库迁移**：集成 Atlas + GORM Provider，支持自动生成和版本化管理迁移。
- **健康检查**：`/api/health` 接口支持应用与数据库检查。
- **配置管理**：集中化 `config/app.yaml`，支持应用、日志、JWT、Zipkin 等配置。
- **容器化**：提供 `Dockerfile` 与 `build.sh` 脚本。
- **可扩展性**：方便集成 Consul、Zipkin/OpenTelemetry、Prometheus 等组件。

---

## 📦 使用方式

### 1. 基于模板创建新仓库
1. 打开本仓库首页：[GoWebTemplate](https://github.com/Programmer-Bugs-Bunny/GoWebTemplate)
2. 点击绿色按钮 **Use this template → Create a new repository**
3. 输入新仓库名称，例如 `CardService`，然后创建。

### 2. 克隆新仓库到本地
```bash
git clone git@github.com:你的用户名/CardService.git
cd CardService
```

### 3. 修改 go.mod 模块名
```bash
go mod edit -module github.com/你的用户名/CardService
go mod tidy
```

一定要修改模块名，否则 import 路径会依然指向模板仓库。

### 4. 配置数据库 & 应用参数
编辑 `config/app.yaml`，修改以下参数：

- `database`: 数据库连接信息（host、user、password、dbname、port）
- `jwt.secret`: JWT 签名密钥（务必替换）
- `app.port`: 服务启动端口（默认 8080）
- `app.timezone`: 时区（默认 `Asia/Shanghai`，可改为 `Asia/Ho_Chi_Minh` 等）

### 5. 启动项目
```bash
go run main.go
```

### 6. 验证
- `GET http://localhost:8080/api/ping` → 返回 `"pong"`
- `GET http://localhost:8080/api/health` → 返回应用与数据库状态

---

## 🗂️ 目录结构
```
.
├── config/         # 配置文件（YAML）
├── database/       # 数据库初始化
├── middlewares/    # 中间件 (CORS/JWT/日志/恢复/Tracing)
├── migrations/     # 数据库迁移文件（Atlas 生成）
├── models/         # 数据模型 (GORM)
├── modules/        # 业务模块目录 (示例: example)
├── routes/         # 路由注册
├── utils/          # 工具方法 (健康检查/时间处理等)
├── cmd/            # 命令行工具
│   └── migrate/    # 数据库迁移工具
├── atlas.hcl       # Atlas 配置文件
├── atlas_loader.go # GORM 模型加载器
├── main.go         # 程序入口
├── Dockerfile      # 容器化支持
└── build.sh        # 构建脚本
```

---

## 🔧 开发建议
- 新建模块时在 `modules/` 下新建目录，例如 `modules/card`，包含：
    - `controller.go`
    - `service.go`
    - `types.go`
- 在 `routes/rest` 中注册模块的公开/私有路由。
- 使用 `zap.L().Info/Error` 记录日志。
- 建议配合 `Makefile` 增加常用命令（run/build/test/lint）。

---

## 🧪 可选：快速自检清单
- `go build ./...` 能通过；
- 启动后 `GET /api/ping` 与 `GET /api/health` 工作正常；
- 私有路由需携带 JWT 才可访问；
- 数据库连接池参数按配置生效。

---

## 📊 数据库迁移管理

本项目集成了 **Atlas + GORM Provider** 来实现自动化的数据库迁移管理。

### 安装 Atlas CLI

```bash
# macOS (推荐)
brew install ariga/tap/atlas

# 或使用 go install
go install ariga.io/atlas/cmd/atlas@latest

# 验证安装
atlas version
```

### 安装 GORM Provider 依赖

需要单独执行，不会被`go mod tidy`命令检测到，因为构建排除了`atlas_loader.go`文件

```bash
go get ariga.io/atlas-provider-gorm
```

### 迁移命令

#### 1. 查看迁移状态
```bash
# 检查当前迁移状态
go run cmd/migrate/main.go -action status

# 指定环境
go run cmd/migrate/main.go -action status -env production
```

#### 2. 生成迁移文件
```bash
# 基于当前模型生成迁移（自动命名）
go run cmd/migrate/main.go -action diff

# 生成带自定义名称的迁移
go run cmd/migrate/main.go -action diff -name create_users_table
```

#### 3. 应用迁移
```bash
# 应用所有待执行的迁移
go run cmd/migrate/main.go -action apply

# 模拟执行（显示将要执行的SQL，不实际执行）
go run cmd/migrate/main.go -action apply -dry-run
```

#### 4. 验证迁移
```bash
# 验证迁移文件的有效性
go run cmd/migrate/main.go -action validate
```

#### 5. 重置迁移历史（危险操作）
```bash
# 显示重置指导（不会直接执行）
go run cmd/migrate/main.go -action reset
```

### 配置说明

- **atlas.hcl**: Atlas 主配置文件，定义数据源和环境
- **atlas_loader.go**: GORM 模型加载器，需要在此文件中注册所有数据模型
- **migrations/**: 存储生成的迁移文件

### 环境配置

#### 本地开发环境 (local)
修改 `atlas.hcl` 中的 `env "local"` 配置：
```hcl
env "local" {
  url = "postgres://username:password@localhost:5432/your_database?sslmode=disable" // 开发数据库
  dev = "postgres://username:password@localhost:5432/dev_database?sslmode=disable"  // 计算差异数据库
}
```

#### 生产环境 (production)
```bash
# 设置环境变量
export DATABASE_URL="postgres://user:pass@host:port/db?sslmode=require"

# 应用迁移
go run cmd/migrate/main.go -action apply -env production
```

### 添加新模型

1. 在 `models/` 目录下创建新的模型文件
2. 在 `atlas_loader.go` 中注册新模型：
   ```go
   stmts, err := gormschema.New("postgres").Load(
       &models.User{},
       &models.NewModel{}, // 添加新模型
   )
   ```
3. 生成迁移：`go run cmd/migrate/main.go -action diff -name add_new_model`
4. 应用迁移：`go run cmd/migrate/main.go -action apply`

### 注意事项

⚠️ **重要提醒**：

1. **备份数据库**: 在生产环境应用迁移前务必备份数据库
2. **测试迁移**: 先在开发环境测试迁移的正确性
3. **版本控制**: 迁移文件应纳入版本控制系统
4. **环境隔离**: 不同环境使用不同的数据库连接
5. **回滚策略**: Atlas 支持迁移回滚，但需要谨慎操作

### 常见问题

**Q: 如何回滚迁移？**
```bash
# 回滚到指定版本
atlas migrate down --env local --to-version 20231201120000
```

**Q: 如何重置迁移历史？**
```bash
# 删除迁移历史表（谨慎操作）
atlas migrate reset --env local
```

**Q: 生产环境迁移失败怎么办？**
1. 检查迁移文件语法
2. 确认数据库权限
3. 查看详细错误日志
4. 必要时手动修复数据库状态

---

## 📜 License
[MIT](./LICENSE)
