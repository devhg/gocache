package gfcache

import (
	"github.com/QXQZX/gofly-cache/gfcache/lru"
	"sync"
)

//并发缓存，对核心lru进行封装
type cache struct {
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

//添加缓存
func (c *cache) add(key string, val ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	//延迟初始化(Lazy Initialization)，一个对象的延迟初始化意味
	//着该对象的创建将会延迟至第一次使用该对象时。主要用于提高性能，并减少程序内存要求。
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, val)
}

//获取缓存
func (c *cache) get(key string) (val ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
