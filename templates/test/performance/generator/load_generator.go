package generator

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"
)

// LoadGenerator 负载生成器
type LoadGenerator struct {
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewLoadGenerator 创建新的负载生成器
func NewLoadGenerator() *LoadGenerator {
	return &LoadGenerator{
		stopCh: make(chan struct{}),
	}
}

// Start 开始生成负载
func (g *LoadGenerator) Start(ctx context.Context) {
	// 模拟多个并发用户
	for i := 0; i < 100; i++ {
		g.wg.Add(1)
		go g.simulateUser(ctx)
	}
}

// Stop 停止负载生成
func (g *LoadGenerator) Stop() {
	close(g.stopCh)
	g.wg.Wait()
}

// simulateUser 模拟用户行为
func (g *LoadGenerator) simulateUser(ctx context.Context) {
	defer g.wg.Done()

	ticker := time.NewTicker(time.Duration(rand.Int63n(1000)) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-g.stopCh:
			return
		case <-ticker.C:
			// 模拟各种操作
			g.simulateOperations()
		}
	}
}

// simulateOperations 模拟各种操作
func (g *LoadGenerator) simulateOperations() {
	operations := []func(){
		func() { time.Sleep(time.Duration(rand.Int63n(100)) * time.Millisecond) }, // 模拟 API 调用
		func() { _ = make([]byte, rand.Intn(1024)) },                              // 模拟内存分配
		func() { log.Printf("模拟日志输出 %d", rand.Int()) },                            // 模拟日志输出
	}

	op := operations[rand.Intn(len(operations))]
	op()
}
