package lru

import "container/list"

type Cache struct {
	maxBytes int64 // 最大使用内存
	nbytes   int64 // 已经使用的内存
	ll       *list.List
	cache    map[string]*list.Element

	// optional and executed when an entry is purged.
	//是某条记录被移除时的回调函数，可以为 nil
	OnEvicted func(key string, value Value)
}

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int
}

//双向链表节点的数据类型，保存key的目的是淘汰队首节点时，
//需要用key删除字典中对应的映射
type entry struct {
	key   string
	value Value
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get look ups a key's value
func (c *Cache) Get(key string) (Value, bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return nil, false
}

// RemoveOldest removes the oldest item
// 删除最近最少使用的缓存节点
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)

		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add adds a value to the cache.
func (c *Cache) Add(key string, val Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(val.Len()) - int64(kv.value.Len())
		kv.value = val
	} else {
		front := c.ll.PushFront(&entry{key: key, value: val})
		c.nbytes += int64(len(key)) + int64(val.Len())
		c.cache[key] = front
	}

	for c.maxBytes != 0 && c.nbytes > c.maxBytes {
		c.RemoveOldest()
	}
}

// Len the number of cache entries
func (c *Cache) Len() int {
	return c.ll.Len()
}
