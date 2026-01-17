# Go 风格观察者模式实现

这是一个完全符合 Go 语言风格的观察者模式实现，充分利用了 Go 的并发特性（goroutine、channel、context）。

## 目录结构

```
example/observer/
├── core/                    # 核心层：接口定义和核心实现
│   ├── event.go            # 事件接口和基础事件
│   ├── observer.go          # 观察者接口和函数式实现
│   ├── eventbus.go          # 事件总线核心实现
│   ├── wrapper.go           # 观察者包装器（管理 goroutine 生命周期）
│   ├── config.go            # 配置定义
│   └── stats.go             # 统计信息
├── events/                  # 事件定义层：具体事件类型
│   ├── user_events.go       # 用户相关事件
│   └── config_events.go     # 配置相关事件
├── observers/               # 观察者实现层：具体观察者实现
│   ├── logger_observer.go   # 日志观察者
│   ├── metrics_observer.go  # 指标观察者
│   └── notification_observer.go  # 通知观察者
├── example.go              # 使用示例
├── cmd/
│   └── main.go             # 可运行的主程序
└── README.md               # 本文档
```

## 设计理念

### 1. Channel 优先原则
- 使用 channel 传递事件，而非直接函数调用
- 每个观察者拥有独立的 channel，实现完全隔离
- 充分利用 Go 的并发原语

### 2. Context 生命周期管理
- 使用 `context.Context` 控制观察者生命周期
- 支持优雅关闭和超时控制
- 与项目现有模式保持一致

### 3. 非阻塞异步设计
- 事件发布完全不阻塞发布者
- 观察者处理失败不影响其他观察者
- 通过缓冲 channel 处理背压

## 核心组件

### EventBus（事件总线）
负责管理事件订阅和分发，是观察者模式的核心。

**特性：**
- 类型安全的事件系统
- 并发安全的订阅管理
- 非阻塞的事件发布
- 优雅关闭支持
- 统计信息收集

### Observer（观察者）
观察者接口，支持两种实现方式：

1. **接口实现**：实现 `Observer` 接口
2. **函数式**：使用 `ObserverFunc` 或 `NamedObserverFunc`（更 Go 风格）

### Event（事件）
所有事件必须实现 `Event` 接口，提供类型和时间戳信息。

## 使用示例

### 基础使用

```go
// 1. 创建事件总线
ctx := context.Background()
bus := core.NewEventBus(ctx,
    core.WithBufferSize(100),
    core.WithTimeout(5*time.Second),
)

// 2. 创建观察者
loggerObserver := observers.NewLoggerObserver("logger-1")

// 3. 订阅事件
bus.Subscribe(events.EventTypeUserLogin, loggerObserver)

// 4. 发布事件
loginEvent := events.NewUserLoginEvent("user-123", "192.168.1.1", "Mozilla/5.0")
bus.Publish(loginEvent)

// 5. 优雅关闭
shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
bus.Shutdown(shutdownCtx)
```

### 函数式观察者（推荐）

```go
bus.Subscribe(events.EventTypeUserRegister, core.ObserverFunc(func(ctx context.Context, event core.Event) error {
    if regEvent, ok := event.(*events.UserRegisterEvent); ok {
        log.Printf("欢迎新用户 %s", regEvent.Username)
    }
    return nil
}))
```

### 多个观察者

```go
loggerObserver := observers.NewLoggerObserver("logger")
auditObserver := observers.NewAuditObserver("audit")
metricsObserver := observers.NewMetricsObserver("metrics")

// 订阅同一个事件
bus.Subscribe(events.EventTypeUserLogin, loggerObserver)
bus.Subscribe(events.EventTypeUserLogin, auditObserver)
bus.Subscribe(events.EventTypeUserLogin, metricsObserver)

// 发布事件，所有观察者都会收到
bus.Publish(events.NewUserLoginEvent("user-123", "192.168.1.1", "Chrome"))
```

## 运行示例

### 运行主程序

```bash
cd example/observer
go run cmd/main.go
```

主程序会：
1. 创建事件总线并注册多个观察者
2. 模拟发布各种事件
3. 定期打印统计信息
4. 等待中断信号（Ctrl+C）后优雅关闭

### 运行示例代码

```bash
cd example/observer
go run example.go
```

## 设计优势

### 1. 类型安全
通过 `EventType` 和 `Event` 接口实现类型安全的事件系统。

### 2. 并发安全
- 使用 `sync.RWMutex` 保护订阅关系
- 使用原子操作更新统计信息
- 每个观察者独立的 goroutine 和 channel

### 3. 非阻塞
- 发布者不等待观察者处理
- 通过 channel 实现异步通信
- 支持缓冲 channel 处理背压

### 4. 错误隔离
- 一个观察者失败不影响其他观察者
- 使用 `recover` 防止 panic 传播
- 支持超时控制

### 5. 优雅关闭
- 支持超时控制
- 等待所有观察者处理完当前事件
- 清空剩余事件（可选）

### 6. Go 风格
- 充分利用 channel 和 goroutine
- 函数式观察者更灵活
- 符合 Go 的并发编程习惯

### 7. 可扩展
- 易于添加新的事件类型
- 易于添加新的观察者
- 支持配置和统计

## 架构说明

### 分层设计

1. **核心层 (core/)**：定义接口和核心实现
   - `Event` 接口：所有事件的基类
   - `Observer` 接口：观察者接口
   - `EventBus`：事件总线实现
   - `observerWrapper`：观察者包装器，管理 goroutine 生命周期

2. **事件层 (events/)**：定义具体事件类型
   - 用户事件：登录、登出、注册、更新
   - 配置事件：配置变更

3. **观察者层 (observers/)**：实现具体观察者
   - 日志观察者：记录所有事件
   - 审计观察者：记录重要事件
   - 指标观察者：收集统计信息
   - 通知观察者：发送通知

### 依赖关系

```
observers/ → core/ (实现 Observer 接口)
events/ → core/ (实现 Event 接口)
cmd/ → core/, events/, observers/ (使用所有层)
```

## 并发模型

### 观察者 goroutine 模型

每个观察者运行在独立的 goroutine 中：

```
EventBus.Publish(event)
    ↓
为每个观察者发送到其 eventCh（非阻塞）
    ↓
观察者 goroutine 从 eventCh 接收事件
    ↓
调用 observer.Handle() 处理事件
```

### 优雅关闭流程

```
1. 标记为关闭，停止接受新订阅
2. 取消所有观察者的 context
3. 等待所有观察者处理完当前事件
4. 清空剩余事件（可选）
5. 关闭完成
```

## 性能考虑

- **缓冲 channel**：通过配置 `BufferSize` 处理背压
- **非阻塞发布**：发布者不等待观察者处理
- **goroutine 复用**：每个观察者一个长期运行的 goroutine
- **统计信息**：使用原子操作，性能影响最小

## 扩展建议

1. **事件过滤**：可以为观察者添加过滤条件
2. **优先级**：可以为观察者添加优先级支持
3. **重试机制**：可以为失败的观察者添加重试
4. **事件持久化**：可以将事件持久化到数据库
5. **分布式支持**：可以扩展为分布式事件总线

## 注意事项

1. **内存管理**：观察者 channel 如果满了，事件可能会丢失（根据配置）
2. **goroutine 数量**：每个观察者一个 goroutine，注意控制数量
3. **超时设置**：合理设置观察者处理超时时间
4. **优雅关闭**：确保在应用退出前调用 `Shutdown()`

## 与项目其他模式的对比

### vs Producer-Consumer 模式

- **观察者模式**：一对多，事件发布给所有订阅者
- **Producer-Consumer**：一对一或多对一，任务分发给消费者

### vs Strategy 模式

- **观察者模式**：关注事件通知，解耦发布者和订阅者
- **Strategy 模式**：关注算法选择，运行时选择策略

## 总结

这个实现充分体现了 Go 语言的并发特性，使用 channel 和 goroutine 实现了高效、安全、易用的观察者模式。代码结构清晰，易于理解和扩展，适合在实际项目中使用。
