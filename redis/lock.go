package redis

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

// Lock 锁
type Lock struct {
	client         redis.UniversalClient // Redis 客户端
	key            string                // 锁键名
	token          string                // 锁的唯一标识
	mu             sync.Mutex            // 本地互斥锁，保护状态
	watchdogCtx    context.Context       // 看门狗上下文
	watchdogCancel context.CancelFunc    // 停止看门狗的 cancel 函数
	channel        string                // 锁释放通知的频道
}

// GetLock 获取 Redis 分布式锁实例
func GetLock(key string) *Lock {
	return GetLockWithClient(key, rdb)
}

// GetLockWithClient 获取 Redis 分布式锁实例，可指定redis客户端，一般情况下不用这个
func GetLockWithClient(key string, client *redis.Client) *Lock {
	return &Lock{
		client:  client,
		key:     key,
		channel: fmt.Sprintf("dc:lock-release:%s", key), // 锁释放通知频道
	}
}

const defaultExpiry = 30 * time.Second

// Lock 阻塞式获取锁
func (l *Lock) Lock(ctx context.Context) error {

	// 创建发布订阅客户端
	pubsub := l.client.Subscribe(ctx, l.channel)
	defer func(pubsub *redis.PubSub) {
		err := pubsub.Close()
		if err != nil {
			return
		}
	}(pubsub)

	// 获取锁释放通知的通道
	ch := pubsub.Channel()

	// 循环尝试获取锁
	for {
		// 尝试获取锁（30秒过期时间）
		acquired, err := l.TryLock(ctx)
		if err != nil {
			return err
		}
		if acquired {
			return nil // 成功获取锁并启动了看门狗
		}

		// 等待锁释放通知
		select {
		case <-ctx.Done():
			return ctx.Err()

		case msg, ok := <-ch:
			if !ok {
				// 通道已关闭，重新订阅
				pubsub = l.client.Subscribe(ctx, l.channel)
				ch = pubsub.Channel()
				continue
			}

			// 收到锁释放通知，继续尝试获取锁
			if msg.Channel == l.channel {
				continue
			}
		}
	}
}

// TryLock 尝试获取锁（推荐使用）
//
// waitTime <= 0 时立即返回，否则最多等待指定时间
//
// 示例：
//
// res, _ := lock.TryLock(ctx)
//
//	if res {
//		defer func(lock *redis.Lock, ctx context.Context) {
//			_, err := lock.Unlock(ctx)
//			if err != nil {
//				return
//			}
//		}(lock, ctx)
//		业务逻辑...
//	} else {
//		return nil
//	}
func (l *Lock) TryLock(ctx context.Context, waitTimeOpt ...time.Duration) (bool, error) {
	var waitTime time.Duration
	if len(waitTimeOpt) == 0 || waitTimeOpt[0] <= 0 {
		// 立即尝试模式
		return l.tryLockOnce(ctx)
	} else {
		waitTime = waitTimeOpt[0]
	}

	// 创建发布订阅客户端
	pubsub := l.client.Subscribe(ctx, l.channel)
	defer func(pubsub *redis.PubSub) {
		err := pubsub.Close()
		if err != nil {
			return
		}
	}(pubsub)

	// 获取锁释放通知的通道
	ch := pubsub.Channel()

	startTime := time.Now()

	// 循环尝试获取锁，直到超时
	for {
		// 尝试获取锁
		acquired, err := l.tryLockOnce(ctx)
		if err != nil {
			return false, err
		}
		if acquired {
			return true, nil
		}

		// 计算剩余等待时间
		remainingTime := waitTime - time.Since(startTime)
		if remainingTime <= 0 {
			return false, nil // 超时返回
		}

		// 等待锁释放通知或超时
		select {
		case <-ctx.Done():
			return false, ctx.Err()

		case <-time.After(remainingTime):
			// 最后一次尝试
			return l.tryLockOnce(ctx)

		case msg, ok := <-ch:
			if !ok {
				// 通道已关闭，重新订阅
				pubsub = l.client.Subscribe(ctx, l.channel)
				ch = pubsub.Channel()
				continue
			}

			// 收到锁释放通知，继续尝试获取锁
			if msg.Channel == l.channel {
				continue
			}
		}
	}
}

// TryLockOnce 执行单次锁获取尝试
func (l *Lock) tryLockOnce(ctx context.Context) (bool, error) {

	l.mu.Lock()
	defer l.mu.Unlock()

	// 如果已经持有锁，直接返回
	if l.watchdogCancel != nil {
		return true, nil
	}

	// 生成唯一 token
	token, err := generateToken()
	if err != nil {
		return false, err
	}

	// 原子性获取锁
	set, err := l.client.SetNX(ctx, l.key, token, defaultExpiry).Result()
	if err != nil {
		return false, err
	}

	if !set {
		return false, nil // 锁已被其他客户端持有
	}

	// 获取锁成功，保存 token 并启动看门狗
	l.token = token
	l.watchdogCtx, l.watchdogCancel = context.WithCancel(context.Background())
	go l.startRefresh(defaultExpiry)

	return true, nil
}

// Unlock 释放锁
func (l *Lock) Unlock(ctx context.Context) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 如果未持有锁，直接返回
	if l.watchdogCancel == nil {
		return true, nil
	} else {
		// 停止看门狗
		l.watchdogCancel()
		l.watchdogCancel = nil
	}

	// 使用 Lua 脚本原子性释放锁（仅当锁存在且 token 匹配时）
	result, err := l.client.Eval(ctx, `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`, []string{l.key}, l.token).Int64()

	if err != nil {
		return false, err
	}

	// 重置状态
	l.token = ""

	return result == 1, nil
}

// IsHeld 检查锁是否被当前持有者持有
func (l *Lock) IsHeld(ctx context.Context) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 如果未持有锁，直接返回
	if l.watchdogCancel == nil {
		return false, nil
	}

	// 检查 Redis 中的 token 是否匹配
	token, err := l.client.Get(ctx, l.key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil // 锁不存在
		}
		return false, err
	}

	return token == l.token, nil
}

// startRefresh 启动看门狗自动续期
func (l *Lock) startRefresh(expiry time.Duration) {
	// 计算续期频率（默认1/3过期时间）
	refreshRate := expiry / 3
	ticker := time.NewTicker(refreshRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 刷新锁的过期时间
			if err := l.refresh(expiry); err != nil {
				// 刷新失败，可能锁已过期或被释放
				l.mu.Lock()
				if l.watchdogCancel != nil {
					l.watchdogCancel()
					l.watchdogCancel = nil
				}
				l.mu.Unlock()
				return
			}
		case <-l.watchdogCtx.Done():
			return
		}
	}
}

// refresh 刷新锁的过期时间
func (l *Lock) refresh(expiry time.Duration) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 使用 Lua 脚本原子性刷新锁的过期时间（仅当锁存在且 token 匹配时）
	result, err := l.client.Eval(l.watchdogCtx, `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("PEXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`, []string{l.key}, l.token, expiry.Milliseconds()).Int64()

	if err != nil {
		return err
	}

	if result == 0 {
		return fmt.Errorf("刷新锁失败，可能已过期或被其他客户端释放")
	}

	return nil
}

// generateToken 生成唯一的随机 token
func generateToken() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
