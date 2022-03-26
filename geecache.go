package mycache

import (
	"fmt"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(k string) ([]byte, error) {
	return f(k)
}

type Group struct {
	name      string
	getter    Getter
	maincache cache
	peers     PeerPicker
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func (g *Group) RegisterPeers(pk PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = pk
}

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("no getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter, //当缓存未命中时，应该从哪里获取数据源
		maincache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(k string) (value ByteView, err error) {
	if k == "" {
		return ByteView{}, fmt.Errorf("key is nil")
	}
	if v, ok := g.maincache.get(k); ok {
		log.Println("[cache hit]")
		return v, nil
	}
	return g.load(k)

}

func (g *Group) load(k string) (bv ByteView, err error) {
	if g.peers != nil {
		//远程加载key
		peerGetter, ok := g.peers.PickPeer(k)
		if ok {
			if bv, err = g.getFromPeer(peerGetter, k); err == nil {
				return bv, nil
			}
			log.Println("[GeeCache] Failed to get from peer", err)
		}
	}
	return g.getLocally(k)
}

/**
PeerGetter 接口的 httpGetter 从访问远程节点，获取缓存值
**/
func (g *Group) getFromPeer(peerGetter PeerGetter, k string) (ByteView, error) {
	bytes, err := peerGetter.Get(g.name, k)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{bytes}, nil
}

func (g *Group) getLocally(k string) (ByteView, error) {
	v, err := g.getter.Get(k) //从数据源获取数据
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{cloneBytes(v)}
	g.populateCache(k, value) //分布式场景下会调用 getFromPeer 从其他节点获取
	return value, nil
}

func (g *Group) populateCache(k string, v ByteView) {
	g.maincache.add(k, v)
}
