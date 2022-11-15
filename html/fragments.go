package html

import "sync"

// Fragment defines a thread safe fragment of HTML.
type Fragment struct {
	mu   sync.RWMutex
	head bool
	val  string
}

var (
	headFragments     []*Fragment
	headFragmentsLock = sync.RWMutex{}

	bodyFragments     []*Fragment
	bodyFragmentsLock = sync.RWMutex{}
)

// Get is used to get the contents of a fragment.
func (f *Fragment) Get() string {
	f.mu.RLock()
	x := f.val
	f.mu.RUnlock()
	return x
}

// Write is used to write a value and trigger a rewrite of whatever fragment chain this belongs to.
func (f *Fragment) Write(val string) {
	f.mu.Lock()
	f.val = val
	f.mu.Unlock()

	htmlPartsLock.Lock()
	res := ""
	if f.head {
		headFragmentsLock.RLock()
		defer headFragmentsLock.RLock()

		for _, v := range headFragments {
			res += v.Get()
		}
		additionalHead = res
	} else {
		bodyFragmentsLock.RLock()
		defer bodyFragmentsLock.RUnlock()

		for _, v := range bodyFragments {
			res += v.Get()
		}
		additionalBody = res
	}
	htmlPartsLock.Unlock()
}

// NewHeadFragment is used to make a new thread safe HTML head element that can be get/set. This will update what is in
// the additional part of the head.
func NewHeadFragment(content string) *Fragment {
	headFragmentsLock.Lock()
	f := &Fragment{val: content}
	headFragments = append(headFragments, f)
	headFragmentsLock.Unlock()
	return f
}

// NewBodyFragment is used to make a new thread safe HTML body element that can be get/set. This will update what is in
// the additional part of the head.
func NewBodyFragment(content string) *Fragment {
	bodyFragmentsLock.Lock()
	f := &Fragment{val: content}
	bodyFragments = append(bodyFragments, f)
	bodyFragmentsLock.Unlock()
	return f
}
