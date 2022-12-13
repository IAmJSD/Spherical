package globalstate

import "sync"

// NOTE: This package should RARELY be used. Other options should ALWAYS be explored
// before the global state. This is for when those are too impractical because thw item
// does not fit anywhere else, might be used by the scheduler, AND is used in multiple
// places.

var (
	map_ = map[string]any{}
	lock = sync.RWMutex{}
)

// Set is used to set a value in the global state.
func Set(key string, value any) {
	lock.Lock()
	map_[key] = value
	lock.Unlock()
}

// Get is used to get a value from the global state.
func Get(key string) (value any) {
	lock.RLock()
	value = map_[key]
	lock.RUnlock()
	return
}
