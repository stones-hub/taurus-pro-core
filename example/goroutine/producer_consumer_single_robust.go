package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// Producer 生产者结构体
type Producer struct {
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	dataCh      chan int
	closed      bool
	mu          sync.RWMutex
	stats       *ProducerStats
	stopTimeout time.Duration

	// 内存泄漏防护相关字段
	metrics           *ProducerMetrics
	processingTimes   []time.Duration
	processingTimesMu sync.RWMutex
}

// Consumer 消费者结构体
type Consumer struct {
	id          int
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	dataCh      chan int
	closed      bool
	mu          sync.RWMutex
	stats       *ConsumerStats
	stopTimeout time.Duration

	// 内存泄漏防护相关字段
	metrics           *ConsumerMetrics
	processingTimes   []time.Duration
	processingTimesMu sync.RWMutex
}

// ProducerStats 生产者统计信息
type ProducerStats struct {
	producedCount int64
	errorCount    int64
	startTime     time.Time
}

// ConsumerStats 消费者统计信息
type ConsumerStats struct {
	processedCount int64
	errorCount     int64
	timeoutCount   int64
	startTime      time.Time
}

// ProducerMetrics 生产者指标
type ProducerMetrics struct {
	MemoryUsage     uint64
	MemoryHeapInuse uint64
	GoroutineCount  int
	QueueSize       int
	AverageSendTime time.Duration
	LastSendTime    time.Duration
	Uptime          time.Duration
}

// ConsumerMetrics 消费者指标
type ConsumerMetrics struct {
	MemoryUsage        uint64
	MemoryHeapInuse    uint64
	GoroutineCount     int
	QueueSize          int
	AverageProcessTime time.Duration
	LastProcessTime    time.Duration
	Uptime             time.Duration
}

// HealthStatus 健康状态
type HealthStatus struct {
	IsHealthy      bool
	ErrorMessage   string
	LastCheckTime  time.Time
	ProducerHealth bool
	ConsumerHealth []bool
	SystemMetrics  *SystemMetrics
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

	// 系统指标
	TotalProduced  int64
	TotalProcessed int64
	TotalErrors    int64
	TotalTimeouts  int64

	// 性能指标
	AverageProcessingTime time.Duration
	LastProcessingTime    time.Duration

	// 时间指标
	Uptime time.Duration
}

// NewProducer 创建生产者
func NewProducer(ctx context.Context, bufferSize int) *Producer {
	ctx, cancel := context.WithCancel(ctx)
	return &Producer{
		ctx:         ctx,
		cancel:      cancel,
		dataCh:      make(chan int, bufferSize),
		stats:       &ProducerStats{startTime: time.Now()},
		stopTimeout: 5 * time.Second,
		metrics:     &ProducerMetrics{},
	}
}

// NewConsumer 创建消费者
func NewConsumer(id int, ctx context.Context, dataCh chan int) *Consumer {
	ctx, cancel := context.WithCancel(ctx)
	return &Consumer{
		id:          id,
		ctx:         ctx,
		cancel:      cancel,
		dataCh:      dataCh,
		stats:       &ConsumerStats{startTime: time.Now()},
		stopTimeout: 5 * time.Second,
		metrics:     &ConsumerMetrics{},
	}
}

// Start 启动生产者
func (p *Producer) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("生产者已关闭，无法启动")
	}

	p.wg.Add(1)
	go p.produce()
	log.Println("生产者已启动")
	return nil
}

// Stop 停止生产者
func (p *Producer) Stop() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return fmt.Errorf("生产者已经停止")
	}
	p.closed = true
	p.mu.Unlock()

	log.Println("开始停止生产者...")
	p.cancel()

	// 等待协程完成，带超时
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("生产者已优雅停止，生产了 %d 条数据", atomic.LoadInt64(&p.stats.producedCount))
	case <-time.After(p.stopTimeout):
		log.Printf("生产者停止超时，已生产 %d 条数据", atomic.LoadInt64(&p.stats.producedCount))
	}

	// 安全关闭通道
	p.safeCloseChannel()
	return nil
}

// safeCloseChannel 安全关闭通道 - 完全修复版本
func (p *Producer) safeCloseChannel() {
	var closeOnce sync.Once
	closeOnce.Do(func() {
		close(p.dataCh)
	})
}

// GetStats 获取生产者统计信息
func (p *Producer) GetStats() ProducerStats {
	return ProducerStats{
		producedCount: atomic.LoadInt64(&p.stats.producedCount),
		errorCount:    atomic.LoadInt64(&p.stats.errorCount),
		startTime:     p.stats.startTime,
	}
}

// GetMetrics 获取生产者指标
func (p *Producer) GetMetrics() *ProducerMetrics {
	p.updateMetrics()
	return p.metrics
}

// updateMetrics 更新生产者指标
func (p *Producer) updateMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	p.metrics.MemoryUsage = memStats.Alloc
	p.metrics.MemoryHeapInuse = memStats.HeapInuse
	p.metrics.GoroutineCount = runtime.NumGoroutine()
	p.metrics.QueueSize = len(p.dataCh)
	p.metrics.Uptime = time.Since(p.stats.startTime)

	// 计算平均发送时间
	p.processingTimesMu.RLock()
	if len(p.processingTimes) > 0 {
		var total time.Duration
		for _, t := range p.processingTimes {
			total += t
		}
		p.metrics.AverageSendTime = total / time.Duration(len(p.processingTimes))
		p.metrics.LastSendTime = p.processingTimes[len(p.processingTimes)-1]
	}
	p.processingTimesMu.RUnlock()
}

// recordSendTime 记录发送时间
func (p *Producer) recordSendTime(duration time.Duration) {
	p.processingTimesMu.Lock()
	defer p.processingTimesMu.Unlock()

	// 保持最近50个发送时间记录
	if len(p.processingTimes) >= 50 {
		p.processingTimes = p.processingTimes[1:]
	}
	p.processingTimes = append(p.processingTimes, duration)
}

// produce 生产数据
func (p *Producer) produce() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("生产者发生panic: %v", r)
			atomic.AddInt64(&p.stats.errorCount, 1)
		}
		p.wg.Done()
		log.Println("生产者协程已停止")
	}()

	counter := 0
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			log.Println("生产者收到停止信号")
			return
		case <-ticker.C:
			startTime := time.Now()
			if err := p.sendData(counter); err != nil {
				log.Printf("生产者发送数据失败: %v", err)
				atomic.AddInt64(&p.stats.errorCount, 1)
				continue
			}
			p.recordSendTime(time.Since(startTime))
			atomic.AddInt64(&p.stats.producedCount, 1)
			counter++
		}
	}
}

// sendData 发送数据，带超时和错误处理
func (p *Producer) sendData(data int) error {
	select {
	case p.dataCh <- data:
		log.Printf("生产者发送数据: %d", data)
		return nil
	case <-p.ctx.Done():
		return fmt.Errorf("生产者已停止")
	case <-time.After(1 * time.Second):
		return fmt.Errorf("发送数据超时")
	}
}

// Start 启动消费者
func (c *Consumer) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("消费者 %d 已关闭，无法启动", c.id)
	}

	c.wg.Add(1)
	go c.consume()
	log.Printf("消费者 %d 已启动", c.id)
	return nil
}

// Stop 停止消费者
func (c *Consumer) Stop() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return fmt.Errorf("消费者 %d 已经停止", c.id)
	}
	c.closed = true
	c.mu.Unlock()

	log.Printf("开始停止消费者 %d...", c.id)
	c.cancel()

	// 等待协程完成，带超时
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		stats := c.GetStats()
		log.Printf("消费者 %d 已优雅停止，处理了 %d 条数据，超时 %d 次",
			c.id, stats.processedCount, stats.timeoutCount)
	case <-time.After(c.stopTimeout):
		stats := c.GetStats()
		log.Printf("消费者 %d 停止超时，已处理 %d 条数据", c.id, stats.processedCount)
	}

	return nil
}

// GetStats 获取消费者统计信息
func (c *Consumer) GetStats() ConsumerStats {
	return ConsumerStats{
		processedCount: atomic.LoadInt64(&c.stats.processedCount),
		errorCount:     atomic.LoadInt64(&c.stats.errorCount),
		timeoutCount:   atomic.LoadInt64(&c.stats.timeoutCount),
		startTime:      c.stats.startTime,
	}
}

// GetMetrics 获取消费者指标
func (c *Consumer) GetMetrics() *ConsumerMetrics {
	c.updateMetrics()
	return c.metrics
}

// updateMetrics 更新消费者指标
func (c *Consumer) updateMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.metrics.MemoryUsage = memStats.Alloc
	c.metrics.MemoryHeapInuse = memStats.HeapInuse
	c.metrics.GoroutineCount = runtime.NumGoroutine()
	c.metrics.QueueSize = len(c.dataCh)
	c.metrics.Uptime = time.Since(c.stats.startTime)

	// 计算平均处理时间
	c.processingTimesMu.RLock()
	if len(c.processingTimes) > 0 {
		var total time.Duration
		for _, t := range c.processingTimes {
			total += t
		}
		c.metrics.AverageProcessTime = total / time.Duration(len(c.processingTimes))
		c.metrics.LastProcessTime = c.processingTimes[len(c.processingTimes)-1]
	}
	c.processingTimesMu.RUnlock()
}

// recordProcessTime 记录处理时间
func (c *Consumer) recordProcessTime(duration time.Duration) {
	c.processingTimesMu.Lock()
	defer c.processingTimesMu.Unlock()

	// 保持最近50个处理时间记录
	if len(c.processingTimes) >= 50 {
		c.processingTimes = c.processingTimes[1:]
	}
	c.processingTimes = append(c.processingTimes, duration)
}

// consume 消费数据
func (c *Consumer) consume() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("消费者 %d 发生panic: %v", c.id, r)
			atomic.AddInt64(&c.stats.errorCount, 1)
		}
		c.wg.Done()
		log.Printf("消费者 %d 协程已停止", c.id)
	}()

	for {
		select {
		case <-c.ctx.Done():
			log.Printf("消费者 %d 收到停止信号", c.id)
			return
		case data, ok := <-c.dataCh:
			if !ok {
				log.Printf("消费者 %d 数据通道已关闭", c.id)
				return
			}

			// 处理数据，带超时控制
			startTime := time.Now()
			if err := c.processData(data); err != nil {
				log.Printf("消费者 %d 处理数据失败: %v", c.id, err)
				atomic.AddInt64(&c.stats.errorCount, 1)
			} else {
				atomic.AddInt64(&c.stats.processedCount, 1)
			}
			c.recordProcessTime(time.Since(startTime))
		}
	}
}

// processData 处理数据，带超时控制
func (c *Consumer) processData(data int) error {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(c.ctx, 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		defer close(done)
		// 模拟处理时间
		time.Sleep(50 * time.Millisecond)
		log.Printf("消费者 %d 处理数据: %d", c.id, data)
		done <- nil
	}()

	select {
	case <-ctx.Done():
		atomic.AddInt64(&c.stats.timeoutCount, 1)
		return fmt.Errorf("处理数据超时")
	case err := <-done:
		return err
	}
}

// SystemManager 系统管理器
type SystemManager struct {
	ctx       context.Context
	cancel    context.CancelFunc
	producer  *Producer
	consumers []*Consumer
	wg        sync.WaitGroup
	// 添加监控字段
	monitorCh chan struct{}
	stats     *SystemStats

	// 内存泄漏防护和健康检查相关字段
	metrics       *SystemMetrics
	startTime     time.Time
	monitorTicker *time.Ticker
	healthCheckCh chan HealthStatus
}

// SystemStats 系统统计信息
type SystemStats struct {
	startTime      time.Time
	runningTime    time.Duration
	totalProduced  int64
	totalProcessed int64
	totalErrors    int64
}

// NewSystemManager 创建系统管理器
func NewSystemManager() *SystemManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &SystemManager{
		ctx:           ctx,
		cancel:        cancel,
		monitorCh:     make(chan struct{}),
		stats:         &SystemStats{startTime: time.Now()},
		metrics:       &SystemMetrics{},
		startTime:     time.Now(),
		healthCheckCh: make(chan HealthStatus, 1),
	}
}

// Start 启动系统
func (sm *SystemManager) Start() error {
	// 创建生产者
	sm.producer = NewProducer(sm.ctx, 100)
	if err := sm.producer.Start(); err != nil {
		return fmt.Errorf("启动生产者失败: %v", err)
	}

	// 创建多个消费者
	sm.consumers = make([]*Consumer, 3)
	for i := 0; i < 3; i++ {
		sm.consumers[i] = NewConsumer(i+1, sm.ctx, sm.producer.dataCh)
		if err := sm.consumers[i].Start(); err != nil {
			return fmt.Errorf("启动消费者 %d 失败: %v", i+1, err)
		}
	}

	// 启动监控协程
	sm.wg.Add(1)
	go sm.monitor()

	// 启动健康检查协程
	sm.wg.Add(1)
	go sm.healthChecker()

	log.Println("系统已启动")
	return nil
}

// Stop 停止系统
func (sm *SystemManager) Stop() error {
	log.Println("开始停止系统...")

	// 更新统计信息
	sm.stats.runningTime = time.Since(sm.stats.startTime)

	// 停止生产者
	if sm.producer != nil {
		if err := sm.producer.Stop(); err != nil {
			log.Printf("停止生产者失败: %v", err)
		}
	}

	// 停止所有消费者
	for _, consumer := range sm.consumers {
		if consumer != nil {
			if err := consumer.Stop(); err != nil {
				log.Printf("停止消费者失败: %v", err)
			}
		}
	}

	// 关闭监控通道
	close(sm.monitorCh)

	// 停止监控
	if sm.monitorTicker != nil {
		sm.monitorTicker.Stop()
	}

	// 等待监控协程完成
	sm.wg.Wait()

	log.Println("系统已停止")
	return nil
}

// monitor 监控协程
func (sm *SystemManager) monitor() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("监控协程发生panic: %v", r)
		}
		sm.wg.Done()
		log.Println("监控协程已停止")
	}()

	sm.monitorTicker = time.NewTicker(2 * time.Second)
	defer sm.monitorTicker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-sm.monitorCh:
			return
		case <-sm.monitorTicker.C:
			sm.updateStats()
			sm.updateMetrics()
			sm.checkMemoryLeak()
			sm.logSystemHealth()
		}
	}
}

// updateStats 更新统计信息
func (sm *SystemManager) updateStats() {
	if sm.producer != nil {
		stats := sm.producer.GetStats()
		atomic.StoreInt64(&sm.stats.totalProduced, stats.producedCount)
		atomic.AddInt64(&sm.stats.totalErrors, stats.errorCount)
	}

	for _, consumer := range sm.consumers {
		if consumer != nil {
			stats := consumer.GetStats()
			atomic.AddInt64(&sm.stats.totalProcessed, stats.processedCount)
			atomic.AddInt64(&sm.stats.totalErrors, stats.errorCount)
		}
	}
}

// updateMetrics 更新系统指标
func (sm *SystemManager) updateMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	sm.metrics.MemoryUsage = memStats.Alloc
	sm.metrics.MemoryAlloc = memStats.Alloc
	sm.metrics.MemorySys = memStats.Sys
	sm.metrics.MemoryHeapAlloc = memStats.HeapAlloc
	sm.metrics.MemoryHeapSys = memStats.HeapSys
	sm.metrics.MemoryHeapIdle = memStats.HeapIdle
	sm.metrics.MemoryHeapInuse = memStats.HeapInuse
	sm.metrics.GoroutineCount = runtime.NumGoroutine()
	sm.metrics.Uptime = time.Since(sm.startTime)

	// 更新系统统计
	if sm.producer != nil {
		stats := sm.producer.GetStats()
		sm.metrics.TotalProduced = stats.producedCount
	}

	totalProcessed := int64(0)
	totalErrors := int64(0)
	totalTimeouts := int64(0)

	for _, consumer := range sm.consumers {
		if consumer != nil {
			stats := consumer.GetStats()
			totalProcessed += stats.processedCount
			totalErrors += stats.errorCount
			totalTimeouts += stats.timeoutCount
		}
	}

	sm.metrics.TotalProcessed = totalProcessed
	sm.metrics.TotalErrors = totalErrors
	sm.metrics.TotalTimeouts = totalTimeouts
}

// checkMemoryLeak 检查内存泄漏
func (sm *SystemManager) checkMemoryLeak() {
	// 检查内存使用是否异常增长
	if sm.metrics.MemoryHeapInuse > 100*1024*1024 { // 100MB
		log.Printf("警告: 内存使用较高: %d MB", sm.metrics.MemoryHeapInuse/1024/1024)
	}

	// 检查协程数量是否异常
	expectedGoroutines := 1 + 1 + len(sm.consumers) + 2 // 主协程 + 生产者 + 消费者 + 监控协程
	if sm.metrics.GoroutineCount > expectedGoroutines*3 {
		log.Printf("警告: 协程数量异常: %d (预期: %d)", sm.metrics.GoroutineCount, expectedGoroutines*3)
	}

	// 检查队列是否积压
	if sm.producer != nil {
		metrics := sm.producer.GetMetrics()
		if metrics.QueueSize > 50 {
			log.Printf("警告: 生产者队列积压: %d", metrics.QueueSize)
		}
	}
}

// healthChecker 健康检查协程
func (sm *SystemManager) healthChecker() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("健康检查协程发生panic: %v", r)
		}
		sm.wg.Done()
		log.Println("健康检查协程已停止")
	}()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			health := sm.HealthCheck()
			select {
			case sm.healthCheckCh <- health:
			default:
				// 通道已满，忽略
			}
		}
	}
}

// HealthCheck 健康检查
func (sm *SystemManager) HealthCheck() HealthStatus {
	sm.updateMetrics()

	health := HealthStatus{
		IsHealthy:      true,
		LastCheckTime:  time.Now(),
		SystemMetrics:  sm.metrics,
		ConsumerHealth: make([]bool, len(sm.consumers)),
	}

	// 检查内存使用
	if sm.metrics.MemoryHeapInuse > 200*1024*1024 { // 200MB
		health.IsHealthy = false
		health.ErrorMessage = fmt.Sprintf("内存使用过高: %d MB", sm.metrics.MemoryHeapInuse/1024/1024)
	}

	// 检查协程泄漏
	expectedGoroutines := 1 + 1 + len(sm.consumers) + 2
	if sm.metrics.GoroutineCount > expectedGoroutines*5 {
		health.IsHealthy = false
		health.ErrorMessage = fmt.Sprintf("协程数量异常: %d", sm.metrics.GoroutineCount)
	}

	// 检查生产者健康状态
	if sm.producer != nil {
		metrics := sm.producer.GetMetrics()
		if metrics.QueueSize > 100 {
			health.IsHealthy = false
			health.ErrorMessage = fmt.Sprintf("生产者队列严重积压: %d", metrics.QueueSize)
		}
		health.ProducerHealth = true
	}

	// 检查消费者健康状态
	for i, consumer := range sm.consumers {
		if consumer != nil {
			metrics := consumer.GetMetrics()
			health.ConsumerHealth[i] = true

			// 检查消费者处理能力
			if metrics.QueueSize > 50 {
				health.IsHealthy = false
				health.ErrorMessage = fmt.Sprintf("消费者 %d 队列积压: %d", i+1, metrics.QueueSize)
			}
		}
	}

	// 检查错误率
	totalProcessed := sm.metrics.TotalProcessed + sm.metrics.TotalErrors + sm.metrics.TotalTimeouts
	if totalProcessed > 0 {
		errorRate := float64(sm.metrics.TotalErrors+sm.metrics.TotalTimeouts) / float64(totalProcessed)
		if errorRate > 0.1 { // 错误率超过10%
			health.IsHealthy = false
			health.ErrorMessage = fmt.Sprintf("错误率过高: %.2f%%", errorRate*100)
		}
	}

	return health
}

// GetHealthStatus 获取健康状态
func (sm *SystemManager) GetHealthStatus() HealthStatus {
	select {
	case health := <-sm.healthCheckCh:
		return health
	default:
		return sm.HealthCheck()
	}
}

// logSystemHealth 记录系统健康状态
func (sm *SystemManager) logSystemHealth() {
	runningTime := time.Since(sm.stats.startTime)
	totalProduced := atomic.LoadInt64(&sm.stats.totalProduced)
	totalProcessed := atomic.LoadInt64(&sm.stats.totalProcessed)
	totalErrors := atomic.LoadInt64(&sm.stats.totalErrors)

	log.Printf("系统健康状态 - 运行时间: %v, 生产: %d, 处理: %d, 错误: %d",
		runningTime, totalProduced, totalProcessed, totalErrors)
}

// PrintStats 打印统计信息
func (sm *SystemManager) PrintStats() {
	sm.updateStats()

	log.Println("=== 系统统计信息 ===")
	log.Printf("系统运行时间: %v", sm.stats.runningTime)
	log.Printf("总生产数据: %d", atomic.LoadInt64(&sm.stats.totalProduced))
	log.Printf("总处理数据: %d", atomic.LoadInt64(&sm.stats.totalProcessed))
	log.Printf("总错误次数: %d", atomic.LoadInt64(&sm.stats.totalErrors))

	if sm.producer != nil {
		stats := sm.producer.GetStats()
		log.Printf("生产者统计: 生产 %d 条数据, 错误 %d 次, 运行时间 %v",
			stats.producedCount, stats.errorCount, time.Since(stats.startTime))
	}

	for _, consumer := range sm.consumers {
		if consumer != nil {
			stats := consumer.GetStats()
			log.Printf("消费者 %d 统计: 处理 %d 条数据, 错误 %d 次, 超时 %d 次, 运行时间 %v",
				consumer.id, stats.processedCount, stats.errorCount, stats.timeoutCount,
				time.Since(stats.startTime))
		}
	}
}

// PrintMetrics 打印系统指标
func (sm *SystemManager) PrintMetrics() {
	sm.updateMetrics()

	log.Println("=== 系统指标 ===")
	log.Printf("运行时间: %v", sm.metrics.Uptime)
	log.Printf("内存使用: %d MB (堆: %d MB)", sm.metrics.MemoryUsage/1024/1024, sm.metrics.MemoryHeapInuse/1024/1024)
	log.Printf("协程数量: %d", sm.metrics.GoroutineCount)
	log.Printf("总生产数据: %d", sm.metrics.TotalProduced)
	log.Printf("总处理数据: %d", sm.metrics.TotalProcessed)
	log.Printf("总错误次数: %d", sm.metrics.TotalErrors)
	log.Printf("总超时次数: %d", sm.metrics.TotalTimeouts)

	if sm.producer != nil {
		metrics := sm.producer.GetMetrics()
		log.Printf("生产者指标: 队列大小 %d, 平均发送时间 %v",
			metrics.QueueSize, metrics.AverageSendTime)
	}

	for i, consumer := range sm.consumers {
		if consumer != nil {
			metrics := consumer.GetMetrics()
			log.Printf("消费者 %d 指标: 队列大小 %d, 平均处理时间 %v",
				i+1, metrics.QueueSize, metrics.AverageProcessTime)
		}
	}
}

func main() {
	// 创建系统管理器
	manager := NewSystemManager()
	defer manager.cancel()

	// 启动系统
	if err := manager.Start(); err != nil {
		log.Fatalf("启动系统失败: %v", err)
	}

	// 定期打印指标和健康状态
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-manager.ctx.Done():
				return
			case <-ticker.C:
				manager.PrintMetrics()

				// 检查健康状态
				health := manager.GetHealthStatus()
				if !health.IsHealthy {
					log.Printf("警告: 系统不健康 - %s", health.ErrorMessage)
				}
			}
		}
	}()

	// 运行一段时间
	time.Sleep(5 * time.Second)

	// 打印统计信息
	manager.PrintStats()

	// 打印最终指标
	manager.PrintMetrics()

	// 检查最终健康状态
	health := manager.GetHealthStatus()
	log.Printf("最终健康状态: %v", health.IsHealthy)
	if !health.IsHealthy {
		log.Printf("健康问题: %s", health.ErrorMessage)
	}

	// 停止系统
	if err := manager.Stop(); err != nil {
		log.Printf("停止系统失败: %v", err)
	}

	log.Println("程序结束")
}
