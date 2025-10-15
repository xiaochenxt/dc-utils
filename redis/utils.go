package redis

import (
	"context"
	"github.com/dc-utils/args"
	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

var rdb *redis.Client

var once sync.Once

func init() {
	enabled := args.GetBool("redis.enabled", true)
	if !enabled {
		return
	}
	Enable()
}

func Enable() {
	pwd := args.GetStr("redis.password", "")
	if pwd == "" {
		return
	}
	EnableWithConfig(&redis.Options{
		Addr:         args.GetStr("redis.addr", "localhost:6379"),
		Password:     args.GetStr("redis.password", ""),
		DB:           args.GetInt("redis.db", 0),
		PoolSize:     args.GetInt("redis.poolSize", 200),   // 连接池最大连接数
		MinIdleConns: args.GetInt("redis.minIdleConns", 1), // 最小空闲连接数
		MaxIdleConns: args.GetInt("redis.maxIdleConns", 8),
		PoolTimeout:  args.GetDuration("redis.poolTimeout", 30*time.Second), // 连接池超时时间（秒）
	})
}

func EnableWithConfig(config *redis.Options) {
	once.Do(func() {
		rdb = redis.NewClient(config)
	})
}

func Get() *redis.Client {
	return rdb
}

// Subscribe 订阅：监听指定频道的消息
func Subscribe(channelName string, callback func(channel string, pattern string, payload string, payloadSlice []string)) {
	client := Get()
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// 订阅频道
		pubSub := client.Subscribe(ctx, channelName)
		defer func(pubSub *redis.PubSub) {
			err := pubSub.Close()
			if err != nil {
				return
			}
		}(pubSub)
		// 等待订阅确认
		_, err := pubSub.Receive(ctx)
		if err != nil {
			log.Errorf("订阅频道失败: %v", err)
		}
		// 接收消息的通道
		ch := pubSub.Channel()
		// 持续监听消息
		log.Debugf("开始监听频道: %s\n", channelName)
		for msg := range ch {
			// 处理接收到的消息
			//log.Printf("收到消息 [频道: %s, 内容: %s]\n", msg.Channel, msg.Payload)
			callback(msg.Channel, msg.Pattern, msg.Payload, msg.PayloadSlice)
		}
	}()
}

// Publish 发布：向指定频道发送消息
func Publish(channelName string, messages []string) {
	client := Get()
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// 发送消息到频道
		for i, msg := range messages {
			// 发送消息
			val, err := client.Publish(ctx, channelName, msg).Result()
			if err != nil {
				log.Errorf("发送消息失败: %v", err)
			}
			log.Debugf("发送消息 #%d [频道: %s, 内容: %s, 订阅者数量: %d]\n", i+1, channelName, msg, val)
		}
	}()
}
