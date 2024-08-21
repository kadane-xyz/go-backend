package cache

import (
	"sync"
	"time"
)

type CacheItem struct {
	Value      interface{}
	Expiration int64
}

type Cache struct {
	items sync.Map
}

func NewCache() *Cache {
	return &Cache{
		items: sync.Map{},
	}
}

func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	expiration := time.Now().Add(duration).UnixNano()
	c.items.Store(key, CacheItem{
		Value:      value,
		Expiration: expiration,
	})
}

func (c *Cache) Get(key string) (interface{}, bool) {
	item, ok := c.items.Load(key)
	if !ok {
		return nil, false
	}

	cacheItem := item.(CacheItem)
	if time.Now().UnixNano() > cacheItem.Expiration {
		c.items.Delete(key)
		return nil, false
	}

	return cacheItem.Value, true
}

func (c *Cache) Delete(key string) {
	c.items.Delete(key)
}
