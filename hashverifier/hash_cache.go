package hashverifier

// HashCache is the interface given for a hashing cacher.
type HashCache interface {
	// LookupPGP is used to look up a PGP public key from the cache. Returns nil if the key was not found.
	LookupPGP(hostname string) []byte

	// WritePGP is used to write a PGP key to the cache. If this wants to error, it should do so silently.
	WritePGP(hostname string, key []byte)

	// Exists checks if a hash exists in the cache. If this wants to error, it should return false.
	Exists(hash []byte) bool

	// Ensure is used to ensure that a hash exists in the cache. Note this should ONLY be called when GPG is validated.
	// If this wants to error, it should do so silently.
	Ensure(hash []byte)
}

type nopCache struct{}

func (nopCache) LookupPGP(string) []byte { return nil }

func (nopCache) WritePGP(string, []byte) {}

func (nopCache) Exists([]byte) bool { return false }

func (nopCache) Ensure([]byte) {}

var _ HashCache = nopCache{}
