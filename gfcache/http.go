package gfcache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/_cache/"

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	basePath string
	self     string
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s\n", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP handle all http requests
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.RequestURI() == "/favicon.ico" {
		return
	}
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}

	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path, "/", 5)

	groupName := parts[2]
	key := parts[3]

	p.Log("%s %s", groupName, key)
	group := GetGroup(groupName)

	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	if get, err := group.Get(key); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(get.ByteSlice())
		return
	}
}
