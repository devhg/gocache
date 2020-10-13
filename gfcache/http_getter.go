package gfcache

import (
	"fmt"
	"github.com/QXQZX/gofly-cache/gfcache/node"
	"io/ioutil"
	"log"
	"net/http"
)

type httpGetter struct {
	baseUrl string //http://10.0.0.1:9305/_cache/
}

func (h *httpGetter) Get(group, key string) ([]byte, error) {
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

var _ node.NodeGetter = (*httpGetter)(nil)
