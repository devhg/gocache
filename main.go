package main

import (
	"fmt"
	"github.com/QXQZX/gofly-cache/gfcache"
	"log"
	"net/http"
)

//用map模仿一个慢的数据库
var db = map[string]string{
	"A": "1",
	"B": "2",
	"C": "3",
}

func main() {
	gfcache.NewGroup("scores", 2<<10,
		gfcache.GetterFunc(func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := ":9305"

	pool := gfcache.NewHTTPPool(addr)
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, pool))

}
