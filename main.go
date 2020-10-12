package main

import (
	"fmt"
	"github.com/QXQZX/gofly-cache/gfcache"
	"github.com/QXQZX/gofly-cache/gfcache/httpget"
	"log"
	"net/http"
)

//用map模仿一个慢的数据库
var db = map[string]string{
	"A": "1",
	"B": "2",
	"C": "3",
}

func createGroup() *gfcache.Group {
	return gfcache.NewGroup("scores", 2<<10, gfcache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, group gfcache.Group) {
	pool := httpget.NewHTTPPool(addr)
	pool.Set(addrs...)

	group.RegisterPicker(pool)

	log.Println("gofly-cache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], pool))
}

func main() {

	//addr := ":9305"

	//pool := http2.NewHTTPPool(addr)
	//log.Println("geecache is running at", addr)
	//log.Fatal(http.ListenAndServe(addr, pool))

}
