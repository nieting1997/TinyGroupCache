# TinyGroupCache
阅读模仿groupcache 构建分布式缓存
* 采用一致性hash算法+虚拟节点方式，达到负载均衡的目的。
* 采用了[2Q算法](http://www.vldb.org/conf/1994/P439.PDF)作为缓存策略。
* 采用singleflight的方法预防缓存击穿。
* 基于HTTP实现分布式缓存，每个节点即可做服务端又可做客户端。

## 查询流程
<div align=center><img src="https://user-images.githubusercontent.com/90097659/132313734-d5045f2b-a62c-4622-8410-89bbda66d028.png" width="500"/> </div>

## 实现细节

### 2Q算法
2Q算法将数据缓存在FIFO队列里面，当数据第二次被访问时，则将数据从FIFO队列移到LRU队列里面，两个队列各自按照自己的方法淘汰数据。
* 新访问的数据插入到FIFO队列；
* 如果数据在FIFO队列中一直没有被再次访问，则最终按照FIFO规则淘汰；
* 如果数据在FIFO队列中被再次访问，则将数据移到LRU队列头部；
* 如果数据在LRU队列再次被访问，则将数据移到LRU队列头部；
* LRU队列淘汰末尾的数据。
具体通过map+list+标志位(标志元素存储位置)

### 一致性hash
引入虚拟节点，解决节点较少的情况下数据容易倾斜的问题。

### singleflight
作用:无论Do被调用多少次，在这段时间内函数fn都只会被调用一次，等待fn调用结束了，返回返回值或错误。
通过sync.WaitGroup与map记录来实现。

## Todo
* groupcache采用mainCache,hotcache保存本地/远程访问得到的值
* groupcache采用proto优化通信细节

## demo
<div align=center><img src="https://user-images.githubusercontent.com/90097659/132310510-32ea3201-4cb5-4e28-aa39-30ae7357ea02.jpg" width="500"/> </div>

code:
```
package main

/*
$ curl "http://localhost:9999/api?key=Tom"
630

$ curl "http://localhost:9999/api?key=kkk"
kkk not exist
*/

import (
	"flag"
	"fmt"
	"gocache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *gocache.Group {
	return gocache.NewGroup("scores", 2<<10, 2<<10, gocache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, group *gocache.Group) {
	peers := gocache.NewHTTPPool(addr)
	peers.Set(addrs...)
	group.RegisterPeers(peers)
	log.Println("gocache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, group *gocache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := group.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())

		}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "gocache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	group := createGroup()
	if api {
		go startAPIServer(apiAddr, group)
	}
	startCacheServer(addrMap[port], addrs, group)
}

```
