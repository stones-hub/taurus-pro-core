# 企业级协程通信与管理标准总结

## 1. 分层停止机制标准

### 1.1 三层停止架构

```go
// 第一层：系统级停止（所有协程）
ctx, cancel := context.WithCancel(context.Background())
wp.cancel()  // 触发 ctx.Done()

// 第二层：功能级停止（特定功能模块）
close(wp.shutdownCh)        // 工作池级停止
close(wp.monitorCh)         // 监控模块停止
close(wp.healthCheckStopCh) // 健康检查停止

// 第三层：通道关闭（依赖型协程）
close(wp.taskCh)    // 任务通道关闭
close(wp.resultCh)  // 结果通道关闭
close(wp.errorCh)   // 错误通道关闭
```

### 1.2 关闭顺序标准

```go
func (wp *WorkerPool) Stop() {
    wp.shutdownOnce.Do(func() {
        // 1. 标记为已关闭
        wp.mu.Lock()
        wp.closed = true
        wp.mu.Unlock()

        // 2. 发送业务级停止信号
        close(wp.shutdownCh)

        // 3. 停止辅助功能
        if wp.monitorTicker != nil {
            wp.monitorTicker.Stop()
        }
        close(wp.monitorCh)
        close(wp.healthCheckStopCh)

        // 4. 取消上下文（系统级保险）
        wp.cancel()

        // 5. 等待协程完成
        wp.wg.Wait()

        // 6. 关闭业务通道
        wp.safeCloseChannels()
    })
}
```

**设计原则**：
- **业务优先**：业务级停止信号优先于系统级停止信号
- **渐进关闭**：从具体到抽象，确保每个阶段都有明确效果
- **双重保险**：即使业务级停止失效，系统级停止也能确保停止

## 2. 协程分类管理标准

### 2.1 主动型协程（需要立即停止）

```go
// 核心业务协程 - 双重停止信号
func (wp *WorkerPool) worker(id int) {
    defer func() {
        if r := recover(); r != nil {
            select {
            case wp.errorCh <- fmt.Errorf("worker %d panic: %v", id, r):
            default:
            }
        }
        atomic.AddInt64(&wp.activeWorkers, -1)
        wp.wg.Done()
    }()

    atomic.AddInt64(&wp.activeWorkers, 1)

    for {
        select {
        case <-wp.ctx.Done():        // 系统级停止
            return
        case <-wp.shutdownCh:        // 业务级停止
            return
        case task, ok := <-wp.taskCh:
            if !ok {
                return
            }
            wp.processTask(id, task)
        }
    }
}
```

### 2.2 被动型协程（依赖数据流）

```go
// 辅助处理协程 - 依赖通道关闭
func (wp *WorkerPool) resultHandler() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("结果处理协程发生panic: %v", r)
        }
        wp.wg.Done()
    }()

    for {
        select {
        case <-wp.ctx.Done():        // 系统级停止
            return
        case result, ok := <-wp.resultCh:
            if !ok {                 // 通道关闭时停止
                return
            }
            processResult(result)
        }
    }
}
```

### 2.3 定时型协程（需要精确控制）

```go
// 监控协程 - 专属停止信号
func (wp *WorkerPool) monitor() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("监控协程发生panic: %v", r)
        }
        wp.wg.Done()
    }()

    wp.monitorTicker = time.NewTicker(5 * time.Second)
    defer wp.monitorTicker.Stop()

    for {
        select {
        case <-wp.ctx.Done():        // 系统级停止
            return
        case <-wp.monitorCh:         // 专属停止信号
            return
        case <-wp.monitorTicker.C:
            wp.updateMetrics()
            wp.checkMemoryLeak()
        }
    }
}
```

## 3. 中间通道使用标准

### 3.1 使用中间通道的场景

```go
// 复杂、耗时、需要超时控制的操作
func (wp *WorkerPool) processTask(workerID int, task Task) {
    startTime := time.Now()
    
    // 创建带超时的上下文
    ctx, cancel := context.WithTimeout(wp.ctx, wp.timeout)
    defer cancel()

    // 使用中间通道
    done := make(chan Result, 1)

    // 启动任务处理协程
    go func() {
        defer func() {
            if r := recover(); r != nil {
                select {
                case done <- Result{
                    TaskID: task.ID,
                    Data:   task.Data,
                    Error:  fmt.Errorf("任务处理panic: %v", r),
                }:
                default:
                }
            }
        }()

        // 模拟复杂操作（网络请求、数据库查询等）
        time.Sleep(time.Duration(task.Priority*100) * time.Millisecond)
        
        var err error
        if task.ID%10 == 0 {
            err = errors.New("模拟任务处理错误")
        }

        select {
        case done <- Result{
            TaskID: task.ID,
            Data:   task.Data + "_processed",
            Error:  err,
        }:
        default:
        }
    }()

    // 等待结果或超时
    select {
    case <-ctx.Done():
        atomic.AddInt64(&wp.timeoutTasks, 1)
        select {
        case wp.resultCh <- Result{
            TaskID: task.ID,
            Data:   task.Data,
            Error:  errors.New("任务处理超时"),
        }:
        case <-wp.ctx.Done():
            return
        }
    case result := <-done:
        processingTime := time.Since(startTime)
        wp.recordProcessingTime(processingTime)
        
        if result.Error != nil {
            atomic.AddInt64(&wp.failedTasks, 1)
        } else {
            atomic.AddInt64(&wp.completedTasks, 1)
        }

        select {
        case wp.resultCh <- result:
        case <-wp.ctx.Done():
            return
        }
    }
}
```

### 3.2 不使用中间通道的场景

```go
// 简单、快速、同步的操作
func (wp *WorkerPool) healthChecker() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("健康检查协程发生panic: %v", r)
        }
        wp.wg.Done()
    }()

    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-wp.ctx.Done():
            return
        case <-wp.healthCheckStopCh:
            return
        case <-ticker.C:
            // 同步调用，立即完成
            health := wp.HealthCheck()
            select {
            case wp.healthCheckCh <- health:
            default:
                // 通道已满，忽略
            }
        }
    }
}
```

### 3.3 判断标准

| 特征 | 使用中间通道 | 不使用中间通道 |
|------|-------------|---------------|
| **执行时间** | >100ms | <100ms |
| **复杂度** | 涉及外部调用 | 纯内存操作 |
| **错误风险** | 可能panic或超时 | 低风险 |
| **控制需求** | 需要精确控制 | 简单处理 |

## 4. 通道设计标准

### 4.1 通道类型选择

```go
type WorkerPool struct {
    // 业务通道 - 缓冲通道，避免阻塞
    taskCh   chan Task   // 容量：workerCount*2
    resultCh chan Result // 容量：workerCount*2
    errorCh  chan error  // 容量：workerCount

    // 控制通道 - 无缓冲通道，立即响应
    shutdownCh        chan struct{}
    monitorCh         chan struct{}
    healthCheckStopCh chan struct{}
    
    // 数据通道 - 缓冲通道，避免阻塞
    healthCheckCh chan HealthStatus // 容量：1
}
```

### 4.2 通道关闭标准

```go
// 使用sync.Once确保只关闭一次
func (wp *WorkerPool) safeCloseChannels() {
    var taskCloseOnce, resultCloseOnce, errorCloseOnce sync.Once

    taskCloseOnce.Do(func() {
        close(wp.taskCh)
    })

    resultCloseOnce.Do(func() {
        close(wp.resultCh)
    })

    errorCloseOnce.Do(func() {
        close(wp.errorCh)
    })
}
```

## 5. 错误处理标准

### 5.1 Panic恢复机制

```go
// 所有协程都必须有panic恢复
func (wp *WorkerPool) worker(id int) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("工作协程 %d 发生panic: %v", id, r)
            // 发送错误到错误通道
            select {
            case wp.errorCh <- fmt.Errorf("工作协程 %d panic: %v", id, r):
            default:
                // 错误通道已满，忽略
            }
        }
        atomic.AddInt64(&wp.activeWorkers, -1)
        wp.wg.Done()
    }()
    
    // 业务逻辑
}
```

### 5.2 错误传播机制

```go
// 错误处理协程
func (wp *WorkerPool) errorHandler() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("错误处理协程发生panic: %v", r)
        }
        wp.wg.Done()
    }()

    for {
        select {
        case <-wp.ctx.Done():
            return
        case err, ok := <-wp.errorCh:
            if !ok {
                return
            }
            log.Printf("处理错误: %v", err)
            // 可以添加错误上报、告警等逻辑
        }
    }
}
```

## 6. 资源管理标准

### 6.1 生命周期管理

```go
// 启动时
func (wp *WorkerPool) Start() {
    // 启动工作协程
    for i := 0; i < wp.workerCount; i++ {
        wp.wg.Add(1)
        go wp.worker(i)
    }

    // 启动辅助协程
    wp.wg.Add(1)
    go wp.errorHandler()

    wp.wg.Add(1)
    go wp.resultHandler()

    wp.wg.Add(1)
    go wp.monitor()

    wp.wg.Add(1)
    go wp.healthChecker()

    log.Printf("工作池已启动，工作协程数: %d", wp.workerCount)
}

// 停止时
func (wp *WorkerPool) Stop() {
    wp.shutdownOnce.Do(func() {
        log.Println("开始停止工作池...")

        // 标记为已关闭
        wp.mu.Lock()
        wp.closed = true
        wp.mu.Unlock()

        // 发送关闭信号
        close(wp.shutdownCh)

        // 停止监控
        if wp.monitorTicker != nil {
            wp.monitorTicker.Stop()
        }
        close(wp.monitorCh)

        // 停止健康检查
        close(wp.healthCheckStopCh)

        // 取消上下文
        wp.cancel()

        // 等待所有协程完成，带超时机制
        timeout := 10 * time.Second
        checkInterval := 100 * time.Millisecond
        elapsed := time.Duration(0)

        for elapsed < timeout {
            done := make(chan struct{})
            go func() {
                wp.wg.Wait()
                close(done)
            }()

            select {
            case <-done:
                log.Println("工作池已优雅停止")
                wp.safeCloseChannels()
                return
            case <-time.After(checkInterval):
                elapsed += checkInterval
                log.Printf("等待协程停止中... (已等待 %v)", elapsed)
            }
        }

        log.Printf("工作池停止超时 (等待了 %v)，强制关闭", timeout)
        wp.safeCloseChannels()
    })
}
```

### 6.2 超时控制

```go
// 任务级超时
func (wp *WorkerPool) processTask(workerID int, task Task) {
    ctx, cancel := context.WithTimeout(wp.ctx, wp.timeout)
    defer cancel()
    
    // 使用中间通道处理超时
    done := make(chan Result, 1)
    
    go func() {
        // 任务处理逻辑
        result := processTask(task)
        done <- result
    }()
    
    select {
    case result := <-done:
        // 正常完成
    case <-ctx.Done():
        // 超时处理
    }
}

// 系统级超时
func (wp *WorkerPool) Stop() {
    timeout := 10 * time.Second
    // 等待协程停止的超时机制
}
```

## 7. 监控和健康检查标准

### 7.1 指标收集

```go
type SystemMetrics struct {
    // 内存指标
    MemoryUsage     uint64
    MemoryAlloc     uint64
    MemorySys       uint64
    MemoryHeapAlloc uint64
    MemoryHeapSys   uint64
    MemoryHeapIdle  uint64
    MemoryHeapInuse uint64

    // 协程指标
    GoroutineCount int

    // 工作池指标
    TaskQueueSize   int
    ResultQueueSize int
    ErrorQueueSize  int
    ActiveWorkers   int64
    TotalTasks      int64
    CompletedTasks  int64
    FailedTasks     int64
    TimeoutTasks    int64

    // 性能指标
    AverageProcessingTime time.Duration
    LastProcessingTime    time.Duration

    // 时间指标
    Uptime time.Duration
}

func (wp *WorkerPool) updateMetrics() {
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)

    wp.mu.Lock()
    defer wp.mu.Unlock()

    wp.metrics.MemoryUsage = memStats.Alloc
    wp.metrics.MemoryAlloc = memStats.Alloc
    wp.metrics.MemorySys = memStats.Sys
    wp.metrics.MemoryHeapAlloc = memStats.HeapAlloc
    wp.metrics.MemoryHeapSys = memStats.HeapSys
    wp.metrics.MemoryHeapIdle = memStats.HeapIdle
    wp.metrics.MemoryHeapInuse = memStats.HeapInuse
    wp.metrics.GoroutineCount = runtime.NumGoroutine()
    wp.metrics.TaskQueueSize = len(wp.taskCh)
    wp.metrics.ResultQueueSize = len(wp.resultCh)
    wp.metrics.ErrorQueueSize = len(wp.errorCh)
    wp.metrics.ActiveWorkers = atomic.LoadInt64(&wp.activeWorkers)
    wp.metrics.TotalTasks = atomic.LoadInt64(&wp.totalTasks)
    wp.metrics.CompletedTasks = atomic.LoadInt64(&wp.completedTasks)
    wp.metrics.FailedTasks = atomic.LoadInt64(&wp.failedTasks)
    wp.metrics.TimeoutTasks = atomic.LoadInt64(&wp.timeoutTasks)
    wp.metrics.Uptime = time.Since(wp.startTime)

    // 计算平均处理时间
    wp.processingTimesMu.RLock()
    if len(wp.processingTimes) > 0 {
        var total time.Duration
        for _, t := range wp.processingTimes {
            total += t
        }
        wp.metrics.AverageProcessingTime = total / time.Duration(len(wp.processingTimes))
        wp.metrics.LastProcessingTime = wp.processingTimes[len(wp.processingTimes)-1]
    }
    wp.processingTimesMu.RUnlock()
}
```

### 7.2 健康检查

```go
type HealthStatus struct {
    IsHealthy     bool
    ErrorMessage  string
    LastCheckTime time.Time
    Metrics       *SystemMetrics
}

func (wp *WorkerPool) HealthCheck() HealthStatus {
    wp.updateMetrics()

    health := HealthStatus{
        IsHealthy:     true,
        LastCheckTime: time.Now(),
        Metrics:       wp.metrics,
    }

    // 检查内存使用
    if wp.metrics.MemoryHeapInuse > 200*1024*1024 { // 200MB
        health.IsHealthy = false
        health.ErrorMessage = fmt.Sprintf("内存使用过高: %d MB", wp.metrics.MemoryHeapInuse/1024/1024)
    }

    // 检查协程泄漏
    if wp.metrics.GoroutineCount > wp.workerCount*5 {
        health.IsHealthy = false
        health.ErrorMessage = fmt.Sprintf("协程数量异常: %d", wp.metrics.GoroutineCount)
    }

    // 检查任务积压
    if wp.metrics.TaskQueueSize > wp.workerCount*5 {
        health.IsHealthy = false
        health.ErrorMessage = fmt.Sprintf("任务队列严重积压: %d", wp.metrics.TaskQueueSize)
    }

    // 检查错误率
    totalProcessed := wp.metrics.CompletedTasks + wp.metrics.FailedTasks + wp.metrics.TimeoutTasks
    if totalProcessed > 0 {
        errorRate := float64(wp.metrics.FailedTasks+wp.metrics.TimeoutTasks) / float64(totalProcessed)
        if errorRate > 0.1 { // 错误率超过10%
            health.IsHealthy = false
            health.ErrorMessage = fmt.Sprintf("错误率过高: %.2f%%", errorRate*100)
        }
    }

    return health
}
```

## 8. 代码组织标准

### 8.1 结构体设计

```go
type WorkerPool struct {
    // 核心字段
    ctx          context.Context
    cancel       context.CancelFunc
    wg           sync.WaitGroup

    // 业务通道
    taskCh       chan Task
    resultCh     chan Result
    errorCh      chan error

    // 控制通道
    shutdownCh        chan struct{}
    monitorCh         chan struct{}
    healthCheckStopCh chan struct{}

    // 数据通道
    healthCheckCh chan HealthStatus

    // 配置参数
    workerCount int
    timeout     time.Duration

    // 状态管理
    closed       bool
    mu           sync.RWMutex
    shutdownOnce sync.Once

    // 监控指标
    metrics           *SystemMetrics
    startTime         time.Time
    activeWorkers     int64
    totalTasks        int64
    completedTasks    int64
    failedTasks       int64
    timeoutTasks      int64
    processingTimes   []time.Duration
    processingTimesMu sync.RWMutex

    // 监控相关
    monitorTicker *time.Ticker
}
```

### 8.2 方法组织

```go
// 生命周期管理
func NewWorkerPool(workerCount int, timeout time.Duration) *WorkerPool
func (wp *WorkerPool) Start()
func (wp *WorkerPool) Stop()

// 业务操作
func (wp *WorkerPool) Submit(task Task) error
func (wp *WorkerPool) worker(id int)
func (wp *WorkerPool) processTask(id int, task Task)

// 辅助协程
func (wp *WorkerPool) errorHandler()
func (wp *WorkerPool) resultHandler()
func (wp *WorkerPool) monitor()
func (wp *WorkerPool) healthChecker()

// 监控和健康检查
func (wp *WorkerPool) HealthCheck() HealthStatus
func (wp *WorkerPool) GetHealthStatus() HealthStatus
func (wp *WorkerPool) GetMetrics() *SystemMetrics
func (wp *WorkerPool) PrintMetrics()

// 工具方法
func (wp *WorkerPool) updateMetrics()
func (wp *WorkerPool) checkMemoryLeak()
func (wp *WorkerPool) recordProcessingTime(duration time.Duration)
func (wp *WorkerPool) safeCloseChannels()
```

## 9. 核心设计原则

### 9.1 分层控制原则
- **系统级** → **功能级** → **通道级**
- 每层都有明确的职责和响应机制
- 确保关闭的完整性和可靠性

### 9.2 按需设计原则
- 根据协程特性选择合适的通信方式
- 避免过度设计，追求简洁高效
- 复杂操作使用中间通道，简单操作直接通信

### 9.3 优雅关闭原则
- 确保数据完整性
- 避免资源泄漏
- 提供快速响应和超时机制

### 9.4 监控可观测原则
- 实时监控系统状态
- 及时发现和处理问题
- 提供健康检查机制和指标收集

### 9.5 错误处理原则
- 优雅处理panic
- 错误传播和记录
- 系统自愈能力

### 9.6 资源管理原则
- 明确的生命周期管理
- 超时控制和重试机制
- 资源清理和释放

这套标准确保了协程系统的**安全性、可靠性、可维护性和可观测性**，是企业级Go应用的基础架构标准。