package lru

import (
	"container/list"
)

const (
	maxBytes   = 1 << 32
	maxEntries = 10 << 10
)

type Cache struct {
	maxBytes int64 // 最大使用内存
	nowBytes int64 // 已经使用的内存

	maxEntries int

	ll    *list.List
	cache map[string]*list.Element

	//是某条记录被移除时的回调函数，可以为 nil
	onEvicted func(key string, value Value)
}

// a config for cache
type CacheConfig struct {
	MaxBytes   int64 // 最大使用内存
	MaxEntries int   // 最大缓存数目

	// 淘汰回调函数
	OnEvicted func(string, Value)
}

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int
}

//双向链表节点的数据类型，保存key的目的是淘汰队首节点时，
type entry struct {
	key   string
	value Value
}

func New(config *CacheConfig) *Cache {
	c := &Cache{
		maxBytes:   maxBytes,
		maxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[string]*list.Element),
	}
	if config == nil {
		return c
	}

	if config.MaxBytes != 0 {
		c.maxBytes = config.MaxBytes
	}
	if config.MaxEntries != 0 {
		c.maxEntries = config.MaxEntries
	}
	if config.OnEvicted != nil {
		c.onEvicted = config.OnEvicted
	}
	return c
}

// 按key 添加缓存
func (c *Cache) Add(key string, val Value) {
	if c.cache == nil {
		c.cache = make(map[string]*list.Element)
		c.ll = list.New()
	}

	//缓存命中
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		// 更新缓存内容
		kv := ele.Value.(*entry).value
		c.nowBytes += int64(val.Len()) - int64(kv.Len())
		return
	}

	// 缓存未命中
	ele := c.ll.PushFront(&entry{key: key, value: val})
	c.nowBytes += int64(val.Len() + len(key))
	c.cache[key] = ele

	// 超过最大缓存数目 淘汰
	if c.maxEntries != 0 && c.ll.Len() > c.maxEntries {
		c.RemoveOldest()
	}

	// 超过最大缓存容量 淘汰
	for c.maxBytes != 0 && c.nowBytes > c.maxBytes {
		c.RemoveOldest()
	}
}

// 删除最近最少使用的缓存
func (c *Cache) RemoveOldest() {
	if c.cache == nil {
		return
	}
	lastEle := c.ll.Back()
	if lastEle != nil {
		c.removeElement(lastEle)
	}
}

// 按key 删除缓存
func (c *Cache) Remove(key string) bool {
	if c.cache == nil {
		return false
	}
	if ele, ok := c.cache[key]; ok {
		c.removeElement(ele)
		return true
	}
	return false
}

func (c *Cache) removeElement(ele *list.Element) {
	c.ll.Remove(ele)
	val := ele.Value.(*entry)

	// 缓存容量减少
	c.nowBytes -= int64(len(val.key)) + int64(val.value.Len())
	delete(c.cache, val.key)

	if c.onEvicted != nil {
		c.onEvicted(val.key, val.value)
	}
}

// 获取缓存
func (c *Cache) Get(key string) (val Value, ok bool) {
	if c.cache == nil {
		return nil, false
	}
	if ele, ok := c.cache[key]; ok {
		c.ll.PushFront(ele)
		val := ele.Value.(*entry)
		return val.value, true
	}
	return
}

// Len the number of cache entries
func (c *Cache) Len() int {
	return c.ll.Len()
}
