package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/devhg/gocache"
)

//用map模仿一个慢的数据库
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *gocache.Group {
	// 创建一个名字为 scores的 group。并注册真正的回调函数
	return gocache.NewGroup("scores", 2<<10, gocache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			log.Println("[SlowDB] key is not exist", key)
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

// startCacheServer 开启一个缓存服务
func startCacheServer(addr string, addrs []string, group *gocache.Group) {
	// 创建一个节点选择器
	pool := gocache.NewHTTPPool(addr)
	pool.SetNodes(addrs...) // 节点选择器设置添加节点

	// 注册节点选择器 到group
	group.RegisterPicker(pool)

	log.Println("gocache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], pool))
}

// startAPIServer 创建一个对外的 REST Full API 服务
func startAPIServer(apiAddr string, group *gocache.Group) {
	// Register statsviz handlers on the default serve mux.
	//statsviz.RegisterDefault()
	//http.ListenAndServe(":8080", nil)

	//http.Handle("/api", http.HandlerFunc(
	//	func(w http.ResponseWriter, r *http.Request) {
	//		key := r.URL.Query().Get("key")
	//		byteView, err := group.Get(key)
	//		if err != nil {
	//			http.Error(w, err.Error(), http.StatusInternalServerError)
	//		}
	//
	//		w.Header().Set("Content-Type", "application/octet-stream")
	//		w.Write(byteView.ByteSlice())
	//	}))

	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		byteView, err := group.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(byteView.ByteSlice())
	})

	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool
	// 分别读取端口  和  是否为api server
	// ./server -port=8003 -api=1 &
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

	// 创建好的cacheGroup
	gfcache := createGroup()

	if api {
		go startAPIServer(apiAddr, gfcache)
	}

	startCacheServer(addrMap[port], addrs, gfcache)
}
