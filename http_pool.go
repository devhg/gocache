package gocache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/devhg/gocache/consistenthash"
	pb "github.com/devhg/gocache/gocachepb"
	"google.golang.org/protobuf/proto"
)

const (
	defaultBasePath   = "/_cache/"
	defaultVirtualNum = 50
)

// NodePicker 节点选择器
type NodePicker interface {
	// 利用一致性哈希算法，根据传入的 key 选择相应节点
	// 并返回节点处理器NodeGetter。
	PickNode(key string) (NodeGetter, bool)
}

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	basePath string // 请求路径基础前缀/_cache/
	selfAddr string // 本节点自身的ip:port

	// 映射远程节点与对应的 httpGetter。每一个远程节点对应一个 httpGetter，
	// 因为 httpGetter 与远程节点的地址 baseURL 有关
	httpGetters map[string]*httpGetter

	// 一致性哈希存放节点，用来根据具体的 key 选择节点
	nodes *consistenthash.Map
	mu    sync.Mutex
}

func NewHTTPPool(selfAddr string) *HTTPPool {
	return &HTTPPool{
		selfAddr: selfAddr,
		basePath: defaultBasePath,
	}
}

// print the Log of HTTPPool
func (p *HTTPPool) Logf(format string, v ...interface{}) {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Printf("[Server %s] %s\n", p.selfAddr, fmt.Sprintf(format, v...))
}

// ServeHTTP handle all http requests
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.RequestURI() == "/favicon.ico" {
		return
	}
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}

	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path, "/", 5)

	groupName := parts[2]
	key := parts[3]

	p.Logf("%s %s -- group=%s key=%s", r.Method, r.URL.Path, groupName, key)
	group := GetGroup(groupName)

	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	byteView, err := group.Get(key)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp, err := proto.Marshal(&pb.Response{Value: byteView.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	_, _ = w.Write(resp)
}

// Set the pool's list of nodes' key.
// example: key=http://10.0.0.1:9305
func (p *HTTPPool) SetNodes(nodeKeys ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 创建添加节点到一致性哈希
	p.nodes = consistenthash.New(defaultVirtualNum, nil)
	p.nodes.Add(nodeKeys...)

	p.httpGetters = make(map[string]*httpGetter)
	for _, nodeKey := range nodeKeys {
		p.httpGetters[nodeKey] = &httpGetter{baseURL: nodeKey + p.basePath}
	}
}

// PickNode method picks a node according to key
// 具体的 key，选择节点，返回节点对应的HTTP处理器(NodeGetter)。
func (p *HTTPPool) PickNode(key string) (NodeGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if nodeKey := p.nodes.Get(key); nodeKey != "" && nodeKey != p.selfAddr {
		p.Logf("pick node %s", nodeKey)
		return p.httpGetters[nodeKey], true
	}
	return nil, false
}

var _ NodePicker = (*HTTPPool)(nil)
