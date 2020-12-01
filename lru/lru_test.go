package lru

import (
	"fmt"
	"log"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

func TestGet(t *testing.T) {
	lru := New(&CacheConfig{
		MaxBytes:   19,
		MaxEntries: 2,
	})

	lru.Add("key", String("value"))
	lru.Add("key2", String("value2"))

	fmt.Println(lru.Len())

	if val, ok := lru.Get("key"); !ok || string(val.(String)) != "value" {
		log.Fatal("get value error")
	}

	if _, ok := lru.Get("key2"); !ok {
		log.Fatal("cache find key2 failed")
	}
}

func TestCache_RemoveOldest(t *testing.T) {
	k1, k2, k3 := "k1", "k2", "k3"
	v1, v2, v3 := "v1", "v2", "v3"

	lru := New(&CacheConfig{
		MaxBytes:   int64(len(k1 + k2 + v1 + v2)),
		MaxEntries: 4,
		OnEvicted:  nil,
	})
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	if _, ok := lru.Get("k1"); ok {
		log.Fatal("remove k1 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	k1, k2, k3 := "k1", "k2", "k3"
	v1, v2, v3 := "v1", "v2", "v3"

	keys := make([]string, 0)
	callback := func(key string, val Value) {
		keys = append(keys, key)
	}

	lru := New(&CacheConfig{
		MaxBytes:   int64(len(k1 + k2 + v1 + v2)),
		MaxEntries: 1,
		OnEvicted:  callback,
	})
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	fmt.Println(keys)
}
