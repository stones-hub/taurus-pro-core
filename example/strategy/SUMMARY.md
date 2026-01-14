# 策略模式实现总结

## 一、策略模式概述

策略模式（Strategy Pattern）是一种行为设计模式，它定义了一系列算法，把它们封装起来，并且使它们可以相互替换。策略模式让算法独立于使用它的客户端而变化。

### 核心思想

- **定义算法族**：将不同的算法封装成独立的策略类
- **封装变化**：将算法的变化部分抽象出来
- **可替换性**：运行时动态选择策略，无需修改调用代码

## 二、实现架构

### 目录结构

```
test/strategy/
├── core/                           # 核心层：接口定义和注册表
│   └── notifier.go                 # Notifier 接口 + NotifierRegistry
├── strategies/                     # 策略层：具体策略实现
│   ├── email_notifier.go          # 邮件通知策略
│   ├── sms_notifier.go            # 短信通知策略
│   └── wechat_notifier.go        # 微信通知策略
├── service/                        # 服务层：调用服务
│   └── notification_service.go   # 通知服务
├── example.go                      # 使用示例
├── strategy_test.go              # 单元测试
└── cmd/
    └── main.go                    # 可运行的主程序
```

### 分层架构

```
┌─────────────────────────────────┐
│  服务层 (service/)               │
│  - 调用策略，提供业务方法         │
│  - 依赖接口，不依赖具体实现       │
└──────────────┬──────────────────┘
               │ 依赖
               ↓
┌─────────────────────────────────┐
│  核心层 (core/)                 │
│  - Notifier 接口定义              │
│  - NotifierRegistry 注册表       │
└──────────────┬──────────────────┘
               │ 实现
               ↓
┌─────────────────────────────────┐
│  策略层 (strategies/)            │
│  - EmailNotifier                │
│  - SMSNotifier                  │
│  - WechatNotifier               │
└─────────────────────────────────┘
```

## 三、核心组件

### 1. 核心层 (core/notifier.go)

#### 1.1 接口定义

```go
// Notifier 通知策略接口
type Notifier interface {
    // Name 返回策略名称，用于注册和识别
    Name() string

    // Send 发送通知的核心方法
    Send(ctx context.Context, notification *Notification) error

    // Validate 验证通知参数是否有效
    Validate(notification *Notification) error
}
```

**设计要点：**
- 接口定义要稳定，避免频繁变更
- 方法签名清晰，职责单一
- 支持 Context 传递，便于控制超时和取消

#### 1.2 数据结构

```go
// Notification 通知消息结构
type Notification struct {
    Title   string            // 标题
    Content string            // 内容
    To      string            // 接收者
    Extras  map[string]string // 扩展信息
}
```

#### 1.3 注册表

```go
// NotifierRegistry 策略注册表
type NotifierRegistry struct {
    strategies map[string]Notifier  // 用 map 存储策略
}

// 核心方法
func NewNotifierRegistry() *NotifierRegistry
func (r *NotifierRegistry) Register(notifier Notifier)
func (r *NotifierRegistry) Get(name string) (Notifier, error)
func (r *NotifierRegistry) List() []string
func (r *NotifierRegistry) Exists(name string) bool
```

**设计要点：**
- 使用 `map[string]Notifier` 存储策略，通过名称快速查找
- 提供注册、获取、列表、检查等方法
- 支持全局注册表（可选）

### 2. 策略层 (strategies/)

#### 2.1 邮件策略 (email_notifier.go)

```go
type EmailNotifier struct {
    smtpHost string
    smtpPort int
}

func (n *EmailNotifier) Name() string {
    return "email"
}

func (n *EmailNotifier) Validate(notification *core.Notification) error {
    // 验证邮箱格式、标题、内容等
    if !strings.Contains(notification.To, "@") {
        return fmt.Errorf("invalid email address: %s", notification.To)
    }
    // ...
    return nil
}

func (n *EmailNotifier) Send(ctx context.Context, notification *core.Notification) error {
    // 1. 验证参数
    if err := n.Validate(notification); err != nil {
        return err
    }
    // 2. 执行发送逻辑
    // ...
    return nil
}
```

#### 2.2 短信策略 (sms_notifier.go)

```go
type SMSNotifier struct {
    apiKey    string
    apiSecret string
}

func (n *SMSNotifier) Name() string {
    return "sms"
}

func (n *SMSNotifier) Validate(notification *core.Notification) error {
    // 验证手机号、内容长度等
    if len(notification.To) < 10 {
        return fmt.Errorf("invalid phone number: %s", notification.To)
    }
    if len(notification.Content) > 500 {
        return fmt.Errorf("sms content too long, max 500 characters")
    }
    return nil
}

func (n *SMSNotifier) Send(ctx context.Context, notification *core.Notification) error {
    // 发送短信逻辑
    // ...
    return nil
}
```

#### 2.3 微信策略 (wechat_notifier.go)

```go
type WechatNotifier struct {
    appID     string
    appSecret string
}

func (n *WechatNotifier) Name() string {
    return "wechat"
}

func (n *WechatNotifier) Validate(notification *core.Notification) error {
    // 验证 OpenID、标题或内容等
    if len(notification.To) < 20 {
        return fmt.Errorf("invalid wechat openid: %s", notification.To)
    }
    return nil
}

func (n *WechatNotifier) Send(ctx context.Context, notification *core.Notification) error {
    // 发送微信消息逻辑
    // ...
    return nil
}
```

**策略层特点：**
- 每个策略独立实现 `Notifier` 接口
- 策略之间完全隔离，互不依赖
- 每个策略有独立的验证逻辑
- 可以独立测试和维护

### 3. 服务层 (service/notification_service.go)

```go
type NotificationService struct {
    registry *core.NotifierRegistry
}

// 单一策略发送
func (s *NotificationService) SendNotification(
    ctx context.Context, 
    strategyName string, 
    notification *core.Notification,
) error {
    // 1. 从注册表获取策略
    notifier, err := s.registry.Get(strategyName)
    if err != nil {
        return err
    }
    // 2. 调用接口方法（多态）
    return notifier.Send(ctx, notification)
}

// 批量发送（多策略并发）
func (s *NotificationService) SendNotificationBatch(
    ctx context.Context,
    strategyNames []string,
    notification *core.Notification,
) map[string]error {
    // 并发执行多个策略
    // ...
}

// 降级策略
func (s *NotificationService) SendNotificationWithFallback(
    ctx context.Context,
    primaryStrategy, fallbackStrategy string,
    notification *core.Notification,
) error {
    // 主策略失败时自动切换备用策略
    // ...
}
```

**服务层特点：**
- 只依赖 `Notifier` 接口，不依赖具体实现
- 通过策略名称动态选择策略
- 提供多种调用方式：单一、批量、降级

## 四、使用示例

### 示例 1：基本使用

```go
package main

import (
    "context"
    "esim/test/strategy/core"
    "esim/test/strategy/service"
    "esim/test/strategy/strategies"
)

func main() {
    // 1. 创建注册表
    registry := core.NewNotifierRegistry()

    // 2. 创建并注册策略
    emailNotifier := strategies.NewEmailNotifier("smtp.example.com", 587)
    smsNotifier := strategies.NewSMSNotifier("api_key", "api_secret")
    wechatNotifier := strategies.NewWechatNotifier("app_id", "app_secret")

    registry.Register(emailNotifier)
    registry.Register(smsNotifier)
    registry.Register(wechatNotifier)

    // 3. 创建服务
    svc := service.NewNotificationService(registry)

    // 4. 使用策略发送通知
    ctx := context.Background()
    notification := &core.Notification{
        Title:   "欢迎注册",
        Content: "感谢您注册我们的服务！",
        To:      "user@example.com",
    }

    // 使用邮件策略
    err := svc.SendNotification(ctx, "email", notification)
    if err != nil {
        fmt.Printf("发送失败: %v\n", err)
    }
}
```

### 示例 2：批量发送

```go
// 同时使用多个策略发送通知
batchNotification := &core.Notification{
    Title:   "系统维护通知",
    Content: "系统将于今晚 22:00-24:00 进行维护",
    To:      "user@example.com",
}

results := svc.SendNotificationBatch(ctx, []string{"email", "sms", "wechat"}, batchNotification)
for strategy, err := range results {
    if err != nil {
        fmt.Printf("策略 '%s' 发送失败: %v\n", strategy, err)
    } else {
        fmt.Printf("策略 '%s' 发送成功\n", strategy)
    }
}
```

### 示例 3：降级策略

```go
// 如果邮件发送失败，自动使用短信
notification := &core.Notification{
    Title:   "重要通知",
    Content: "这是一条重要消息",
    To:      "user@example.com",
}

err := svc.SendNotificationWithFallback(ctx, "email", "sms", notification)
if err != nil {
    fmt.Printf("所有策略都失败: %v\n", err)
}
```

### 示例 4：使用 Context 控制

```go
// 创建带超时的 context
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

notification := &core.Notification{
    Title:   "超时测试",
    Content: "这条消息可能会因为超时而取消",
    To:      "user@example.com",
}

err := svc.SendNotification(ctx, "email", notification)
if err != nil {
    fmt.Printf("发送失败: %v\n", err)
}
```

### 示例 5：参数验证

```go
// 使用指定策略的验证逻辑
invalidNotification := &core.Notification{
    Content: "测试",
    To:      "", // 缺少接收者
}

err := svc.ValidateNotification("email", invalidNotification)
if err != nil {
    fmt.Printf("验证失败: %v\n", err)
}
```

## 五、设计要点

### 1. 接口设计要稳定

```go
// ✅ 好的接口：行为明确，职责单一
type Notifier interface {
    Name() string
    Send(ctx context.Context, notification *Notification) error
    Validate(notification *Notification) error
}

// ❌ 不好的接口：包含实现细节
type BadNotifier interface {
    Send(...) error
    GetConfig() map[string]string  // 暴露实现细节
    SetTimeout(time.Duration)      // 这是配置，不是策略行为
}
```

### 2. 策略的可替换性

```go
// ✅ 好的：依赖接口
func (s *NotificationService) SendNotification(
    ctx context.Context,
    strategyName string,
    notification *core.Notification,
) error {
    notifier, _ := s.registry.Get(strategyName)
    return notifier.Send(ctx, notification)  // 多态调用
}

// ❌ 不好的：依赖具体类型
func BadSend(notifier *strategies.EmailNotifier, ...) {
    // 这样就绑定死了，无法替换
}
```

### 3. 策略的隔离性

- 每个策略独立文件
- 策略之间无依赖
- 可以独立测试和维护

### 4. 运行时选择

```go
// 通过字符串名称动态选择策略
svc.SendNotification(ctx, "email", notification)  // 邮件策略
svc.SendNotification(ctx, "sms", notification)   // 短信策略
svc.SendNotification(ctx, "wechat", notification) // 微信策略
```

## 六、扩展新策略

要添加新的通知策略，只需：

### 步骤 1：实现接口

```go
// strategies/push_notifier.go
package strategies

import (
    "context"
    "esim/test/strategy/core"
)

type PushNotifier struct {
    appKey string
}

func NewPushNotifier(appKey string) *PushNotifier {
    return &PushNotifier{appKey: appKey}
}

func (n *PushNotifier) Name() string {
    return "push"
}

func (n *PushNotifier) Validate(notification *core.Notification) error {
    // 验证逻辑
    return nil
}

func (n *PushNotifier) Send(ctx context.Context, notification *core.Notification) error {
    // 发送逻辑
    return nil
}
```

### 步骤 2：注册策略

```go
registry := core.NewNotifierRegistry()
pushNotifier := strategies.NewPushNotifier("app_key")
registry.Register(pushNotifier)
```

### 步骤 3：使用策略

```go
svc.SendNotification(ctx, "push", notification)
```

**无需修改现有代码！**

## 七、优势总结

### 1. 开闭原则
- ✅ 对扩展开放：可以轻松添加新策略
- ✅ 对修改关闭：添加新策略无需修改现有代码

### 2. 依赖倒置
- ✅ 依赖抽象（接口），不依赖具体实现
- ✅ 调用方只依赖 `Notifier` 接口

### 3. 单一职责
- ✅ 每个策略只负责一种通知方式
- ✅ 服务层只负责调用策略

### 4. 可测试性
- ✅ 可以轻松 mock 策略进行单元测试
- ✅ 每个策略可以独立测试

### 5. 可维护性
- ✅ 策略之间解耦，易于维护
- ✅ 代码结构清晰，职责分明

## 八、运行和测试

### 运行示例

```bash
go run ./test/strategy/cmd/main.go
```

### 运行测试

```bash
go test -v ./test/strategy
```

### 运行基准测试

```bash
go test -bench=. ./test/strategy
```

## 九、总结

本实现展示了策略模式在 Go 语言中的典型应用：

1. **核心层**：定义接口和注册表，提供策略管理能力
2. **策略层**：实现具体策略，每个策略独立且可替换
3. **服务层**：调用策略，通过接口实现多态，不依赖具体实现

通过这种分层设计，我们实现了：
- ✅ 策略的可替换性
- ✅ 代码的可扩展性
- ✅ 良好的可维护性
- ✅ 清晰的职责划分

策略模式的核心在于**接口设计**，一个好的接口设计可以让整个系统更加灵活和可扩展。
