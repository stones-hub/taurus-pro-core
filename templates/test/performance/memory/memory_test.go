package memory

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"{{.ProjectName}}/test/performance/generator"
)

func TestMemoryLeak(t *testing.T) {
	// 强制进行 GC
	runtime.GC()

	// 记录初始内存状态
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	initialAlloc := stats.Alloc
	t.Logf("初始内存使用: %d MB", initialAlloc/1024/1024)

	// 启动应用
	cmd := exec.Command("go", "run", "../../../bin/taurus.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("启动应用失败: %v", err)
	}

	// 确保程序在测试结束时关闭
	defer func() {
		if err := cmd.Process.Kill(); err != nil {
			t.Errorf("关闭应用失败: %v", err)
		}
	}()

	// 等待应用启动
	time.Sleep(3 * time.Second)

	// 创建上下文和负载生成器
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	g := generator.NewLoadGenerator()
	g.Start(ctx)
	defer g.Stop()

	// 每10秒记录一次内存使用情况
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var maxMemory uint64
	var samples []uint64

	t.Log("开始内存泄漏测试...")
	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			// 测试结束，进行最终检查
			runtime.GC()
			runtime.ReadMemStats(&stats)
			finalAlloc := stats.Alloc

			t.Logf("测试完成，耗时: %v", time.Since(startTime))
			t.Logf("最终内存使用: %d MB", finalAlloc/1024/1024)
			t.Logf("内存增长: %d MB", (finalAlloc-initialAlloc)/1024/1024)
			t.Logf("最大内存使用: %d MB", maxMemory/1024/1024)

			// 检查内存增长趋势
			if len(samples) > 0 {
				growthRate := (finalAlloc - samples[0]) / uint64(len(samples))
				t.Logf("平均每次采样内存增长: %d KB", growthRate/1024)

				// 如果内存持续增长超过阈值，报告可能的泄漏
				if growthRate > 1024*1024 { // 1MB
					t.Errorf("检测到可能的内存泄漏: 平均每次采样增长 %d KB", growthRate/1024)
				}
			}
			return

		case <-ticker.C:
			runtime.GC()
			runtime.ReadMemStats(&stats)
			currentAlloc := stats.Alloc
			samples = append(samples, currentAlloc)

			if currentAlloc > maxMemory {
				maxMemory = currentAlloc
			}

			t.Logf("当前内存使用: %d MB", currentAlloc/1024/1024)

			// 检查单次内存增长是否异常
			if len(samples) > 1 {
				lastSample := samples[len(samples)-2]
				growth := currentAlloc - lastSample
				if growth > 10*1024*1024 { // 10MB
					t.Logf("警告: 检测到大幅内存增长: %d MB", growth/1024/1024)
				}
			}

			// 尝试获取堆快照
			snapshotFile := fmt.Sprintf("heap_%s.prof", time.Now().Format("150405"))
			if err := exec.Command("go", "tool", "pprof", "-proto",
				"http://localhost:6060/debug/pprof/heap",
				snapshotFile).Run(); err != nil {
				t.Logf("获取堆快照失败: %v", err)
			}
		}
	}
}
