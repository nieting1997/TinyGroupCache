package twoqueues

import (
	"log"
	"testing"
)

func TestAddAndGet(t *testing.T) {
	c := New(10, 10)
	c.Add("a", "1")
	log.Println(c.lru.Len(), c.fifo.Len())// 0 1

	c.Add("b", "2")
	log.Println(c.lru.Len(), c.fifo.Len()) // 0 2
	log.Println(c.Get("b"))

	c.Add("c", "3")
	log.Println(c.lru.Len(), c.fifo.Len()) // 1 2
	log.Println(c.Get("c"))
	log.Println(c.lru.Len(), c.fifo.Len()) // 2 1
}
