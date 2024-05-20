package cache

import (
	"sync"
	"time"
)

// Cache is a thread-safe kv cache with ttl expiration.
// The data will be cleared at the next read operation when ttl is expired.
// In some use cases, e.g. high volume of temporary data are stored,
// and then there are no more reading operations, memory leak may happen.
// So we need another gc routine to do the clean job.
type Cache struct {
	lock sync.Mutex
	kv   map[string]entry
}

type entry struct {
	data        interface{}
	expiredTime time.Time
}

func NewCache() *Cache {
	return &Cache{
		kv: make(map[string]entry),
	}
}

func (c *Cache) Save(k string, v interface{}, ttl time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()
	expired := time.Now().Add(ttl)
	c.kv[k] = entry{
		data:        v,
		expiredTime: expired,
	}
}

func (c *Cache) Get(k string) (interface{}, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	e, ok := c.kv[k]
	if !ok {
		return nil, false
	}
	if e.expiredTime.Before(time.Now()) {
		delete(c.kv, k)
		return nil, false
	}
	return e.data, true
}

func (c *Cache) GetOrSet(k string, renew func() interface{}, ttl time.Duration) interface{} {
	c.lock.Lock()
	defer c.lock.Unlock()
	e, ok := c.kv[k]
	if !ok {
		v := renew()
		if v == nil {
			return nil
		}
		expired := time.Now().Add(ttl)
		c.kv[k] = entry{
			data:        v,
			expiredTime: expired,
		}
		return v
	}
	if e.expiredTime.Before(time.Now()) {
		v := renew()
		if v == nil {
			// clear cache
			delete(c.kv, k)
			return nil
		}
		expired := time.Now().Add(ttl)
		c.kv[k] = entry{
			data:        v,
			expiredTime: expired,
		}
		return v
	}
	return e.data
}
