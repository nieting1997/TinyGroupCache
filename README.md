# TinyGroupCache
阅读模仿groupcache 构建分布式缓存
* 采用一致性hash算法+虚拟节点方式，达到负载均衡的目的。
* 采用了[2Q算法](http://www.vldb.org/conf/1994/P439.PDF)作为缓存策略 
* 采用singleflight的方法预防缓存击穿
* 基于HTTP实现分布式缓存

