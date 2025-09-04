# Go Context 使用手册

## 目录
- [概述](#概述)
- [Context 接口](#context-接口)
- [创建 Context](#创建-context)
- [Context 的父子关系](#context-的父子关系)
- [常用方法详解](#常用方法详解)
- [最佳实践](#最佳实践)
- [常见陷阱](#常见陷阱)
- [实际应用示例](#实际应用示例)

## 概述

Context 是 Go 语言中用于管理请求生命周期、取消信号和跨 API 边界传递请求范围值的重要机制。它主要用于：

- **取消操作**：传播取消信号
- **超时控制**：设置操作超时时间
- **值传递**：在请求范围内传递键值对
- **请求追踪**：跟踪请求的完整生命周期

## Context 接口

```go
type Context interface {
    // 返回 context 的截止时间
    Deadline() (deadline time.Time, ok bool)
    
    // 返回一个 channel，当 context 被取消或超时时会关闭
    Done() <-chan struct{}
    
    // 返回 context 被取消的原因
    Err() error
    
    // 返回与 key 关联的值
    Value(key interface{}) interface{}
}
```

## 创建 Context

### 1. 根 Context

```go
// 创建根 context
ctx := context.Background()

// 或者使用 TODO（当不确定使用哪个 context 时）
ctx := context.TODO()
```

### 2. 带取消的 Context

```go
// 创建可取消的 context
ctx, cancel := context.WithCancel(context.Background())
defer cancel() // 确保资源被释放

// 手动取消
cancel()
```

### 3. 带超时的 Context

```go
// 创建带超时的 context
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// 或者使用 WithDeadline
deadline := time.Now().Add(5 * time.Second)
ctx, cancel := context.WithDeadline(context.Background(), deadline)
defer cancel()
```

### 4. 带值的 Context

```go
// 创建带值的 context
ctx := context.WithValue(context.Background(), "userID", "12345")
ctx = context.WithValue(ctx, "requestID", "req-001")

// 获取值
userID := ctx.Value("userID")
requestID := ctx.Value("requestID")
```

## Context 的父子关系

### 核心原则
- **单向传播**：取消信号只能从父 context 向子 context 传播
- **级联取消**：父 context 被取消时，所有子 context 都会被取消
- **独立取消**：子 context 被取消不会影响父 context

### 关系图
```
根 Context
    ↓
父 Context (取消)
    ↓ ↓
子 Context1  子 Context2 (取消)
    ↓
孙 Context
```

### 示例代码

```go
func demonstrateContextHierarchy() {
    // 创建根 context
    rootCtx := context.Background()
    
    // 创建父 context
    parentCtx, cancelParent := context.WithCancel(rootCtx)
    
    // 创建子 context
    childCtx, cancelChild := context.WithCancel(parentCtx)
    
    // 取消父 context
    cancelParent()
    
    // 检查状态
    select {
    case <-parentCtx.Done():
        fmt.Println("父 context 被取消")
    default:
        fmt.Println("父 context 仍然活跃")
    }
    
    select {
    case <-childCtx.Done():
        fmt.Println("子 context 被取消") // 会输出这个
    default:
        fmt.Println("子 context 仍然活跃")
    }
}
```

## 常用方法详解

### 1. Done() 方法

```go
func worker(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            fmt.Printf("工作被取消: %v\n", ctx.Err())
            return
        default:
            // 执行工作
            fmt.Println("正在工作...")
            time.Sleep(1 * time.Second)
        }
    }
}
```

### 2. Err() 方法

```go
func checkContextStatus(ctx context.Context) {
    select {
    case <-ctx.Done():
        switch ctx.Err() {
        case context.Canceled:
            fmt.Println("Context 被手动取消")
        case context.DeadlineExceeded:
            fmt.Println("Context 超时")
        default:
            fmt.Printf("Context 错误: %v\n", ctx.Err())
        }
    default:
        fmt.Println("Context 仍然活跃")
    }
}
```

### 3. Deadline() 方法

```go
func checkDeadline(ctx context.Context) {
    if deadline, ok := ctx.Deadline(); ok {
        fmt.Printf("Context 将在 %v 过期\n", deadline)
        fmt.Printf("剩余时间: %v\n", time.Until(deadline))
    } else {
        fmt.Println("Context 没有设置截止时间")
    }
}
```

### 4. Value() 方法

```go
// 定义键类型
type key string

const (
    UserIDKey    key = "userID"
    RequestIDKey key = "requestID"
)

func setAndGetValues() {
    ctx := context.Background()
    ctx = context.WithValue(ctx, UserIDKey, "12345")
    ctx = context.WithValue(ctx, RequestIDKey, "req-001")
    
    // 获取值
    userID := ctx.Value(UserIDKey)
    requestID := ctx.Value(RequestIDKey)
    
    fmt.Printf("用户ID: %v\n", userID)
    fmt.Printf("请求ID: %v\n", requestID)
}
```

## 最佳实践

### 1. 函数签名

```go
// ✅ 正确：将 context 作为第一个参数
func ProcessRequest(ctx context.Context, data []byte) error {
    // 实现
}

// ❌ 错误：将 context 放在其他位置
func ProcessRequest(data []byte, ctx context.Context) error {
    // 实现
}
```

### 2. 超时设置

```go
// ✅ 正确：为长时间运行的操作设置超时
func LongRunningOperation(ctx context.Context) error {
    // 设置合理的超时时间
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // 执行操作
    return doWork(ctx)
}
```

### 3. 资源清理

```go
// ✅ 正确：同步使用 context 时确保 cancel 函数被调用
func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
    return context.WithTimeout(ctx, timeout)
}

func main() {
    ctx, cancel := WithTimeout(context.Background(), 5*time.Second)
    defer cancel() // 确保资源被释放
    
    // 同步使用 ctx
    doWork(ctx)
}

// ✅ 正确：协程使用 context 时的资源清理
func goroutineWithContext() {
    ctx, cancel := context.WithCancel(context.Background())
    
    // 使用 WaitGroup 等待协程完成
    var wg sync.WaitGroup
    
    // 启动协程
    wg.Add(1)
    go func() {
        defer wg.Done()
        doWork(ctx)
    }()
    
    // 等待协程完成
    wg.Wait()
    
    // 现在可以安全地取消 context
    cancel()
}
```

### 4. 值传递

```go
// ✅ 正确：使用自定义类型作为键
type key string

const UserIDKey key = "userID"

func SetUserID(ctx context.Context, userID string) context.Context {
    return context.WithValue(ctx, UserIDKey, userID)
}

func GetUserID(ctx context.Context) (string, bool) {
    userID, ok := ctx.Value(UserIDKey).(string)
    return userID, ok
}
```

## 常见陷阱

### 1. 忘记调用 cancel

```go
// ❌ 错误：忘记调用 cancel
func badExample() {
    ctx, cancel := context.WithCancel(context.Background())
    // 忘记调用 cancel()
    // 这会导致资源泄漏
}

// ✅ 正确：同步使用 context 时使用 defer
func goodExample() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    // 同步使用 ctx
    doWork(ctx) // 假设这是同步调用
}

// ⚠️ 注意：如果 context 要传递给协程，需要谨慎使用 defer
func goroutineExample() {
    ctx, cancel := context.WithCancel(context.Background())
    // 不要立即 defer cancel()，因为协程可能还在运行
    
    // 启动协程
    go doWork(ctx)
    
    // 等待协程完成或设置超时
    select {
    case <-time.After(5 * time.Second):
        cancel() // 超时后取消
    case <-done: // 假设有 done channel
        // 协程正常完成
    }
    
    // 确保最终调用 cancel
    defer cancel()
}
```

### 2. 在结构体中存储 context

```go
// ❌ 错误：在结构体中存储 context
type BadService struct {
    ctx context.Context
}

// ✅ 正确：将 context 作为参数传递
type GoodService struct {
    // 其他字段
}

func (s *GoodService) Process(ctx context.Context) error {
    // 实现
}
```

### 3. 使用内置类型作为键

```go
// ❌ 错误：使用内置类型作为键
ctx := context.WithValue(context.Background(), "userID", "12345")

// ✅ 正确：使用自定义类型作为键
type key string
const UserIDKey key = "userID"
ctx := context.WithValue(context.Background(), UserIDKey, "12345")
```

## 实际应用示例

### 1. HTTP 请求处理

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    // 从请求创建 context
    ctx := r.Context()
    
    // 设置超时
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    
    // 添加请求信息
    ctx = context.WithValue(ctx, "requestID", generateRequestID())
    ctx = context.WithValue(ctx, "userID", getUserID(r))
    
    // 处理请求
    result, err := processRequest(ctx, r.Body)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Write(result)
}
```

### 2. 数据库操作

```go
func (db *Database) QueryWithTimeout(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    // 设置数据库查询超时
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    // 执行查询
    return db.conn.QueryContext(ctx, query, args...)
}
```

### 3. 并发任务管理

```go
func processConcurrentTasks(ctx context.Context, tasks []Task) error {
    // 创建带超时的 context
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // 创建结果通道
    results := make(chan error, len(tasks))
    
    // 启动 goroutine 处理任务
    for _, task := range tasks {
        go func(t Task) {
            select {
            case <-ctx.Done():
                results <- ctx.Err()
            default:
                results <- t.Process(ctx)
            }
        }(task)
    }
    
    // 收集结果
    for i := 0; i < len(tasks); i++ {
        if err := <-results; err != nil {
            return err
        }
    }
    
    return nil
}
```

### 4. 中间件模式

```go
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 创建带请求信息的 context
        ctx := context.WithValue(r.Context(), "requestID", generateRequestID())
        ctx = context.WithValue(ctx, "startTime", time.Now())
        
        // 创建新的请求
        req := r.WithContext(ctx)
        
        // 调用下一个处理器
        next.ServeHTTP(w, req)
        
        // 记录日志
        logRequest(ctx)
    })
}
```

### 5. 优雅关闭

```go
func gracefulShutdown(server *http.Server) {
    // 创建 context 用于优雅关闭
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // 关闭服务器
    if err := server.Shutdown(ctx); err != nil {
        log.Printf("服务器关闭失败: %v", err)
    } else {
        log.Println("服务器已优雅关闭")
    }
}
```

## 总结

Context 是 Go 语言中管理请求生命周期的重要工具。正确使用 Context 可以：

1. **提高代码的可维护性**：统一的取消和超时机制
2. **避免资源泄漏**：及时取消不需要的操作
3. **增强可观测性**：通过值传递跟踪请求信息
4. **简化并发编程**：统一的取消信号传播

记住这些关键点：
- 总是将 context 作为函数的第一个参数
- 同步使用 context 时使用 defer cancel() 确保资源被释放
- 协程使用 context 时，需要等待协程完成后再调用 cancel()
- 使用自定义类型作为 context 的键
- 为长时间运行的操作设置合理的超时时间
- 避免在结构体中存储 context
- 使用 WaitGroup 或 channel 来协调协程的完成时机

通过遵循这些最佳实践，您可以编写出更加健壮和可维护的 Go 代码。
