# Taurus Pro 企业级Go框架

[![Go Version](https://img.shields.io/badge/Go-1.24.2+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://github.com/stones-hub/taurus-pro)

> 🚀 **Taurus Pro** 是一个基于Go语言构建的企业级Web应用框架，集成了现代化的架构设计、完整的开发工具链和丰富的企业级特性。框架采用分层架构、依赖注入、中间件机制等最佳实践，为开发者提供高效、可维护、可扩展的应用开发体验。

## 📋 目录

- [✨ 核心特性](#-核心特性)
- [🏗️ 架构设计](#️-架构设计)
- [🚀 快速开始](#-快速开始)
- [📚 详细文档](#-详细文档)
- [🔧 配置说明](#-配置说明)
- [📦 项目结构](#-项目结构)
- [🛠️ 开发指南](#️-开发指南)
- [🧪 测试与性能](#-测试与性能)
- [🐳 部署方案](#-部署方案)
- [📈 性能监控](#-性能监控)
- [🤝 贡献指南](#-贡献指南)
- [📄 许可证](#-许可证)

## ✨ 核心特性

### 🌟 企业级架构
- **分层架构设计**: 清晰的分层结构，职责分离，易于维护
- **组件化设计**: 基于模块化组件架构，每个组件职责明确，可独立升级和维护
- **依赖注入**: 基于Google Wire的自动依赖注入系统，自动扫描Provider Set
- **中间件机制**: 灵活的中间件链，支持认证、日志、监控等，内置常用中间件，支持自定义开发
- **配置管理**: 模块化配置设计，支持环境变量注入，多环境配置，配置热重载
- **日志系统**: 结构化日志，支持多种输出格式和级别

### 🚀 高性能特性
- **多协议支持**: HTTP/2、gRPC、WebSocket、TCP等协议完整支持
- **HTTP/2支持**: 原生HTTP/2协议支持
- **gRPC集成**: 完整的gRPC服务支持，支持流式RPC
- **WebSocket**: 实时双向通信支持，内置连接管理
- **TCP服务**: 自定义协议支持，底层网络控制
- **连接池管理**: 数据库和Redis连接池优化
- **异步处理**: 支持异步任务和消息队列
- **队列系统**: 基于Redis的分布式队列，支持重试、失败处理、批量操作

### 🗄️ 数据存储
- **多数据库支持**: MySQL、PostgreSQL、SQLite
- **ORM集成**: 基于GORM的数据访问层
- **泛型Repository**: 基于Go泛型的通用数据访问层，减少90%重复代码
- **Redis支持**: 缓存和会话存储
- **对象存储**: 支持阿里云OSS、腾讯云COS等
- **事务管理**: 完整的事务支持

### 🔧 开发工具
- **代码生成**: 自动生成CRUD代码
- **命令行工具**: 丰富的CLI命令支持
- **定时任务**: Cron表达式支持的定时任务系统
- **钩子系统**: 应用生命周期钩子管理
- **热重载**: 开发环境下的代码热重载

### 📨 队列系统
- **分布式队列**: 基于Redis的可靠队列实现
- **多队列支持**: 源队列、处理队列、失败队列、重试队列
- **重试机制**: 可配置的重试策略和延迟策略
- **批量处理**: 支持批量读取和处理数据
- **失败处理**: 自动失败队列管理和错误统计
- **监控统计**: 实时处理统计和性能指标

### 📊 监控与运维
- **性能分析**: 集成pprof性能分析工具
- **健康检查**: 应用健康状态监控
- **指标收集**: Prometheus指标收集
- **链路追踪**: OpenTelemetry集成
- **容器化**: Docker和Kubernetes支持

## 🏗️ 架构设计

### 整体架构
```
┌─────────────────────────────────────────────────────────────┐
│                        Taurus Pro                          │
├─────────────────────────────────────────────────────────────┤
│  HTTP Server  │  gRPC Server  │  WebSocket  │  Cli Commond │
├─────────────────────────────────────────────────────────────┤
│                    Middleware Layer                        │
│  Built-in     │  Custom       │  Auth        │  Logging    │
├─────────────────────────────────────────────────────────────┤
│  Controller   │  Service      │  Repository  │  Model      │
├─────────────────────────────────────────────────────────────┤
│                    Data Access Layer                       │
├─────────────────────────────────────────────────────────────┤
│  Database     │  Redis        │  Object Storage │  Queue   │
├─────────────────────────────────────────────────────────────┤
│                    Queue Processing Layer                  │
│  Source Queue │  Processing   │  Retry Queue │  Failed    │
└─────────────────────────────────────────────────────────────┘
```

### 🔗 组件依赖关系

```
┌─────────────────────────────────────────────────────────────┐
│                    应用层 (Application)                     │
│  Controller ←→ Service ←→ Repository ←→ Model             │
├─────────────────────────────────────────────────────────────┤
│                    框架层 (Framework)                       │
│  taurus-pro-http    │  taurus-pro-grpc    │  taurus-pro-tcp │
│  taurus-pro-storage │  taurus-pro-consul  │  taurus-pro-otel│
├─────────────────────────────────────────────────────────────┤
│                    基础层 (Foundation)                      │
│  taurus-pro-common  │  taurus-pro-config  │  Google Wire   │
│  GORM               │  gRPC               │  Redis         │
└─────────────────────────────────────────────────────────────┘
```

### 核心组件
- **Container**: 应用容器，管理所有组件实例
- **Router**: 路由管理器，支持RESTful API设计
- **Middleware**: 中间件系统，支持链式调用，内置常用中间件，支持自定义开发
- **Service**: 业务逻辑层，封装核心业务逻辑
- **Repository**: 数据访问层，抽象数据操作，基于泛型实现减少重复代码
- **Model**: 数据模型，定义业务实体结构
- **Queue Manager**: 队列管理器，处理异步任务和消息队列

## 🚀 快速开始

### 环境要求
- Go 1.24.2+
- MySQL 8.0+ / PostgreSQL 12+ / SQLite 3
- Redis 6.0+
- Docker (可选)

### 🧩 框架组件架构

Taurus Pro 框架基于模块化设计，集成了多个专业组件，每个组件都有明确的职责和能力边界。

#### **核心组件** (`github.com/stones-hub/taurus-pro-*`)

##### **taurus-pro-common**
- **地址**: https://github.com/stones-hub/taurus-pro-common
- **能力**: 框架核心公共库
  - 命令行工具 (`cmd`): 支持多种数据类型的CLI框架
  - 定时任务 (`cron`): 支持秒级精度的任务调度系统
  - 工具函数 (`util`): IP处理、字符串处理等通用工具
  - 错误处理 (`recovery`): 全局panic恢复机制
  - 配置管理 (`config`): 配置读取和环境变量处理

##### **taurus-pro-config**
- **地址**: https://github.com/stones-hub/taurus-pro-config
- **能力**: 配置管理核心
  - 多格式支持: YAML、TOML、JSON、环境变量
  - 配置热重载: 运行时动态更新配置
  - 配置验证: 启动时验证配置有效性
  - 环境隔离: 开发、测试、生产环境配置分离

##### **taurus-pro-http**
- **地址**: https://github.com/stones-hub/taurus-pro-http
- **能力**: HTTP服务框架
  - 路由管理: RESTful API路由和中间件支持
  - 内置中间件: CORS、日志、恢复、限流、压缩、安全头
  - 响应处理: 统一的响应格式和错误处理
  - 优雅关闭: 支持优雅关闭和超时控制
  - 性能优化: HTTP/2支持、连接池管理

##### **taurus-pro-grpc**
- **地址**: https://github.com/stones-hub/taurus-pro-grpc
- **能力**: gRPC服务框架
  - 服务注册: 自动服务发现和注册
  - 负载均衡: 多种负载均衡策略
  - 拦截器: 认证、日志、监控拦截器
  - 流控制: 双向流和流式RPC支持
  - TLS支持: 安全传输层配置

##### **taurus-pro-storage**
- **地址**: https://github.com/stones-hub/taurus-pro-storage
- **能力**: 数据存储抽象层
  - 数据库支持: MySQL、PostgreSQL、SQLite
  - ORM集成: 基于GORM的泛型Repository
  - 连接池: 可配置的连接池管理
  - 事务支持: 完整的事务和回滚机制
  - 对象存储: 阿里云OSS、腾讯云COS等
  - 队列系统: 基于Redis的分布式队列

##### **taurus-pro-consul**
- **地址**: https://github.com/stones-hub/taurus-pro-consul
- **能力**: 服务发现和配置中心
  - 服务注册: 自动服务注册和健康检查
  - 服务发现: 服务查询和负载均衡
  - KV存储: 配置键值对管理
  - 健康检查: HTTP、TCP、脚本健康检查
  - 数据中心: 多数据中心支持

##### **taurus-pro-opentelemetry**
- **地址**: https://github.com/stones-hub/taurus-pro-opentelemetry
- **能力**: 可观测性框架
  - 链路追踪: 分布式请求追踪
  - 指标收集: Prometheus指标收集
  - 日志聚合: 结构化日志和日志关联
  - 采样控制: 可配置的采样策略
  - 导出支持: gRPC、HTTP导出协议

##### **taurus-pro-tcp**
- **地址**: https://github.com/stones-hub/taurus-pro-tcp
- **能力**: TCP服务框架
  - 连接管理: 连接池和生命周期管理
  - 协议支持: 自定义协议解析
  - 心跳机制: Keep-Alive和连接保活
  - 缓冲区管理: 可配置的读写缓冲区
  - 并发控制: 最大连接数限制

#### **第三方组件**

##### **Google Wire**
- **地址**: https://github.com/google/wire
- **能力**: 依赖注入框架
  - 编译时依赖注入: 类型安全的依赖管理
  - 自动代码生成: 减少手写样板代码
  - 循环依赖检测: 编译时发现依赖问题

##### **GORM**
- **地址**: https://gorm.io/
- **能力**: Go语言ORM库
  - 数据库抽象: 支持多种数据库
  - 自动迁移: 数据库结构自动更新
  - 关联查询: 复杂关系查询支持
  - 钩子机制: 生命周期钩子
  - 事务支持: 完整的事务管理

##### **gRPC**
- **地址**: https://grpc.io/
- **能力**: 高性能RPC框架
  - 协议缓冲: 高效的序列化协议
  - 多语言支持: 跨语言服务调用
  - 流式RPC: 双向流通信
  - 拦截器: 中间件机制
  - 负载均衡: 客户端负载均衡

### 安装依赖
```bash
# 安装Go依赖
go mod download

# 安装Wire工具
go install github.com/google/wire/cmd/wire@latest
```

### 🔍 **Wire工具安装说明**
Wire是Google开发的Go依赖注入工具，框架集成了自动扫描功能：
- **自动扫描**: 无需手动配置，自动发现所有Provider Set
- **代码生成**: 自动生成类型安全的依赖注入代码
- **编译时检查**: 在编译时验证依赖关系，避免运行时错误

### 配置环境
```bash
# 复制环境配置文件
cp .env.example .env.local

# 编辑环境变量
vim .env.local
```

### 启动应用
```bash
# 生成依赖注入代码
make wire

# 运行应用
make run

# 或者直接运行
go run ./bin/taurus.go
```

### 🔧 Wire依赖注入系统

框架使用Google Wire进行自动依赖注入，`make wire` 命令会自动扫描并生成依赖注入代码：

#### 📁 **自动扫描机制**
- 自动扫描 `app/` 目录下的所有Go文件
- 识别以 `WireSet`、`Set`、`ProviderSet` 结尾的变量
- 自动生成 `wire_gen.go` 文件

#### 🎯 **Provider Set命名规范**
```go
// 控制器依赖注入集合
var IndexControllerSet = wire.NewSet(wire.Struct(new(IndexController), "*"))
var UserControllerSet = wire.NewSet(wire.Struct(new(UserController), "*"))

// 服务层依赖注入集合  
var IndexServiceSet = wire.NewSet(wire.Struct(new(IndexService), "*"))
var UserServiceSet = wire.NewSet(wire.Struct(new(UserService), "*"))

// 中间件依赖注入集合
var AuthMiddlewareSet = wire.NewSet(wire.Struct(new(AuthMiddleware), "*"))
```

### 访问服务
- **HTTP服务**: http://localhost:8080
- **gRPC服务**: localhost:9000
- **WebSocket服务**: ws://localhost:8080/ws
- **TCP服务**: localhost:8081
- **性能分析**: http://localhost:6060/debug/pprof/
- **健康检查**: http://localhost:8080/health

## 📚 详细文档

### 基础使用
- [路由配置](docs/routing.md)
- [中间件开发](docs/middleware.md)
- [数据库操作](docs/database.md)
- [缓存使用](docs/cache.md)

### 高级特性
- [依赖注入](docs/dependency-injection.md)
- [定时任务](docs/cron.md)
- [钩子系统](docs/hooks.md)
- [性能优化](docs/performance.md)

### API文档
- [RESTful API](docs/api/rest.md)
- [gRPC API](docs/api/grpc.md)
- [WebSocket API](docs/api/websocket.md)

## 🔧 配置说明

### 主配置文件
```yaml
# config/config.yaml
version: "${VERSION:v1.0.0}"
app_name: "${APP_NAME:taurus}"
pprof_enabled: true
go:
  max_procs: 8
  gc: 150
  memory_limit: 12
```

### HTTP配置
```yaml
# config/autoload/http/http.yaml
http:
  address: "${SERVER_ADDRESS:0.0.0.0}"
  port: ${SERVER_PORT:8080}
  shutdown_timeout: 5
  read_timeout: 30
  write_timeout: 30
  idle_timeout: 120
  authorization: "Bearer ${AUTHORIZATION:123456}"
```

### 数据库配置
```yaml
# config/autoload/db/db.yaml
databases:
  enable: true
  list:
    - dbname: "${DB_NAME:kf_ai_demo}"
      dbtype: "mysql"
      dsn: "${DB_DSN:user:pass@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local}"
      max_open_conns: 500
      max_idle_conns: 50
      conn_max_lifetime: 600
```

### 🔧 完整配置说明

框架采用模块化配置设计，所有配置都在 `config/autoload/` 目录下，支持环境变量注入。

#### **主配置文件** (`config/config.yaml`)
```yaml
version: "${VERSION:v1.0.0}"        # 应用版本
app_name: "${APP_NAME:taurus}"      # 应用名称
pprof_enabled: true                 # 是否启用性能分析
go:
  max_procs: 8                      # 最大CPU核心数
  gc: 150                           # 垃圾回收比例
  memory_limit: 12                  # 内存限制(GB)
```

#### **HTTP服务配置** (`config/autoload/http/http.yaml`)
```yaml
http:
  address: "${SERVER_ADDRESS:0.0.0.0}"  # 服务监听地址
  port: ${SERVER_PORT:8080}             # 服务端口
  shutdown_timeout: 5                    # 优雅关闭超时时间(秒)
  read_timeout: 30                       # 读取超时时间(秒)
  write_timeout: 30                      # 写入超时时间(秒)
  idle_timeout: 120                      # 空闲连接超时时间(秒)
  authorization: "Bearer ${AUTHORIZATION:123456}"  # 授权码
```

#### **数据库配置** (`config/autoload/db/db.yaml`)
```yaml
databases:
  enable: true
  list:
    - dbname: "${DB_NAME:kf_ai_demo}"    # 数据库名称标识
      dbtype: "mysql"                    # 数据库类型(mysql/postgres/sqlite)
      dsn: "${DB_DSN:user:pass@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local}"
      max_open_conns: 500                # 最大连接数
      max_idle_conns: 50                 # 最大空闲连接数
      conn_max_lifetime: 600             # 连接最大生命周期(秒)
      max_retries: 10                    # 最大重试次数
      retry_delay: 15                    # 重试延迟时间(秒)
      log_path: "./logs/db/db.log"       # 日志路径
      log_level: "info"                  # 日志级别
      log_formatter: "default"           # 日志格式
```

#### **Redis配置** (`config/autoload/redis/redis.toml`)
```toml
[redis]
enable = true                                    # 是否启用Redis
address = ["${REDIS_HOST:redis_demo}:${REDIS_PORT:6379}"]  # Redis地址列表
password = ${REDIS_PASSWORD:""}                 # Redis密码
db = 0                                          # 数据库索引
pool_size = 500                                 # 连接池大小
min_idle_conns = 50                             # 最小空闲连接数
dial_timeout = 5                                # 连接超时时间(秒)
read_timeout = 3                                # 读取超时时间(秒)
write_timeout = 3                               # 写入超时时间(秒)
max_retries = 3                                 # 最大重试次数
logger_fomatter = "default"                     # 日志格式
logger_path = "./logs/redis/redis.log"          # 日志路径
logger_level = "info"                           # 日志级别
logger_max_size = 100                           # 单个日志文件大小(MB)
logger_max_backups = 10                         # 保留的旧日志文件数量
logger_max_age = 7                              # 日志文件最大保存天数
```

#### **日志配置** (`config/autoload/logger/logger.yaml`)
```yaml
loggers:
  - name: default                               # 日志器名称
    prefix: ""                                  # 日志前缀
    log_level: 0                                # 日志级别(0:debug, 1:info, 2:warn, 3:error, 4:fatal, 5:none)
    output_type: file                           # 输出类型(console/file)
    log_file_path: logs/app.log                 # 日志文件路径
    max_size: 100                               # 单个日志文件最大大小(MB)
    max_backups: 5                              # 保留的旧日志文件数量
    max_age: 30                                 # 日志文件最大保存天数
    compress: true                              # 是否压缩旧日志文件
    formatter: default                          # 日志格式(default/json)
```

#### **OpenTelemetry配置** (`config/autoload/otel/otel.yaml`)
```yaml
otel:
  enable: true                                 # 是否启用链路追踪
  service:
    name: Taurus                               # 服务名称
    version: v0.1.0                            # 服务版本
    environment: dev                            # 环境标识
  export:
    protocol: grpc                             # 导出协议(grpc/http)
    endpoint: 192.168.3.240:4317              # 导出端点
    insecure: true                             # 是否跳过TLS验证
    timeout: 10s                               # 导出超时时间
  sampling:
    ratio: 1.0                                 # 采样比例(0.0-1.0)
  batch:
    timeout: 10s                               # 批处理超时时间
    max_size: 10                               # 批处理最大大小
    max_queue_size: 10                         # 批处理队列最大大小
    export_timeout: 10s                        # 导出超时时间
  tracers: ["http-server", "grpc-server"]      # 追踪器列表
```

#### **Consul服务发现配置** (`config/autoload/consul/consul.yaml`)
```yaml
consul:
  enable: true                                 # 是否启用Consul
  client:
    address: "192.168.3.240:8500"             # Consul服务地址
    token: ""                                  # 认证Token
    timeout: 5                                 # 客户端超时时间(秒)
    scheme: "http"                             # 协议类型
    datacenter: "dc1"                          # 数据中心
    wait_time: 10                              # 等待返回结果时间(秒)
    retry_time: 3                              # 重试间隔时间(秒)
    max_retrys: 3                              # 最大重试次数
  service:
    name: "taurus"                             # 服务名称
    id: "taurus-1"                             # 服务ID
    tags: ["http", "tcp", "https"]             # 服务标签
    address: "192.168.40.30"                   # 服务地址
    port: 8080                                 # 服务端口
    meta: {version: "v0.0.1", type: "http"}   # 服务元数据
    healths:                                   # 健康检查配置
      - http: "http://192.168.40.30:8080/health"  # 健康检查URL
        http_method: "GET"                     # 健康检查方法
        interval: 10                           # 检查间隔时间(秒)
        timeout: 5                             # 检查超时时间(秒)
        deregister_after: 10                   # 服务下线后移除时间(秒)
  watch:                                       # KV监听配置
    wait_time: 10                              # 获取KV等待时间(秒)
    retry_time: 3                              # 重试间隔时间(秒)
  invoke:                                      # 服务调用配置
    load_balance_strategy: 0                   # 负载均衡策略(0:随机, 1:轮询, 2:最少连接)
    timeout: 5                                 # 请求超时时间(秒)
    retry_count: 3                             # 重试次数
    retry_interval: 1                          # 重试间隔时间(秒)
```

#### **定时任务配置** (`config/autoload/cron/cron.yaml`)
```yaml
cron:
  enable: true                                 # 是否启用定时任务
  location: "Asia/Shanghai"                    # 时区设置
  enable_seconds: true                         # 是否启用秒级定时
  concurrency_mode: 1                          # 并发模式(0:允许并发, 1:跳过重复, 2:等待完成)
```

#### **gRPC配置** (`config/autoload/gRPC/server.yaml`)
```yaml
grpc:
  enable: true                                 # 是否启用gRPC服务
  address: "${GRPC_ADDRESS:0.0.0.0}"          # gRPC监听地址
  port: ${GRPC_PORT:9000}                     # gRPC监听端口
  tls_enable: false                            # 是否启用TLS
  cert_file: ""                                # 证书文件路径
  key_file: ""                                 # 私钥文件路径
  max_concurrent_streams: 100                 # 最大并发流数
  max_connection_idle: 300                    # 最大空闲连接时间(秒)
  max_connection_age: 600                     # 最大连接存活时间(秒)
```

#### **WebSocket配置** (`config/autoload/websocket/ws.yaml`)
```yaml
websocket:
  enable: true                                 # 是否启用WebSocket
  path: "/ws"                                  # WebSocket路径
  max_connections: 1000                        # 最大连接数
  read_buffer_size: 1024                       # 读取缓冲区大小
  write_buffer_size: 1024                      # 写入缓冲区大小
  ping_interval: 30                            # Ping间隔时间(秒)
  pong_wait: 60                                # Pong等待时间(秒)
  max_message_size: 512                        # 最大消息大小
```

#### **TCP配置** (`config/autoload/tcp/tcp.yaml`)
```yaml
tcp:
  enable: true                                 # 是否启用TCP服务
  address: "${TCP_ADDRESS:0.0.0.0}"           # TCP监听地址
  port: ${TCP_PORT:8081}                      # TCP监听端口
  max_connections: 1000                        # 最大连接数
  read_buffer_size: 1024                       # 读取缓冲区大小
  write_buffer_size: 1024                      # 写入缓冲区大小
  keep_alive: true                             # 是否启用Keep-Alive
  keep_alive_period: 30                       # Keep-Alive周期(秒)
```

#### **模板配置** (`config/autoload/templates/templates.yaml`)
```yaml
templates:
  enable: true                                 # 是否启用模板
  path: "./templates"                          # 模板文件路径
  extension: ".html"                           # 模板文件扩展名
  reload: true                                 # 是否启用热重载
  minify: false                                # 是否压缩输出
  cache: true                                  # 是否启用缓存
```

#### **MCP配置** (`config/autoload/mcp/mcp.yaml`)
```yaml
mcp:
  enable: true                                 # 是否启用MCP服务
  address: "${MCP_ADDRESS:0.0.0.0}"           # MCP监听地址
  port: ${MCP_PORT:8082}                      # MCP监听端口
  protocol: "tcp"                              # 协议类型
  timeout: 30                                  # 超时时间(秒)
  max_connections: 100                         # 最大连接数
```

### 💡 **配置管理最佳实践**

#### **环境变量注入**
框架支持环境变量注入，使用 `${ENV_NAME:default_value}` 格式：
```bash
# 环境变量设置
export VERSION=v1.0.1
export APP_NAME=myapp
export SERVER_PORT=9090
export DB_DSN="user:pass@tcp(localhost:3306)/mydb"
export REDIS_HOST=redis.example.com
export REDIS_PORT=6379
```

#### **配置文件优先级**
1. **环境变量**: 最高优先级，运行时动态修改
2. **配置文件**: 中等优先级，部署时配置
3. **默认值**: 最低优先级，代码中硬编码

#### **配置热重载**
部分配置支持热重载，无需重启服务：
- 日志级别调整
- 连接池大小修改
- 限流参数调整

#### **配置验证**
框架在启动时会验证配置的有效性：
- 必填字段检查
- 数值范围验证
- 文件路径存在性检查
- 网络连接测试


## 📦 项目结构

```
demo/
├── app/                    # 应用层
│   ├── bootstrap.go       # 应用启动引导
│   ├── command/           # 命令行工具
│   ├── controller/        # 控制器层
│   ├── crontab/           # 定时任务
│   ├── hooks/             # 生命周期钩子
│   ├── model/             # 数据模型
│   ├── process/           # 队列处理
│   ├── service/           # 业务服务层
│   ├── wire.go            # 依赖注入配置
│   └── wire_gen.go        # 自动生成的依赖注入代码
├── bin/                    # 可执行文件
│   └── taurus.go          # 主程序入口
├── config/                 # 配置文件
│   ├── autoload/          # 自动加载配置
│   └── config.yaml        # 主配置文件
├── internal/               # 内部包
│   └── taurus/            # 框架核心
├── pkg/                    # 公共包
│   └── middleware/        # 自定义中间件
├── static/                 # 静态资源
├── templates/              # 模板文件
├── test/                   # 测试文件
│   └── performance/       # 性能测试
├── benchmark/              # 基准测试
├── scripts/                # 脚本文件
├── Makefile                # 构建脚本
├── go.mod                  # Go模块文件
└── README.md               # 项目说明
```

## 🛠️ 开发指南

### 创建新控制器
```go
package controller

import (
    "net/http"
    "github.com/stones-hub/taurus-pro-http/pkg/httpx"
    "github.com/google/wire"
)

type UserController struct {
    UserService *service.UserService
}

var UserControllerSet = wire.NewSet(wire.Struct(new(UserController), "*"))

func (c *UserController) CreateUser(w http.ResponseWriter, r *http.Request) {
    // 实现创建用户逻辑
    httpx.SendResponse(w, http.StatusCreated, user, nil)
}
```

### 创建新服务
```go
package service

import (
    "context"
    "demo/app/model"
    "github.com/google/wire"
)

type UserService struct {
    userRepo *model.UserRepository
}

// UserServiceSet 必须使用WireSet、Set或ProviderSet结尾，才能被自动扫描
var UserServiceSet = wire.NewSet(wire.Struct(new(UserService), "*"))

func NewUserService() *UserService {
    userRepo, err := model.NewUserRepository()
    if err != nil {
        panic("创建UserRepository失败: " + err.Error())
    }
    return &UserService{userRepo: userRepo}
}

func (s *UserService) CreateUser(ctx context.Context, name, password string) (*model.User, error) {
    user := &model.User{Name: name, Password: password}
    err := s.userRepo.Create(ctx, user)
    return user, err
}
```

### 🔒 中间件开发与使用

框架提供了丰富的中间件支持，包括内置中间件和自定义中间件开发。

#### **使用内置中间件**
框架在 `taurus-pro-http` 包中提供了常用中间件，开箱即用：

```go
import (
    "github.com/stones-hub/taurus-pro-http/pkg/middleware"
)

// 在路由中使用内置中间件
router.Use(middleware.CorsMiddleware())           // CORS跨域支持
router.Use(middleware.LoggingMiddleware())        // 请求日志记录
router.Use(middleware.RecoveryMiddleware())       // 异常恢复
router.Use(middleware.RateLimitMiddleware())     // 限流控制
router.Use(middleware.CompressionMiddleware())    // 响应压缩
router.Use(middleware.SecurityMiddleware())       // 安全头设置
```

#### **自定义中间件开发**
在 `pkg/middleware` 包下创建自定义中间件：

```go
package middleware

import (
    "net/http"
    "time"
    "github.com/stones-hub/taurus-pro-http/pkg/httpx"
)

// 方式1: 函数式中间件
func LoggingMiddleware() func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // 调用下一个处理器
            next.ServeHTTP(w, r)
            
            // 记录请求信息
            duration := time.Since(start)
            log.Printf("请求: %s %s, 耗时: %v", r.Method, r.URL.Path, duration)
        })
    }
}

// 方式2: 结构体中间件
type AuthMiddleware struct {
    RequiredRoles []string
}

func NewAuthMiddleware(roles ...string) *AuthMiddleware {
    return &AuthMiddleware{RequiredRoles: roles}
}

func (m *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.Handler) {
    token := r.Header.Get("Authorization")
    if token == "" {
        httpx.SendResponse(w, http.StatusUnauthorized, "未授权访问", nil)
        return
    }
    
    // 验证token和角色权限
    if !m.validateToken(token) {
        httpx.SendResponse(w, http.StatusForbidden, "权限不足", nil)
        return
    }
    
    next.ServeHTTP(w, r)
}

func (m *AuthMiddleware) validateToken(token string) bool {
    // 实现token验证逻辑
    return true
}

// 方式3: 配置化中间件
type RateLimitConfig struct {
    RequestsPerMinute int
    BurstSize         int
}

func RateLimitMiddleware(config RateLimitConfig) func(next http.Handler) http.Handler {
    limiter := rate.NewLimiter(rate.Limit(config.RequestsPerMinute), config.BurstSize)
    
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                httpx.SendResponse(w, http.StatusTooManyRequests, "请求过于频繁", nil)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

#### **中间件注册与使用**
```go
// 全局中间件
router.Use(middleware.LoggingMiddleware())
router.Use(middleware.RecoveryMiddleware())

// 路由组中间件
apiGroup := router.Group("/api")
apiGroup.Use(middleware.AuthMiddleware())
apiGroup.Use(middleware.RateLimitMiddleware(RateLimitConfig{
    RequestsPerMinute: 100,
    BurstSize:         10,
}))

// 单个路由中间件
router.GET("/admin", middleware.AdminOnlyMiddleware(), adminHandler)
```

### 创建新模型
```go
package model

import (
    "context"
    "time"
    "demo/internal/taurus"
    "github.com/stones-hub/taurus-pro-storage/pkg/db/dao"
    "gorm.io/gorm"
)

// User 用户实体
type User struct {
    ID        uint      `gorm:"primarykey" json:"id"`
    Name      string    `gorm:"column:username;type:varchar(100);not null;comment:用户名" json:"name"`
    Password  string    `gorm:"column:password_hash;type:varchar(255);not null;comment:密码" json:"-"`
    CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
    UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
}

// TableName 实现Entity接口 - 返回数据库表名
func (u User) TableName() string {
    return "users"
}

// DB 实现Entity接口 - 返回数据库连接实例
func (u User) DB() *gorm.DB {
    db, exists := taurus.Container.DbList["default"]
    if !exists {
        // 如果默认数据库不存在，尝试获取第一个可用的数据库
        for _, d := range taurus.Container.DbList {
            db = d
            break
        }
    }
    return db
}

// UserRepository 用户数据访问层，继承泛型Repository减少重复代码
type UserRepository struct {
    dao.Repository[User] // 泛型实现，自动提供基础的CRUD操作
}

// NewUserRepository 创建User Repository实例
func NewUserRepository() (*UserRepository, error) {
    repo, err := dao.NewBaseRepositoryWithDB[User]()
    if err != nil {
        return nil, err
    }
    return &UserRepository{Repository: repo}, nil
}

// NewUserRepositoryWithDB 使用指定数据库连接创建Repository
func NewUserRepositoryWithDB(db *gorm.DB) *UserRepository {
    return &UserRepository{
        Repository: dao.NewBaseRepository[User](db),
    }
}

// ==================== 便捷查询方法 ====================
// 这些方法基于泛型Repository，无需重复实现基础CRUD逻辑

// FindByName 根据用户名查找用户
func (r *UserRepository) FindByName(ctx context.Context, name string) (*User, error) {
    return r.FindOneByCondition(ctx, "username = ?", name)
}

// FindByNameLike 根据用户名模糊查找用户
func (r *UserRepository) FindByNameLike(ctx context.Context, namePattern string) ([]User, error) {
    return r.FindByCondition(ctx, "username LIKE ?", "%"+namePattern+"%")
}

// FindByCreatedTimeRange 根据创建时间范围查找用户
func (r *UserRepository) FindByCreatedTimeRange(ctx context.Context, startTime, endTime time.Time) ([]User, error) {
    return r.FindByCondition(ctx, "created_at BETWEEN ? AND ?", startTime, endTime)
}

// FindActiveUsers 查找激活状态的用户
func (r *UserRepository) FindActiveUsers(ctx context.Context) ([]User, error) {
    return r.FindByCondition(ctx, "status = ?", "active")
}

// CountByDepartment 按部门统计用户数量
func (r *UserRepository) CountByDepartment(ctx context.Context, department string) (int64, error) {
    return r.CountByCondition(ctx, "department = ?", department)
}

// ==================== 业务逻辑方法 ====================
// 这些方法利用泛型Repository的高级功能

// CreateUserIfNotExists 如果用户不存在则创建
func (r *UserRepository) CreateUserIfNotExists(ctx context.Context, user *User) (*User, error) {
    exists, err := r.ExistsByCondition(ctx, "username = ?", user.Name)
    if err != nil {
        return nil, err
    }
    
    if exists {
        return r.FindByName(ctx, user.Name)
    }
    
    err = r.Create(ctx, user)
    return user, err
}

// UpdateUserStatus 更新用户状态
func (r *UserRepository) UpdateUserStatus(ctx context.Context, id uint, status string) error {
    return r.UpdateByCondition(ctx, map[string]interface{}{"status": status}, "id = ?", id)
}

// DeleteInactiveUsers 删除非活跃用户
func (r *UserRepository) DeleteInactiveUsers(ctx context.Context, days int) error {
    cutoffDate := time.Now().AddDate(0, 0, -days)
    return r.DeleteByCondition(ctx, "last_login_at < ? AND status = ?", cutoffDate, "inactive")
}

// ==================== 批量操作方法 ====================
// 利用泛型Repository的批量操作功能

// CreateUsersBatch 批量创建用户
func (r *UserRepository) CreateUsersBatch(ctx context.Context, users []User) error {
    return r.CreateBatch(ctx, users)
}

// UpdateUsersBatch 批量更新用户
func (r *UserRepository) UpdateUsersBatch(ctx context.Context, users []User) error {
    return r.UpdateBatch(ctx, users)
}

// DeleteUsersBatch 批量删除用户
func (r *UserRepository) DeleteUsersBatch(ctx context.Context, ids []uint) error {
    return r.DeleteBatch(ctx, ids)
}

// ==================== 高级查询方法 ====================
// 利用泛型Repository的分页和聚合功能

// FindUsersWithPagination 分页查询用户
func (r *UserRepository) FindUsersWithPagination(ctx context.Context, page, pageSize int, orderBy string, desc bool) ([]User, int64, error) {
    return r.FindWithPagination(ctx, page, pageSize, orderBy, desc, nil)
}

// FindUsersByConditionWithPagination 条件分页查询
func (r *UserRepository) FindUsersByConditionWithPagination(ctx context.Context, page, pageSize int, condition interface{}, args ...interface{}) ([]User, int64, error) {
    return r.FindWithPagination(ctx, page, pageSize, "created_at", true, condition, args...)
}

// GetUserStatistics 获取用户统计信息
func (r *UserRepository) GetUserStatistics(ctx context.Context) (map[string]interface{}, error) {
    stats := make(map[string]interface{})
    
    // 总用户数
    total, err := r.Count(ctx)
    if err != nil {
        return nil, err
    }
    stats["total"] = total
    
    // 今日新增用户数
    today := time.Now().Truncate(24 * time.Hour)
    tomorrow := today.Add(24 * time.Hour)
    todayCount, err := r.CountByCondition(ctx, "created_at BETWEEN ? AND ?", today, tomorrow)
    if err != nil {
        return nil, err
    }
    stats["today_new"] = todayCount
    
    return stats, nil
}
```

### 🎯 泛型Repository的优势

通过继承 `dao.Repository[User]`，您的模型自动获得以下功能，无需重复实现：

#### ✨ **基础CRUD操作**
- `Create(ctx, entity)` - 创建实体
- `FindByID(ctx, id)` - 根据ID查找
- `FindAll(ctx)` - 查找所有
- `Update(ctx, entity)` - 更新实体
- `Delete(ctx, entity)` - 删除实体
- `DeleteByID(ctx, id)` - 根据ID删除

#### 🔍 **条件查询**
- `FindByCondition(ctx, condition, args...)` - 条件查询
- `FindOneByCondition(ctx, condition, args...)` - 单条条件查询
- `CountByCondition(ctx, condition, args...)` - 条件统计
- `ExistsByCondition(ctx, condition, args...)` - 条件存在性检查

#### 📊 **分页和聚合**
- `FindWithPagination(ctx, page, pageSize, orderBy, desc, condition, args...)` - 分页查询
- `Count(ctx)` - 总数统计
- `QueryToMap(ctx, sql, args...)` - 原生SQL查询

#### ⚡ **批量操作**
- `CreateBatch(ctx, entities)` - 批量创建
- `UpdateBatch(ctx, entities)` - 批量更新
- `DeleteBatch(ctx, ids)` - 批量删除

#### 🔄 **事务支持**
- `WithTransaction(ctx, fn)` - 事务执行
- `Exec(ctx, sql, args...)` - 执行SQL

#### 💡 **使用建议**
1. **继承泛型Repository**: 减少90%的重复代码
2. **添加业务方法**: 基于泛型方法构建业务逻辑
3. **利用条件查询**: 使用 `FindByCondition` 等灵活查询
4. **批量操作优化**: 使用批量方法提高性能
5. **事务管理**: 利用 `WithTransaction` 确保数据一致性

### 🎯 **依赖注入最佳实践**

#### **Provider Set命名规范**
```go
// ✅ 正确 - 使用Set结尾，会被自动扫描
var UserControllerSet = wire.NewSet(wire.Struct(new(UserController), "*"))
var UserServiceSet = wire.NewSet(wire.Struct(new(UserService), "*"))

// ❌ 错误 - 不符合命名规范，不会被扫描
var UserControllerProvider = wire.NewSet(wire.Struct(new(UserController), "*"))
var UserServiceProvider = wire.NewSet(wire.Struct(new(UserService), "*"))
```

#### **依赖注入结构**
```go
// 控制器依赖服务
type UserController struct {
    UserService *service.UserService  // 自动注入
}

// 服务依赖Repository
type UserService struct {
    userRepo *model.UserRepository    // 自动注入
}

// Repository依赖数据库连接
type UserRepository struct {
    dao.Repository[User]              // 泛型实现
}
```

#### **自动扫描流程**
1. **开发阶段**: 创建新的Controller/Service/Repository
2. **命名规范**: 使用 `WireSet`、`Set` 或 `ProviderSet` 结尾
3. **执行命令**: 运行 `make wire` 自动扫描和生成
4. **编译运行**: 依赖关系自动解析和注入

### 创建定时任务
```go
package crontab

import (
    "context"
    "log"
    "github.com/stones-hub/taurus-pro-common/pkg/cron"
)

func init() {
    businessGroup := GetOrCreateTaskGroup("business", "core", "monitoring")
    
    statusCheckTask := cron.NewTask(
        "status_check",
        "* * * * * *", // 每1秒执行一次
        func(ctx context.Context) error {
            log.Println("执行状态检查...")
            return nil
        },
        cron.WithTimeout(10*time.Second),
        cron.WithRetry(3, time.Second),
        cron.WithGroup(businessGroup),
    )
    
    Register(statusCheckTask)
}
```

### 创建命令行工具
```go
package command

import (
    "github.com/stones-hub/taurus-pro-common/pkg/cmd"
)

type ExampleCommand struct {
    cmd.BaseCommand
}

func (c *ExampleCommand) Run(args []string) error {
    ctx, err := c.ParseOptions(args)
    if err != nil {
        return err
    }
    
    name := ctx.Options["name"].(string)
    fmt.Printf("Hello, %s!\n", name)
    return nil
}

func init() {
    baseCommand, err := cmd.NewBaseCommand(
        "hello",
        "示例命令",
        "[options]",
        []cmd.Option{
            {
                Name:        "name",
                Shorthand:   "n",
                Description: "用户名",
                Type:        cmd.OptionTypeString,
                Required:    true,
            },
        },
    )
    
    Register(&ExampleCommand{BaseCommand: *baseCommand})
}
```

### 创建队列处理器
```go
package process

import (
    "context"
    "log"
    "time"
    "github.com/stones-hub/taurus-pro-storage/pkg/queue"
    "github.com/stones-hub/taurus-pro-storage/pkg/queue/engine"
)

// 自定义数据处理器
type CustomProcessor struct {
    processedCount int64
}

func (p *CustomProcessor) Process(ctx context.Context, data []byte) error {
    // 处理业务逻辑
    log.Printf("Processing data: %s", string(data))
    
    // 模拟处理时间
    time.Sleep(100 * time.Millisecond)
    
    p.processedCount++
    return nil
}

// 初始化队列管理器
func init() {
    config := &queue.Config{
        EngineType:        engine.TypeRedis,
        Source:            "custom_source",
        Failed:            "custom_failed",
        Processing:        "custom_processing",
        Retry:             "custom_retry",
        ReaderCount:       2,
        WorkerCount:       3,
        WorkerTimeout:     time.Second * 10,
        MaxRetries:        3,
        RetryDelay:        time.Second * 5,
        EnableFailedQueue: true,
        EnableRetryQueue:  true,
    }
    
    processor := &CustomProcessor{}
    queueManager, err := queue.NewManager(processor, config)
    if err != nil {
        log.Fatalf("Failed to create queue manager: %v", err)
    }
    
    if err = queueManager.Start(); err != nil {
        log.Fatalf("Failed to start queue manager: %v", err)
    }
}

// 添加数据到队列
func AddToQueue(data []byte) error {
    return TestQueue.AddData(context.Background(), data)
}
```

## 🧪 测试与性能

### 单元测试
```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./app/service

# 运行特定测试函数
go test -v -run TestCreateUser
```

### 性能测试
```bash
# 运行性能测试
go test -v ./test/performance -run TestMemoryLeak count=1

# 运行基准测试
go test -bench=. ./test/performance
```

### API性能测试
```bash
# 安装k6
brew install k6  # macOS
# 或
sudo apt-get install k6  # Ubuntu

# 运行HTTP测试
k6 run benchmark/http.js

# 运行gRPC测试
k6 run benchmark/grpc.js

# 运行WebSocket测试
k6 run benchmark/websocket.js
```

### 内存分析
```bash
# 启动pprof服务
curl http://localhost:6060/debug/pprof/heap > heap.prof

# 分析内存使用
go tool pprof heap.prof

# 生成火焰图
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/profile
```

## 🐳 部署方案

### 本地运行
```bash
# 构建应用
make build

# 后台运行
make local-run

# 停止应用
make local-stop

# 查看日志
tail -f logs/app.log
```

### 🔧 **Makefile命令详解**

#### **Wire依赖注入生成**
```bash
make wire
```
**执行步骤**:
1. **生成扫描工具**: 编译 `./internal/wire/gen_wire.go` 为可执行文件
2. **自动扫描**: 扫描 `app/` 目录下所有以 `WireSet`、`Set`、`ProviderSet` 结尾的变量
3. **生成配置**: 自动生成 `app/wire.go` 文件
4. **代码生成**: 使用Google Wire生成最终的 `wire_gen.go` 文件

**扫描规则**:
- 自动发现所有符合命名规范的Provider Set
- 无需手动配置依赖关系
- 支持嵌套依赖和循环依赖检测

### Docker部署
```bash
# 构建并运行Docker容器
make docker-run

# 停止Docker容器
make docker-stop

# 使用Docker Compose
make docker-compose-up

# 停止Docker Compose
make docker-compose-down
```

### Docker Swarm部署
```bash
# 推送到镜像仓库
make docker-image-push

# 部署到Swarm集群
make docker-swarm-up

# 更新应用服务
make docker-update-app

# 删除Swarm服务
make docker-swarm-down
```

### 生产环境部署
```bash
# 创建发布包
make local-release

# 启动发布包
make local-release-start

# 停止发布包
make local-release-stop
```

## 📈 性能监控

### PProf性能分析
框架集成了Go的pprof性能分析工具，提供以下分析端点：

- **内存分析**: `/debug/pprof/heap`
- **CPU分析**: `/debug/pprof/profile`
- **Goroutine分析**: `/debug/pprof/goroutine`
- **阻塞分析**: `/debug/pprof/block`
- **互斥锁分析**: `/debug/pprof/mutex`

### 性能监控最佳实践
1. **内存泄漏检测**: 定期检查堆内存使用情况
2. **CPU性能分析**: 识别性能瓶颈和热点代码
3. **Goroutine监控**: 防止Goroutine泄漏
4. **数据库性能**: 监控查询性能和连接池状态
5. **网络性能**: 监控HTTP请求响应时间和吞吐量

### 监控指标
- 应用启动时间
- 请求响应时间
- 内存使用情况
- 数据库连接数
- 错误率和异常统计

## 🤝 贡献指南

### 开发流程
1. Fork项目到个人仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建Pull Request

### 代码规范
- 遵循Go官方代码规范
- 使用有意义的变量和函数名
- 添加必要的注释和文档
- 编写完整的测试用例
- 确保代码通过所有检查

### 提交规范
- 使用清晰的提交信息
- 每个提交只包含一个功能或修复
- 使用适当的提交类型前缀

## 📄 许可证

本项目采用 [Apache License 2.0](LICENSE) 许可证。

## 🙏 致谢

感谢所有为Taurus Pro框架做出贡献的开发者和用户。

---

**Taurus Pro** - 让Go开发更简单、更高效、更专业！

如有问题或建议，请通过以下方式联系：
- 📧 Email: 61647649@qq.com
- 🐛 Issues: [GitHub Issues](https://github.com/stones-hub/taurus-pro/issues)
- �� 文档: [项目文档](docs/)
