package profile

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// 采样配置
const (
	sampleInterval = 2 * time.Second // 采样间隔
	sampleDuration = 2 * time.Minute // 总采样时间
)

// 基础指标结构
type BaseMetrics struct {
	Timestamp time.Time `json:"timestamp"`
}

// 内存指标
type MemoryMetrics struct {
	BaseMetrics
	HeapAlloc          float64          `json:"heap_alloc"`          // 已分配的堆内存(MB)
	HeapInUse          float64          `json:"heap_in_use"`         // 正在使用的堆内存(MB)
	HeapIdle           float64          `json:"heap_idle"`           // 空闲的堆内存(MB)
	HeapReleased       float64          `json:"heap_released"`       // 释放回操作系统的内存(MB)
	HeapObjects        int64            `json:"heap_objects"`        // 堆对象数量
	AllocBySize        map[string]int64 `json:"alloc_by_size"`       // 按大小分类的分配
	LargeObjects       []ObjectInfo     `json:"large_objects"`       // 大对象列表(>1MB)
	FragmentationRatio float64          `json:"fragmentation_ratio"` // 内存碎片率
	AllocationHotspots []string         `json:"allocation_hotspots"` // 内存分配热点
}

// GC指标
type GCMetrics struct {
	BaseMetrics
	NumGC       int64   `json:"num_gc"`       // GC次数
	PauseNs     int64   `json:"pause_ns"`     // GC暂停时间
	CPUFraction float64 `json:"cpu_fraction"` // GC CPU占用比例
	NextGC      int64   `json:"next_gc"`      // 下次GC阈值
	ForcedGC    int64   `json:"forced_gc"`    // 强制GC次数
}

// Goroutine指标
type GoroutineMetrics struct {
	BaseMetrics
	Total    int64            `json:"total"`     // 总数
	Blocked  int64            `json:"blocked"`   // 阻塞数
	IOWait   int64            `json:"io_wait"`   // IO等待
	Running  int64            `json:"running"`   // 运行中
	States   map[string]int64 `json:"states"`    // 各种状态的数量
	HotSpots []string         `json:"hot_spots"` // 热点位置(top 5)
}

// CPU指标
type CPUMetrics struct {
	BaseMetrics
	UsagePercent float64 `json:"usage_percent"` // CPU使用率
	ThreadCount  int     `json:"thread_count"`  // 线程数
}

// 函数性能指标
type FunctionMetric struct {
	Name      string  `json:"name"`       // 函数名
	FileLine  string  `json:"file_line"`  // 文件和行号
	CPUTime   float64 `json:"cpu_time"`   // CPU时间
	CallCount int64   `json:"call_count"` // 调用次数
}

// 对象信息
type ObjectInfo struct {
	Type     string `json:"type"`     // 对象类型
	Size     int64  `json:"size"`     // 对象大小
	Count    int64  `json:"count"`    // 对象数量
	Location string `json:"location"` // 分配位置
}

// 完整的性能分析报告
type ProfilingReport struct {
	StartTime  time.Time           `json:"start_time"`
	EndTime    time.Time           `json:"end_time"`
	Memory     []*MemoryMetrics    `json:"memory"`     // 内存采样数据
	GC         []*GCMetrics        `json:"gc"`         // GC采样数据
	Goroutines []*GoroutineMetrics `json:"goroutines"` // Goroutine采样数据
	CPU        []*CPUMetrics       `json:"cpu"`        // CPU采样数据
	Trends     *TrendAnalysis      `json:"trends"`     // 趋势分析
}

// 趋势分析
type TrendAnalysis struct {
	Memory struct {
		Start     float64 `json:"start"`      // 起始内存
		End       float64 `json:"end"`        // 结束内存
		Peak      float64 `json:"peak"`       // 峰值内存
		AvgGrowth float64 `json:"avg_growth"` // 平均增长率
	} `json:"memory"`
	GC struct {
		AvgPause    float64 `json:"avg_pause"`    // 平均暂停时间
		MaxPause    float64 `json:"max_pause"`    // 最大暂停时间
		PauseCount  int64   `json:"pause_count"`  // 暂停次数
		AvgInterval float64 `json:"avg_interval"` // 平均间隔
	} `json:"gc"`
	Goroutines struct {
		MinCount  int64   `json:"min_count"`  // 最小数量
		MaxCount  int64   `json:"max_count"`  // 最大数量
		AvgCount  float64 `json:"avg_count"`  // 平均数量
		BlockRate float64 `json:"block_rate"` // 阻塞率
	} `json:"goroutines"`
	CPU struct {
		AvgUsage     float64 `json:"avg_usage"`      // 平均使用率
		PeakUsage    float64 `json:"peak_usage"`     // 峰值使用率
		UserSysRatio float64 `json:"user_sys_ratio"` // 用户态/系统态比例
	} `json:"cpu"`
}

// 从pprof获取原始数据
func fetchPprofData(endpoint string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:6060%s", endpoint))
	if err != nil {
		return nil, fmt.Errorf("获取pprof数据失败: %v", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 打印原始数据
	// log.Printf("pprof原始数据 (%s):\n%s\n", endpoint, string(data))
	return data, nil
}

// 收集内存指标
func collectMemoryMetrics() (*MemoryMetrics, error) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics := &MemoryMetrics{
		BaseMetrics: BaseMetrics{
			Timestamp: time.Now(),
		},
		HeapAlloc:    float64(memStats.HeapAlloc) / 1024 / 1024, // 转换为MB
		HeapInUse:    float64(memStats.HeapInuse) / 1024 / 1024,
		HeapIdle:     float64(memStats.HeapIdle) / 1024 / 1024,
		HeapReleased: float64(memStats.HeapReleased) / 1024 / 1024,
		HeapObjects:  int64(memStats.HeapObjects),
	}

	// 计算内存碎片率
	if memStats.HeapInuse > 0 {
		metrics.FragmentationRatio = float64(memStats.HeapInuse-memStats.HeapAlloc) / float64(memStats.HeapInuse)
	}

	// 获取内存消耗的top信息
	heapData, err := fetchPprofData("/debug/pprof/heap?debug=1")
	if err != nil {
		return metrics, fmt.Errorf("获取堆profile失败: %v", err)
	}

	// 解析pprof输出，提取所有内存消耗top信息
	lines := strings.Split(string(heapData), "\n")
	var hotspots []string
	var largeObjects []ObjectInfo
	var suspiciousAllocations []string               // 可疑的内存分配
	var functionAllocations = make(map[string]int64) // 按函数统计分配

	// 添加调试信息，显示前几行原始数据
	log.Printf("调试: pprof heap数据前10行:")
	for i, line := range lines {
		if i < 10 {
			log.Printf("  行%d: %s", i+1, line)
		}
	}

	// 统计解析到的数据
	var parsedCount int
	var totalParsedSize int64

	for _, line := range lines {
		// 跳过空行和注释行
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 跳过头部信息行
		if strings.HasPrefix(line, "heap profile:") {
			continue
		}

		// 解析内存分配记录
		// 格式: 0: 0 [431: 4067426304] @ 0x106806636 ...
		// 注意：第一个数字是采样计数，第二个是分配计数，方括号内是对象数和总大小
		if strings.Contains(line, "[") && strings.Contains(line, "] @") {
			parts := strings.Split(line, "[")
			if len(parts) >= 2 {
				allocPart := strings.Split(parts[1], "]")[0]
				allocInfo := strings.Split(allocPart, ":")
				if len(allocInfo) >= 2 {
					objectCount, _ := strconv.ParseInt(strings.TrimSpace(allocInfo[0]), 10, 64)
					totalSize, _ := strconv.ParseInt(strings.TrimSpace(allocInfo[1]), 10, 64)

					// 计算平均对象大小
					var avgSize int64
					if objectCount > 0 {
						avgSize = totalSize / objectCount
					}

					// 添加调试信息，显示解析的原始数据
					if parsedCount < 5 {
						log.Printf("调试: 解析行 '%s' -> 对象数=%d, 总大小=%d, 平均大小=%d", line, objectCount, totalSize, avgSize)
					}

					parsedCount++
					totalParsedSize += totalSize

					// 提取函数调用栈信息
					var functionName string
					if idx := strings.Index(line, "@"); idx != -1 {
						stack := strings.TrimSpace(line[idx+1:])
						if parts := strings.Fields(stack); len(parts) > 0 {
							functionName = parts[0]

							// 过滤掉明显不是函数名的地址
							if !strings.HasPrefix(functionName, "0x") &&
								!strings.HasPrefix(functionName, "heap/") &&
								len(functionName) > 3 {
								functionAllocations[functionName] += totalSize
							} else {
								// 对于地址，尝试提取更多信息
								if len(parts) > 1 {
									// 尝试找到更具体的函数名
									for _, part := range parts[1:] {
										if !strings.HasPrefix(part, "0x") &&
											!strings.HasPrefix(part, "heap/") &&
											len(part) > 3 {
											functionAllocations[part] += totalSize
											break
										}
									}
								}
							}
						}
					}

					// 记录大对象(>1MB)
					if totalSize > 1024*1024 {
						largeObjects = append(largeObjects, ObjectInfo{
							Type:     fmt.Sprintf("Object_%d", len(largeObjects)+1),
							Size:     avgSize,
							Count:    objectCount,
							Location: line,
						})
					}

					// 检测可疑的内存分配模式
					if objectCount > 1000 && avgSize > 1024 {
						suspiciousAllocations = append(suspiciousAllocations,
							fmt.Sprintf("大量分配: %d个对象, 每个%d字节, 总计%.2f MB - %s",
								objectCount, avgSize, float64(totalSize)/1024/1024, functionName))
					}

					// 检测字符串相关的大量分配
					if strings.Contains(functionName, "string") && objectCount > 500 {
						suspiciousAllocations = append(suspiciousAllocations,
							fmt.Sprintf("字符串大量分配: %d个, 总计%.2f MB - %s",
								objectCount, float64(totalSize)/1024/1024, functionName))
					}

					// 检测切片相关的大量分配
					if strings.Contains(functionName, "slice") || strings.Contains(functionName, "append") {
						if objectCount > 300 {
							suspiciousAllocations = append(suspiciousAllocations,
								fmt.Sprintf("切片大量分配: %d个, 总计%.2f MB - %s",
									objectCount, float64(totalSize)/1024/1024, functionName))
						}
					}

					// 收集所有热点信息（不限制于应用代码）
					hotspots = append(hotspots, fmt.Sprintf("Size: %d bytes, Count: %d, Total: %.2f MB - %s",
						avgSize, objectCount, float64(totalSize)/1024/1024, line))
				}
			}
		}
	}

	// 分析函数级别的内存分配
	var topFunctions []string
	for funcName, totalSize := range functionAllocations {
		// 添加调试信息
		if totalSize > 1024*1024 { // 超过1MB的函数
			log.Printf("调试: 函数 %s 累计分配 %d 字节 (%.2f MB)", funcName, totalSize, float64(totalSize)/1024/1024)
		}

		if totalSize > 10*1024*1024 { // 超过10MB的函数
			topFunctions = append(topFunctions,
				fmt.Sprintf("函数 %s 分配了 %.2f MB", funcName, float64(totalSize)/1024/1024))
		}
	}

	// 如果有可疑分配，记录到日志
	if len(suspiciousAllocations) > 0 {
		log.Printf("⚠️ 发现可疑内存分配模式:")
		for i, allocation := range suspiciousAllocations {
			if i < 10 { // 只显示前10个
				log.Printf("  %d. %s", i+1, allocation)
			}
		}
	}

	// 输出解析统计信息
	log.Printf("调试: 解析了 %d 条分配记录，总大小 %.2f MB", parsedCount, float64(totalParsedSize)/1024/1024)
	log.Printf("调试: runtime.MemStats显示: HeapAlloc=%.2f MB, HeapInUse=%.2f MB",
		float64(memStats.HeapAlloc)/1024/1024, float64(memStats.HeapInuse)/1024/1024)

	// 分析内存回收情况
	if totalParsedSize > 0 {
		recoveryRate := float64(totalParsedSize-int64(memStats.HeapAlloc)) / float64(totalParsedSize) * 100
		log.Printf("♻️ 内存回收情况: 累计分配 %.2f MB, 当前使用 %.2f MB, 回收率 %.1f%%",
			float64(totalParsedSize)/1024/1024, float64(memStats.HeapAlloc)/1024/1024, recoveryRate)

		if recoveryRate > 90 {
			log.Printf("✅ GC工作正常，内存回收效率很高")
		} else if recoveryRate < 50 {
			log.Printf("⚠️ 内存回收率较低，可能存在内存泄漏")
		}
	}

	// 如果有大量分配的函数，记录到日志
	if len(topFunctions) > 0 {
		log.Printf("🔍 内存分配热点函数:")
		for i, funcInfo := range topFunctions {
			if i < 5 { // 只显示前5个
				log.Printf("  %d. %s", i+1, funcInfo)
			}
		}
	} else {
		log.Printf("调试: 没有找到超过10MB的函数分配")
	}

	// 按内存占用排序大对象
	sort.Slice(largeObjects, func(i, j int) bool {
		return largeObjects[i].Size*largeObjects[i].Count > largeObjects[j].Size*largeObjects[j].Count
	})

	// 只保留top 10的大对象
	if len(largeObjects) > 10 {
		largeObjects = largeObjects[:10]
	}

	// 只保留top 20的热点信息
	if len(hotspots) > 20 {
		hotspots = hotspots[:20]
	}

	metrics.LargeObjects = largeObjects
	metrics.AllocationHotspots = hotspots

	// 简单的内存泄漏检测
	if memStats.HeapObjects > 100000 { // 对象数量超过10万
		log.Printf("🚨 警告: 堆对象数量过多 (%d), 可能存在内存泄漏", memStats.HeapObjects)

		// 分析对象类型分布
		var objectTypes = make(map[string]int64)
		for _, line := range lines {
			if strings.Contains(line, "[") && strings.Contains(line, "] @") {
				// 尝试提取对象类型信息
				if idx := strings.Index(line, "@"); idx != -1 {
					stack := strings.TrimSpace(line[idx+1:])
					if parts := strings.Fields(stack); len(parts) > 0 {
						funcName := parts[0]
						// 根据函数名推测对象类型
						switch {
						case strings.Contains(funcName, "string"):
							objectTypes["string"]++
						case strings.Contains(funcName, "slice") || strings.Contains(funcName, "array"):
							objectTypes["slice/array"]++
						case strings.Contains(funcName, "map"):
							objectTypes["map"]++
						case strings.Contains(funcName, "struct"):
							objectTypes["struct"]++
						case strings.Contains(funcName, "channel"):
							objectTypes["channel"]++
						default:
							objectTypes["other"]++
						}
					}
				}
			}
		}

		// 输出对象类型分布
		log.Printf("📊 对象类型分布:")
		for objType, count := range objectTypes {
			if count > 1000 {
				log.Printf("  %s: %d 个对象", objType, count)
			}
		}
	}

	// 检测内存碎片化
	if memStats.HeapInuse > 0 {
		fragmentation := float64(memStats.HeapInuse-memStats.HeapAlloc) / float64(memStats.HeapInuse)
		if fragmentation > 0.3 { // 碎片率超过30%
			log.Printf("⚠️ 内存碎片化严重: %.2f%%, 建议进行GC", fragmentation*100)
		}
	}

	return metrics, nil
}

// 收集GC指标
func collectGCMetrics() (*GCMetrics, error) {
	metrics := &GCMetrics{
		BaseMetrics: BaseMetrics{Timestamp: time.Now()},
	}

	// 从pprof heap profile中获取GC信息
	heapData, err := fetchPprofData("/debug/pprof/heap?debug=1")
	if err != nil {
		return nil, fmt.Errorf("获取堆profile失败: %v", err)
	}

	// 解析pprof输出中的GC信息
	lines := strings.Split(string(heapData), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// GC信息是以#开头的注释行
		if !strings.HasPrefix(line, "# ") {
			continue
		}

		// 去掉#前缀
		line = strings.TrimPrefix(line, "# ")

		switch {
		case strings.HasPrefix(line, "NumGC = "):
			if num, err := strconv.ParseInt(strings.Fields(line)[2], 10, 64); err == nil {
				metrics.NumGC = num
			}
		case strings.HasPrefix(line, "GCCPUFraction = "):
			if fraction, err := strconv.ParseFloat(strings.Fields(line)[2], 64); err == nil {
				metrics.CPUFraction = fraction
			}
		case strings.HasPrefix(line, "NextGC = "):
			if next, err := strconv.ParseInt(strings.Fields(line)[2], 10, 64); err == nil {
				metrics.NextGC = next
			}
		case strings.HasPrefix(line, "NumForcedGC = "):
			if forced, err := strconv.ParseInt(strings.Fields(line)[2], 10, 64); err == nil {
				metrics.ForcedGC = forced
			}
		case strings.HasPrefix(line, "LastGC = "):
			// 可以解析LastGC时间戳，但这里我们主要关注其他指标
			// 暂时跳过LastGC解析
			continue
		}
	}

	// 添加调试信息
	// log.Printf("GC调试信息: NumGC=%d, CPUFraction=%.6f, NextGC=%d, ForcedGC=%d", metrics.NumGC, metrics.CPUFraction, metrics.NextGC, metrics.ForcedGC)

	return metrics, nil
}

// 收集Goroutine指标
func collectGoroutineMetrics() (*GoroutineMetrics, error) {
	metrics := &GoroutineMetrics{
		BaseMetrics: BaseMetrics{Timestamp: time.Now()},
		States:      make(map[string]int64),
	}

	// 获取goroutine profile
	gorData, err := fetchPprofData("/debug/pprof/goroutine?debug=2")
	if err != nil {
		return nil, fmt.Errorf("获取goroutine profile失败: %v", err)
	}

	gorStr := string(gorData)
	goroutines := strings.Split(gorStr, "\n\ngoroutine ")

	// 计算总数（跳过第一个空白部分）
	metrics.Total = int64(len(goroutines) - 1)

	// 用于统计创建位置
	creationLocations := make(map[string]int)

	for _, g := range goroutines[1:] {
		lines := strings.Split(g, "\n")
		if len(lines) == 0 {
			continue
		}

		// 解析goroutine状态 - 格式: goroutine 52016 [running]:
		stateMatch := regexp.MustCompile(`\[(.*?)\]:`).FindStringSubmatch(lines[0])
		if len(stateMatch) > 1 {
			state := stateMatch[1]
			metrics.States[state]++

			// 更详细的状态分类
			switch {
			case strings.Contains(state, "IO wait"):
				metrics.IOWait++
			case strings.Contains(state, "running"):
				metrics.Running++
			case strings.Contains(state, "runnable"):
				metrics.Running++
			case strings.Contains(state, "select"):
				metrics.Blocked++
			case strings.Contains(state, "chan receive") ||
				strings.Contains(state, "chan send"):
				metrics.Blocked++
			case strings.Contains(state, "semacquire") ||
				strings.Contains(state, "semrelease"):
				metrics.Blocked++
			case strings.Contains(state, "sleep"):
				metrics.Blocked++
			case strings.Contains(state, "syscall"):
				metrics.IOWait++
			case strings.Contains(state, "sync.WaitGroup.Wait"):
				metrics.Blocked++
			default:
				// 其他状态归类为阻塞
				metrics.Blocked++
			}
		}

		// 提取创建位置和上下文
		var createdBy string
		var creationStack []string
		inCreationStack := false

		for i, line := range lines {
			if strings.Contains(line, "created by") {
				inCreationStack = true
				createdBy = strings.TrimSpace(line)
				continue
			}

			if inCreationStack && i < len(lines) && len(strings.TrimSpace(line)) > 0 {
				// 添加文件和行号信息
				creationStack = append(creationStack, strings.TrimSpace(line))
			}
		}

		if len(creationStack) > 0 {
			// 组合创建位置信息
			location := createdBy
			if len(creationStack) > 0 {
				location += " at " + creationStack[0]
			}
			creationLocations[location]++
		}
	}

	// 获取top 5热点创建位置
	type hotSpot struct {
		loc   string
		count int
	}
	var hotSpots []hotSpot
	for loc, count := range creationLocations {
		hotSpots = append(hotSpots, hotSpot{loc, count})
	}
	sort.Slice(hotSpots, func(i, j int) bool {
		return hotSpots[i].count > hotSpots[j].count
	})

	// 只保留前5个热点
	for i := 0; i < len(hotSpots) && i < 5; i++ {
		metrics.HotSpots = append(metrics.HotSpots, fmt.Sprintf("%s (count: %d)",
			hotSpots[i].loc, hotSpots[i].count))
	}

	// 添加调试信息
	// log.Printf("Goroutine调试信息: Total=%d, Running=%d, IOWait=%d, Blocked=%d", metrics.Total, metrics.Running, metrics.IOWait, metrics.Blocked)

	return metrics, nil
}

// 收集CPU指标
func collectCPUMetrics() (*CPUMetrics, error) {
	metrics := &CPUMetrics{
		BaseMetrics: BaseMetrics{Timestamp: time.Now()},
	}

	// 获取线程数
	metrics.ThreadCount = runtime.NumCPU()

	// 获取CPU使用率 - 使用更准确的方法
	// 方法1: 使用runtime包获取基本信息
	goroutineCount := runtime.NumGoroutine()

	// 方法2: 尝试从pprof获取一些系统信息
	// 注意：CPU profile是二进制格式，我们无法直接解析
	// 但我们可以获取一些系统级别的信息

	// 获取内存统计作为负载指标的一部分
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 综合多个指标来估算CPU使用率
	// 1. Goroutine数量
	// 2. 内存使用情况
	// 3. GC活动

	gcLoad := memStats.GCCPUFraction * 100                                    // GC CPU占比
	memLoad := float64(memStats.HeapInuse) / float64(memStats.HeapSys) * 50   // 内存使用占比
	goroutineLoad := float64(goroutineCount) / float64(runtime.NumCPU()) * 20 // Goroutine负载

	// 综合CPU使用率估算
	cpuUsage := gcLoad + memLoad + goroutineLoad
	if cpuUsage > 100 {
		cpuUsage = 100
	}

	metrics.UsagePercent = cpuUsage

	// 添加调试信息
	// log.Printf("CPU调试信息: UsagePercent=%.2f%%, ThreadCount=%d, GoroutineCount=%d, GCLoad=%.2f%%, MemLoad=%.2f%%, GoroutineLoad=%.2f%%", metrics.UsagePercent, metrics.ThreadCount, goroutineCount, gcLoad, memLoad, goroutineLoad)

	return metrics, nil
}

// 分析趋势
func analyzeTrends(memMetrics []*MemoryMetrics, gcMetrics []*GCMetrics,
	gorMetrics []*GoroutineMetrics, cpuMetrics []*CPUMetrics) *TrendAnalysis {

	trends := &TrendAnalysis{}

	// 分析内存趋势
	if len(memMetrics) > 0 {
		start := memMetrics[0]
		end := memMetrics[len(memMetrics)-1]

		trends.Memory.Start = start.HeapInUse
		trends.Memory.End = end.HeapInUse

		// 计算内存变化
		memoryChange := end.HeapInUse - start.HeapInUse
		memoryChangePercent := 0.0
		if start.HeapInUse > 0 {
			memoryChangePercent = (memoryChange / start.HeapInUse) * 100
		}

		// 找出峰值
		var peak float64
		var totalGrowth float64
		for i := 1; i < len(memMetrics); i++ {
			growth := memMetrics[i].HeapInUse - memMetrics[i-1].HeapInUse
			totalGrowth += growth
			if memMetrics[i].HeapInUse > peak {
				peak = memMetrics[i].HeapInUse
			}
		}
		trends.Memory.Peak = peak

		if len(memMetrics) > 1 {
			trends.Memory.AvgGrowth = totalGrowth / float64(len(memMetrics)-1)
		}

		// 内存趋势分析总结
		log.Printf("=== 内存趋势分析 ===")
		log.Printf("采样次数: %d", len(memMetrics))
		log.Printf("起始内存: %.2f MB", start.HeapInUse)
		log.Printf("结束内存: %.2f MB", end.HeapInUse)
		log.Printf("内存变化: %.2f MB (%.2f%%)", memoryChange, memoryChangePercent)
		log.Printf("峰值内存: %.2f MB", peak)
		log.Printf("平均增长: %.2f MB/采样", trends.Memory.AvgGrowth)

		// 显示前几次采样的内存变化
		if len(memMetrics) >= 3 {
			log.Printf("内存变化详情:")
			for i := 0; i < len(memMetrics) && i < 5; i++ {
				log.Printf("  采样%d: %.2f MB", i+1, memMetrics[i].HeapInUse)
			}
		}

		// 内存健康度评估 - 更合理的判断逻辑
		if len(memMetrics) >= 3 {
			// 跳过前两次采样，使用稳定后的数据
			stableStart := memMetrics[2].HeapInUse
			stableEnd := end.HeapInUse
			stableChange := stableEnd - stableStart
			stableChangePercent := 0.0
			if stableStart > 0 {
				stableChangePercent = (stableChange / stableStart) * 100
			}

			log.Printf("稳定后内存变化: %.2f MB (%.2f%%)", stableChange, stableChangePercent)

			if stableChangePercent > 20 {
				log.Printf("⚠️  警告: 稳定后内存增长超过20%%，可能存在内存泄漏")
			} else if stableChangePercent > 10 {
				log.Printf("⚠️  注意: 稳定后内存增长超过10%%，需要关注")
			} else if stableChangePercent < -10 {
				log.Printf("✅ 良好: 内存使用减少，GC工作正常")
			} else {
				log.Printf("✅ 稳定: 内存使用相对稳定")
			}
		} else {
			// 如果采样次数不足，使用原始逻辑但调整阈值
			if memoryChangePercent > 50 {
				log.Printf("⚠️  注意: 内存增长较大，但可能是初始化阶段")
			} else if memoryChangePercent > 20 {
				log.Printf("⚠️  注意: 内存增长超过20%%，需要关注")
			} else {
				log.Printf("✅ 稳定: 内存使用相对稳定")
			}
		}
	}

	// 分析GC趋势
	if len(gcMetrics) > 0 {
		start := gcMetrics[0]
		end := gcMetrics[len(gcMetrics)-1]

		// 计算GC活动变化
		gcCountChange := end.NumGC - start.NumGC
		gcCpuChange := end.CPUFraction - start.CPUFraction

		var totalPause float64
		var maxPause float64
		for _, gc := range gcMetrics {
			pause := float64(gc.PauseNs) / 1000000 // 转换为毫秒
			totalPause += pause
			if pause > maxPause {
				maxPause = pause
			}
		}

		trends.GC.MaxPause = maxPause
		trends.GC.PauseCount = int64(len(gcMetrics))

		if trends.GC.PauseCount > 0 {
			trends.GC.AvgPause = totalPause / float64(trends.GC.PauseCount)
			trends.GC.AvgInterval = sampleDuration.Seconds() / float64(trends.GC.PauseCount)
		}

		// GC趋势分析总结
		log.Printf("=== GC趋势分析 ===")
		log.Printf("起始GC次数: %d", start.NumGC)
		log.Printf("结束GC次数: %d", end.NumGC)
		log.Printf("GC次数变化: %d", gcCountChange)
		log.Printf("起始GC CPU占比: %.4f", start.CPUFraction)
		log.Printf("结束GC CPU占比: %.4f", end.CPUFraction)
		log.Printf("GC CPU占比变化: %.4f", gcCpuChange)
		log.Printf("平均暂停时间: %.2f ms", trends.GC.AvgPause)
		log.Printf("最大暂停时间: %.2f ms", trends.GC.MaxPause)

		// GC健康度评估
		if trends.GC.MaxPause > 100 {
			log.Printf("⚠️  警告: GC暂停时间过长(>100ms)，可能影响性能")
		} else if trends.GC.MaxPause > 50 {
			log.Printf("⚠️  注意: GC暂停时间较长(>50ms)")
		} else {
			log.Printf("✅ 良好: GC暂停时间正常")
		}

		if gcCpuChange > 0.01 {
			log.Printf("⚠️  注意: GC CPU占比增加，可能存在内存压力")
		}
	}

	// 分析Goroutine趋势
	if len(gorMetrics) > 0 {
		start := gorMetrics[0]
		end := gorMetrics[len(gorMetrics)-1]

		trends.Goroutines.MinCount = start.Total
		trends.Goroutines.MaxCount = start.Total

		var totalCount, totalBlocked int64
		for _, g := range gorMetrics {
			totalCount += g.Total
			totalBlocked += g.Blocked
			if g.Total < trends.Goroutines.MinCount {
				trends.Goroutines.MinCount = g.Total
			}
			if g.Total > trends.Goroutines.MaxCount {
				trends.Goroutines.MaxCount = g.Total
			}
		}

		trends.Goroutines.AvgCount = float64(totalCount) / float64(len(gorMetrics))
		if totalCount > 0 {
			trends.Goroutines.BlockRate = float64(totalBlocked) / float64(totalCount)
		}

		// Goroutine趋势分析总结
		log.Printf("=== Goroutine趋势分析 ===")
		log.Printf("起始Goroutine数: %d", start.Total)
		log.Printf("结束Goroutine数: %d", end.Total)
		log.Printf("Goroutine变化: %d", end.Total-start.Total)
		log.Printf("最小数量: %d", trends.Goroutines.MinCount)
		log.Printf("最大数量: %d", trends.Goroutines.MaxCount)
		log.Printf("平均数量: %.2f", trends.Goroutines.AvgCount)
		log.Printf("阻塞率: %.2f%%", trends.Goroutines.BlockRate*100)

		// Goroutine健康度评估
		goroutineChange := end.Total - start.Total
		if goroutineChange > 100 {
			log.Printf("⚠️  警告: Goroutine数量大幅增加，可能存在goroutine泄漏")
		} else if goroutineChange > 50 {
			log.Printf("⚠️  注意: Goroutine数量增加较多")
		} else if goroutineChange < -50 {
			log.Printf("✅ 良好: Goroutine数量减少，资源释放正常")
		} else {
			log.Printf("✅ 稳定: Goroutine数量相对稳定")
		}

		if trends.Goroutines.BlockRate > 0.3 {
			log.Printf("⚠️  警告: Goroutine阻塞率过高(>30%%)")
		} else if trends.Goroutines.BlockRate > 0.1 {
			log.Printf("⚠️  注意: Goroutine阻塞率较高(>10%%)")
		} else {
			log.Printf("✅ 良好: Goroutine阻塞率正常")
		}
	}

	// 分析CPU趋势
	if len(cpuMetrics) > 0 {
		start := cpuMetrics[0]
		end := cpuMetrics[len(cpuMetrics)-1]

		var totalUsage float64
		for _, cpu := range cpuMetrics {
			totalUsage += cpu.UsagePercent
			if cpu.UsagePercent > trends.CPU.PeakUsage {
				trends.CPU.PeakUsage = cpu.UsagePercent
			}
		}

		trends.CPU.AvgUsage = totalUsage / float64(len(cpuMetrics))
		trends.CPU.UserSysRatio = end.UsagePercent / 100 // 简化的用户态/系统态比例

		// CPU趋势分析总结
		log.Printf("=== CPU趋势分析 ===")
		log.Printf("起始CPU使用率: %.2f%%", start.UsagePercent)
		log.Printf("结束CPU使用率: %.2f%%", end.UsagePercent)
		log.Printf("CPU使用率变化: %.2f%%", end.UsagePercent-start.UsagePercent)
		log.Printf("平均使用率: %.2f%%", trends.CPU.AvgUsage)
		log.Printf("峰值使用率: %.2f%%", trends.CPU.PeakUsage)
		log.Printf("线程数: %d", start.ThreadCount)

		// CPU健康度评估
		cpuChange := end.UsagePercent - start.UsagePercent
		if trends.CPU.PeakUsage > 80 {
			log.Printf("⚠️  警告: CPU使用率峰值过高(>80%%)")
		} else if trends.CPU.AvgUsage > 60 {
			log.Printf("⚠️  注意: CPU平均使用率较高(>60%%)")
		} else {
			log.Printf("✅ 良好: CPU使用率正常")
		}

		if cpuChange > 20 {
			log.Printf("⚠️  注意: CPU使用率显著增加")
		} else if cpuChange < -20 {
			log.Printf("✅ 良好: CPU使用率减少")
		}
	}

	// 整体系统健康度评估
	log.Printf("=== 整体系统健康度评估 ===")
	var issues []string
	var goodPoints []string

	// 检查各项指标
	if len(memMetrics) > 0 {
		start := memMetrics[0]
		end := memMetrics[len(memMetrics)-1]
		memoryChangePercent := (end.HeapInUse - start.HeapInUse) / start.HeapInUse * 100
		if memoryChangePercent > 20 {
			issues = append(issues, "内存泄漏风险")
		} else {
			goodPoints = append(goodPoints, "内存使用稳定")
		}
	}

	if len(gcMetrics) > 0 && trends.GC.MaxPause > 100 {
		issues = append(issues, "GC暂停时间过长")
	} else if len(gcMetrics) > 0 {
		goodPoints = append(goodPoints, "GC性能良好")
	}

	if len(gorMetrics) > 0 && trends.Goroutines.BlockRate > 0.3 {
		issues = append(issues, "Goroutine阻塞严重")
	} else if len(gorMetrics) > 0 {
		goodPoints = append(goodPoints, "Goroutine运行正常")
	}

	if len(cpuMetrics) > 0 && trends.CPU.PeakUsage > 80 {
		issues = append(issues, "CPU负载过高")
	} else if len(cpuMetrics) > 0 {
		goodPoints = append(goodPoints, "CPU使用合理")
	}

	if len(issues) > 0 {
		log.Printf("⚠️  发现的问题:")
		for _, issue := range issues {
			log.Printf("  - %s", issue)
		}
	}

	if len(goodPoints) > 0 {
		log.Printf("✅ 系统优点:")
		for _, point := range goodPoints {
			log.Printf("  - %s", point)
		}
	}

	if len(issues) == 0 {
		log.Printf("🎉 系统运行状态良好，无明显性能问题")
	}

	return trends
}

// 保存报告到文件
func saveReport(report *ProfilingReport) error {
	reportDir := "reports"
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return fmt.Errorf("创建报告目录失败: %v", err)
	}

	reportFile := filepath.Join(reportDir,
		fmt.Sprintf("profile_%s.json", time.Now().Format("20060102_150405")))

	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化报告失败: %v", err)
	}

	if err := os.WriteFile(reportFile, reportJSON, 0644); err != nil {
		return fmt.Errorf("保存报告失败: %v", err)
	}

	log.Printf("报告已保存: %s", reportFile)
	return nil
}

// 记录内存突增
func logMemorySpike(memMetrics *MemoryMetrics, gcMetrics *GCMetrics) error {
	if memMetrics.HeapObjects <= 10000 {
		return nil
	}

	// 获取详细的pprof数据来解析文件路径
	heapData, err := fetchPprofData("/debug/pprof/heap?debug=1")
	if err != nil {
		return fmt.Errorf("获取堆profile失败: %v", err)
	}

	// 解析pprof数据获取文件路径信息
	fileLocations := parseFileLocations(string(heapData))

	details := fmt.Sprintf("内存突增详情 (时间: %s):\n",
		memMetrics.Timestamp.Format("2006-01-02 15:04:05"))
	details += "=== 内存状态 ===\n"
	details += fmt.Sprintf("- 堆内存使用: %.2f MB\n", memMetrics.HeapInUse)
	details += fmt.Sprintf("- 堆内存分配: %.2f MB\n", memMetrics.HeapAlloc)
	details += fmt.Sprintf("- 对象数量: %d\n", memMetrics.HeapObjects)
	details += fmt.Sprintf("- 内存碎片率: %.2f%%\n", memMetrics.FragmentationRatio*100)

	// 添加GC状态信息
	if gcMetrics != nil {
		details += "\n=== GC状态 ===\n"
		details += fmt.Sprintf("- GC次数: %d\n", gcMetrics.NumGC)
		details += fmt.Sprintf("- GC CPU占比: %.4f%%\n", gcMetrics.CPUFraction*100)
		details += fmt.Sprintf("- 下次GC阈值: %d\n", gcMetrics.NextGC)
		details += fmt.Sprintf("- 强制GC次数: %d\n", gcMetrics.ForcedGC)

		// GC健康度评估
		if gcMetrics.CPUFraction > 0.1 {
			details += fmt.Sprintf("- GC状态: 高负载 (CPU占比%.2f%%)\n", gcMetrics.CPUFraction*100)
		} else if gcMetrics.CPUFraction > 0.05 {
			details += fmt.Sprintf("- GC状态: 中等负载 (CPU占比%.2f%%)\n", gcMetrics.CPUFraction*100)
		} else {
			details += fmt.Sprintf("- GC状态: 正常 (CPU占比%.2f%%)\n", gcMetrics.CPUFraction*100)
		}
	}

	// 添加大对象信息（包含文件路径）
	if len(memMetrics.LargeObjects) > 0 {
		details += "\n=== 大对象 (>1MB) ===\n"
		for i, obj := range memMetrics.LargeObjects {
			if i >= 5 { // 只显示前5个
				break
			}

			// 尝试从Location中提取文件路径信息
			fileInfo := extractFileInfo(obj.Location, fileLocations)

			details += fmt.Sprintf("- 对象%d: %d个, 每个%d字节, 总计%.2f MB\n",
				i+1, obj.Count, obj.Size, float64(obj.Count*obj.Size)/1024/1024)
			if fileInfo != "" {
				details += fmt.Sprintf("  位置: %s\n", fileInfo)
			}
		}
	}

	if len(memMetrics.AllocationHotspots) > 0 {
		details += "\n=== 内存分配热点 ===\n"
		for i, hotspot := range memMetrics.AllocationHotspots {
			if i >= 10 { // 只显示前10个
				break
			}
			details += fmt.Sprintf("- %s\n", hotspot)
		}
	}

	// 添加文件位置分析
	if len(fileLocations) > 0 {
		details += "\n=== 文件位置分析 ===\n"
		topFiles := getTopFileLocations(fileLocations, 10)
		for _, fileInfo := range topFiles {
			details += fmt.Sprintf("- %s\n", fileInfo)
		}
	}

	// 添加分析建议
	details += "\n=== 分析建议 ===\n"
	if memMetrics.HeapObjects > 100000 {
		details += "- 对象数量过多，可能存在对象泄漏\n"
	}
	if memMetrics.FragmentationRatio > 0.3 {
		details += "- 内存碎片化严重，建议手动触发GC\n"
	}
	if gcMetrics != nil && gcMetrics.CPUFraction > 0.1 {
		details += "- GC负载过高，可能存在内存压力\n"
	}
	if len(memMetrics.LargeObjects) > 0 {
		details += "- 存在大对象分配，检查是否合理\n"
	}

	// 使用append方式记录到单个日志文件
	logDir := "reports"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	logFile := filepath.Join(logDir, "memory_spike.log")

	// 添加分隔线
	details += "\n" + strings.Repeat("=", 80) + "\n\n"

	// 以append模式写入文件
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString(details); err != nil {
		return fmt.Errorf("写入日志失败: %v", err)
	}

	log.Printf("内存突增日志已追加到: %s", logFile)
	return nil
}

// 解析文件位置信息
func parseFileLocations(pprofData string) map[string]string {
	fileLocations := make(map[string]string)
	lines := strings.Split(pprofData, "\n")

	for _, line := range lines {
		// 查找包含文件路径的行
		if strings.Contains(line, ".go:") {
			// 格式通常是: 函数名 文件路径:行号
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				// 查找包含.go:的部分
				for _, part := range parts {
					if strings.Contains(part, ".go:") {
						// 提取函数名和文件路径
						if len(parts) > 0 {
							funcName := parts[0]
							fileLocations[funcName] = part
						}
						break
					}
				}
			}
		}
	}

	return fileLocations
}

// 从Location中提取文件信息
func extractFileInfo(location string, fileLocations map[string]string) string {
	// 从location中提取函数名
	if idx := strings.Index(location, "@"); idx != -1 {
		stack := strings.TrimSpace(location[idx+1:])
		if parts := strings.Fields(stack); len(parts) > 0 {
			funcName := parts[0]
			if filePath, exists := fileLocations[funcName]; exists {
				return fmt.Sprintf("%s -> %s", funcName, filePath)
			}
			return funcName
		}
	}
	return ""
}

// 获取top文件位置
func getTopFileLocations(fileLocations map[string]string, limit int) []string {
	type fileInfo struct {
		file  string
		count int
	}

	fileCounts := make(map[string]int)
	for _, filePath := range fileLocations {
		fileCounts[filePath]++
	}

	var files []fileInfo
	for file, count := range fileCounts {
		files = append(files, fileInfo{file, count})
	}

	// 按计数排序
	sort.Slice(files, func(i, j int) bool {
		return files[i].count > files[j].count
	})

	var result []string
	for i := 0; i < len(files) && i < limit; i++ {
		result = append(result, fmt.Sprintf("%s (引用次数: %d)", files[i].file, files[i].count))
	}

	return result
}

func TestSystemProfile(t *testing.T) {
	// 检查pprof服务
	resp, err := http.Get("http://localhost:6060/debug/pprof/")
	if err != nil {
		t.Fatal("pprof HTTP服务未启动:", err)
	}
	defer resp.Body.Close()

	// 验证响应
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Types of profiles available") {
		t.Fatal("pprof服务响应异常")
	}

	t.Log("=== 开始系统性能分析 ===")
	t.Logf("采样配置: 间隔=%v, 时长=%v", sampleInterval, sampleDuration)

	startTime := time.Now()
	var memoryMetrics []*MemoryMetrics
	var gcMetrics []*GCMetrics
	var goroutineMetrics []*GoroutineMetrics
	var cpuMetrics []*CPUMetrics

	// 持续采样
	sampleCount := 0
	for time.Since(startTime) < sampleDuration {
		sampleCount++
		t.Logf("\n[采样 #%d]", sampleCount)

		// 1. 收集内存指标
		t.Log("收集内存指标...")
		if metrics, err := collectMemoryMetrics(); err != nil {
			t.Logf("警告: 收集内存指标失败: %v", err)
		} else {
			memoryMetrics = append(memoryMetrics, metrics)
			t.Logf("- 堆内存: %.2f MB", metrics.HeapInUse)
			t.Logf("- 对象数: %d", metrics.HeapObjects)

			// 检查内存突增
			if metrics.HeapObjects > 10000 {
				if gc, err := collectGCMetrics(); err == nil {
					if err := logMemorySpike(metrics, gc); err != nil {
						t.Logf("警告: 记录内存突增失败: %v", err)
					}
				}
			}
		}

		// 2. 收集GC指标
		t.Log("收集GC指标...")
		if metrics, err := collectGCMetrics(); err != nil {
			t.Logf("警告: 收集GC指标失败: %v", err)
		} else {
			gcMetrics = append(gcMetrics, metrics)
			t.Logf("- GC次数: %d", metrics.NumGC)
			t.Logf("- GC CPU占比: %.2f%%", metrics.CPUFraction*100)
		}

		// 3. 收集Goroutine指标
		t.Log("收集Goroutine指标...")
		if metrics, err := collectGoroutineMetrics(); err != nil {
			t.Logf("警告: 收集Goroutine指标失败: %v", err)
		} else {
			goroutineMetrics = append(goroutineMetrics, metrics)
			t.Logf("- 总数: %d (运行=%d, IO等待=%d, 阻塞=%d)",
				metrics.Total, metrics.Running,
				metrics.IOWait, metrics.Blocked)
		}

		// 4. 收集CPU指标
		t.Log("收集CPU指标...")
		if metrics, err := collectCPUMetrics(); err != nil {
			t.Logf("警告: 收集CPU指标失败: %v", err)
		} else {
			cpuMetrics = append(cpuMetrics, metrics)
			// 输出CPU指标
			t.Logf("- CPU使用率: %.2f%%", metrics.UsagePercent)
			t.Logf("- 线程数: %d", metrics.ThreadCount)
		}

		time.Sleep(sampleInterval)
	}

	// 生成报告
	report := &ProfilingReport{
		StartTime:  startTime,
		EndTime:    time.Now(),
		Memory:     memoryMetrics,
		GC:         gcMetrics,
		Goroutines: goroutineMetrics,
		CPU:        cpuMetrics,
	}

	// 分析趋势
	report.Trends = analyzeTrends(report.Memory, report.GC, report.Goroutines, report.CPU)

	// 保存报告
	if err := saveReport(report); err != nil {
		t.Fatalf("保存报告失败: %v", err)
	}

	// 输出趋势分析结果
	t.Log("\n=== 趋势分析报告 ===")

	t.Log("\n1. 内存趋势:")
	t.Logf("- 起始: %.2f MB", report.Trends.Memory.Start)
	t.Logf("- 结束: %.2f MB", report.Trends.Memory.End)
	t.Logf("- 峰值: %.2f MB", report.Trends.Memory.Peak)
	t.Logf("- 平均增长: %.2f MB/采样", report.Trends.Memory.AvgGrowth)

	t.Log("\n2. GC趋势:")
	t.Logf("- 平均暂停: %.2f ms", report.Trends.GC.AvgPause)
	t.Logf("- 最大暂停: %.2f ms", report.Trends.GC.MaxPause)
	t.Logf("- GC次数: %d", report.Trends.GC.PauseCount)
	t.Logf("- 平均间隔: %.2f s", report.Trends.GC.AvgInterval)

	t.Log("\n3. Goroutine趋势:")
	t.Logf("- 最小数量: %d", report.Trends.Goroutines.MinCount)
	t.Logf("- 最大数量: %d", report.Trends.Goroutines.MaxCount)
	t.Logf("- 平均数量: %.2f", report.Trends.Goroutines.AvgCount)
	t.Logf("- 阻塞率: %.2f%%", report.Trends.Goroutines.BlockRate*100)

	t.Log("\n4. CPU趋势:")
	t.Logf("- 平均使用率: %.2f%%", report.Trends.CPU.AvgUsage)
	t.Logf("- 峰值使用率: %.2f%%", report.Trends.CPU.PeakUsage)
	t.Logf("- 用户态/系统态比例: %.2f", report.Trends.CPU.UserSysRatio)

	// 输出警告
	var warnings []string
	if report.Trends.Memory.AvgGrowth > 5 { // 平均每次采样增长超过5MB
		warnings = append(warnings, "内存持续增长,可能存在内存泄漏")
	}
	if report.Trends.GC.MaxPause > 100 { // GC暂停超过100ms
		warnings = append(warnings, "GC暂停时间过长")
	}
	if report.Trends.Goroutines.BlockRate > 0.2 { // 超过20%的goroutine处于阻塞状态
		warnings = append(warnings, "Goroutine阻塞率过高")
	}
	if report.Trends.CPU.AvgUsage > 80 { // CPU使用率超过80%
		warnings = append(warnings, "CPU使用率过高")
	}

	if len(warnings) > 0 {
		t.Log("\n⚠️ 警告:")
		for i, warning := range warnings {
			t.Logf("%d. %s", i+1, warning)
		}
	}

	t.Log("\n=== 性能分析完成 ===")
}
