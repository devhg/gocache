package gocache

import (
	"github.com/cddgo/gocache/lru"
	"sync"
)

//并发缓存，对核心lru进行封装
type cache struct {
	sync.RWMutex
	lru        *lru.Cache
	cacheBytes int64
	nhit, nget int64
	nevict     int64 // number of evictions
}

//添加缓存
func (c *cache) add(key string, val ByteView) {
	c.Lock()
	defer c.Unlock()

	//延迟初始化(Lazy Initialization)，一个对象的延迟初始化意味
	//着该对象的创建将会延迟至第一次使用该对象时。主要用于提高性能，并减少程序内存要求。
	if c.lru == nil {
		c.lru = lru.New(&lru.CacheConfig{
			MaxBytes: c.cacheBytes,
			OnEvicted: func(s string, value lru.Value) {
				c.nevict++
			},
		})
	}
	c.lru.Add(key, val)
	c.cacheBytes += int64(val.Len())
}

//获取缓存
func (c *cache) get(key string) (val ByteView, ok bool) {
	c.RLock()
	defer c.RUnlock()

	if c.lru == nil {
		return
	}

	c.nget++
	if v, hit := c.lru.Get(key); hit {
		c.nhit++ // 命中返回true
		return v.(ByteView), hit
	}
	return
}
