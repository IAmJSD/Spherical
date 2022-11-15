package tlru

import (
	"sync"
	"time"
)

type tlruItem[V any] struct {
	val   V
	timer *time.Timer
}

// Cache is used to define the TLRU cache.
type Cache[V any] struct {
	lock  sync.RWMutex
	items map[string]tlruItem[V]
}

func (c *Cache[V]) spawnDeleter(key string) func() {
	return func() {
		c.lock.Lock()
		t := c.items[key]
		if t.timer != nil {
			// Just in case we were raced and this is someone else.
			t.timer.Stop()
		}
		delete(c.items, key)
		c.lock.Unlock()
	}
}

// Set is used to set an item in the cache and delete the existing item if it already exists.
func (c *Cache[V]) Set(key string, value V, ttl time.Duration) {
	// Lock the cache.
	c.lock.Lock()
	defer c.lock.Unlock()

	// Make sure neither map is nil.
	if c.items == nil {
		c.items = map[string]tlruItem[V]{}
	}

	// Check if the timer exists.
	val := c.items[key]
	if val.timer != nil {
		// Stop the timer.
		val.timer.Stop()
	}

	// Write the value and timer to the map.
	timer := time.AfterFunc(ttl, c.spawnDeleter(key))
	c.items[key] = tlruItem[V]{
		val:   value,
		timer: timer,
	}
}

// Get is used to get and extend the life of an item from the TLRU cache if it exists. Returns an empty value if not.
func (c *Cache[V]) Get(key string, ttl time.Duration) (value V) {
	// Read lock the cache.
	c.lock.RLock()
	defer c.lock.RUnlock()

	// Return if the map is nil.
	if c.items == nil {
		return
	}

	// Get the item from the cache.
	if res, ok := c.items[key]; ok {
		// Extend the length of the TLRU cache.
		res.timer.Reset(ttl)

		// Set the value.
		value = res.val
	}
	return
}
