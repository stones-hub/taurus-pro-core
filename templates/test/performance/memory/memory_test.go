package memory

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"testing"
	"time"

	"{{.ProjectName}}/test/performance/generator"
)

// MemoryStats 存储内存统计信息
type MemoryStats struct {
	Alloc         uint64
	TotalAlloc    uint64
	Sys           uint64
	NumGC         uint32
	NumGoroutine  int
	HeapObjects   uint64
	GCCPUFraction float64
	PauseTotalNs  uint64
}

func collectMemoryStats() MemoryStats {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	return MemoryStats{
		Alloc:         stats.Alloc,
		TotalAlloc:    stats.TotalAlloc,
		Sys:           stats.Sys,
		NumGC:         stats.NumGC,
		NumGoroutine:  runtime.NumGoroutine(),
		HeapObjects:   stats.HeapObjects,
		GCCPUFraction: stats.GCCPUFraction,
		PauseTotalNs:  stats.PauseTotalNs,
	}
}

func TestMemoryLeak(t *testing.T) {
	// 设置GC参数以获得更好的监控效果
	debug.SetGCPercent(10)

	// 强制进行 GC
	runtime.GC()

	// 记录初始内存状态
	initialStats := collectMemoryStats()
	t.Logf("初始状态:")
	t.Logf("- 内存使用: %d MB", initialStats.Alloc/1024/1024)
	t.Logf("- 系统内存: %d MB", initialStats.Sys/1024/1024)
	t.Logf("- Goroutine数量: %d", initialStats.NumGoroutine)
	t.Logf("- 堆对象数量: %d", initialStats.HeapObjects)

	// 启动应用
	cmd := exec.Command("go", "run", "bin/taurus.go")
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

	var (
		maxMemory uint64
		samples   []MemoryStats
		lastGC    uint32
	)

	t.Log("开始内存泄漏测试...")
	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			// 测试结束，进行最终检查
			runtime.GC()
			finalStats := collectMemoryStats()

			// 计算关键指标
			duration := time.Since(startTime)
			allocRate := float64(finalStats.TotalAlloc-initialStats.TotalAlloc) / duration.Seconds() / 1024 / 1024
			gcRate := float64(finalStats.NumGC-initialStats.NumGC) / duration.Seconds()
			avgPauseTime := float64(finalStats.PauseTotalNs-initialStats.PauseTotalNs) / float64(finalStats.NumGC-initialStats.NumGC) / 1000000 // 转换为毫秒

			// 输出详细报告
			t.Log("\n=== 测试报告 ===")
			t.Logf("测试持续时间: %v", duration)
			t.Log("\n1. 内存使用情况:")
			t.Logf("- 最终内存使用: %d MB", finalStats.Alloc/1024/1024)
			t.Logf("- 内存增长量: %d MB", (finalStats.Alloc-initialStats.Alloc)/1024/1024)
			t.Logf("- 最大内存使用: %d MB", maxMemory/1024/1024)
			t.Logf("- 系统内存使用: %d MB", finalStats.Sys/1024/1024)
			t.Logf("- 内存分配速率: %.2f MB/s", allocRate)

			t.Log("\n2. GC 统计:")
			t.Logf("- GC次数: %d", finalStats.NumGC-initialStats.NumGC)
			t.Logf("- GC频率: %.2f 次/秒", gcRate)
			t.Logf("- 平均GC暂停时间: %.2f ms", avgPauseTime)
			t.Logf("- GC CPU占用比例: %.2f%%", finalStats.GCCPUFraction*100)

			t.Log("\n3. 其他指标:")
			t.Logf("- Goroutine数量变化: %d -> %d", initialStats.NumGoroutine, finalStats.NumGoroutine)
			t.Logf("- 堆对象数量变化: %d -> %d", initialStats.HeapObjects, finalStats.HeapObjects)

			// 分析内存增长趋势
			if len(samples) > 0 {
				growthRate := (finalStats.Alloc - samples[0].Alloc) / uint64(len(samples))
				t.Logf("\n4. 内存增长分析:")
				t.Logf("- 平均每次采样内存增长: %d KB", growthRate/1024)

				// 检测内存泄漏
				if growthRate > 1024*1024 { // 1MB
					t.Errorf("警告: 检测到可能的内存泄漏")
					t.Errorf("- 平均每次采样增长: %d KB", growthRate/1024)
				}
			}
			return

		case <-ticker.C:
			runtime.GC()
			currentStats := collectMemoryStats()
			samples = append(samples, currentStats)

			if currentStats.Alloc > maxMemory {
				maxMemory = currentStats.Alloc
			}

			// 输出实时监控数据
			t.Logf("\n=== %s ===", time.Now().Format("15:04:05"))
			t.Logf("内存使用: %d MB", currentStats.Alloc/1024/1024)
			t.Logf("Goroutines: %d", currentStats.NumGoroutine)
			t.Logf("堆对象: %d", currentStats.HeapObjects)

			// 检查是否发生了新的GC
			if currentStats.NumGC > lastGC {
				gcCount := currentStats.NumGC - lastGC
				t.Logf("发生GC: %d次", gcCount)
				lastGC = currentStats.NumGC
			}

			// 检查单次内存增长是否异常
			if len(samples) > 1 {
				lastSample := samples[len(samples)-2]
				growth := currentStats.Alloc - lastSample.Alloc
				if growth > 10*1024*1024 { // 10MB
					t.Logf("警告: 检测到大幅内存增长: %d MB", growth/1024/1024)
				}
			}

			// 获取堆快照
			snapshotFile := fmt.Sprintf("heap_%s.prof", time.Now().Format("150405"))
			if err := exec.Command("go", "tool", "pprof", "-proto",
				"http://localhost:6060/debug/pprof/heap",
				snapshotFile).Run(); err != nil {
				t.Logf("获取堆快照失败: %v", err)
			}
		}
	}
}
