package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

/**
一致性哈希 ---从单节点走向分布式节点的一个重要的环节

算法原理：
一致性hash算法将key映射到2^32的空间，一个环状结构 0 - 2^32-1
* 计算节点/机器(通常为节点的编号、名称或IP地址)的哈希值，放到环上
* 计算key的哈希值，放到环上，顺时针寻找到key后面的第一个节点就是key存储的位置，就是选取的机器

这样也就是说，在新增或者删除节点时，只需要重新定位节点附近的一小块数据区域，而不需要重新定位所有
数据

如果服务器的节点过少，容易引起 key 的倾斜。容易造成一大部分 分布在环的上半部分，下半部分是空的。
那么映射到环下半部分的 key 都会被分配给 peer2，key 过度向 peer2 倾斜，缓存节点间负载不均。

为了解决这个问题，引入了虚拟节点的概念，一个真实节点对应多个虚拟节点。

假设 1 个真实节点对应 3 个虚拟节点，那么 peer1 对应的虚拟节点是 peer1-1、 peer1-2、 peer1-3（通常以添加编号的方式实现），
其余节点也以相同的方式操作。

第一步，计算虚拟节点的 Hash 值，放置在环上。
第二步，计算 key 的 Hash 值，在环上顺时针寻找到应选取的虚拟节点，例如是 peer2-1，那么就对应真实节点 peer2。
虚拟节点扩充了节点的数量，解决了节点较少的情况下数据容易倾斜的问题。而且代价非常小，
只需要增加一个字典(map)维护真实节点与虚拟节点的映射关系即可。


*/

//a hash maps bytes to uint32
type Hash func(data []byte) uint32

// Map contains all hashed keys
type Map struct {
	hash       Hash     // 注入哈希处理函数
	keys       []uint32 // 哈希环 sorted
	virtualNum int      // 每个节点对应虚拟节点的数目

	// 存放虚拟节点和真实节点的数目 key=hash(b"i-真实节点id")  value=真实节点id
	hashMap map[uint32]string
}

func New(vNum int, hash Hash) *Map {
	m := &Map{
		hash:       hash,
		virtualNum: vNum,
		hashMap:    make(map[uint32]string),
	}

	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 添加真实/虚拟节点函数
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.virtualNum; i++ {
			// 计算虚拟节点的hash值
			hash := m.hash([]byte(strconv.Itoa(i) + key))
			//将虚拟节点的hash值添加到换上
			m.keys = append(m.keys, hash)
			//增加虚拟节点和真实节点的映射关系
			m.hashMap[hash] = key
		}
	}

	sort.Slice(m.keys, func(i, j int) bool {
		return m.keys[i] < m.keys[j]
	})
}

// 实现选择节点的get方法
func (m *Map) Get(key string) string {
	if len(key) == 0 {
		return ""
	}

	hash := m.hash([]byte(key))
	// 采用二分查找顺时针查找第一个匹配的虚拟节点
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	// idx == len(m.keys) 在没有找到的情况下返回len(m.keys)
	// 说明应选择 m.keys[0]，因为 m.keys 是一个环状结构，所以用取余数的方式来处理这种情况。
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
