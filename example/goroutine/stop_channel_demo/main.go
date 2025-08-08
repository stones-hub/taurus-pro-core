package main

import (
	"context"
	"log"
	"sync"
	"time"
)

// Worker 工作协程
type Worker struct {
	id     int
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	stopCh chan struct{}
}

func NewWorker(id int, ctx context.Context) *Worker {
	ctx, cancel := context.WithCancel(ctx)
	return &Worker{
		id:     id,
		ctx:    ctx,
		cancel: cancel,
		stopCh: make(chan struct{}),
	}
}

func (w *Worker) Start() {
	w.wg.Add(1)
	go w.work()
}

func (w *Worker) Stop() {
	log.Printf("Worker %d: 发送专属停止信号", w.id)
	close(w.stopCh)
	w.wg.Wait()
}

func (w *Worker) StopByContext() {
	log.Printf("Worker %d: 通过Context停止", w.id)
	w.cancel()
	w.wg.Wait()
}

func (w *Worker) work() {
	defer func() {
		w.wg.Done()
		log.Printf("Worker %d: 协程结束", w.id)
	}()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			log.Printf("Worker %d: 收到专属停止信号", w.id)
			return
		case <-w.ctx.Done():
			log.Printf("Worker %d: 收到Context停止信号", w.id)
			return
		case <-ticker.C:
			log.Printf("Worker %d: 正在工作...", w.id)
		}
	}
}

func main() {
	log.Println("=== 演示 stopCh 和 Context 的区别 ===")

	// 场景1: 使用 stopCh 精确停止特定Worker
	log.Println("\n--- 场景1: 使用 stopCh 精确停止 ---")
	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()

	worker1 := NewWorker(1, ctx1)
	worker2 := NewWorker(2, ctx1)
	worker3 := NewWorker(3, ctx1)

	worker1.Start()
	worker2.Start()
	worker3.Start()

	// 让它们工作一段时间
	time.Sleep(2 * time.Second)

	// 只停止 worker2
	log.Println("只停止 worker2...")
	worker2.Stop()

	// 其他worker继续工作
	time.Sleep(2 * time.Second)

	// 停止所有
	worker1.Stop()
	worker3.Stop()

	// 场景2: 使用 Context 停止所有Worker
	log.Println("\n--- 场景2: 使用 Context 停止所有 ---")
	ctx2, cancel2 := context.WithCancel(context.Background())

	worker4 := NewWorker(4, ctx2)
	worker5 := NewWorker(5, ctx2)
	worker6 := NewWorker(6, ctx2)

	worker4.Start()
	worker5.Start()
	worker6.Start()

	// 让它们工作一段时间
	time.Sleep(2 * time.Second)

	// 通过Context停止所有
	log.Println("通过Context停止所有worker...")
	cancel2()

	// 等待所有worker停止
	time.Sleep(1 * time.Second)

	// 场景3: 演示Context超时
	log.Println("\n--- 场景3: Context超时停止 ---")
	ctx3, cancel3 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel3()

	worker7 := NewWorker(7, ctx3)
	worker7.Start()

	// 等待Context超时
	time.Sleep(4 * time.Second)

	log.Println("=== 演示结束 ===")
}
