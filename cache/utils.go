package cache

import (
	"github.com/coocood/freecache"
	"github.com/dc-utils/args"
	"github.com/dc-utils/datasize"
	"log"
)

type Cache struct {
	cacheName string
	cache     *freecache.Cache
}

var caches = make(map[string]*Cache)

func init() {
	cache := freecache.NewCache(args.GetDataBytes("cache.default.size", datasize.MB*100))
	caches["default"] = &Cache{cacheName: "default", cache: cache}
}

func Default() *Cache {
	return caches["default"]
}

func GetCache(cacheName string) *Cache {
	return caches[cacheName]
}

func Create(cacheName string, size int) *freecache.Cache {
	cache := freecache.NewCache(size)
	caches[cacheName] = &Cache{cacheName: cacheName, cache: cache}
	return cache
}

func (c *Cache) Get(key string) []byte {
	res, _ := c.cache.Get([]byte(key))
	return res
}

func (c *Cache) GetOrSet(key string, value func() []byte, expireSeconds int) []byte {
	var v []byte
	_, _, err := c.cache.Update([]byte(key), func(existing []byte, exists bool) ([]byte, bool, int) {
		if exists {
			v = existing
			return nil, false, 0
		}
		v = value()
		return v, true, expireSeconds
	})
	if err != nil {
		log.Println("缓存设置失败：", err)
	}
	return v
}

func (c *Cache) Set(key string, value []byte, expireSeconds int) {
	err := c.cache.Set([]byte(key), value, expireSeconds)
	if err != nil {
		log.Println("缓存设置失败：", err)
		return
	}
}

func (c *Cache) Del(key string) {
	c.cache.Del([]byte(key))
}

func (c *Cache) Clear() {
	c.cache.Clear()
}

func Clear() {
	for _, cache := range caches {
		cache.Clear()
	}
}
