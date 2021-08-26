package gocache

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"
)

func TestGetterFunc_Get(t *testing.T) {
	var g DataGetter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("222")
	if get, _ := g.Get("222"); reflect.DeepEqual(get, expect) {
		fmt.Println(string(get))
	}
}

func TestA(t *testing.T) {
	parts := strings.SplitN("/c/g/k", "/", 4)
	fmt.Println(parts[2])
}

// 用map模仿一个慢的数据库
var db = map[string]string{
	"A": "1",
	"B": "2",
	"C": "3",
}

func TestGetGroup(t *testing.T) {
	// 统计慢数据库中没一个数据的访问次数，>1则表示调用了多次回调函数，没有缓存。
	loadCounts := make(map[string]int, len(db))

	group := NewGroup("test1", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[slow DB] search key", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key]++
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s is not found", key)
		}))

	for k, v := range db {
		if get, err := group.Get(k); err != nil || get.String() != v {
			t.Fatal("failed to get value")
		}

		if i := loadCounts[k]; i > 1 {
			log.Fatal("cache miss", i)
		}
	}

	if get, err := group.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", get)
	}
}
