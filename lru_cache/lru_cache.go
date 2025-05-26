package lru_cache

import (
	"container/list"
	"sync"
)

type CacheEntry[V any] struct {
	value V
	lp    *list.Element
}

type ListEntry[K comparable, V any] struct {
	Key   K
	Value V
}

func New[K comparable, V any](size int) *Cache[K, V] {
	return &Cache[K, V]{
		data:      make(map[K]*CacheEntry[V]),
		evictList: list.New(),
		size:      size,
	}
}

type Cache[K comparable, V any] struct {
	sync.RWMutex
	data      map[K]*CacheEntry[V]
	evictList *list.List
	size      int
}

func (c *Cache[K, V]) Add(key K, value V) {
	c.Lock()
	defer c.Unlock()

	if cacheEntry, ok := c.data[key]; ok {
		c.evictList.MoveToFront(cacheEntry.lp)
		cacheEntry.value = value
		return
	}

	listEntry := ListEntry[K, V]{key, value}
	cacheEntry := &CacheEntry[V]{value: value}

	lp := c.evictList.PushFront(listEntry)
	cacheEntry.lp = lp
	c.data[key] = cacheEntry

	if c.evictList.Len() > c.size {
		c.removeOldest()
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.Lock()
	defer c.Unlock()
	var v V
	entry, ok := c.data[key]
	if ok {
		v = entry.value
		c.evictList.MoveToFront(entry.lp)
		return v, true
	}
	return v, false
}

func (c *Cache[K, V]) removeOldest() {
	ent := c.evictList.Back()
	if ent != nil {
		removed := c.evictList.Remove(ent)
		cKey := removed.(ListEntry[K, V]).Key
		delete(c.data, cKey)
	}
}
