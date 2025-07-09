package generator

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
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

	// 确保至少有 100ms 的间隔
	ticker := time.NewTicker(time.Duration(rand.Int63n(900)+100) * time.Millisecond)
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
		// 调用首页API
		func() {
			resp, err := http.Get("http://localhost:9080/")
			if err != nil {
				log.Printf("调用首页API失败: %v", err)
				return
			}
			defer resp.Body.Close()
			log.Printf("调用首页API成功")
		},

		// 测试内存分配
		func() {
			size := rand.Intn(10) + 1 // 1-10MB
			resp, err := http.Get(fmt.Sprintf("http://localhost:9080/memory/allocate?size=%d", size))
			if err != nil {
				log.Printf("内存分配测试失败: %v", err)
				return
			}
			defer resp.Body.Close()
			log.Printf("内存分配测试成功")
		},

		// 测试内存泄漏
		func() {
			count := rand.Intn(1000) + 100 // 100-1100个对象
			resp, err := http.Get(fmt.Sprintf("http://localhost:9080/memory/leak?count=%d", count))
			if err != nil {
				log.Printf("内存泄漏测试失败: %v", err)
				return
			}
			defer resp.Body.Close()
			log.Printf("内存泄漏测试成功")
		},

		// 偶尔释放内存（模拟GC）
		func() {
			if rand.Float32() < 0.1 { // 10%的概率执行释放
				resp, err := http.Get("http://localhost:9080/memory/free")
				if err != nil {
					log.Printf("内存释放失败: %v", err)
					return
				}
				defer resp.Body.Close()
				log.Printf("内存释放成功")
			}
		},
	}

	op := operations[rand.Intn(len(operations))]
	op()
}
