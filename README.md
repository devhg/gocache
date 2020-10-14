# 分布式缓存

用go语言实现的一个分布式缓存


### 进度
- [x] LRU缓存淘汰策略
- [x] 单机并发缓存
- [x] http客户端支持
- [x] 一致性哈希算法
- [x] 分布式节点支持
- [ ] 缓存击穿和缓存穿透问题
- [ ] Protobuf通信
- [ ] 其他问题。。


* 缓存雪崩：缓存在同一时刻全部失效，造成瞬时DB请求量大、压力骤增，引起雪崩。缓存雪崩通常因为缓存服务器宕机、缓存的 key 设置了相同的过期时间等引起。

    解决：一致性哈希算法

* 缓存击穿：一个存在的key，在缓存过期的一刻，同时有大量的请求，这些请求都会击穿到 DB ，造成瞬时DB请求量大、压力骤增。

    解决：

* 缓存穿透：查询一个不存在的数据，因为不存在则不会写到缓存中，所以每次都会去请求 DB，如果瞬间流量过大，穿透到 DB，导致宕机。

    解决：
    

###
项目结构
![](https://cdn.jsdelivr.net/gh/QXQZX/CDN@latest/images/go/gfcache/framework.png)
使用流程
![](https://cdn.jsdelivr.net/gh/QXQZX/CDN@latest/images/go/gfcache/runAndUse.png)

<hr>
仅用于学习