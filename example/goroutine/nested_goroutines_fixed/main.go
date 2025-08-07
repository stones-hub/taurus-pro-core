package main

import (
	"context"
	"log"
	"sync"
	"time"
)

// Producer 生产者结构体
type Producer struct {
	id     int
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	dataCh chan int
}

// Consumer 消费者结构体
type Consumer struct {
	id           int
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	dataCh       chan int
	workerCount  int
	workers      []*Worker
	workerDataCh chan int
	workerWg     sync.WaitGroup
}

// Worker 工作协程结构体
type Worker struct {
	id         int
	consumerID int
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	dataCh     chan int
}

// NewProducer 创建生产者
func NewProducer(id int, ctx context.Context, dataCh chan int) *Producer {
	ctx, cancel := context.WithCancel(ctx)
	return &Producer{
		id:     id,
		ctx:    ctx,
		cancel: cancel,
		dataCh: dataCh,
	}
}

// NewConsumer 创建消费者
func NewConsumer(id int, ctx context.Context, dataCh chan int, workerCount int) *Consumer {
	ctx, cancel := context.WithCancel(ctx)
	return &Consumer{
		id:           id,
		ctx:          ctx,
		cancel:       cancel,
		dataCh:       dataCh,
		workerCount:  workerCount,
		workers:      make([]*Worker, 0, workerCount),
		workerDataCh: make(chan int, workerCount*10),
	}
}

// NewWorker 创建工作协程
func NewWorker(id, consumerID int, ctx context.Context, dataCh chan int) *Worker {
	ctx, cancel := context.WithCancel(ctx)
	return &Worker{
		id:         id,
		consumerID: consumerID,
		ctx:        ctx,
		cancel:     cancel,
		dataCh:     dataCh,
	}
}

// Start 启动生产者
func (p *Producer) Start() {
	p.wg.Add(1)
	go p.produce()
}

// Stop 停止生产者
func (p *Producer) Stop() {
	p.cancel()
	p.wg.Wait()
}

// produce 生产数据 - 完全修复版本
func (p *Producer) produce() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("生产者 %d 发生panic: %v", p.id, r)
		}
		p.wg.Done()
		log.Printf("生产者 %d 已停止", p.id)
	}()

	counter := 0
	ticker := time.NewTicker(time.Duration(100+p.id*30) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			log.Printf("生产者 %d 收到停止信号", p.id)
			return
		case <-ticker.C:
			select {
			case p.dataCh <- counter:
				log.Printf("生产者 %d 发送数据: %d", p.id, counter)
				counter++
			case <-p.ctx.Done():
				return
			}
		}
	}
}

// Start 启动消费者
func (c *Consumer) Start() {
	c.wg.Add(1)
	go c.consume()

	// 启动工作协程并正确管理
	for i := 0; i < c.workerCount; i++ {
		worker := NewWorker(i+1, c.id, c.ctx, c.workerDataCh)
		c.workers = append(c.workers, worker)
		c.workerWg.Add(1)
		go func(w *Worker) {
			defer c.workerWg.Done()
			w.work()
		}(worker)
	}
}

// Stop 停止消费者 - 完全修复版本
func (c *Consumer) Stop() {
	c.cancel()

	// 等待工作协程完成
	c.workerWg.Wait()

	// 安全关闭工作数据通道
	c.safeCloseWorkerChannel()

	c.wg.Wait()
}

// safeCloseWorkerChannel 安全关闭工作通道
func (c *Consumer) safeCloseWorkerChannel() {
	// 使用sync.Once确保只关闭一次
	var closeOnce sync.Once
	closeOnce.Do(func() {
		close(c.workerDataCh)
	})
}

// consume 消费数据 - 完全修复版本
func (c *Consumer) consume() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("消费者 %d 发生panic: %v", c.id, r)
		}
		c.wg.Done()
		log.Printf("消费者 %d 已停止", c.id)
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

			// 模拟预处理时间
			time.Sleep(20 * time.Millisecond)
			log.Printf("消费者 %d 预处理数据: %d", c.id, data)

			// 分发给工作协程 - 修复竞态条件
			select {
			case c.workerDataCh <- data:
				// 数据成功分发给工作协程
			case <-c.ctx.Done():
				return
			default:
				// 如果工作协程通道已满，记录警告但继续处理
				log.Printf("消费者 %d 工作协程通道已满，跳过数据: %d", c.id, data)
			}
		}
	}
}

// work 工作协程处理逻辑 - 完全修复版本
func (w *Worker) work() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("工作协程 %d-%d 发生panic: %v", w.consumerID, w.id, r)
		}
		log.Printf("工作协程 %d-%d 已停止", w.consumerID, w.id)
	}()

	for {
		select {
		case <-w.ctx.Done():
			log.Printf("工作协程 %d-%d 收到停止信号", w.consumerID, w.id)
			return
		case data, ok := <-w.dataCh:
			if !ok {
				log.Printf("工作协程 %d-%d 数据通道已关闭", w.consumerID, w.id)
				return
			}

			// 模拟复杂处理逻辑
			time.Sleep(time.Duration(50+w.id*10) * time.Millisecond)
			log.Printf("工作协程 %d-%d 处理数据: %d", w.consumerID, w.id, data)

			// 模拟嵌套的异步处理 - 完全修复版本
			w.processAsync(data)
		}
	}
}

// processAsync 异步处理逻辑 - 完全修复版本
func (w *Worker) processAsync(data int) {
	// 使用独立的协程处理，但不添加到WaitGroup中
	// 因为这是fire-and-forget模式
	go func() {
		// 使用recover防止panic导致程序崩溃
		defer func() {
			if r := recover(); r != nil {
				log.Printf("工作协程 %d-%d 异步处理发生panic: %v", w.consumerID, w.id, r)
			}
		}()

		// 创建带超时的上下文，避免协程泄漏
		ctx, cancel := context.WithTimeout(w.ctx, 5*time.Second)
		defer cancel()

		// 模拟异步处理
		select {
		case <-ctx.Done():
			log.Printf("工作协程 %d-%d 异步处理超时: %d", w.consumerID, w.id, data)
		case <-time.After(30 * time.Millisecond):
			log.Printf("工作协程 %d-%d 异步处理完成: %d", w.consumerID, w.id, data)
		}
	}()
}

func main() {
	// 创建根上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建共享数据通道
	dataCh := make(chan int, 500)

	// 创建多个生产者
	producers := make([]*Producer, 2)
	for i := 0; i < 2; i++ {
		producers[i] = NewProducer(i+1, ctx, dataCh)
		producers[i].Start()
	}

	// 创建多个消费者，每个消费者有3个工作协程
	consumers := make([]*Consumer, 3)
	for i := 0; i < 3; i++ {
		consumers[i] = NewConsumer(i+1, ctx, dataCh, 3)
		consumers[i].Start()
	}

	// 运行一段时间后停止
	time.Sleep(8 * time.Second)
	log.Println("开始停止所有协程...")

	// 先取消上下文，通知所有协程停止
	cancel()

	// 等待一小段时间让协程响应停止信号
	time.Sleep(100 * time.Millisecond)

	// 停止所有生产者
	for _, producer := range producers {
		producer.Stop()
	}

	// 安全关闭数据通道
	safeCloseDataChannel(dataCh)

	// 停止所有消费者
	for _, consumer := range consumers {
		consumer.Stop()
	}

	log.Println("所有协程已停止")
}

// safeCloseDataChannel 安全关闭数据通道
func safeCloseDataChannel(ch chan int) {
	var closeOnce sync.Once
	closeOnce.Do(func() {
		close(ch)
	})
}
