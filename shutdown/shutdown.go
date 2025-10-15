package shutdown

import (
	"github.com/gofiber/fiber/v2/log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	mu    sync.Mutex
	hooks []func()
)

func Add(hook func()) {
	mu.Lock()
	defer mu.Unlock()
	hooks = append([]func(){hook}, hooks...) // 插入到切片头部实现LIFO
}

// 执行所有注册的钩子
func executeHooks() {
	mu.Lock()
	defer mu.Unlock()
	log.Info("正在释放资源...")
	for _, hook := range hooks {
		hook()
	}
	log.Info("资源释放完毕")
}

// 启动信号监听
func init() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		<-sigChan
		executeHooks()
	}()
}
