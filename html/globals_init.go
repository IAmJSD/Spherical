package html

import (
	"io/fs"
	"sync"

	publicFs "github.com/jakemakesstuff/spherical/public"
	"github.com/jakemakesstuff/spherical/utils/tlru"
)

var (
	public    fs.FS
	tlruCache *tlru.Cache[[]byte]
)

// Setup is used to initialize the public variable. This should be called immediately at boot before any templating
// and then never called again!
func Setup(isDev bool) {
	if public != nil {
		panic("html Setup called twice!")
	}
	public = publicFs.GetFS(isDev)
	if !isDev {
		tlruCache = &tlru.Cache[[]byte]{}
	}
}

var (
	additionalHead string
	additionalBody string

	htmlPartsLock sync.RWMutex
)

func getHtmlParts() (head, body string) {
	htmlPartsLock.RLock()
	head = additionalHead
	body = additionalBody
	htmlPartsLock.RUnlock()
	return
}
