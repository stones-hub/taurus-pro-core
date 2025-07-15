package profile

import (
	"log"
	"runtime"
	"testing"
	"time"
)

// MemoryLeakDetector 内存泄漏检测器
type MemoryLeakDetector struct {
	initialMemory  float64
	initialObjects int64
	checkInterval  time.Duration
	threshold      float64 // 内存增长阈值百分比
}

// NewMemoryLeakDetector 创建内存泄漏检测器
func NewMemoryLeakDetector(checkInterval time.Duration, threshold float64) *MemoryLeakDetector {
	return &MemoryLeakDetector{
		checkInterval: checkInterval,
		threshold:     threshold,
	}
}

// Start 开始检测
func (d *MemoryLeakDetector) Start(duration time.Duration) {
	log.Println("=== 开始内存泄漏检测 ===")
	log.Printf("检测配置: 间隔=%v, 时长=%v, 阈值=%.1f%%", d.checkInterval, duration, d.threshold)

	// 预热系统
	log.Println("预热系统...")
	time.Sleep(5 * time.Second)

	// 记录初始状态
	d.recordInitialState()

	// 开始监控
	startTime := time.Now()
	ticker := time.NewTicker(d.checkInterval)
	defer ticker.Stop()

	var checkCount int
	for time.Since(startTime) < duration {
		select {
		case <-ticker.C:
			checkCount++
			d.checkMemoryLeak(checkCount)
		}
	}

	// 最终分析
	d.finalAnalysis()
}

// recordInitialState 记录初始状态
func (d *MemoryLeakDetector) recordInitialState() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	d.initialMemory = float64(memStats.HeapInuse) / 1024 / 1024
	d.initialObjects = int64(memStats.HeapObjects)

	log.Printf("初始状态: 内存=%.2f MB, 对象数=%d", d.initialMemory, d.initialObjects)
}

// checkMemoryLeak 检查内存泄漏
func (d *MemoryLeakDetector) checkMemoryLeak(checkCount int) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	currentMemory := float64(memStats.HeapInuse) / 1024 / 1024
	currentObjects := int64(memStats.HeapObjects)

	memoryChange := currentMemory - d.initialMemory
	memoryChangePercent := (memoryChange / d.initialMemory) * 100
	objectChange := currentObjects - d.initialObjects

	log.Printf("检查 #%d: 内存=%.2f MB (变化: %.2f MB, %.1f%%), 对象数=%d (变化: %d)",
		checkCount, currentMemory, memoryChange, memoryChangePercent, currentObjects, objectChange)

	// 检查是否超过阈值
	if memoryChangePercent > d.threshold {
		log.Printf("⚠️ 警告: 内存增长超过阈值 %.1f%%", d.threshold)
		d.analyzeMemoryGrowth(memStats)
	}

	// 检查对象数量异常增长
	if objectChange > 10000 {
		log.Printf("⚠️ 警告: 对象数量异常增长 %d", objectChange)
	}
}

// analyzeMemoryGrowth 分析内存增长
func (d *MemoryLeakDetector) analyzeMemoryGrowth(memStats runtime.MemStats) {
	log.Println("=== 内存增长分析 ===")

	// 分析内存分配情况
	allocRate := float64(memStats.TotalAlloc) / 1024 / 1024
	sysRate := float64(memStats.Sys) / 1024 / 1024

	log.Printf("总分配: %.2f MB", allocRate)
	log.Printf("系统内存: %.2f MB", sysRate)
	log.Printf("堆内存: %.2f MB", float64(memStats.HeapInuse)/1024/1024)
	log.Printf("空闲内存: %.2f MB", float64(memStats.HeapIdle)/1024/1024)

	// 分析GC情况
	if memStats.NumGC > 0 {
		gcFraction := memStats.GCCPUFraction * 100
		log.Printf("GC次数: %d, GC CPU占比: %.2f%%", memStats.NumGC, gcFraction)

		if gcFraction > 10 {
			log.Printf("⚠️ GC负载过高，可能存在内存压力")
		}
	}

	// 分析内存碎片
	if memStats.HeapInuse > 0 {
		fragmentation := float64(memStats.HeapInuse-memStats.HeapAlloc) / float64(memStats.HeapInuse) * 100
		log.Printf("内存碎片率: %.2f%%", fragmentation)

		if fragmentation > 30 {
			log.Printf("⚠️ 内存碎片化严重")
		}
	}
}

// finalAnalysis 最终分析
func (d *MemoryLeakDetector) finalAnalysis() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	finalMemory := float64(memStats.HeapInuse) / 1024 / 1024
	finalObjects := int64(memStats.HeapObjects)

	totalMemoryChange := finalMemory - d.initialMemory
	totalMemoryChangePercent := (totalMemoryChange / d.initialMemory) * 100
	totalObjectChange := finalObjects - d.initialObjects

	log.Println("=== 最终分析 ===")
	log.Printf("检测完成")
	log.Printf("起始内存: %.2f MB", d.initialMemory)
	log.Printf("结束内存: %.2f MB", finalMemory)
	log.Printf("内存变化: %.2f MB (%.1f%%)", totalMemoryChange, totalMemoryChangePercent)
	log.Printf("对象变化: %d", totalObjectChange)

	// 评估结果
	if totalMemoryChangePercent > d.threshold {
		log.Printf("🚨 检测到内存泄漏风险!")
		log.Printf("   内存增长 %.1f%% 超过阈值 %.1f%%", totalMemoryChangePercent, d.threshold)
	} else if totalMemoryChangePercent > 5 {
		log.Printf("⚠️ 内存有轻微增长 (%.1f%%)", totalMemoryChangePercent)
	} else {
		log.Printf("✅ 内存使用稳定 (变化: %.1f%%)", totalMemoryChangePercent)
	}

	if totalObjectChange > 10000 {
		log.Printf("⚠️ 对象数量显著增加 (%d)", totalObjectChange)
	} else if totalObjectChange > 0 {
		log.Printf("📊 对象数量增加 (%d)", totalObjectChange)
	} else {
		log.Printf("✅ 对象数量稳定")
	}
}

// TestMemoryLeakDetection 内存泄漏和对象增长监控测试
func TestMemoryLeakDetection(t *testing.T) {
	// 创建检测器
	detector := NewMemoryLeakDetector(10*time.Second, 10.0) // 10秒间隔，10%阈值

	// 开始检测，持续1分钟
	detector.Start(1 * time.Minute)
}
