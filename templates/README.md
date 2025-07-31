# Taurus Pro Core 框架使用指南

---

## 一、环境配置

### 1.1、环境变量文件

项目支持多种环境变量文件配置：

- **`.env.local`**：本地部署的默认环境变量文件
- **`.env.docker-compose`**：Docker-Compose 部署所需的环境变量文件

### 1.2、必需的环境变量

以下环境变量必须在环境变量文件中设置：

```bash
VERSION=1.0.0                    # 应用版本号
APP_NAME=taurus-pro              # 应用名称
SERVER_ADDRESS=0.0.0.0          # 服务器地址
SERVER_PORT=8080                # 服务器端口
APP_CONFIG=config/config.yaml   # 应用配置文件路径
HOST_PORT=8080                  # 主机端口（Docker）
CONTAINER_PORT=8080             # 容器端口（Docker）
WORKDIR=/app                    # 工作目录
REGISTRY_URL=your-registry.com  # Docker镜像注册中心地址
```

---

## 二、Wire 依赖注入

### 2.1、生成 Wire 代码

Wire 是 Google 的依赖注入工具，用于自动生成依赖注入代码：

```shell
make wire
```

此命令会：
1. 生成 app 目录的 wire 扫描工具
2. 扫描 app 目录下的 provider set，并生成 wire.go 文件
3. 生成 wire_gen.go 文件

### 2.2、运行应用（包含 Wire 生成）

```shell
make run
```

此命令会先执行 `make wire`，然后运行应用。

---

## 三、构建和清理

### 3.1、构建应用

```shell
make build
```

此命令会：
1. 执行 `make wire` 生成依赖注入代码
2. 在 `build/` 目录下生成可执行文件

### 3.2、清理构建文件

```shell
make clean
```

此命令会清理：
- `build/` 目录
- `release/` 目录

---

## 四、本地部署

### 4.1、本地运行（后台模式）

```shell
make local-run
```

此命令会：
1. 清理并重新构建应用
2. 在后台启动应用
3. 将日志输出到 `logs/app.log`
4. 将进程PID保存到 `logs/app.pid`

**查看日志：**
```shell
tail -f logs/app.log
```

**停止应用：**
```shell
make local-stop
```

### 4.2、发布包管理

#### 创建发布包

```shell
make local-release
```

此命令会：
1. 清理并重新构建应用
2. 在 `release/` 目录下创建完整的发布包，包含：
   - 可执行文件
   - 环境配置文件
   - 配置目录（config、templates、static）
   - 文档和脚本（docs、scripts）
   - 测试和基准测试目录（test、benchmark）
   - 示例配置（example）
   - 重要根目录文件（Makefile、README.md、LICENSE、Dockerfile等）
3. 创建压缩包 `release/$(APP_NAME)-$(VERSION).tar.gz`

#### 从发布包启动应用

```shell
make local-release-start
```

**停止发布包应用：**
```shell
make local-release-stop
```

---

## 五、Docker 部署

### 5.1、Docker 单容器部署

#### 构建并运行 Docker 容器

```shell
make docker-run
```

此命令会：
1. 构建 Docker 镜像
2. 创建 Docker 网络
3. 创建日志和数据卷
4. 运行容器

#### 停止并清理 Docker 资源

```shell
make docker-stop
```

此命令会：
1. 停止并删除容器
2. 删除网络
3. 删除镜像

### 5.2、Docker Compose 部署

#### 启动 Docker Compose 服务

```shell
make docker-compose-up env_file=.env.docker-compose
```

#### 停止 Docker Compose 服务

```shell
make docker-compose-down env_file=.env.docker-compose
```

#### 启动已存在的服务

```shell
make docker-compose-start env_file=.env.docker-compose
```

#### 停止服务（保持容器）

```shell
make docker-compose-stop env_file=.env.docker-compose
```

---

## 六、Docker Swarm 集群部署

### 6.1、推送镜像到注册中心

```shell
make docker-image-push env_file=.env.docker-compose
```

### 6.2、集群管理

#### 部署到 Swarm 集群

```shell
make docker-swarm-up env_file=.env.docker-compose
```

此命令会：
1. 推送镜像到注册中心
2. 部署应用到 Swarm 集群

#### 停止整个集群

```shell
make docker-swarm-down env_file=.env.docker-compose
```

### 6.3、服务更新

#### 更新应用服务（零停机）

```shell
make docker-update-app env_file=.env.docker-compose
```

> **重要注意事项：**
> 
> **适用条件（缺一不可）：**
> - 副本是VIP（虚拟IP）模式
> - `docker-compose-swarm.yml` 文件未被修改
> - Nginx负载均衡模式不使用`ip_hash`
> 
> **限制说明：**
> - 更新会导致应用服务的IP发生变化
> - 如果上下游服务依赖IP连接，可能导致服务不可用
> - `docker service update` 不支持通过`env-file`传递环境变量
> - 需要手动读取环境变量文件并构建`--env-add`参数

#### 重新部署应用服务（完全重建）

```shell
make docker-swarm-deploy-app env_file=.env.docker-compose
```

> **重要注意事项：**
> - 适用于app或nginx修改了任意配置的情况
> - 整个服务会被删掉重建，会导致集群暂时不可用
> - 恢复需要时间，请谨慎操作
> - 不支持 `docker service scale` 扩缩容

---

## 七、配置文件指南

### 7.1、配置目录结构

- **`config/`**：应用内各种组件的配置文件
- **`config/autoload/`**：自动加载的配置文件
  - `consul/`：Consul 配置
  - `cron/`：定时任务配置
  - `db/`：数据库配置
  - `gRPC/`：gRPC 服务配置
  - `http/`：HTTP 服务配置
  - `logger/`：日志配置
  - `mcp/`：MCP 配置
  - `otel/`：OpenTelemetry 配置
  - `redis/`：Redis 配置
  - `tcp/`：TCP 服务配置
  - `templates/`：模板配置
  - `websocket/`：WebSocket 配置

### 7.2、环境变量文件

- **`.env.local`**：本地部署的默认环境变量
- **`.env.docker-compose`**：Docker-Compose 部署所需的环境变量

### 7.3、Docker 配置文件

- **`docker-compose.yml`**：Docker-Compose 单机部署配置
- **`docker-compose-swarm.yml`**：Swarm 集群部署配置

---

## 八、MCP(SSE模式)注意事项

### 8.1、Nginx 配置要求

1. 需要使用 `ip_hash` 模式，确保同一客户端的请求分配到同一 Nginx 实例
2. 需要支持 `/see` 特殊协议

### 8.2、应用副本路由模式

1. 需要使用 `dnsrr` 模式，确保同一客户端的请求分配到同一 App 实例

### 8.3、更新注意事项

1. 更新 App 时需要重新部署 Nginx 和 App
2. `dnsrr` 模式下，仅更新 App 时 Nginx 不会自动更新到新的副本 IP
3. 不支持 `docker service scale` 扩缩容

---

## 九、项目结构

```
taurus-pro-core/
├── app/                   # 应用主目录
│   ├── bootstrap.gotmpl   # 应用启动模板
│   ├── constants/         # 常量定义
│   ├── controller/        # 控制器层
│   ├── crontab/          # 定时任务
│   ├── helper/           # 辅助函数
│   ├── hooks/            # 生命周期钩子
│   ├── model/            # 数据模型
│   ├── process/          # 进程管理
│   └── service/          # 服务层
├── bin/                   # 可执行文件目录
├── config/               # 配置文件目录
│   ├── config.yaml       # 主配置文件
│   └── autoload/         # 自动加载配置
│       ├── consul/       # Consul 配置
│       ├── cron/         # 定时任务配置
│       ├── db/           # 数据库配置
│       ├── gRPC/         # gRPC 服务配置
│       ├── http/         # HTTP 服务配置
│       ├── logger/       # 日志配置
│       ├── mcp/          # MCP 配置
│       ├── otel/         # OpenTelemetry 配置
│       ├── redis/        # Redis 配置
│       ├── tcp/          # TCP 服务配置
│       ├── templates/    # 模板配置
│       └── websocket/    # WebSocket 配置
├── internal/             # 内部包
│   ├── taurus/           # Taurus 核心包
│   └── wire/             # Wire 依赖注入
├── pkg/                  # 公共包
├── scripts/              # 脚本文件
│   ├── data/             # 数据脚本
│   ├── nginx/            # Nginx 配置脚本
│   └── wire/             # Wire 生成脚本
├── static/               # 静态资源
├── templates/            # 模板文件
├── test/                 # 测试文件
├── benchmark/            # 性能测试
├── docs/                 # 文档
├── downloads/            # 下载文件目录
├── example/              # 示例配置
├── logs/                 # 日志目录
├── Dockerfile            # Docker 构建文件
├── docker-compose.yml    # Docker Compose 配置
├── docker-compose-swarm.yml # Docker Swarm 配置
├── Makefile              # 构建脚本
├── README.md             # 项目说明
├── LICENSE               # 许可证
├── .dockerignore         # Docker 忽略文件
├── .gitignore            # Git 忽略文件
└── .env.local            # 本地环境变量（需要创建）
```

---

## 十、最佳实践

### 10.1、环境变量管理

- **优先级**：环境变量中的配置会覆盖 `config` 文件内的配置
- **自定义配置路径**：可在环境变量文件中修改配置文件目录
- **集中管理**：建议将配置集中在 `config` 目录和环境变量中进行管理

### 10.2、部署建议

- **本地开发**：使用 `make local-run` 进行快速开发和测试
- **单机部署**：使用 `make docker-compose-up` 进行容器化部署
- **生产环境**：使用 `make docker-swarm-up` 进行集群部署

### 10.3、更新流程

1. **开发环境**：修改代码后使用 `make run` 快速测试
2. **测试环境**：使用 `make local-release` 创建发布包进行测试
3. **生产环境**：使用 `make docker-swarm-deploy-app` 进行完整部署

---

## 十一、故障排除

### 11.1、常见问题

1. **环境变量文件不存在**
   ```
   Error: Environment file '.env.local' not found
   ```
   解决：创建对应的环境变量文件并设置必需的环境变量

2. **必需环境变量未设置**
   ```
   Error: VERSION is not set in the environment variables
   ```
   解决：在环境变量文件中设置所有必需的环境变量

3. **Docker 镜像构建失败**
   解决：检查 Dockerfile 和构建上下文是否正确

### 11.2、日志查看

- **本地运行**：`tail -f logs/app.log`
- **Docker 运行**：`docker logs $(APP_NAME)`
- **Docker Compose**：`docker-compose logs -f`
- **Docker Swarm**：`docker service logs $(APP_NAME)_app`

### 11.3、服务状态检查

- **本地进程**：`ps aux | grep $(APP_NAME)`
- **Docker 容器**：`docker ps | grep $(APP_NAME)`
- **Docker Compose**：`docker-compose ps`
- **Docker Swarm**：`docker service ls | grep $(APP_NAME)`