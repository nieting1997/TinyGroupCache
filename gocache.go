package gocache

import (
	"fmt"
	"gocache/singleflight"
	"log"
	"sync"
)

type Group struct {
	name   		string
	getter 		Getter
	mainCache   cache
	peers  		PeerPicker
	loader 		*singleflight.Group
}


type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunction func(key string) ([]byte, error)

func (f GetterFunction) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, lMaxNum int, fMaxNum int, getter Getter) *Group {
	if getter == nil {
		panic("the Group getter is nil .")
	}
	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:   	name,
		getter: 	getter,
		mainCache:  cache{lEntries:lMaxNum, fEntries:fMaxNum},
		loader: 	&singleflight.Group{},
	}

	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	return groups[name]
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("Group err: key is \"\"")
	}
	if value, ok := g.mainCache.get(key); ok {
		log.Printf("[Group %s] Hit cache,key is %s\n", g.name, key)
		return value, nil
	}
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	bv, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if bytes, err := peer.(*httpGetter).Get(g.name, key); err == nil {
					return ByteView{b: bytes}, nil
				}
				log.Println("[Group] Failed to get from peer", g.name)
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return bv.(ByteView), nil
	}
	return
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err

	}
	c := make([]byte, len(bytes))
	copy(c, bytes)
	value := ByteView{b: c}
	g.Set(key, value)
	return value, nil
}

func (g *Group) Set(key string, value ByteView) {
	g.mainCache.add(key, value)
	log.Printf("[Group %s] Add cache,key is %s\n", g.name, key)
}


func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}
