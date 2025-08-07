package main

import (
	"context"
	"log"
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
}

// Consumer 消费者结构体
type Consumer struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	dataCh chan int
	closed bool
	mu     sync.RWMutex
	stats  *ConsumerStats
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

// NewProducer 创建生产者
func NewProducer(id int, ctx context.Context, dataCh chan int) *Producer {
	ctx, cancel := context.WithCancel(ctx)
	return &Producer{
		id:     id,
		ctx:    ctx,
		cancel: cancel,
		dataCh: dataCh,
		stats:  &ProducerStats{startTime: time.Now()},
	}
}

// NewConsumer 创建消费者
func NewConsumer(ctx context.Context, dataCh chan int) *Consumer {
	ctx, cancel := context.WithCancel(ctx)
	return &Consumer{
		ctx:    ctx,
		cancel: cancel,
		dataCh: dataCh,
		stats:  &ConsumerStats{startTime: time.Now()},
	}
}

// Start 启动生产者
func (p *Producer) Start() {
	p.wg.Add(1)
	go p.produce()
}

// Stop 停止生产者 - 完全修复版本
func (p *Producer) Stop() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	p.mu.Unlock()

	p.cancel()

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

// produce 生产数据 - 完全修复版本
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
	ticker := time.NewTicker(time.Duration(100+p.id*50) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			log.Printf("生产者 %d 收到停止信号", p.id)
			return
		case <-ticker.C:
			if err := p.sendData(counter); err != nil {
				log.Printf("生产者 %d 发送数据失败: %v", p.id, err)
				atomic.AddInt64(&p.stats.errorCount, 1)
				continue
			}
			atomic.AddInt64(&p.stats.producedCount, 1)
			counter++
		}
	}
}

// sendData 发送数据，带超时和错误处理
func (p *Producer) sendData(data int) error {
	select {
	case p.dataCh <- data:
		log.Printf("生产者 %d 发送数据: %d", p.id, data)
		return nil
	case <-p.ctx.Done():
		return context.Canceled
	case <-time.After(1 * time.Second):
		return context.DeadlineExceeded
	}
}

// Start 启动消费者
func (c *Consumer) Start() {
	c.wg.Add(1)
	go c.consume()
}

// Stop 停止消费者 - 完全修复版本
func (c *Consumer) Stop() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	c.mu.Unlock()

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
		log.Printf("消费者已优雅停止，处理了 %d 条数据", stats.processedCount)
	case <-time.After(5 * time.Second):
		stats := c.GetStats()
		log.Printf("消费者停止超时，已处理 %d 条数据", stats.processedCount)
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

// consume 消费数据 - 完全修复版本
func (c *Consumer) consume() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("消费者发生panic: %v", r)
			atomic.AddInt64(&c.stats.errorCount, 1)
		}
		c.wg.Done()
		log.Println("消费者已停止")
	}()

	for {
		select {
		case <-c.ctx.Done():
			log.Println("消费者收到停止信号")
			return
		case data, ok := <-c.dataCh:
			if !ok {
				log.Println("消费者数据通道已关闭")
				return
			}

			// 处理数据，带超时控制
			if err := c.processData(data); err != nil {
				log.Printf("消费者处理数据失败: %v", err)
				atomic.AddInt64(&c.stats.errorCount, 1)
			} else {
				atomic.AddInt64(&c.stats.processedCount, 1)
			}
		}
	}
}

// processData 处理数据，带超时控制 - 完全修复版本
func (c *Consumer) processData(data int) error {
	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(c.ctx, 3*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		defer close(done)
		// 模拟处理时间
		time.Sleep(100 * time.Millisecond)
		log.Printf("消费者处理数据: %d", data)
		done <- nil
	}()

	select {
	case <-ctx.Done():
		atomic.AddInt64(&c.stats.timeoutCount, 1)
		return ctx.Err()
	case err := <-done:
		return err
	}
}

func main() {
	// 创建根上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建共享数据通道
	dataCh := make(chan int, 200)

	// 创建多个生产者
	producers := make([]*Producer, 3)
	for i := 0; i < 3; i++ {
		producers[i] = NewProducer(i+1, ctx, dataCh)
		producers[i].Start()
	}

	// 创建单个消费者
	consumer := NewConsumer(ctx, dataCh)
	consumer.Start()

	// 运行一段时间后停止
	time.Sleep(5 * time.Second)
	log.Println("开始停止所有协程...")

	// 停止所有生产者
	for _, producer := range producers {
		producer.Stop()
	}

	// 安全关闭数据通道
	safeCloseDataChannel(dataCh)

	// 停止消费者
	consumer.Stop()

	// 打印统计信息
	log.Println("=== 统计信息 ===")
	for _, producer := range producers {
		stats := producer.GetStats()
		log.Printf("生产者 %d: 生产 %d 条数据, 错误 %d 次, 运行时间 %v",
			producer.id, stats.producedCount, stats.errorCount, time.Since(stats.startTime))
	}

	consumerStats := consumer.GetStats()
	log.Printf("消费者: 处理 %d 条数据, 错误 %d 次, 超时 %d 次, 运行时间 %v",
		consumerStats.processedCount, consumerStats.errorCount, consumerStats.timeoutCount,
		time.Since(consumerStats.startTime))

	log.Println("所有协程已停止")
}

// safeCloseDataChannel 安全关闭数据通道
func safeCloseDataChannel(ch chan int) {
	var closeOnce sync.Once
	closeOnce.Do(func() {
		close(ch)
	})
}
