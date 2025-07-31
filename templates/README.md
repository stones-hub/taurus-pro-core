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

### 2.2、依赖注入自动注入规则

为了确保依赖注入能够自动注入，必须遵循以下严格规则：

#### 命名规则
Provider Set 变量名必须以以下后缀之一结尾：
- **`Set`** - 例如：`UserServiceSet`
- **`ProviderSet`** - 例如：`UserServiceProviderSet`  
- **`WireSet`** - 例如：`UserServiceWireSet`

#### 结构体匹配规则
- **严格匹配**：Provider Set 变量对应的结构体必须在**当前文件**中定义
- **命名对应**：Provider Set 变量名去掉后缀后，必须与结构体名称完全匹配

#### 示例

**✅ 正确的示例：**

```go
// user_service.go
type UserService struct {
    // 结构体定义
}

// 以下任意一种命名都可以被自动识别
var UserServiceSet = wire.NewSet(NewUserService)
var UserServiceProviderSet = wire.NewSet(NewUserService)
var UserServiceWireSet = wire.NewSet(NewUserService)
```

**❌ 错误的示例：**

```go
// 错误1：变量名不以指定后缀结尾
var UserServiceProvider = wire.NewSet(NewUserService)  // 缺少 "Set"

// 错误2：结构体不在当前文件中定义
// user_service.go 中没有 UserService 结构体定义
var UserServiceSet = wire.NewSet(NewUserService)

// 错误3：命名不匹配
type UserController struct {}
var UserServiceSet = wire.NewSet(NewUserController)  // 名称不匹配
```

#### 自动注入流程
1. 扫描器识别符合命名规则的 `wire.NewSet` 变量
2. 从变量名推断对应的结构体名称（去掉后缀）
3. 在当前文件中查找对应的结构体定义
4. 只有找到匹配的结构体时，才会被添加到依赖注入中
5. 生成正确的 wire 代码和包路径引用

### 2.3、运行应用（包含 Wire 生成）

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

## 十一、Hook 生命周期钩子

### 11.1、Hook 概述

Hook 是应用生命周期管理的重要组件，允许在应用启动和停止时执行自定义逻辑。支持优先级控制，确保钩子按正确顺序执行。

### 11.2、Hook 类型

- **`HookTypeStart`**：应用启动时执行的钩子
- **`HookTypeStop`**：应用停止时执行的钩子

### 11.3、注册 Hook

#### 基本注册方式

```go
package hooks

import (
    "context"
    "log"
)

func init() {
    // 注册启动钩子
    RegisterHook("my_start_hook", HookTypeStart, func(ctx context.Context) error {
        log.Println("应用启动时执行...")
        // 执行启动逻辑
        return nil
    }, 100) // 优先级 100

    // 注册停止钩子
    RegisterHook("my_stop_hook", HookTypeStop, func(ctx context.Context) error {
        log.Println("应用停止时执行...")
        // 执行清理逻辑
        return nil
    }, 100)
}
```

#### 默认优先级注册

```go
func init() {
    // 使用默认优先级 100
    RegisterDefaultHook("simple_hook", HookTypeStart, func(ctx context.Context) error {
        log.Println("简单钩子执行...")
        return nil
    })
}
```

### 11.4、Hook 优先级

- **优先级范围**：0-10，数值越大优先级越高
- **执行顺序**：高优先级钩子先执行
- **默认优先级**：100

### 11.5、Hook 最佳实践

1. **资源初始化**：在启动钩子中初始化数据库连接、缓存等
2. **优雅关闭**：在停止钩子中关闭连接、保存状态等
3. **错误处理**：钩子函数应返回错误，框架会记录日志
4. **超时控制**：钩子执行有10秒超时限制

---

## 十二、Crontab 定时任务

### 12.1、Crontab 概述

Crontab 提供强大的定时任务管理功能，支持任务分组、标签、重试、超时等特性。

### 12.2、创建定时任务

#### 基本任务创建

```go
package crontab

import (
    "context"
    "log"
    "time"

    "github.com/stones-hub/taurus-pro-common/pkg/cron"
)

func init() {
    // 创建简单任务
    simpleTask := cron.NewTask(
        "simple_task",
        "*/5 * * * * *", // 每5秒执行一次
        func(ctx context.Context) error {
            log.Println("执行简单任务...")
            return nil
        },
    )

    Register(simpleTask)
}
```

#### 高级任务配置

```go
func init() {
    // 创建任务组
    businessGroup := GetOrCreateTaskGroup("business", "core", "monitoring")

    // 创建复杂任务
    complexTask := cron.NewTask(
        "complex_task",
        "0 */5 * * * *", // 每5分钟执行一次
        func(ctx context.Context) error {
            log.Println("开始执行复杂任务...")
            
            // 检查上下文取消
            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-time.After(30 * time.Second):
                log.Println("任务执行完成")
                return nil
            }
        },
        cron.WithTimeout(45*time.Second),        // 设置超时时间
        cron.WithRetry(3, time.Second),          // 设置重试次数和间隔
        cron.WithGroup(businessGroup),           // 设置任务组
        cron.WithTag("data_sync"),               // 添加标签
        cron.WithTag("periodic"),                // 添加多个标签
    )

    Register(complexTask)
}
```

### 12.3、Cron 表达式

支持标准的 cron 表达式格式：`秒 分 时 日 月 星期`

#### 常用表达式示例

```go
"* * * * * *"     // 每秒执行
"*/5 * * * * *"   // 每5秒执行
"0 * * * * *"     // 每分钟执行
"0 */5 * * * *"   // 每5分钟执行
"0 0 * * * *"     // 每小时执行
"0 0 0 * * *"     // 每天0点执行
"0 0 12 * * *"    // 每天12点执行
"0 0 0 * * 1"     // 每周一0点执行
```

### 12.4、任务配置选项

#### 超时控制

```go
cron.WithTimeout(30 * time.Second)  // 任务超时时间
```

#### 重试机制

```go
cron.WithRetry(3, time.Second)      // 失败重试3次，间隔1秒
```

#### 任务分组

```go
// 创建任务组
group := GetOrCreateTaskGroup("business", "core", "monitoring")

// 使用任务组
cron.WithGroup(group)
```

#### 标签管理

```go
cron.WithTag("data_sync")           // 添加单个标签
cron.WithTag("periodic", "core")    // 添加多个标签
```

### 12.5、任务管理最佳实践

1. **任务命名**：使用有意义的任务名称
2. **错误处理**：在任务函数中正确处理错误
3. **上下文检查**：定期检查 `ctx.Done()` 以支持优雅停止
4. **资源管理**：合理设置超时和重试参数
5. **日志记录**：记录任务执行的关键信息

---

## 十三、Command 命令行工具

### 13.1、Command 概述

Command 提供强大的命令行工具支持，可以创建自定义命令用于系统管理、数据处理等场景。

### 13.2、创建自定义命令

#### 基本命令结构

```go
package command

import (
    "fmt"
    "log"

    "github.com/stones-hub/taurus-pro-common/pkg/cmd"
)

// 继承 BaseCommand 和 Command 接口
type MyCommand struct {
    cmd.BaseCommand
}

// Run 执行命令逻辑
func (c *MyCommand) Run(args []string) error {
    ctx, err := c.ParseOptions(args)
    if err != nil {
        return err
    }

    // 获取选项值
    name := ctx.Options["name"].(string)
    age := ctx.Options["age"].(int)

    fmt.Printf("Hello %s, you are %d years old\n", name, age)
    return nil
}

func init() {
    // 创建基础命令
    baseCommand, err := cmd.NewBaseCommand(
        "hello",                    // 命令名称
        "Say hello to someone",     // 命令描述
        "[options]",                // 使用说明
        []cmd.Option{               // 选项定义
            {
                Name:        "name",
                Shorthand:   "n",
                Description: "Your name",
                Type:        cmd.OptionTypeString,
                Required:    true,
            },
            {
                Name:        "age",
                Shorthand:   "a",
                Description: "Your age",
                Type:        cmd.OptionTypeInt,
                Default:     25,
            },
        },
    )
    if err != nil {
        log.Printf("NewBaseCommand failed: %v\n", err)
        return
    }

    // 注册命令
    Register(&MyCommand{
        BaseCommand: *baseCommand,
    })
}
```

### 13.3、选项类型支持

#### 字符串选项

```go
{
    Name:        "name",
    Shorthand:   "n",
    Description: "用户名",
    Type:        cmd.OptionTypeString,
    Required:    true,
}
```

#### 整数选项

```go
{
    Name:        "age",
    Shorthand:   "a",
    Description: "年龄",
    Type:        cmd.OptionTypeInt,
    Default:     25,
}
```

#### 浮点数选项

```go
{
    Name:        "score",
    Shorthand:   "s",
    Description: "评分",
    Type:        cmd.OptionTypeFloat,
    Default:     85.5,
}
```

#### 布尔选项

```go
{
    Name:        "verbose",
    Shorthand:   "v",
    Description: "详细输出",
    Type:        cmd.OptionTypeBool,
    Default:     false,
}
```

### 13.4、运行命令

#### 脚本模式运行

```bash
# 使用 --script 参数启用脚本模式
./taurus --script user --name "张三" --age 30 --verbose

# 使用短参数
./taurus --script user -n "张三" -a 30 -v

# 查看帮助
./taurus --script user --help
```

#### 命令示例

```bash
# 基本使用
./taurus --script user --name "李四" --email "lisi@example.com"

# 使用所有选项
./taurus --script user \
  --name "王五" \
  --email "wangwu@example.com" \
  --age 28 \
  --active \
  --score 92.5 \
  --verbose \
  --roles "admin,user" \
  --department "技术部" \
  --level 3 \
  --verified \
  --salary 25000.0
```

### 13.5、命令开发最佳实践

1. **命令命名**：使用简洁、有意义的命令名称
2. **选项设计**：合理设计必填和可选选项
3. **错误处理**：提供清晰的错误信息和帮助
4. **输出格式**：使用统一的输出格式
5. **文档说明**：为每个选项提供清晰的描述

---

## 十四、故障排除

### 14.1、常见问题

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

4. **Hook 执行失败**
   ```
   Hook execution failed: context deadline exceeded
   ```
   解决：检查钩子函数是否超时，优化执行逻辑

5. **定时任务注册失败**
   ```
   register task failed: invalid cron expression
   ```
   解决：检查 cron 表达式格式是否正确

6. **命令执行失败**
   ```
   Command run failed: required option not provided
   ```
   解决：检查是否提供了所有必需的选项

### 14.2、日志查看

- **本地运行**：`tail -f logs/app.log`
- **Docker 运行**：`docker logs $(APP_NAME)`
- **Docker Compose**：`docker-compose logs -f`
- **Docker Swarm**：`docker service logs $(APP_NAME)_app`

### 14.3、服务状态检查

- **本地进程**：`ps aux | grep $(APP_NAME)`
- **Docker 容器**：`docker ps | grep $(APP_NAME)`
- **Docker Compose**：`docker-compose ps`
- **Docker Swarm**：`docker service ls | grep $(APP_NAME)`

### 14.4、性能分析

#### PProf 访问

应用内置了 PProf 性能分析工具，默认在 `localhost:6060` 端口提供服务：

```bash
# 内存分析
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# CPU 分析
curl http://localhost:6060/debug/pprof/profile > cpu.prof
go tool pprof cpu.prof

# Goroutine 分析
curl http://localhost:6060/debug/pprof/goroutine?debug=2 > goroutine.txt
```

#### 常用分析命令

```bash
# 查看内存使用情况
(pprof) top

# 查看特定函数的内存分配
(pprof) list <function_name>

# 生成火焰图
go tool pprof -http=:8081 http://localhost:6060/debug/pprof/profile

# 实时监控
watch -n 5 'curl -s http://localhost:6060/debug/pprof/ | grep -E "goroutine|heap"'
```