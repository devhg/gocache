package singlereq

import "sync"

// 解决缓存击穿，缓存雪崩
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type ReqGroup struct {
	sync.Mutex // 保护map
	keyCall    map[string]*call
}

// 确保并发环境下，相同的key只会被请求一次
// 使用 sync.WaitGroup锁 避免重入。
// 无论Do并发被调用多少次，fn只会执行一次
// 等待 fn 调用结束了，返回返回值或错误。
// 同步锁mu的目的是保护map
func (rg *ReqGroup) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	rg.Lock()
	if rg.keyCall == nil {
		rg.keyCall = make(map[string]*call)
	}

	// 已经有一个请求在进行
	if call, ok := rg.keyCall[key]; ok {
		rg.Unlock()
		call.wg.Wait()            // 有请求正在进行中，等待已经进行的请求的结果
		return call.val, call.err // 所有的并发请求都会在此返回
	}

	c := new(call)

	c.wg.Add(1)         // 发起请求前加入任务
	rg.keyCall[key] = c // 添加call， 表明key已经有请求在处理
	rg.Unlock()

	c.val, c.err = fn() // 发起请求
	c.wg.Done()         // 请求结束

	rg.Lock()
	delete(rg.keyCall, key)
	rg.Unlock()
	return c.val, c.err
}
