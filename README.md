# Taurus Pro Core

这是一个基于 taurus-pro-http 的项目脚手架，用于快速创建和启动新的 Web 服务项目。

## 项目结构

```
.
├── cmd                     # 命令行工具
│   └── taurus             # 主程序入口
├── internal               # 内部代码
│   ├── config            # 配置
│   ├── handler           # HTTP 处理器
│   ├── middleware        # 中间件
│   ├── model            # 数据模型
│   ├── repository       # 数据访问层
│   └── service          # 业务逻辑层
├── pkg                   # 可重用的包
├── scripts              # 脚本文件
└── test                 # 测试文件
```

## 快速开始

1. 创建新项目：
```bash
go mod init your-project-name
```

2. 添加依赖：
```bash
go get github.com/stones-hub/taurus-pro-http
```

3. 启动服务：
```bash
go run cmd/taurus/main.go
```

## 配置说明

配置文件位于 `internal/config` 目录下，支持以下配置：

- HTTP 服务器配置
- 数据库配置
- 日志配置
- 中间件配置

## 使用示例

```go
package main

import (
    "github.com/stones-hub/taurus-pro-http/server"
)

func main() {
    srv := server.New()
    srv.Start()
}
```

## 贡献指南

欢迎提交 Issue 和 Pull Request。

## 许可证

Apache-2.0 license
