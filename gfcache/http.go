package gfcache

import (
	"fmt"
	"github.com/QXQZX/gofly-cache/consistenthash"
	"github.com/QXQZX/gofly-cache/gfcache/node"
	"log"
	"net/http"
	"strings"
	"sync"
)

const (
	defaultBasePath   = "/_cache/"
	defaultVirtualNum = 50
)

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	basePath string // 请求路径基础前缀ip:port/_cache/
	self     string // 本节点自身的ip:port

	// 映射远程节点与对应的 httpGetter。每一个远程节点对应一个 httpGetter，
	// 因为 httpGetter 与远程节点的地址 baseURL 有关
	httpGetters map[string]*httpGetter //
	// 一致性哈希存放节点，用来根据具体的 key 选择节点
	nodes *consistenthash.Map
	mu    sync.Mutex
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s\n", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP handle all httpget requests
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.RequestURI() == "/favicon.ico" {
		return
	}
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}

	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path, "/", 5)

	groupName := parts[2]
	key := parts[3]

	p.Log("%s %s", groupName, key)
	group := GetGroup(groupName)

	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	if get, err := group.Get(key); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(get.ByteSlice())
		return
	}
}

// Set updates the pool's list of nodes' key.
// example: key=http://10.0.0.1:9305
func (p *HTTPPool) Set(nodeKeys ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 创建添加节点到一致性哈希
	p.nodes = consistenthash.New(defaultVirtualNum, nil)
	p.nodes.Add(nodeKeys...)

	p.httpGetters = make(map[string]*httpGetter)
	for _, key := range nodeKeys {
		p.httpGetters[key] = &httpGetter{baseUrl: key + p.basePath}
	}
}

// PickNode method picks a node according to key
//具体的 key，选择节点，返回节点对应的 HTTP 客户端。
func (p *HTTPPool) PickNode(key string) (node.NodeGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if nodeKey := p.nodes.Get(key); nodeKey != "" && nodeKey != p.self {
		p.Log("pick node %s", nodeKey)
		return p.httpGetters[nodeKey], true
	}

	return nil, false
}

var _ node.NodePicker = (*HTTPPool)(nil)
