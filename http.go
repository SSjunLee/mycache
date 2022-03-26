package mycache

import (
	"fmt"
	"io/ioutil"
	"log"
	"mycache/consistanthash"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const defaultBasePath = "/_geecache/"
const defaultReplicats = 50

type HttpPool struct {
	self        string //self，当前节点的ip+端口地址
	basePath    string
	peers       *consistanthash.Map //一致性哈希选择器
	mu          sync.Mutex
	httpGetters map[string]*httpGetter //每个远程节点对应一个httpclient
}

func NewHttpPool(self string) *HttpPool {
	return &HttpPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HttpPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, v)
}

func (p *HttpPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	url := r.URL.Path
	p.Log("%s %s", r.Method, r.URL.Path)
	parts := strings.Split(url[len(p.basePath):], "/")
	if len(parts) != 2 {
		http.Error(w, "bad request", 403)
	}
	groupName, key := parts[0], parts[1]
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no group"+groupName, 403)
		return
	}
	v, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), 403)

		return
	}
	w.Header().Add("Content-Type", "application/octet-stream")
	w.Write(v.ByteSlice())
}

//更新peer List
func (p *HttpPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistanthash.New(defaultReplicats, nil) //新建一个一致性哈希选择器
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter)
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseUrl: peer + p.basePath}
	}
}

//根据key选择一个节点，返回访问该节点的http客户端
func (p *HttpPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	peer := p.peers.Get(key)
	if peer != "" && peer != p.self {
		return p.httpGetters[peer], true
	}
	return nil, false
}

type httpGetter struct {
	baseUrl string
}

func (h *httpGetter) Get(group, key string) ([]byte, error) {
	u := fmt.Sprintf("%v%v/%v", h.baseUrl, url.QueryEscape(group), url.QueryEscape(key))
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	return bytes, nil
}

var _Pregetter = (*httpGetter)(nil)
