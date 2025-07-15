package crontab

import (
	"context"
	"log"
	"time"

	"github.com/stones-hub/taurus-pro-common/pkg/cron"
)

func init() {
	// 创建任务组
	businessGroup := GetOrCreateTaskGroup("business", "core", "monitoring")

	// 创建一个每5秒执行一次的状态检查任务
	statusCheckTask := cron.NewTask(
		"status_check",
		"* * * * * *", // 每1秒执行一次
		func(ctx context.Context) error {
			log.Println("执行状态检查...")
			// 模拟任务执行
			time.Sleep(2 * time.Second)
			return nil
		},
		cron.WithTimeout(10*time.Second),
		cron.WithRetry(3, time.Second),
		cron.WithGroup(businessGroup),
		cron.WithTag("check"),
		cron.WithTag("periodic"),
	)

	// 创建一个每分钟执行一次的数据同步任务
	dataSyncTask := cron.NewTask(
		"data_sync",
		"* * * * * *", // 每1秒执行一次
		func(ctx context.Context) error {
			log.Println("开始数据同步...")
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(30 * time.Second):
				log.Println("数据同步完成")
				return nil
			}
		},
		cron.WithTimeout(45*time.Second),
		cron.WithGroup(businessGroup),
		cron.WithTag("sync"),
		cron.WithTag("data"),
	)

	// 注册任务
	Register(statusCheckTask, dataSyncTask)
}
