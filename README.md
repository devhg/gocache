# 分布式缓存

用go语言实现的一个分布式缓存


### TODO
- [x] LRU缓存淘汰策略
- [x] 单机并发缓存
- [x] http客户端及请求支持
- [x] 实现一致性哈希算法
- [x] 利用一致性哈希算法，从单一节点走向分布式
- [x] 缓存击穿问题
- [ ] 缓存穿透问题
- [x] Protobuf通信
- [ ] 支持统计信息展示
- [ ] 其他问题


* 缓存雪崩：缓存在同一时刻全部失效，造成瞬时DB请求量大、压力骤增，引起雪崩。缓存雪崩通常因为缓存服务器宕机、缓存的 key 设置了相同的过期时间等引起。

    解决：不支持过期时间，只维护了一个最近使用的缓存队列，暂时无法解决雪崩问题

* 缓存击穿：一个存在的key，在缓存过期的一刻，同时有大量的请求，这些请求都会击穿到 DB ，造成瞬时DB请求量大、压力骤增。

    解决：使用sync.WaitGroup锁来避免重入，保证并发的时候只有一个请求在工作，详见singlereq/single_req.go的Do()

* 缓存穿透：查询一个不存在的数据，因为不存在则不会写到缓存中，所以每次都会去请求 DB，如果瞬间流量过大，穿透到 DB，导致宕机。

    解决：
    

###
项目结构
![](https://cdn.jsdelivr.net/gh/QXQZX/CDN@latest/images/go/gfcache/framework.png)
创建流程
![](https://cdn.jsdelivr.net/gh/QXQZX/CDN@latest/images/go/gfcache/runAndUse.png)

<hr>
仅用于学习