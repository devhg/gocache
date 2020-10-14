package main

import (
	"flag"
	"fmt"
	"github.com/QXQZX/gofly-cache/gfcache"
	"log"
	"net/http"
)

//用map模仿一个慢的数据库
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *gfcache.Group {
	return gfcache.NewGroup("scores", 2<<10, gfcache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			log.Println("[SlowDB] key is not exist", key)
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, group *gfcache.Group) {
	pool := gfcache.NewHTTPPool(addr)
	pool.SetNodes(addrs...)

	group.RegisterPicker(pool)

	log.Println("gofly-cache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], pool))
}

func startAPIServer(apiAddr string, group *gfcache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			byteView, err := group.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(byteView.ByteSlice())
		}))

	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {

	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"

	// 缓存节点
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	gfcache := createGroup()

	if api {
		go startAPIServer(apiAddr, gfcache)
	}

	startCacheServer(addrMap[port], addrs, gfcache)
}
