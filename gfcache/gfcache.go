package gfcache

import (
	"fmt"
	"github.com/QXQZX/gofly-cache/gfcache/node"
	"log"
	"sync"
)

//对外回调函数接口
// A Getter loads data for a key.
type Getter interface {
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
	name      string
	mainCache cache  // 并发缓存实现
	getter    Getter //缓存未命中时获取数据源的回调

	picker node.NodePicker
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("getter is needed")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		mainCache: cache{cacheBytes: cacheBytes},
		getter:    getter,
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

	if byteView, ok := g.mainCache.get(key); ok {
		log.Printf("read from cache %p", &byteView)
		return byteView, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (byteView ByteView, err error) {
	//log.Println("get resources from other nodes or callback")
	if g.picker != nil {
		if nodeGetter, ok := g.picker.PickNode(key); ok {
			if byteView, err = g.getFromNode(nodeGetter, key); err == nil {
				return byteView, nil
			}
			log.Println("[gofly-Cache] Failed to get from peer", err)
		}
	}
	return g.getLocally(key)
}

//从自定义的回调函数中获取缓存中没有的资源
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	byteView := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, byteView)
	return byteView, nil
}

//迁移缓存到当前group
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
	bytes, err := getter.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}
