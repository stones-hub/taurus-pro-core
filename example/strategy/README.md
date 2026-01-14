# 策略模式实现

## 目录结构

```
test/strategy/
├── core/                    # 核心层：接口定义和注册表
│   └── notifier.go
├── strategies/              # 策略层：具体策略实现
│   ├── email_notifier.go
│   ├── sms_notifier.go
│   └── wechat_notifier.go
├── service/                 # 服务层：调用服务
│   └── notification_service.go
├── example.go              # 使用示例
├── strategy_test.go        # 单元测试
└── cmd/
    └── main.go             # 可运行的主程序
```

## 分层说明

### 1. 核心层 (core/)
- **职责**：定义策略接口和注册表
- **文件**：`notifier.go`
- **内容**：
  - `Notifier` 接口：定义统一策略行为
  - `NotifierRegistry`：策略注册表，管理策略的注册和查找
  - `Notification`：通知数据结构

### 2. 策略层 (strategies/)
- **职责**：实现具体策略
- **文件**：`email_notifier.go`, `sms_notifier.go`, `wechat_notifier.go`
- **特点**：
  - 每个策略独立实现 `Notifier` 接口
  - 策略之间完全隔离，互不依赖
  - 可以独立测试和维护

### 3. 服务层 (service/)
- **职责**：调用策略，提供业务方法
- **文件**：`notification_service.go`
- **特点**：
  - 依赖 `Notifier` 接口，不依赖具体实现
  - 通过注册表获取策略
  - 实现策略的动态选择和调用

## 依赖关系

```
service/ → core/ (依赖接口和注册表)
strategies/ → core/ (实现接口)
service/ → strategies/ (无直接依赖，通过接口)
```

## 使用示例

```go
// 1. 创建注册表
registry := core.NewNotifierRegistry()

// 2. 创建并注册策略
emailNotifier := strategies.NewEmailNotifier("smtp.com", 587)
registry.Register(emailNotifier)

// 3. 创建服务
service := service.NewNotificationService(registry)

// 4. 使用策略
notification := &core.Notification{
    Title:   "测试",
    Content: "内容",
    To:      "user@example.com",
}
service.SendNotification(ctx, "email", notification)
```

## 运行

```bash
# 运行示例
go run ./test/strategy/cmd/main.go

# 运行测试
go test -v ./test/strategy

# 运行基准测试
go test -bench=. ./test/strategy
```
