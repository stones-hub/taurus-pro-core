package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Task 任务结构体
type Task struct {
	ID       int
	Data     string
	Priority int
}

// Result 结果结构体
type Result struct {
	TaskID int
	Data   string
	Error  error
}

// HealthStatus 健康状态
type HealthStatus struct {
	IsHealthy     bool
	ErrorMessage  string
	LastCheckTime time.Time
	Metrics       *SystemMetrics
}

// SystemMetrics 系统指标
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

// WorkerPool 工作池结构体
type WorkerPool struct {
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	taskCh       chan Task
	resultCh     chan Result
	workerCount  int
	timeout      time.Duration
	errorCh      chan error
	shutdownCh   chan struct{}
	shutdownOnce sync.Once
	closed       bool
	mu           sync.RWMutex

	// 内存泄漏防护和健康检查相关字段
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
	monitorCh         chan struct{}
	monitorTicker     *time.Ticker
	healthCheckCh     chan HealthStatus
	healthCheckStopCh chan struct{} // 健康检查专属停止信号
}

// NewWorkerPool 创建工作池
func NewWorkerPool(workerCount int, timeout time.Duration) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		ctx:               ctx,
		cancel:            cancel,
		taskCh:            make(chan Task, workerCount*2),
		resultCh:          make(chan Result, workerCount*2),
		workerCount:       workerCount,
		timeout:           timeout,
		errorCh:           make(chan error, workerCount),
		shutdownCh:        make(chan struct{}),
		metrics:           &SystemMetrics{},
		startTime:         time.Now(),
		monitorCh:         make(chan struct{}),
		healthCheckCh:     make(chan HealthStatus, 1),
		healthCheckStopCh: make(chan struct{}),
	}
}

// Start 启动工作池
func (wp *WorkerPool) Start() {
	log.Printf("启动工作池，工作协程数量: %d", wp.workerCount)

	// 启动工作协程
	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i + 1)
	}

	// 启动错误处理协程
	wp.wg.Add(1)
	go wp.errorHandler()

	// 启动结果处理协程
	wp.wg.Add(1)
	go wp.resultHandler()

	// 启动监控协程
	wp.wg.Add(1)
	go wp.monitor()

	// 启动健康检查协程
	wp.wg.Add(1)
	go wp.healthChecker()
}

// Stop 停止工作池 - 完全修复版本
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

		// 等待所有协程完成，带超时和重试机制
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

// safeCloseChannels 安全关闭通道 - 完全修复版本
func (wp *WorkerPool) safeCloseChannels() {
	// 使用sync.Once确保每个通道只关闭一次
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

// Submit 提交任务 - 完全修复版本
func (wp *WorkerPool) Submit(task Task) error {
	wp.mu.RLock()
	if wp.closed {
		wp.mu.RUnlock()
		return errors.New("工作池已关闭")
	}
	wp.mu.RUnlock()

	select {
	case <-wp.ctx.Done():
		return errors.New("工作池已关闭")
	case <-wp.shutdownCh:
		return errors.New("工作池正在关闭")
	case wp.taskCh <- task:
		atomic.AddInt64(&wp.totalTasks, 1)
		return nil
	case <-time.After(wp.timeout):
		return errors.New("提交任务超时")
	}
}

// worker 工作协程 - 完全修复版本
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
		log.Printf("工作协程 %d 已停止", id)
	}()

	atomic.AddInt64(&wp.activeWorkers, 1)

	for {
		select {
		case <-wp.ctx.Done():
			log.Printf("工作协程 %d 收到停止信号", id)
			return
		case <-wp.shutdownCh:
			log.Printf("工作协程 %d 收到关闭信号", id)
			return
		case task, ok := <-wp.taskCh:
			if !ok {
				log.Printf("工作协程 %d 任务通道已关闭", id)
				return
			}

			// 处理任务
			wp.processTask(id, task)
		}
	}
}

// processTask 处理任务 - 完全修复版本
func (wp *WorkerPool) processTask(workerID int, task Task) {
	startTime := time.Now()

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(wp.ctx, wp.timeout)
	defer cancel()

	// 使用带缓冲的通道避免协程泄漏
	done := make(chan Result, 1)

	// 启动任务处理协程
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("工作协程 %d 处理任务时发生panic: %v", workerID, r)
				select {
				case done <- Result{
					TaskID: task.ID,
					Data:   task.Data,
					Error:  fmt.Errorf("任务处理panic: %v", r),
				}:
				default:
					// 通道已满或已关闭，忽略
				}
			}
		}()

		// 模拟处理时间
		time.Sleep(time.Duration(task.Priority*100) * time.Millisecond)

		// 模拟可能的错误
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
			// 通道已满，忽略结果
		}
	}()

	select {
	case <-ctx.Done():
		// 超时或取消
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
		// 任务完成
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

// recordProcessingTime 记录处理时间
func (wp *WorkerPool) recordProcessingTime(duration time.Duration) {
	wp.processingTimesMu.Lock()
	defer wp.processingTimesMu.Unlock()

	// 保持最近100个处理时间记录
	if len(wp.processingTimes) >= 100 {
		wp.processingTimes = wp.processingTimes[1:]
	}
	wp.processingTimes = append(wp.processingTimes, duration)
}

// errorHandler 错误处理协程 - 完全修复版本
func (wp *WorkerPool) errorHandler() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("错误处理协程发生panic: %v", r)
		}
		wp.wg.Done()
		log.Println("错误处理协程已停止")
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
		}
	}
}

// resultHandler 结果处理协程 - 完全修复版本
func (wp *WorkerPool) resultHandler() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("结果处理协程发生panic: %v", r)
		}
		wp.wg.Done()
		log.Println("结果处理协程已停止")
	}()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case result, ok := <-wp.resultCh:
			if !ok {
				return
			}

			if result.Error != nil {
				log.Printf("任务 %d 处理失败: %v", result.TaskID, result.Error)
			} else {
				log.Printf("任务 %d 处理成功: %s", result.TaskID, result.Data)
			}
		}
	}
}

// monitor 监控协程
func (wp *WorkerPool) monitor() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("监控协程发生panic: %v", r)
		}
		wp.wg.Done()
		log.Println("监控协程已停止")
	}()

	wp.monitorTicker = time.NewTicker(5 * time.Second)
	defer wp.monitorTicker.Stop()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case <-wp.monitorCh:
			return
		case <-wp.monitorTicker.C:
			wp.updateMetrics()
			wp.checkMemoryLeak()
		}
	}
}

// updateMetrics 更新指标
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

// checkMemoryLeak 检查内存泄漏
func (wp *WorkerPool) checkMemoryLeak() {
	// 检查内存使用是否异常增长
	if wp.metrics.MemoryHeapInuse > 100*1024*1024 { // 100MB
		log.Printf("警告: 内存使用较高: %d MB", wp.metrics.MemoryHeapInuse/1024/1024)
	}

	// 检查协程数量是否异常
	if wp.metrics.GoroutineCount > wp.workerCount*3 {
		log.Printf("警告: 协程数量异常: %d (预期: %d)", wp.metrics.GoroutineCount, wp.workerCount*3)
	}

	// 检查队列是否积压
	if wp.metrics.TaskQueueSize > wp.workerCount*2 {
		log.Printf("警告: 任务队列积压: %d", wp.metrics.TaskQueueSize)
	}
}

// healthChecker 健康检查协程
func (wp *WorkerPool) healthChecker() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("健康检查协程发生panic: %v", r)
		}
		wp.wg.Done()
		log.Println("健康检查协程已停止")
	}()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case <-wp.healthCheckStopCh: // 健康检查专属停止信号
			return
		case <-ticker.C:
			health := wp.HealthCheck()
			select {
			case wp.healthCheckCh <- health:
			default:
				// 通道已满，忽略
			}
		}
	}
}

// HealthCheck 健康检查
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

// GetHealthStatus 获取健康状态
func (wp *WorkerPool) GetHealthStatus() HealthStatus {
	select {
	case health := <-wp.healthCheckCh:
		return health
	default:
		return wp.HealthCheck()
	}
}

// GetMetrics 获取系统指标
func (wp *WorkerPool) GetMetrics() *SystemMetrics {
	wp.updateMetrics()
	return wp.metrics
}

// PrintMetrics 打印指标
func (wp *WorkerPool) PrintMetrics() {
	metrics := wp.GetMetrics()

	log.Println("=== 系统指标 ===")
	log.Printf("运行时间: %v", metrics.Uptime)
	log.Printf("内存使用: %d MB (堆: %d MB)", metrics.MemoryUsage/1024/1024, metrics.MemoryHeapInuse/1024/1024)
	log.Printf("协程数量: %d", metrics.GoroutineCount)
	log.Printf("活跃工作协程: %d", metrics.ActiveWorkers)
	log.Printf("任务队列大小: %d", metrics.TaskQueueSize)
	log.Printf("结果队列大小: %d", metrics.ResultQueueSize)
	log.Printf("总任务数: %d", metrics.TotalTasks)
	log.Printf("完成任务数: %d", metrics.CompletedTasks)
	log.Printf("失败任务数: %d", metrics.FailedTasks)
	log.Printf("超时任务数: %d", metrics.TimeoutTasks)
	log.Printf("平均处理时间: %v", metrics.AverageProcessingTime)
	log.Printf("最后处理时间: %v", metrics.LastProcessingTime)
}

// TaskGenerator 任务生成器
type TaskGenerator struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	pool   *WorkerPool
}

// NewTaskGenerator 创建任务生成器
func NewTaskGenerator(ctx context.Context, pool *WorkerPool) *TaskGenerator {
	ctx, cancel := context.WithCancel(ctx)
	return &TaskGenerator{
		ctx:    ctx,
		cancel: cancel,
		pool:   pool,
	}
}

// Start 启动任务生成器
func (tg *TaskGenerator) Start() {
	tg.wg.Add(1)
	go tg.generate()
}

// Stop 停止任务生成器
func (tg *TaskGenerator) Stop() {
	tg.cancel()
	tg.wg.Wait()
}

// generate 生成任务
func (tg *TaskGenerator) generate() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("任务生成器发生panic: %v", r)
		}
		tg.wg.Done()
		log.Println("任务生成器已停止")
	}()

	counter := 0
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-tg.ctx.Done():
			return
		case <-ticker.C:
			task := Task{
				ID:       counter,
				Data:     fmt.Sprintf("task_%d", counter),
				Priority: (counter % 5) + 1,
			}

			if err := tg.pool.Submit(task); err != nil {
				log.Printf("提交任务失败: %v", err)
				return
			}

			counter++
		}
	}
}

func main() {
	// 创建根上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建工作池
	pool := NewWorkerPool(5, 2*time.Second)
	pool.Start()

	// 创建任务生成器
	generator := NewTaskGenerator(ctx, pool)
	generator.Start()

	// 定期打印指标
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pool.PrintMetrics()

				// 检查健康状态
				health := pool.GetHealthStatus()
				if !health.IsHealthy {
					log.Printf("警告: 系统不健康 - %s", health.ErrorMessage)
				}
			}
		}
	}()

	// 运行一段时间
	time.Sleep(15 * time.Second)

	// 优雅停止
	log.Println("开始优雅停止...")
	generator.Stop()
	pool.Stop()

	// 最终指标报告
	log.Println("=== 最终指标报告 ===")
	pool.PrintMetrics()

	health := pool.GetHealthStatus()
	log.Printf("最终健康状态: %v", health.IsHealthy)
	if !health.IsHealthy {
		log.Printf("健康问题: %s", health.ErrorMessage)
	}

	log.Println("程序结束")
}
