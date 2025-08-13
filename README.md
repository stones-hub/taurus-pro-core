# Taurus Pro 脚手架项目

## 项目简介

Taurus Pro 是一个强大的 Go 微服务项目脚手架工具，它可以帮助您快速创建具有完整架构的 Go 微服务项目。该工具基于组件化设计，支持多种可选组件的灵活配置，自动生成项目结构、依赖注入代码和配置文件。

## 脚手架工具使用指南

### 1. 安装脚手架工具

```bash
# 克隆项目
git clone <repository-url>
cd taurus-pro-core

# 构建脚手架工具
go build -o taurus cmd/taurus/main.go

# 将工具添加到 PATH 或直接使用
./taurus create my-project
```

### 2. 创建新项目

```bash
# 基本用法
taurus create <项目名称>

# 示例
taurus create my-microservice
```

创建过程中，脚手架会交互式地询问：
- 项目路径（默认为当前目录下的项目名）
- 要包含的可选组件

### 3. 组件系统

#### 必需组件（自动包含）
- **wire** - 依赖注入工具
- **config** - 配置管理
- **http** - HTTP 服务
- **common** - 通用组件

#### 可选组件（用户选择）
- **grpc** - gRPC 服务
- **storage** - 数据库和 Redis 存储
- **tcp** - TCP 服务
- **otel** - OpenTelemetry 监控
- **consul** - 服务发现

## 项目模板目录结构详解

### 核心应用结构 (`templates/app/`)

#### 1. **bootstrap.gotmpl** - 应用启动引导
- 应用的主入口点
- 支持 HTTP 服务和脚本命令两种运行模式
- 集成 pprof 性能分析
- 优雅关闭和信号处理
- 全局 panic 恢复机制

#### 2. **command/** - 命令行工具
- `command.gotmpl` - 基础命令框架
- `example_cmd.gotmpl` - 示例命令实现
- 支持脚本模式运行，可通过 `--script` 参数启用

#### 3. **controller/** - HTTP 控制器层
- `index_controller.gotmpl` - 首页控制器
- `user_controller.gotmpl` - 用户管理控制器
- `pprof/memory_controller.gotmpl` - 内存分析控制器

#### 4. **service/** - 业务逻辑层
- `index_service.gotmpl` - 首页服务
- `user_service.gotmpl` - 用户管理服务

#### 5. **model/** - 数据模型层
- `user_model.gotmpl` - 用户数据模型

#### 6. **process/** - 队列处理
- `process.gotmpl` - 队列处理管理

#### 7. **crontab/** - 定时任务
- `crontab.gotmpl` - 定时任务框架
- `example_task.gotmpl` - 示例定时任务

#### 8. **hooks/** - 生命周期钩子
- `hook.gotmpl` - 钩子基础框架
- `example_hook.gotmpl` - 示例钩子实现

#### 9. **constants/** - 常量定义
- 应用级别的常量定义

#### 10. **helper/** - 辅助工具
- 通用辅助函数和工具

### 配置管理 (`templates/config/`)

#### 1. **config.yaml** - 主配置文件
```yaml
version: "${VERSION:v1.0.0}"
app_name: "${APP_NAME:taurus}"
pprof_enabled: true
go:
  max_procs: 8        # 最大CPU核心数
  gc: 150             # 垃圾回收比例
  memory_limit: 12    # 内存限制(GB)
```

#### 2. **autoload/** - 自动加载配置
- **http/http.yaml** - HTTP 服务配置
  - 地址、端口、超时设置
  - 授权码配置
  
- **db/db.yaml** - 数据库配置
  - 支持多数据库实例
  - 连接池配置
  - 重试机制
  - 日志配置
  
- **redis/redis.toml** - Redis 配置
  - 连接池设置
  - 超时配置
  - 日志配置
  
- **gRPC/server.yaml** - gRPC 服务配置
- **tcp/tcp.yaml** - TCP 服务配置
- **otel/otel.yaml** - OpenTelemetry 配置
- **consul/consul.yaml** - Consul 配置
- **cron/cron.yaml** - 定时任务配置
- **logger/logger.yaml** - 日志配置
- **websocket/ws.yaml** - WebSocket 配置
- **templates/templates.yaml** - 模板配置
- **mcp/mcp.yaml** - MCP 配置

### 中间件和工具 (`templates/pkg/`)

#### 1. **middleware/** - HTTP 中间件
- `auth_middleware.gotmpl` - 认证中间件
- `host_middleware.gotmpl` - 主机中间件

### 部署和运维 (`templates/`)

#### 1. **Dockerfile** - 容器化配置
- 多阶段构建
- 支持环境变量配置
- 时区设置
- 最小化运行时镜像

#### 2. **docker-compose.yml** - 本地开发环境
- 应用服务
- MySQL 数据库
- Redis 缓存
- 健康检查
- 数据卷挂载
- 网络配置

#### 3. **docker-compose-swarm.yml** - 生产环境部署
- 支持 Docker Swarm 集群部署

#### 4. **Makefile** - 构建和部署脚本
- 环境变量检查
- 代码生成（wire）
- 构建目标
- Docker 操作
- 本地运行
- 发布管理

### 测试和性能 (`templates/test/`)

#### 1. **performance/** - 性能测试
- **generator/** - 负载生成器
- **memory/** - 内存测试
- **profile/** - 性能分析
  - 内存泄漏测试
  - pprof 测试

#### 2. **test_user_api/** - API 测试
- 用户 API 测试脚本

### 基准测试 (`templates/benchmark/`)

- HTTP、gRPC、WebSocket 性能测试
- 配置文件和运行脚本
- 测试报告生成

### 静态资源和模板 (`templates/static/`, `templates/templates/`)

- CSS 样式文件
- JavaScript 脚本
- HTML 模板
- 图片资源

## 生成项目的使用指南

### 1. 项目结构概览

使用脚手架创建项目后，您将得到一个完整的 Go 微服务项目，包含以下主要部分：

```
my-project/
├── app/                    # 应用核心代码
│   ├── bootstrap.go       # 应用启动引导
│   ├── command/           # 命令行工具
│   ├── controller/        # HTTP 控制器
│   ├── service/           # 业务逻辑层
│   ├── model/             # 数据模型
│   ├── process/           # 队列处理
│   ├── crontab/           # 定时任务
│   ├── hooks/             # 生命周期钩子
│   ├── constants/         # 常量定义
│   └── helper/            # 辅助工具
├── internal/               # 内部包
│   └── taurus/            # 核心组件
│       └── wire.go        # 依赖注入配置
├── config/                 # 配置文件
│   ├── config.yaml        # 主配置
│   └── autoload/          # 自动加载配置
├── pkg/                    # 公共包
│   └── middleware/        # 中间件
├── static/                 # 静态资源
├── templates/              # HTML 模板
├── logs/                   # 日志目录
├── downloads/              # 下载目录
├── scripts/                # 脚本文件
├── test/                   # 测试文件
├── benchmark/              # 性能测试
├── Dockerfile              # 容器配置
├── docker-compose.yml      # 本地开发环境
├── docker-compose-swarm.yml # 生产环境部署
├── Makefile                # 构建脚本
├── go.mod                  # Go 模块文件
└── go.sum                  # 依赖校验文件
```

### 2. 环境配置

#### 创建环境变量文件

```bash
# 复制示例环境文件
cp .env.example .env.local

# 编辑环境变量
vim .env.local
```

#### 主要环境变量配置

```bash
# 应用配置
VERSION=v1.0.0
APP_NAME=taurus
APP_CONFIG=./config

# 服务配置
SERVER_ADDRESS=0.0.0.0
SERVER_PORT=8080
HOST_PORT=8080
CONTAINER_PORT=8080

# 数据库配置
DB_DSN=user:password@tcp(host:port)/database
DB_NAME=database_name
DB_USER=username
DB_PASSWORD=password

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# 工作目录
WORKDIR=/app
```

### 3. 开发和运行

#### 本地开发运行

```bash
# 1. 安装依赖
go mod tidy

# 2. 生成依赖注入代码
make wire

# 3. 运行项目
make run

# 或者直接运行
go run ./bin/taurus.go
```

#### 脚本命令模式

```bash
# 启用脚本模式
go run ./bin/taurus.go --script

# 查看可用命令
go run ./bin/taurus.go --script --help
```

#### 性能分析

项目集成了 pprof 性能分析工具：

```bash
# 访问性能分析页面
http://localhost:8080/debug/pprof/

# 内存分析
http://localhost:8080/debug/pprof/heap

# CPU 分析
http://localhost:8080/debug/pprof/profile
```

### 4. 数据库操作

#### 启动数据库服务

```bash
# 使用 Docker Compose 启动数据库
docker-compose up mysql redis -d

# 或者单独启动
docker run -d --name mysql -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=kf_ai -p 3306:3306 mysql:8
docker run -d --name redis -p 6379:6379 redis:6
```

#### 数据库初始化

```bash
# 执行初始化 SQL 脚本
mysql -u root -p < scripts/data/init_mysql/kf_ai.sql
```

### 5. 构建和部署

#### 本地构建

```bash
# 构建项目
make build

# 构建二进制文件
go build -o taurus ./bin/taurus.go
```

#### Docker 部署

```bash
# 构建 Docker 镜像
make docker-build

# 本地运行 Docker 容器
make docker-run

# 停止容器
make docker-stop

# 使用 Docker Compose
make docker-compose-up
```

#### 生产环境部署

```bash
# Docker Swarm 部署
make docker-swarm-up

# 更新应用
make docker-swarm-deploy-app
```

### 6. 测试和性能

#### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定测试
go test ./app/controller

# 运行性能测试
go test -bench=. ./test/performance/
```

#### 性能基准测试

```bash
# 运行 HTTP 性能测试
cd benchmark
node http.js

# 运行 gRPC 性能测试
node grpc.js

# 运行 WebSocket 性能测试
node websocket.js
```

#### 内存泄漏测试

```bash
# 运行内存泄漏测试
go test -v ./test/performance/profile/
```

### 7. 日志和监控

#### 日志配置

项目支持结构化日志，日志文件位于 `logs/` 目录：

```bash
# 查看应用日志
tail -f logs/app.log

# 查看数据库日志
tail -f logs/db/db.log

# 查看 Redis 日志
tail -f logs/redis/redis.log
```

#### 监控和追踪

如果启用了 OpenTelemetry 组件：

```bash
# 查看指标
http://localhost:8080/metrics

# 查看追踪信息
http://localhost:8080/traces
```

### 8. 常用 Makefile 命令

```bash
# 代码生成
make wire          # 生成依赖注入代码

# 构建和运行
make build         # 构建项目
make run           # 运行项目
make clean         # 清理构建文件

# Docker 操作
make docker-build  # 构建 Docker 镜像
make docker-run    # 运行 Docker 容器
make docker-stop   # 停止容器

# 部署
make docker-compose-up      # 启动开发环境
make docker-compose-down    # 停止开发环境
make docker-swarm-up        # 启动生产环境
make docker-swarm-down      # 停止生产环境

# 发布
make local-release          # 本地发布
make local-release-start    # 启动发布版本
make local-release-stop     # 停止发布版本
```

### 9. 开发最佳实践

#### 代码组织

1. **控制器层** (`app/controller/`): 处理 HTTP 请求和响应
2. **服务层** (`app/service/`): 实现业务逻辑
3. **模型层** (`app/model/`): 定义数据结构和数据库操作
4. **中间件** (`pkg/middleware/`): 处理跨切面关注点

#### 依赖注入

项目使用 Google Wire 进行依赖注入：

```bash
# 添加新的 Provider 后，需要重新生成
make wire

# 或者手动执行
wire ./internal/taurus/wire.go
wire ./app/wire.go
```

#### 配置管理

1. 使用环境变量覆盖默认配置
2. 敏感信息通过环境变量传递
3. 配置文件支持模板语法 `${VARIABLE:default_value}`

#### 错误处理

1. 使用统一的错误处理机制
2. 记录详细的错误日志
3. 实现优雅的错误恢复

### 10. 故障排除

#### 常见问题

1. **依赖注入失败**
   ```bash
   # 重新生成 wire 代码
   make wire
   ```

2. **配置文件找不到**
   ```bash
   # 检查环境变量 APP_CONFIG 设置
   echo $APP_CONFIG
   ```

3. **数据库连接失败**
   ```bash
   # 检查数据库服务状态
   docker ps | grep mysql
   docker ps | grep redis
   ```

4. **端口被占用**
   ```bash
   # 检查端口使用情况
   lsof -i :8080
   ```

#### 调试技巧

1. **启用详细日志**
   ```bash
   # 设置日志级别
   export LOG_LEVEL=debug
   ```

2. **使用 pprof 分析性能**
   ```bash
   # 生成 CPU 分析文件
   curl -o cpu.prof http://localhost:8080/debug/pprof/profile
   
   # 分析 CPU 性能
   go tool pprof cpu.prof
   ```

3. **内存分析**
   ```bash
   # 生成内存分析文件
   curl -o mem.prof http://localhost:8080/debug/pprof/heap
   
   # 分析内存使用
   go tool pprof mem.prof
   ```

## 总结

Taurus Pro 脚手架工具为您提供了一个完整的 Go 微服务开发框架，包含：

- **完整的项目结构**: 遵循 Go 项目最佳实践
- **组件化架构**: 支持灵活的功能组件选择
- **自动代码生成**: 自动生成依赖注入和配置代码
- **完整的部署支持**: Docker、Docker Compose、Docker Swarm
- **性能监控**: 集成 pprof 和 OpenTelemetry
- **测试框架**: 单元测试、性能测试、基准测试
- **开发工具**: Makefile、脚本、配置管理

通过这个脚手架，您可以快速启动 Go 微服务开发，专注于业务逻辑实现，而不是基础设施搭建。
