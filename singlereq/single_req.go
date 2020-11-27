package singlereq

import "sync"

// 解决缓存击穿
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type ReqGroup struct {
	mu      sync.Mutex // 保护map
	keyCall map[string]*call
}

// 将原来的load逻辑用Do函数包裹起来，这样确保并发环境下，相同的key只会被请求一次
// 使用 sync.WaitGroup 锁避免重入。
// 无论Do并发被调用多少次，fn只会执行一次，除非 非并发的间隔请求
// 等待 fn 调用结束了，返回返回值或错误。
// 同步锁mu的目的是保护map
func (rg *ReqGroup) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	rg.mu.Lock()
	if rg.keyCall == nil {
		rg.keyCall = make(map[string]*call)
	}

	if call, ok := rg.keyCall[key]; ok {
		rg.mu.Unlock()
		call.wg.Wait()            // 有请求正在进行中，等待
		return call.val, call.err // 所有的并发请求都会在此返回
	}

	c := new(call)

	c.wg.Add(1)         // 发起请求前加锁
	rg.keyCall[key] = c // 添加call， 表明key已经有请求在处理
	rg.mu.Unlock()
	c.val, c.err = fn() // 发起请求
	c.wg.Done()         // 请求结束

	rg.mu.Lock()
	delete(rg.keyCall, key)
	rg.mu.Unlock()
	return c.val, c.err
}
