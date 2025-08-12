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
	id     int
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	dataCh chan int
	closed bool
	mu     sync.RWMutex
	stats  *ProducerStats

	// 内存泄漏防护相关字段
	metrics           *ProducerMetrics
	processingTimes   []time.Duration
	processingTimesMu sync.RWMutex

	// 专属停止channel
	stopCh chan struct{}

	// 企业级增强字段
	config            *ProducerConfig
	retryCount        int64
	backpressureLimit int
	rateLimiter       *time.Ticker
}

// ProducerConfig 生产者配置
type ProducerConfig struct {
	ID                int
	RateLimit         time.Duration
	RetryAttempts     int
	BackpressureLimit int
	EnableEncryption  bool
	EnableAudit       bool
}

// CircuitBreaker 熔断器状态
type CircuitBreakerState int

const (
	CircuitBreakerClosed   CircuitBreakerState = iota // 关闭状态：正常工作
	CircuitBreakerOpen                                // 开启状态：快速失败
	CircuitBreakerHalfOpen                            // 半开状态：尝试恢复
)

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	state           CircuitBreakerState
	failureCount    int64
	successCount    int64
	lastFailureTime time.Time
	mu              sync.RWMutex

	// 配置参数
	failureThreshold int64         // 失败阈值
	timeout          time.Duration // 熔断时间
	successThreshold int64         // 成功阈值（半开状态）
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(failureThreshold int64, timeout time.Duration, successThreshold int64) *CircuitBreaker {
	return &CircuitBreaker{
		state:            CircuitBreakerClosed,
		failureThreshold: failureThreshold,
		timeout:          timeout,
		successThreshold: successThreshold,
	}
}

// CanExecute 检查是否可以执行
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		// 检查是否超时，可以进入半开状态
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = CircuitBreakerHalfOpen
			cb.mu.Unlock()
			cb.mu.RLock()
			return true
		}
		return false
	case CircuitBreakerHalfOpen:
		return true
	default:
		return false
	}
}

// OnSuccess 记录成功
func (cb *CircuitBreaker) OnSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitBreakerClosed:
		// 重置失败计数
		cb.failureCount = 0
	case CircuitBreakerHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			// 恢复到关闭状态
			cb.state = CircuitBreakerClosed
			cb.failureCount = 0
			cb.successCount = 0
		}
	}
}

// OnFailure 记录失败
func (cb *CircuitBreaker) OnFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case CircuitBreakerClosed:
		if cb.failureCount >= cb.failureThreshold {
			// 进入开启状态
			cb.state = CircuitBreakerOpen
		}
	case CircuitBreakerHalfOpen:
		// 半开状态下失败，立即回到开启状态
		cb.state = CircuitBreakerOpen
		cb.successCount = 0
	}
}

// GetState 获取当前状态
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Consumer 消费者结构体
type Consumer struct {
	id     int
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	dataCh chan int
	closed bool
	mu     sync.RWMutex
	stats  *ConsumerStats

	// 内存泄漏防护相关字段
	metrics           *ConsumerMetrics
	processingTimes   []time.Duration
	processingTimesMu sync.RWMutex

	// 专属停止channel
	stopCh chan struct{}

	// 熔断器
	circuitBreaker *CircuitBreaker
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
	MemoryUsage         uint64
	MemoryHeapInuse     uint64
	GoroutineCount      int
	QueueSize           int
	AverageProcessTime  time.Duration
	LastProcessTime     time.Duration
	Uptime              time.Duration
	CircuitBreakerState CircuitBreakerState // 熔断器状态
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

// HealthStatus 健康状态
type HealthStatus struct {
	IsHealthy      bool
	ErrorMessage   string
	LastCheckTime  time.Time
	ProducerHealth []bool
	ConsumerHealth []bool
	SystemMetrics  *SystemMetrics
}

// NewProducer 创建生产者
func NewProducer(id int, ctx context.Context, dataCh chan int, config *ProducerConfig) *Producer {
	ctx, cancel := context.WithCancel(ctx)
	if config == nil {
		config = &ProducerConfig{
			ID:                id,
			RateLimit:         100 * time.Millisecond,
			RetryAttempts:     3,
			BackpressureLimit: 100,
			EnableEncryption:  false,
			EnableAudit:       true,
		}
	}
	return &Producer{
		id:                id,
		ctx:               ctx,
		cancel:            cancel,
		dataCh:            dataCh,
		stats:             &ProducerStats{startTime: time.Now()},
		metrics:           &ProducerMetrics{},
		stopCh:            make(chan struct{}),
		config:            config,
		backpressureLimit: config.BackpressureLimit,
		rateLimiter:       time.NewTicker(config.RateLimit),
	}
}

// NewConsumer 创建消费者
func NewConsumer(id int, ctx context.Context, dataCh chan int) *Consumer {
	ctx, cancel := context.WithCancel(ctx)
	return &Consumer{
		id:             id,
		ctx:            ctx,
		cancel:         cancel,
		dataCh:         dataCh,
		stats:          &ConsumerStats{startTime: time.Now()},
		metrics:        &ConsumerMetrics{},
		stopCh:         make(chan struct{}),
		circuitBreaker: NewCircuitBreaker(5, 10*time.Second, 3), // 5次失败后熔断，10秒后尝试恢复，3次成功恢复
	}
}

// Start 启动生产者
func (p *Producer) Start() {
	p.wg.Add(1)
	go p.produce()
}

// Stop 停止生产者 - 使用专属停止channel
func (p *Producer) Stop() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	p.mu.Unlock()

	// 发送停止信号到专属channel
	close(p.stopCh)

	// 等待协程完成，带超时
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		stats := p.GetStats()
		log.Printf("生产者 %d 已优雅停止，生产了 %d 条数据", p.id, stats.producedCount)
	case <-time.After(5 * time.Second):
		stats := p.GetStats()
		log.Printf("生产者 %d 停止超时，已生产 %d 条数据", p.id, stats.producedCount)
	}
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

// produce 生产数据 - 使用专属停止channel
func (p *Producer) produce() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("生产者 %d 发生panic: %v", p.id, r)
			atomic.AddInt64(&p.stats.errorCount, 1)
		}
		p.wg.Done()
		log.Printf("生产者 %d 已停止", p.id)
	}()

	counter := 0
	ticker := time.NewTicker(time.Duration(100+p.id*30) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopCh:
			log.Printf("生产者 %d 收到专属停止信号", p.id)
			return
		case <-p.ctx.Done():
			log.Printf("生产者 %d 收到context停止信号", p.id)
			return
		case <-ticker.C:
			// 在发送数据前检查是否已经停止
			p.mu.RLock()
			if p.closed {
				p.mu.RUnlock()
				log.Printf("生产者 %d 检测到已停止标记，退出生产循环", p.id)
				return
			}
			p.mu.RUnlock()

			startTime := time.Now()
			if err := p.sendData(counter); err != nil {
				log.Printf("生产者 %d 发送数据失败: %v", p.id, err)
				atomic.AddInt64(&p.stats.errorCount, 1)
				continue
			}
			p.recordSendTime(time.Since(startTime))
			atomic.AddInt64(&p.stats.producedCount, 1)
			counter++
		}
	}
}

// sendData 发送数据，带超时、错误处理、背压控制和重试机制
func (p *Producer) sendData(data int) error {
	// 背压控制：检查队列大小
	if len(p.dataCh) >= p.backpressureLimit {
		log.Printf("生产者 %d 检测到背压，队列大小: %d", p.id, len(p.dataCh))
		return fmt.Errorf("背压限制: 队列已满")
	}

	// 审计日志
	if p.config.EnableAudit {
		log.Printf("生产者 %d 审计: 准备发送数据 %d", p.id, data)
	}

	// 重试机制
	for attempt := 0; attempt <= p.config.RetryAttempts; attempt++ {
		select {
		case p.dataCh <- data:
			if p.config.EnableAudit {
				log.Printf("生产者 %d 审计: 成功发送数据 %d (尝试 %d)", p.id, data, attempt+1)
			}
			log.Printf("生产者 %d 发送数据: %d", p.id, data)
			return nil
		case <-p.stopCh:
			return context.Canceled
		case <-p.ctx.Done():
			return context.Canceled
		case <-time.After(1 * time.Second):
			if attempt < p.config.RetryAttempts {
				log.Printf("生产者 %d 发送超时，重试 %d/%d", p.id, attempt+1, p.config.RetryAttempts)
				atomic.AddInt64(&p.retryCount, 1)
				continue
			}
			return context.DeadlineExceeded
		}
	}
	return fmt.Errorf("发送失败，已重试 %d 次", p.config.RetryAttempts)
}

// Start 启动消费者
func (c *Consumer) Start() {
	c.wg.Add(1)
	go c.consume()
}

// Stop 停止消费者 - 使用专属停止channel
func (c *Consumer) Stop() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	c.mu.Unlock()

	// 发送停止信号到专属channel
	close(c.stopCh)

	// 等待协程完成，带超时
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		stats := c.GetStats()
		log.Printf("消费者 %d 已优雅停止，处理了 %d 条数据", c.id, stats.processedCount)
	case <-time.After(5 * time.Second):
		stats := c.GetStats()
		log.Printf("消费者 %d 停止超时，已处理 %d 条数据", c.id, stats.processedCount)
	}
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
	c.metrics.CircuitBreakerState = c.circuitBreaker.GetState()

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

// consume 消费数据 - 使用专属停止channel和熔断机制
func (c *Consumer) consume() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("消费者 %d 发生panic: %v", c.id, r)
			atomic.AddInt64(&c.stats.errorCount, 1)
			c.circuitBreaker.OnFailure()
		}
		c.wg.Done()
		log.Printf("消费者 %d 已停止", c.id)
	}()

	for {
		select {
		case <-c.stopCh:
			log.Printf("消费者 %d 收到专属停止信号", c.id)
			return
		case <-c.ctx.Done():
			log.Printf("消费者 %d 收到context停止信号", c.id)
			return
		case data, ok := <-c.dataCh:
			if !ok {
				log.Printf("消费者 %d 数据通道已关闭", c.id)
				return
			}

			// 检查熔断器状态
			if !c.circuitBreaker.CanExecute() {
				state := c.circuitBreaker.GetState()
				log.Printf("消费者 %d 熔断器状态: %v，跳过处理数据: %d", c.id, state, data)
				atomic.AddInt64(&c.stats.errorCount, 1)
				continue
			}

			// 处理数据，带超时控制
			startTime := time.Now()
			if err := c.processData(data); err != nil {
				log.Printf("消费者 %d 处理数据失败: %v", c.id, err)
				atomic.AddInt64(&c.stats.errorCount, 1)
				c.circuitBreaker.OnFailure()
			} else {
				atomic.AddInt64(&c.stats.processedCount, 1)
				c.circuitBreaker.OnSuccess()
			}
			c.recordProcessTime(time.Since(startTime))
		}
	}
}

// processData 处理数据，带超时控制和故障模拟 - 使用专属停止channel
func (c *Consumer) processData(data int) error {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(c.ctx, 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		defer close(done)
		// 模拟处理时间
		time.Sleep(time.Duration(50+c.id*20) * time.Millisecond)

		// 模拟故障：消费者2在处理特定数据时失败
		if c.id == 2 && data%7 == 0 {
			log.Printf("消费者 %d 模拟处理失败: %d", c.id, data)
			done <- fmt.Errorf("模拟处理失败")
			return
		}

		log.Printf("消费者 %d 处理数据: %d", c.id, data)
		done <- nil
	}()

	select {
	case <-c.stopCh:
		return context.Canceled
	case <-ctx.Done():
		atomic.AddInt64(&c.stats.timeoutCount, 1)
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// SystemManager 系统管理器
type SystemManager struct {
	ctx       context.Context
	cancel    context.CancelFunc
	producers []*Producer
	consumers []*Consumer
	wg        sync.WaitGroup

	// 内存泄漏防护和健康检查相关字段
	metrics       *SystemMetrics
	startTime     time.Time
	monitorTicker *time.Ticker
	healthCheckCh chan HealthStatus
}

// NewSystemManager 创建系统管理器
func NewSystemManager() *SystemManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &SystemManager{
		ctx:           ctx,
		cancel:        cancel,
		metrics:       &SystemMetrics{},
		startTime:     time.Now(),
		healthCheckCh: make(chan HealthStatus, 1),
	}
}

// Start 启动系统
func (sm *SystemManager) Start() error {
	// 创建共享数据通道
	dataCh := make(chan int, 300)

	// 创建多个生产者
	sm.producers = make([]*Producer, 3)
	for i := 0; i < 3; i++ {
		config := &ProducerConfig{
			ID:                i + 1,
			RateLimit:         time.Duration(100+i*30) * time.Millisecond,
			RetryAttempts:     3,
			BackpressureLimit: 100,
			EnableEncryption:  false,
			EnableAudit:       true,
		}
		sm.producers[i] = NewProducer(i+1, sm.ctx, dataCh, config)
		sm.producers[i].Start()
	}

	// 创建多个消费者
	sm.consumers = make([]*Consumer, 4)
	for i := 0; i < 4; i++ {
		sm.consumers[i] = NewConsumer(i+1, sm.ctx, dataCh)
		sm.consumers[i].Start()
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

	// 首先取消context，通知所有协程停止
	sm.cancel()

	// 停止所有生产者
	for _, producer := range sm.producers {
		if producer != nil {
			producer.Stop()
		}
	}

	// 停止所有消费者
	for _, consumer := range sm.consumers {
		if consumer != nil {
			consumer.Stop()
		}
	}

	// 停止监控
	if sm.monitorTicker != nil {
		sm.monitorTicker.Stop()
	}

	// 等待监控协程完成
	sm.wg.Wait()

	log.Println("系统已停止")
	return nil
}

// StopProducer 停止指定的生产者
func (sm *SystemManager) StopProducer(id int) error {
	if id < 1 || id > len(sm.producers) {
		return fmt.Errorf("无效的生产者ID: %d", id)
	}

	producer := sm.producers[id-1]
	if producer != nil {
		log.Printf("停止生产者 %d", id)
		producer.Stop()
	}
	return nil
}

// StopConsumer 停止指定的消费者
func (sm *SystemManager) StopConsumer(id int) error {
	if id < 1 || id > len(sm.consumers) {
		return fmt.Errorf("无效的消费者ID: %d", id)
	}

	consumer := sm.consumers[id-1]
	if consumer != nil {
		log.Printf("停止消费者 %d", id)
		consumer.Stop()
	}
	return nil
}

// StopAllProducers 停止所有生产者
func (sm *SystemManager) StopAllProducers() {
	log.Println("停止所有生产者...")
	for _, producer := range sm.producers {
		if producer != nil {
			producer.Stop()
		}
	}
}

// StopAllConsumers 停止所有消费者
func (sm *SystemManager) StopAllConsumers() {
	log.Println("停止所有消费者...")
	for _, consumer := range sm.consumers {
		if consumer != nil {
			consumer.Stop()
		}
	}
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

	sm.monitorTicker = time.NewTicker(3 * time.Second)
	defer sm.monitorTicker.Stop()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-sm.monitorTicker.C:
			sm.updateMetrics()
			sm.checkMemoryLeak()
			sm.logSystemHealth()
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
	totalProduced := int64(0)
	totalProcessed := int64(0)
	totalErrors := int64(0)
	totalTimeouts := int64(0)

	for _, producer := range sm.producers {
		if producer != nil {
			stats := producer.GetStats()
			totalProduced += stats.producedCount
			totalErrors += stats.errorCount
		}
	}

	for _, consumer := range sm.consumers {
		if consumer != nil {
			stats := consumer.GetStats()
			totalProcessed += stats.processedCount
			totalErrors += stats.errorCount
			totalTimeouts += stats.timeoutCount
		}
	}

	sm.metrics.TotalProduced = totalProduced
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
	expectedGoroutines := 1 + len(sm.producers) + len(sm.consumers) + 2 // 主协程 + 生产者 + 消费者 + 监控协程
	if sm.metrics.GoroutineCount > expectedGoroutines*3 {
		log.Printf("警告: 协程数量异常: %d (预期: %d)", sm.metrics.GoroutineCount, expectedGoroutines*3)
	}

	// 检查队列是否积压
	for _, producer := range sm.producers {
		if producer != nil {
			metrics := producer.GetMetrics()
			if metrics.QueueSize > 50 {
				log.Printf("警告: 生产者 %d 队列积压: %d", producer.id, metrics.QueueSize)
			}
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
		ProducerHealth: make([]bool, len(sm.producers)),
		ConsumerHealth: make([]bool, len(sm.consumers)),
	}

	// 检查内存使用
	if sm.metrics.MemoryHeapInuse > 200*1024*1024 { // 200MB
		health.IsHealthy = false
		health.ErrorMessage = fmt.Sprintf("内存使用过高: %d MB", sm.metrics.MemoryHeapInuse/1024/1024)
	}

	// 检查协程泄漏
	expectedGoroutines := 1 + len(sm.producers) + len(sm.consumers) + 2
	if sm.metrics.GoroutineCount > expectedGoroutines*5 {
		health.IsHealthy = false
		health.ErrorMessage = fmt.Sprintf("协程数量异常: %d", sm.metrics.GoroutineCount)
	}

	// 检查生产者健康状态
	for i, producer := range sm.producers {
		if producer != nil {
			metrics := producer.GetMetrics()
			health.ProducerHealth[i] = true

			if metrics.QueueSize > 100 {
				health.IsHealthy = false
				health.ErrorMessage = fmt.Sprintf("生产者 %d 队列严重积压: %d", i+1, metrics.QueueSize)
			}
		}
	}

	// 检查消费者健康状态
	for i, consumer := range sm.consumers {
		if consumer != nil {
			metrics := consumer.GetMetrics()
			health.ConsumerHealth[i] = true

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
	log.Printf("系统健康状态 - 运行时间: %v, 生产: %d, 处理: %d, 错误: %d",
		sm.metrics.Uptime, sm.metrics.TotalProduced, sm.metrics.TotalProcessed, sm.metrics.TotalErrors)
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

	for _, producer := range sm.producers {
		if producer != nil {
			metrics := producer.GetMetrics()
			log.Printf("生产者 %d 指标: 队列大小 %d, 平均发送时间 %v",
				producer.id, metrics.QueueSize, metrics.AverageSendTime)
		}
	}

	for _, consumer := range sm.consumers {
		if consumer != nil {
			metrics := consumer.GetMetrics()
			stateStr := "Closed"
			switch metrics.CircuitBreakerState {
			case CircuitBreakerOpen:
				stateStr = "Open"
			case CircuitBreakerHalfOpen:
				stateStr = "HalfOpen"
			}
			log.Printf("消费者 %d 指标: 队列大小 %d, 平均处理时间 %v, 熔断器状态 %s",
				consumer.id, metrics.QueueSize, metrics.AverageProcessTime, stateStr)
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
	metricsDone := make(chan struct{})
	go func() {
		defer close(metricsDone)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-manager.ctx.Done():
				log.Println("指标监控协程收到停止信号")
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

	// 运行一段时间后演示单独停止功能
	time.Sleep(3 * time.Second)

	// 演示单独停止功能
	log.Println("=== 演示单独停止功能 ===")

	// 停止第一个生产者
	if err := manager.StopProducer(1); err != nil {
		log.Printf("停止生产者1失败: %v", err)
	}

	time.Sleep(1 * time.Second)

	// 停止第一个消费者
	if err := manager.StopConsumer(1); err != nil {
		log.Printf("停止消费者1失败: %v", err)
	}

	time.Sleep(1 * time.Second)

	// 停止所有生产者
	manager.StopAllProducers()

	time.Sleep(1 * time.Second)

	// 停止所有消费者
	manager.StopAllConsumers()

	log.Println("开始停止所有协程...")

	// 停止系统
	if err := manager.Stop(); err != nil {
		log.Printf("停止系统失败: %v", err)
	}

	// 等待指标监控协程完成
	select {
	case <-metricsDone:
		log.Println("指标监控协程已停止")
	case <-time.After(5 * time.Second):
		log.Println("等待指标监控协程停止超时")
	}

	// 打印最终指标
	manager.PrintMetrics()

	// 检查最终健康状态
	health := manager.GetHealthStatus()
	log.Printf("最终健康状态: %v", health.IsHealthy)
	if !health.IsHealthy {
		log.Printf("健康问题: %s", health.ErrorMessage)
	}

	log.Println("所有协程已停止")
}
