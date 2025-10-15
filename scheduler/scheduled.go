package scheduler

import (
	"context"
	"github.com/dc-utils/redis"
	"github.com/robfig/cron/v3"
	"time"
)

func init() {
	cronInstance = cron.New(cron.WithSeconds())
	cronInstance.Start()
}

var cronInstance *cron.Cron

// AddDelayTask 添加延迟任务
func AddDelayTask(delay time.Duration, task func()) {
	go func() {
		<-time.After(delay)
		task()
	}()
}

// AddPeriodicTask 添加周期执行任务 (带初始延迟)
func AddPeriodicTask(initialDelay, interval time.Duration, task func()) {
	go func() {
		<-time.After(initialDelay)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			task()
		}
	}()
}

// AddCronTask 添加Cron任务
func AddCronTask(spec string, task func()) cron.EntryID {
	id, _ := cronInstance.AddFunc(spec, task)
	return id
}

func AddPeriodicClusterTask(initialDelay, interval time.Duration, task func(), taskName string) {
	go func() {
		<-time.After(initialDelay)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		for range ticker.C {
			if redis.Get().SetNX(ctx, "go:task:"+taskName, "doing", 20*time.Second).Val() {
				task()
			}
		}
	}()
}

func AddCronClusterTask(spec string, task func(), taskName string) cron.EntryID {
	id, _ := cronInstance.AddFunc(spec, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if redis.Get().SetNX(ctx, "go:task:"+taskName, "doing", 20*time.Second).Val() {
			task()
		}
	})
	return id
}
