package twoqueues

import (
	"container/list"
)

type Cache struct {
	lruMaxEntries  int
	fifoMaxEntries int
	lru            *list.List
	fifo           *list.List
	cache          map[string]*list.Element
}

type entry struct {
	key   string
	value interface{}
	isLru bool
}

func New(lMaxEntries, fMaxEntries int) *Cache {
	return &Cache{
		lruMaxEntries:  lMaxEntries,
		fifoMaxEntries: fMaxEntries,
		lru:            list.New(),
		fifo:           list.New(),
		cache:          make(map[string]*list.Element),
	}
}

func (c *Cache) Add(key string, value interface{}) {
	if c.cache == nil {
		c.lru = list.New()
		c.fifo = list.New()
		c.cache = make(map[string]*list.Element)
		c.fifoMaxEntries = 2 << 10
		c.lruMaxEntries = 2 << 10
	}

	if element, ok := c.cache[key]; ok {
		ele := element.Value.(*entry)
		ele.value = value
		if ele.isLru {
			c.lru.MoveToFront(element)
		} else {
			c.fifo.Remove(element)
			ele.isLru = true
			updateElement := c.lru.PushFront(ele)
			c.cache[key] = updateElement
		}

	} else {
		element := c.fifo.PushBack(&entry{key, value, false})
		c.cache[key] = element
	}

	if c.lru.Len() > c.lruMaxEntries {
		c.RemoveLruBack()
	}

	if c.fifo.Len() > c.fifoMaxEntries {
		c.RemoveFifoFront()
	}
}

func (c *Cache) Get(key string) (value interface{}, ok bool) {
	if c.cache == nil {
		return
	}

	if element, hit := c.cache[key]; hit {
		ele := element.Value.(*entry)
		if ele.isLru {
			c.lru.MoveToFront(element)
		} else {
			c.fifo.Remove(element)
			ele.isLru = true
			updateElement := c.lru.PushFront(ele)
			c.cache[key] = updateElement
		}

		if c.lru.Len() > c.lruMaxEntries {
			c.RemoveLruBack()
		}
		return ele.value, true
	}
	return
}

func (c *Cache) RemoveLruBack() {
	if c.cache == nil {
		return
	}

	element := c.lru.Back()

	if element != nil {
		c.lru.Remove(element)
		entry := element.Value.(*entry)
		delete(c.cache, entry.key)

	}
}

func (c *Cache) RemoveFifoFront() {
	if c.cache == nil {
		return
	}

	element := c.fifo.Front()

	if element != nil {
		c.fifo.Remove(element)
		entry := element.Value.(*entry)
		delete(c.cache, entry.key)

	}
}
