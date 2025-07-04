# Taurus Pro Core

Taurus Pro Core 是一个企业级Go微服务框架（脚手架），提供完整的微服务开发解决方案。框架采用组件化设计，支持多种通信协议，并提供全面的服务治理、可观测性和存储解决方案。

## 🌟 核心特性

### 1. 组件化架构
- 基于Wire的依赖注入系统
- 可插拔的组件设计
- 遵循Clean Architecture原则
- 组件生命周期管理
- 组件间依赖关系自动解析

### 2. 服务治理
- Consul服务注册与发现
- 多种健康检查机制
  - HTTP检查
  - gRPC检查
  - TCP检查
  - 自定义脚本检查
- 优雅启动和关闭
- 熔断和限流保护
- 服务负载均衡

### 3. 多协议支持
- HTTP/HTTPS服务
  - RESTful API支持
  - 中间件机制
  - 路由管理
  - 参数验证
- gRPC服务
  - 双向流支持
  - 拦截器机制
  - 自动代码生成
- WebSocket
  - 实时通信支持
  - 心跳机制
  - 连接管理
- TCP服务
  - 原生TCP支持
  - 自定义协议
  - 长连接管理

### 4. 可观测性
- OpenTelemetry集成
  - 分布式追踪
  - 指标收集
  - 日志关联
- 统一日志管理
  - 多日志实例
  - 日志分割
  - 日志级别控制
  - 自定义格式化
- 性能指标
  - 系统指标
  - 业务指标
  - 自定义指标
- 监控告警
  - 阈值告警
  - 异常检测
  - 告警通知

### 5. 配置管理
- 多环境配置
  - 开发环境
  - 测试环境
  - 生产环境
- 配置热更新
- 配置加密
- 分布式配置中心
- 配置版本管理
- 支持多种格式
  - YAML
  - TOML
  - JSON

### 6. 企业存储
- 数据库支持
  - MySQL
  - PostgreSQL
  - SQLite
- Redis缓存
  - 连接池管理
  - 分布式锁
  - 缓存策略
- 连接池优化
- 事务管理
- 读写分离

### 7. 开发工具
- 项目脚手架
- 代码生成
- 中间件支持
- 统一错误处理
- 测试支持
  - 单元测试
  - 集成测试
  - 性能测试

### 8. 安全特性
- 认证授权
  - JWT支持
  - OAuth2集成
  - 自定义认证
- 数据加密
- 安全配置
- 审计日志

### 9. 性能优化
- 连接池管理
- 内存复用
- 协程调度
- 资源管理

## 🚀 快速开始

### 安装

```bash
# 安装taurus命令行工具
go install github.com/stones-hub/taurus-pro-core/cmd/taurus@latest

# 创建新项目
taurus new myproject
```

### 基础配置

项目创建后，主要配置文件位于 `config/` 目录：

```yaml
# config/config.yaml
app:
  name: myproject
  version: v1.0.0

# 服务配置
server:
  http:
    enable: true
    port: 8080
  grpc:
    enable: true
    port: 9090

# 更多配置见config/autoload/目录
```

### 示例代码

1. HTTP服务示例

```go
// app/controller/index_controller.go
package controller

import (
    "net/http"
    "github.com/stones-hub/taurus-pro-http/pkg/server"
)

type IndexController struct {
    server.Controller
}

func NewIndexController() *IndexController {
    return &IndexController{}
}

// 定义路由
func (c *IndexController) Routes() []server.Route {
    return []server.Route{
        {
            Method:  "GET",
            Path:    "/hello",
            Handler: c.Hello,
        },
    }
}

func (c *IndexController) Hello(w http.ResponseWriter, r *http.Request) {
    c.JSON(w, http.StatusOK, map[string]interface{}{
        "message": "Hello Taurus",
    })
}
```

2. gRPC服务示例

```go
// proto/hello.proto
syntax = "proto3";

package hello;
option go_package = "myproject/proto/hello";

service HelloService {
    rpc SayHello (HelloRequest) returns (HelloResponse) {}
}

message HelloRequest {
    string name = 1;
}

message HelloResponse {
    string message = 1;
}

// app/service/hello_service.go
package service

import (
    "context"
    "github.com/stones-hub/taurus-pro-grpc/pkg/server"
    pb "myproject/proto/hello"
)

type HelloService struct {
    server.Service
    pb.UnimplementedHelloServiceServer
}

func NewHelloService() *HelloService {
    return &HelloService{}
}

// 注册gRPC服务
func (s *HelloService) Register(srv *server.Server) {
    pb.RegisterHelloServiceServer(srv.GRPCServer(), s)
}

func (s *HelloService) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
    // 获取请求上下文信息
    md := server.GetMetadata(ctx)
    requestID := md.Get("request-id")
    
    // 使用框架提供的日志
    s.Logger().Info("received request",
        "request_id", requestID,
        "name", req.Name,
    )
    
    return &pb.HelloResponse{
        Message: "Hello " + req.Name,
    }, nil
}

// cmd/main.go
package main

import (
    "github.com/stones-hub/taurus-pro-grpc/pkg/server"
    "myproject/app/service"
)

func main() {
    // 创建gRPC服务器
    srv := server.NewServer(
        server.WithAddress(":9090"),
        server.WithUnaryInterceptors(
            server.RecoveryInterceptor(),
            server.LoggingInterceptor(),
            server.TracingInterceptor(),
        ),
    )
    
    // 注册服务
    helloSrv := service.NewHelloService()
    helloSrv.Register(srv)
    
    // 启动服务器
    if err := srv.Start(); err != nil {
        panic(err)
    }
}
```

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详细信息。

## 🙋 获取帮助

- 提交Issue
- 查看Wiki文档
- 加入社区讨论 