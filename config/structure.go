package config

import "sync"

type Structure struct {
	// Setup is set to true when the out of box experience is complete.
	Setup bool `config:"setup"`
}

var (
	configVar     Structure
	configVarLock = sync.RWMutex{}
)

// Config is used to return the current configuration.
func Config() Structure {
	configVarLock.RLock()
	c := configVar
	configVarLock.RUnlock()
	return c
}
