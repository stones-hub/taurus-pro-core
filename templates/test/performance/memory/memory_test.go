package memory

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

	// 获取项目根目录路径
	projectRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取当前目录失败: %v", err)
	}
	projectRoot = projectRoot + "/../../.." // 从 test/performance/memory 回到项目根目录

	// 切换到项目根目录
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("切换到项目根目录失败: %v", err)
	}

	// 创建临时目录用于存储堆快照
	snapshotDir, err := os.MkdirTemp("", "heap_snapshots")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(snapshotDir) // 清理临时目录

	// 使用 make build 编译程序
	buildCmd := exec.Command("make", "build")
	buildCmd.Dir = projectRoot // 设置工作目录为项目根目录
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("编译程序失败: %v", err)
	}

	// 运行编译后的程序
	cmd := exec.Command("./build/taurus", "-c", "config", "-e", ".env.local")
	cmd.Dir = projectRoot

	// 直接重定向到标准输出和标准错误
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 启动进程
	if err := cmd.Start(); err != nil {
		t.Fatalf("启动应用失败: %v", err)
	}

	// 等待应用启动
	if err := waitForAppReady(t, 30*time.Second); err != nil {
		t.Fatalf("等待应用启动失败: %v", err)
	}

	// 确保程序在测试结束时关闭
	defer func() {
		// 首先尝试优雅关闭
		if err := cmd.Process.Signal(os.Interrupt); err != nil {
			t.Logf("发送中断信号失败: %v", err)
		}

		// 使用channel来控制超时
		done := make(chan error, 1)
		go func() {
			// 当进程退出以后，会返回一个错误，给done赋值
			done <- cmd.Wait()
		}()

		// 等待进程退出，设置超时
		select {
		case err := <-done:
			if err != nil {
				t.Logf("进程退出: %v", err)
			}
		case <-time.After(30 * time.Second):
			t.Log("等待进程退出超时，强制终止")
			if err := cmd.Process.Kill(); err != nil {
				t.Errorf("强制终止进程失败: %v", err)
			}
		}

		// 清理堆快照文件
		files, err := filepath.Glob(filepath.Join(snapshotDir, "heap_*.prof"))
		if err == nil {
			for _, f := range files {
				os.Remove(f)
			}
		}
	}()

	// 创建上下文和负载生成器
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
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
			snapshotFile := filepath.Join(snapshotDir, fmt.Sprintf("heap_%s.prof", time.Now().Format("150405")))
			heapCmd := exec.Command("go", "tool", "pprof", "-proto",
				"http://localhost:6060/debug/pprof/heap",
				snapshotFile)

			if err := heapCmd.Run(); err != nil {
				t.Logf("获取堆快照失败: %v", err)
			}
		}
	}
}

// 实现健康检查
func waitForAppReady(t *testing.T, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		// 尝试连接应用健康检查接口
		resp, err := http.Get("http://localhost:9080/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("应用未在%v内启动完成", timeout)
}
