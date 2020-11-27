package main

import (
	"flag"
	"fmt"
	"github.com/arl/statsviz"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

//用map模仿一个慢的数据库
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *Group {
	return NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			log.Println("[SlowDB] key is not exist", key)
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, group *Group) {
	pool := NewHTTPPool(addr)
	pool.SetNodes(addrs...)

	group.RegisterPicker(pool)

	log.Println("gocache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], pool))
}

func startAPIServer(apiAddr string, group *Group) {
	// Force the GC to work to make the plots "move".
	go work()

	// Register statsviz handlers on the default serve mux.
	statsviz.RegisterDefault()
	//http.ListenAndServe(":8080", nil)

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

func work() {
	// Generate some allocations
	m := map[string][]byte{}

	for {
		b := make([]byte, 512+rand.Intn(16*1024))
		m[strconv.Itoa(len(m)%(10*100))] = b

		if len(m)%(10*100) == 0 {
			m = make(map[string][]byte)
		}

		time.Sleep(10 * time.Millisecond)
	}
}
