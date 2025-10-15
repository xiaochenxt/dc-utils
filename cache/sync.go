package cache

import (
	"encoding/json"
	"github.com/dc-utils/redis"
	"github.com/gofiber/fiber/v2/log"
)

func init() {
	if redis.Get() == nil {
		return
	}
	redis.Subscribe("dc:cache:sync", func(channel string, pattern string, payload string, payloadSlice []string) {
		type message struct {
			Type      int      `json:"type"`
			CacheName []string `json:"cacheName"`
			Key       string   `json:"key"`
		}
		var msg message
		err := json.Unmarshal([]byte(payload), &msg)
		if err != nil {
			return
		}
		log.Debugf("本地缓存同步监听，channel：%v，pattern：%v，payload：%v，payloadSlice：%v", channel, pattern, msg, payloadSlice)
		if msg.Type == 1 {
			for _, s := range msg.CacheName {
				cache := GetCache(s)
				if cache != nil {
					cache.Del(msg.Key)
				}
			}
		} else if msg.Type == 2 {
			if msg.CacheName == nil || len(msg.CacheName) == 0 {
				Clear()
			} else {
				for _, s := range msg.CacheName {
					cache := GetCache(s)
					if cache != nil {
						cache.Clear()
					}
				}
			}
		}
	})
}
