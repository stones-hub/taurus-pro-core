package controller

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/google/wire"
	"github.com/stones-hub/taurus-pro-http/pkg/httpx"
)

// MemoryController 内存测试控制器
type MemoryController struct {
	// 用于模拟内存泄漏的map
	leakyMap sync.Map
	// 用于模拟大对象分配
	largeObjects [][]byte
}

func NewMemoryController() *MemoryController {
	return &MemoryController{
		leakyMap:     sync.Map{},
		largeObjects: [][]byte{},
	}
}

// MemoryControllerSet wire provider set
var MemoryControllerSet = wire.NewSet(NewMemoryController)

// AllocateMemory 分配指定大小的内存
func (c *MemoryController) AllocateMemory(w http.ResponseWriter, r *http.Request) {
	// 获取要分配的内存大小（MB）
	sizeStr := r.URL.Query().Get("size")
	if sizeStr == "" {
		sizeStr = "1" // 默认1MB
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		httpx.SendResponse(w, http.StatusBadRequest, nil, map[string]string{
			"error": err.Error(),
		})
		return
	}

	// 分配内存
	data := make([]byte, size*1024*1024) // 转换为MB
	c.largeObjects = append(c.largeObjects, data)

	httpx.SendResponse(w, http.StatusOK, map[string]interface{}{
		"allocated_size_mb": size,
		"total_objects":     len(c.largeObjects),
	}, nil)
}

// SimulateMemoryLeak 模拟内存泄漏
func (c *MemoryController) SimulateMemoryLeak(w http.ResponseWriter, r *http.Request) {
	// 获取要生成的对象数量
	countStr := r.URL.Query().Get("count")
	if countStr == "" {
		countStr = "1000" // 默认1000个对象
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		httpx.SendResponse(w, http.StatusBadRequest, nil, map[string]string{
			"error": err.Error(),
		})
		return
	}

	// 生成并存储对象（模拟内存泄漏）
	for i := 0; i < count; i++ {
		key := "key_" + strconv.Itoa(i)
		// 每个对象约1KB
		value := make([]byte, 1024)
		c.leakyMap.Store(key, value)
	}

	httpx.SendResponse(w, http.StatusOK, map[string]interface{}{
		"objects_created": count,
	}, nil)
}

// FreeMemory 释放已分配的内存
func (c *MemoryController) FreeMemory(w http.ResponseWriter, r *http.Request) {
	c.largeObjects = nil // 释放大对象

	// 清理泄漏的对象
	c.leakyMap.Range(func(key, value interface{}) bool {
		c.leakyMap.Delete(key)
		return true
	})

	httpx.SendResponse(w, http.StatusOK, map[string]interface{}{
		"status": "memory freed",
	}, nil)
}
