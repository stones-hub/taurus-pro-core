# {{.ProjectName}}

这是一个基于 taurus-pro-http 的 Web 服务项目。

## 项目结构

```
.
├── app                     # 应用代码
│   ├── command            # 命令行工具
│   ├── constants          # 常量定义
│   ├── controller         # HTTP 控制器
│   ├── crontab           # 定时任务
│   ├── gRPC              # gRPC 服务
│   ├── helper            # 辅助函数
│   ├── middleware        # 中间件
│   ├── model             # 数据模型
│   ├── queue             # 队列处理
│   ├── service           # 业务逻辑
│   └── tcp               # TCP 服务
├── bin                    # 编译产物
├── config                 # 配置文件
├── docs                   # 文档
├── downloads              # 下载文件
├── example               # 示例代码
├── logs                  # 日志文件
├── scripts               # 脚本文件
├── static                # 静态文件
├── templates             # 模板文件
└── test                  # 测试文件
```

## 快速开始

### 编译项目

```bash
go build -o bin/app cmd/main.go
```

### 运行服务器

```bash
./bin/app
```

## 配置说明

配置文件位于 `config` 目录下，支持以下配置：

- HTTP 服务器配置
- 数据库配置
- 日志配置
- 中间件配置

## 许可证

Apache-2.0 license 