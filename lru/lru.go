package lru

import "container/list"

type Cache struct {
	maxBytes  int64
	nbytes    int64 //已经使用的内存
	ll        *list.List
	cache     map[string]*list.Element
	OnEvicted func(k string, value Value) //当有数据被移除时，触发回调
}

type entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes int64, OnEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		OnEvicted: OnEvicted,
		nbytes:    0,
		cache:     make(map[string]*list.Element),
		ll:        list.New(),
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.ll.Remove(ele)
		c.nbytes -= int64(kv.value.Len()) + int64(len(kv.key))
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len() - kv.value.Len())
		kv.value = value
		c.ll.MoveToFront(ele)
	} else {
		c.nbytes += int64(len(key) + value.Len())
		ele = c.ll.PushFront(&entry{key: key,
			value: value})
		ele.Value = &entry{key, value}
		c.cache[key] = ele
	}

	for c.maxBytes != 0 && c.nbytes > c.maxBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
