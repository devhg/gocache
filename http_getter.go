package gocache

import (
	"fmt"
	pb "github.com/cddgo/gocache/gocachepb"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

//节点处理器
type NodeGetter interface {
	//用于从对应 group 查找对应key的缓存值
	HttpGet(group, key string) ([]byte, error)
	Get(*pb.Request, *pb.Response) error
}

type httpGetter struct {
	baseUrl string //http://10.0.0.1:9305/_cache/
}

// 普通http通信
func (h *httpGetter) HttpGet(group, key string) ([]byte, error) {
	url := fmt.Sprintf("%v%v/%v", h.baseUrl, group, key)
	log.Println(url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", resp.Status)
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	return bytes, nil
}

// protobuf通信
func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	URL := fmt.Sprintf(
		"%v%v/%v",
		h.baseUrl,
		url.QueryEscape(in.GetGroup()),
		url.QueryEscape(in.GetKey()),
	)
	res, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}

	return nil
}

var _ NodeGetter = (*httpGetter)(nil)
