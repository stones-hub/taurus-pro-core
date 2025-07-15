package crontab

import (
	"fmt"
	"log"

	"{{.ProjectName}}/internal/taurus"

	"github.com/stones-hub/taurus-pro-common/pkg/cron"
)

// 预定义的任务组
var (
	// 任务组映射
	taskGroups = make(map[string]*cron.TaskGroup)
	// 任务列表
	tasks = make([]*cron.Task, 0)
)

// GetOrCreateTaskGroup 获取或创建任务组
func GetOrCreateTaskGroup(name string, tags ...string) *cron.TaskGroup {
	if group, exists := taskGroups[name]; exists {
		return group
	}

	group := cron.NewTaskGroup(name)
	for _, tag := range tags {
		group.AddTag(tag)
	}
	taskGroups[name] = group
	return group
}

// Register 注册一个定时任务
func Register(task ...*cron.Task) {
	log.Printf("注册任务: %v\n", task)
	tasks = append(tasks, task...)
}

// StartTasks 启动所有注册的定时任务
func StartTasks() error {
	// 获取cron管理器
	cm := taurus.Container.Cron
	if cm == nil {
		return fmt.Errorf("cron manager is nil, please check the configuration")
	}

	// 注册所有任务
	for _, task := range tasks {
		log.Printf("register task: %s\n", task.Name)
		taskID, err := cm.AddTask(task)
		if err != nil {
			log.Printf("register task failed: %v\n", err)
			continue
		}
		log.Printf("register task success (ID: %d)\n", taskID)
	}
	cm.Start()
	return nil
}
