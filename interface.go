package gocache

// // 节点选择器
// type NodePicker interface {
// 	// 利用一致性哈希算法，根据传入的 key 选择相应节点
// 	// 并返回节点处理器NodeGetter。
// 	PickNode(key string) (NodeGetter, bool)
// }

// // 节点处理器
// type NodeGetter interface {
// 	// 用于从对应 group 查找对应key的缓存值
// 	// Get(group, key string) ([]byte, error)
// 	Get(*pb.Request, *pb.Response) error
// }
