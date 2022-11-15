package config

import "sync"

type Structure struct {
	// Setup is set to true when the out of box experience is complete.
	Setup bool `config:"setup"`

	// JobCount is used to define the number of jobs. Defaults to 120.
	JobCount uint `config:"job_count"`

	// SchedulerSleepTime is how number of milliseconds the scheduler sleeps for. Defaults to 1000ms.
	SchedulerSleepTime uint `config:"scheduler_sleep_time"`
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
