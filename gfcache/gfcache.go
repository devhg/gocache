package gfcache

import (
	"fmt"
	"github.com/QXQZX/gofly-cache/gfcache/node"
	pb "github.com/QXQZX/gofly-cache/gfcache/proto"
	"github.com/QXQZX/gofly-cache/gfcache/singlereq"
	"log"
	"sync"
)

//对外回调函数接口
// A Getter loads data from DB or etc. for a key.
type DataGetter interface {
	Get(key string) ([]byte, error)
}

//定义一个函数类型 F，并且实现接口 A 的方法，然后在这个方法中调用自己。
//这是 Go 语言中将其他函数（参数返回值定义与 F 一致）转换为接口 A 的常用技巧。

// A GetterFunc implements Getter with a function.
type GetterFunc func(string) ([]byte, error)

// Get implements Getter interface function
func (g GetterFunc) Get(key string) ([]byte, error) {
	return g(key)
}

//一个group可以被认为一个缓存的命名空间
//每一个group拥有一个唯一的name，这样可以创建多个group
//score得分，info个人信息，courses课程
type Group struct {
	name       string
	mainCache  cache      // 并发缓存实现
	dataGetter DataGetter //缓存未命中时获取数据源的回调

	loader *singlereq.ReqGroup
	picker node.NodePicker
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter DataGetter) *Group {
	if getter == nil {
		panic("getter is needed")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:       name,
		mainCache:  cache{cacheBytes: cacheBytes},
		dataGetter: getter,
		loader:     &singlereq.ReqGroup{},
	}
	groups[name] = g
	return groups[name]
}

func GetGroup(name string) *Group {
	//共享锁
	mu.RLock()
	defer mu.RUnlock()
	g := groups[name]
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	//在本机缓存中查找
	if byteView, ok := g.mainCache.get(key); ok {
		log.Printf("read from local cache %p", &byteView)
		return byteView, nil
	}

	//去其他节点查找或者从数据库从新缓存
	return g.load(key)
}

func (g *Group) load(key string) (byteView ByteView, err error) {
	// 每一个key只允许请求一次远程服务器或者db  防止缓存击穿和缓存穿透
	val, err := g.loader.Do(key, func() (i interface{}, err error) {
		if g.picker != nil {
			if nodeGetter, ok := g.picker.PickNode(key); ok {
				if byteView, err = g.getFromNode(nodeGetter, key); err == nil {
					return byteView, nil
				}
				log.Println("[gfCache] Failed to get from other node", err)
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return val.(ByteView), nil
	}
	return
}

//从自定义的回调函数中获取缓存中没有的资源
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.dataGetter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	byteView := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, byteView)
	return byteView, nil
}

//缓存到当前节点的group
func (g *Group) populateCache(key string, val ByteView) {
	g.mainCache.add(key, val)
}

// 将实现了 NodePicker接口的 HTTPPool 注入到 Group 中
func (g *Group) RegisterPicker(picker node.NodePicker) {
	if g.picker != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.picker = picker
}

// 用实现了 NodeGetter 接口的 httpGetter 从访问远程节点，获取缓存值
func (g *Group) getFromNode(getter node.NodeGetter, key string) (ByteView, error) {
	if getter == nil {
		panic("NodeGetter is required")
	}
	request := &pb.Request{Group: g.name, Key: key}
	response := &pb.Response{}
	err := getter.Get(request, response)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: response.Value}, nil
}
