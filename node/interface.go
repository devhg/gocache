package node

import pb "github.com/cddgo/gocache/proto"

//节点选择器
type NodePicker interface {
	//用于根据传入的 key 选择相应节点 NodeGetter。
	PickNode(key string) (NodeGetter, bool)
}

//节点处理器
//类似于 HTTP 客户端
type NodeGetter interface {
	//用于从对应 group 查找对应key的缓存值
	//Get(group, key string) ([]byte, error)
	Get(*pb.Request, *pb.Response) error
}
